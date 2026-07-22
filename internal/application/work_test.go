package application_test

import (
	"context"
	"errors"
	"testing"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/application/inmemory"
	"ai-studio-os/internal/domain/event"
	"ai-studio-os/internal/domain/execution"
	"ai-studio-os/internal/domain/executor"
	"ai-studio-os/internal/domain/shared"
	"ai-studio-os/internal/domain/workflow"
)

func newWorkFixture(t *testing.T) (*application.WorkService, *inmemory.EventBus, *application.TaskPlanningService) {
	t.Helper()
	projects := inmemory.NewProjectStore()
	tasks := inmemory.NewTaskStore()
	executors := inmemory.NewExecutorStore()
	executions := inmemory.NewExecutionStore()
	bus := inmemory.NewEventBus()
	rules := workflow.Machine{}

	planning := &application.TaskPlanningService{Projects: projects, Tasks: tasks, Events: bus, Rules: rules}
	work := &application.WorkService{Tasks: tasks, Executors: executors, Executions: executions, Events: bus, Rules: rules}

	newActiveProject(t, projects)
	return work, bus, planning
}

func newReadyTask(t *testing.T, planning *application.TaskPlanningService) {
	t.Helper()
	ctx := context.Background()
	if _, err := planning.CreateTask(ctx, application.CreateTaskParams{
		ID: "task-1", ProjectID: "proj-1", Title: "Задача", Type: "feature",
	}); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	if err := planning.PlanTask(ctx, "proj-1", "task-1", ""); err != nil {
		t.Fatalf("PlanTask: %v", err)
	}
}

func saveExecutor(t *testing.T, store application.ExecutorStore, active bool, roles ...shared.Role) *executor.Executor {
	t.Helper()
	ctx := context.Background()
	e, _, err := executor.New("executor-1", "claude-code-instance", roles)
	if err != nil {
		t.Fatalf("executor.New: %v", err)
	}
	if active {
		if _, err := e.Activate(); err != nil {
			t.Fatalf("Activate: %v", err)
		}
	}
	if err := store.Save(ctx, e); err != nil {
		t.Fatalf("Save: %v", err)
	}
	return e
}

func TestStartTask_Success(t *testing.T) {
	ctx := context.Background()
	work, bus, planning := newWorkFixture(t)
	newReadyTask(t, planning)
	saveExecutor(t, work.Executors, true, shared.RoleDeveloper)

	run, err := work.StartTask(ctx, application.StartTaskParams{ProjectID: "proj-1", TaskID: "task-1", ExecutorID: "executor-1", Actor: "pm:executor-2"})
	if err != nil {
		t.Fatalf("StartTask: %v", err)
	}
	if run.TaskID() != "task-1" || run.ExecutorID() != "executor-1" {
		t.Errorf("Execution refs = task=%q executor=%q", run.TaskID(), run.ExecutorID())
	}
	if run.State() != execution.StateRunning {
		t.Errorf("Execution.State() = %v, want %v (Accept was called)", run.State(), execution.StateRunning)
	}

	tsk, err := work.Tasks.Get(ctx, "proj-1", "task-1")
	if err != nil {
		t.Fatalf("Tasks.Get: %v", err)
	}
	if tsk.State() != shared.StateInProgress {
		t.Errorf("Task.State() = %v, want %v", tsk.State(), shared.StateInProgress)
	}

	// TaskCreated, TaskPlanned (from the fixture), then TaskStarted,
	// ExecutionQueued, ExecutionStarted.
	published := bus.Published()
	if len(published) != 5 {
		t.Fatalf("published = %d events, want 5", len(published))
	}
	wantTail := []string{event.TaskStarted, event.ExecutionQueued, event.ExecutionStarted}
	if len(published) < len(wantTail) {
		t.Fatalf("published = %d events, want at least %d", len(published), len(wantTail))
	}
	tail := published[len(published)-len(wantTail):]
	for i, want := range wantTail {
		if tail[i].Type() != want {
			t.Errorf("event[%d].Type() = %q, want %q", i, tail[i].Type(), want)
		}
	}
}

