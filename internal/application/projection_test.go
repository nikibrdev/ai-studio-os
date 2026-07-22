package application_test

import (
	"context"
	"testing"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/application/inmemory"
	"ai-studio-os/internal/domain/shared"
	"ai-studio-os/internal/domain/workflow"
)

func TestTaskProjection_IncrementalUpdatesTrackState(t *testing.T) {
	ctx := context.Background()
	projects := inmemory.NewProjectStore()
	tasks := inmemory.NewTaskStore()
	bus := inmemory.NewEventBus()
	rules := workflow.Machine{}
	newActiveProject(t, projects)

	proj := application.NewTaskProjection()
	if err := proj.Subscribe(bus); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	planning := &application.TaskPlanningService{Projects: projects, Tasks: tasks, Events: bus, Rules: rules}
	if _, err := planning.CreateTask(ctx, application.CreateTaskParams{ID: "task-1", ProjectID: "proj-1", Title: "Задача", Type: "feature"}); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	view, ok := proj.Get("proj-1", "task-1")
	if !ok || view.State != shared.StateBacklog || view.ProjectID != "proj-1" {
		t.Fatalf("Get() after CreateTask = (%+v, %v), want Backlog/proj-1", view, ok)
	}

	if err := planning.PlanTask(ctx, "proj-1", "task-1", ""); err != nil {
		t.Fatalf("PlanTask: %v", err)
	}
	view, _ = proj.Get("proj-1", "task-1")
	if view.State != shared.StateReady {
		t.Errorf("State() after PlanTask = %v, want %v", view.State, shared.StateReady)
	}
}

func TestTaskProjection_ReviewCompletedDisambiguatesOutcome(t *testing.T) {
	ctx := context.Background()
	projects := inmemory.NewProjectStore()
	tasks := inmemory.NewTaskStore()
	executors := inmemory.NewExecutorStore()
	executions := inmemory.NewExecutionStore()
	bus := inmemory.NewEventBus()
	rules := workflow.Machine{}
	newActiveProject(t, projects)

	proj := application.NewTaskProjection()
	if err := proj.Subscribe(bus); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	planning := &application.TaskPlanningService{Projects: projects, Tasks: tasks, Events: bus, Rules: rules}
	work := &application.WorkService{Tasks: tasks, Executors: executors, Executions: executions, Events: bus, Rules: rules}
	completion := &application.CompletionService{Tasks: tasks, Repositories: inmemory.NewRepositoryProvider(), Events: bus, Rules: rules}

	if _, err := planning.CreateTask(ctx, application.CreateTaskParams{ID: "task-1", ProjectID: "proj-1", Title: "Задача", Type: "feature"}); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	if err := planning.PlanTask(ctx, "proj-1", "task-1", ""); err != nil {
		t.Fatalf("PlanTask: %v", err)
	}
	saveExecutor(t, executors, true, shared.RoleDeveloper)
	if _, err := work.StartTask(ctx, application.StartTaskParams{ProjectID: "proj-1", TaskID: "task-1", ExecutorID: "executor-1"}); err != nil {
		t.Fatalf("StartTask: %v", err)
	}
	if err := completion.RequestReview(ctx, "proj-1", "task-1", ""); err != nil {
		t.Fatalf("RequestReview: %v", err)
	}

	// Changes requested: the projection must land back in In Progress,
	// not guess Testing just because a ReviewCompleted event fired.
	if err := completion.CompleteReview(ctx, "proj-1", "task-1", false, ""); err != nil {
		t.Fatalf("CompleteReview(false): %v", err)
	}
	view, _ := proj.Get("proj-1", "task-1")
	if view.State != shared.StateInProgress {
		t.Fatalf("State() after changes-requested = %v, want %v", view.State, shared.StateInProgress)
	}

	if err := completion.RequestReview(ctx, "proj-1", "task-1", ""); err != nil {
		t.Fatalf("second RequestReview: %v", err)
	}
	if err := completion.CompleteReview(ctx, "proj-1", "task-1", true, ""); err != nil {
		t.Fatalf("CompleteReview(true): %v", err)
	}
	view, _ = proj.Get("proj-1", "task-1")
	if view.State != shared.StateTesting {
		t.Fatalf("State() after approval = %v, want %v", view.State, shared.StateTesting)
	}
}

func TestTaskProjection_RebuildFromJournalMatchesIncremental(t *testing.T) {
	ctx := context.Background()
	projects := inmemory.NewProjectStore()
	tasks := inmemory.NewTaskStore()
	bus := inmemory.NewEventBus()
	rules := workflow.Machine{}
	newActiveProject(t, projects)

	live := application.NewTaskProjection()
	if err := live.Subscribe(bus); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	planning := &application.TaskPlanningService{Projects: projects, Tasks: tasks, Events: bus, Rules: rules}
	if _, err := planning.CreateTask(ctx, application.CreateTaskParams{ID: "task-1", ProjectID: "proj-1", Title: "Задача", Type: "feature"}); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	if err := planning.PlanTask(ctx, "proj-1", "task-1", ""); err != nil {
		t.Fatalf("PlanTask: %v", err)
	}

	rebuilt := application.NewTaskProjection()
	if err := rebuilt.Rebuild(ctx, bus.Published()); err != nil {
		t.Fatalf("Rebuild: %v", err)
	}

	liveView, liveOK := live.Get("proj-1", "task-1")
	rebuiltView, rebuiltOK := rebuilt.Get("proj-1", "task-1")
	if liveOK != rebuiltOK || liveView != rebuiltView {
		t.Errorf("rebuilt = (%+v, %v), want it to match live = (%+v, %v)", rebuiltView, rebuiltOK, liveView, liveOK)
	}
}

func TestTaskProjection_GetUnknownTask(t *testing.T) {
	proj := application.NewTaskProjection()
	if _, ok := proj.Get("proj-1", "missing"); ok {
		t.Error("Get() for unseen task ok = true, want false")
	}
}