func TestStartTask_RejectedWhenExecutorNotActive(t *testing.T) {
	ctx := context.Background()
	work, bus, planning := newWorkFixture(t)
	newReadyTask(t, planning)
	saveExecutor(t, work.Executors, false, shared.RoleDeveloper) // Registered, not Active
	before := len(bus.Published())

	if _, err := work.StartTask(ctx, application.StartTaskParams{ProjectID: "proj-1", TaskID: "task-1", ExecutorID: "executor-1"}); !errors.Is(err, application.ErrExecutorNotAssignable) {
		t.Errorf("StartTask() error = %v, want %v", err, application.ErrExecutorNotAssignable)
	}
	if len(bus.Published()) != before {
		t.Errorf("published grew on rejection: %d -> %d", before, len(bus.Published()))
	}
}

func TestStartTask_RejectedWhenExecutorLacksRole(t *testing.T) {
	ctx := context.Background()
	work, _, planning := newWorkFixture(t)
	newReadyTask(t, planning)
	saveExecutor(t, work.Executors, true, shared.RoleReviewer) // Active, wrong role

	if _, err := work.StartTask(ctx, application.StartTaskParams{ProjectID: "proj-1", TaskID: "task-1", ExecutorID: "executor-1"}); !errors.Is(err, application.ErrExecutorNotAssignable) {
		t.Errorf("StartTask() error = %v, want %v", err, application.ErrExecutorNotAssignable)
	}
}

func TestStartTask_RejectedWhenTaskNotReady(t *testing.T) {
	ctx := context.Background()
	work, _, planning := newWorkFixture(t)
	// Create but do not plan: task stays in Backlog.
	if _, err := planning.CreateTask(ctx, application.CreateTaskParams{
		ID: "task-1", ProjectID: "proj-1", Title: "Задача", Type: "feature",
	}); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	saveExecutor(t, work.Executors, true, shared.RoleDeveloper)

	if _, err := work.StartTask(ctx, application.StartTaskParams{ProjectID: "proj-1", TaskID: "task-1", ExecutorID: "executor-1"}); err == nil {
		t.Error("StartTask() from Backlog error = nil, want the workflow.Machine's rejection")
	}
}

func TestStartTask_TaskNotFound(t *testing.T) {
	work, _, _ := newWorkFixture(t)
	saveExecutor(t, work.Executors, true, shared.RoleDeveloper)
	if _, err := work.StartTask(context.Background(), application.StartTaskParams{ProjectID: "proj-1", TaskID: "missing", ExecutorID: "executor-1"}); !errors.Is(err, application.ErrNotFound) {
		t.Errorf("StartTask() error = %v, want %v", err, application.ErrNotFound)
	}
}

func TestStartTask_ExecutorNotFound(t *testing.T) {
	work, _, planning := newWorkFixture(t)
	newReadyTask(t, planning)
	if _, err := work.StartTask(context.Background(), application.StartTaskParams{ProjectID: "proj-1", TaskID: "task-1", ExecutorID: "missing"}); !errors.Is(err, application.ErrNotFound) {
		t.Errorf("StartTask() error = %v, want %v", err, application.ErrNotFound)
	}
}

// failingExecutionStore always fails Save, to verify StartTask propagates
// a storage failure instead of silently reporting success (a regression a
// coverage number alone would not catch).
type failingExecutionStore struct{ application.ExecutionStore }

func (failingExecutionStore) Save(context.Context, *execution.Execution) error {
	return errors.New("boom: execution store unavailable")
}

func TestStartTask_PropagatesExecutionStoreFailure(t *testing.T) {
	ctx := context.Background()
	work, _, planning := newWorkFixture(t)
	newReadyTask(t, planning)
	saveExecutor(t, work.Executors, true, shared.RoleDeveloper)
	work.Executions = failingExecutionStore{ExecutionStore: work.Executions}

	if _, err := work.StartTask(ctx, application.StartTaskParams{ProjectID: "proj-1", TaskID: "task-1", ExecutorID: "executor-1"}); err == nil {
		t.Fatal("StartTask() error = nil, want the store's failure propagated")
	}

	// The Task transition and its event were already committed before the
	// Execution step failed — Task Ready -> In Progress is not rolled
	// back. This documents current behaviour (no cross-aggregate
	// transaction in the Application Layer yet) rather than asserting it
	// is desirable; EPIC-005 infrastructure may need a saga or outbox to
	// close this gap when a real store can actually fail mid-sequence.
	tsk, err := work.Tasks.Get(ctx, "proj-1", "task-1")
	if err != nil {
		t.Fatalf("Tasks.Get: %v", err)
	}
	if tsk.State() != shared.StateInProgress {
		t.Errorf("Task.State() = %v, want %v (already committed before the failure)", tsk.State(), shared.StateInProgress)
	}
}
