package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/domain/event"
	"ai-studio-os/internal/domain/executor"
	"ai-studio-os/internal/domain/shared"
	"ai-studio-os/internal/domain/task"
	"ai-studio-os/internal/domain/workflow"
	"ai-studio-os/internal/platform"
)

// fakePlanner is a fakeBackend that also reports a Definition of Ready proposal.
type fakePlanner struct {
	*fakeBackend
	scope      string
	criteria   []string
	proposeErr error
}

func (p *fakePlanner) Propose(context.Context) (string, []string, error) {
	if p.proposeErr != nil {
		return "", nil, p.proposeErr
	}
	return p.scope, p.criteria, nil
}

// ownProjectManager is this orchestrator's own registry entry for the Project
// Manager role, found by derived identifier like every other role (TASK-086).
func ownProjectManager(t *testing.T) *executor.Executor {
	t.Helper()
	e, _, err := executor.New(
		executorIDForRole(shared.RoleProjectManager), "claude-code",
		[]shared.Role{shared.RoleProjectManager},
	)
	if err != nil {
		t.Fatalf("executor.New: %v", err)
	}
	if _, err := e.Activate(); err != nil {
		t.Fatalf("Activate: %v", err)
	}
	return e
}

// planFixture creates a task and returns the TaskCreated event that a Project
// Manager run reacts to. The task is left exactly as created — in Backlog.
func planFixture(t *testing.T, planner *fakePlanner) (*dispatchFixture, platform.Event) {
	t.Helper()
	ctx := context.Background()

	f := newDispatchFixture(t, []string{"github.com/org/repo"}, ownDeveloper(t), ownProjectManager(t))
	planner.fakeBackend = f.backend
	f.dispatcher.NewPlanner = func() (PlanExecutor, error) { return planner, nil }

	planning := &application.TaskPlanningService{
		Projects: mustProjects(f), Tasks: f.tasks, Events: f.bus, Rules: workflow.Machine{},
	}
	f.dispatcher.Planning = planning

	if _, err := planning.CreateTask(ctx, application.CreateTaskParams{
		ID: "TASK-001", ProjectID: "proj-1", Title: "Заголовок", Type: "feature",
		Scope: "исходный scope", AcceptanceCriteria: []string{"исходный критерий"},
	}); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	var created platform.Event
	for _, e := range f.bus.Published() {
		if e.Type() == event.TaskCreated {
			created = e
		}
	}
	if created == nil {
		t.Fatal("no TaskCreated event was published")
	}
	return f, created
}

// planTask reads the task from the store — the source of truth.
//
// Deliberately not the projection: this fixture's projection is only fed by
// events passed through Handle, so it would not see TaskRefined (in production
// the Poller delivers it on a later tick). Asserting on the view would let a
// test pass because the read model was stale rather than because the task was
// left alone.
func planTask(t *testing.T, f *dispatchFixture) *task.Task {
	t.Helper()
	tsk, err := f.tasks.Get(context.Background(), "proj-1", "TASK-001")
	if err != nil {
		t.Fatalf("Get task: %v", err)
	}
	return tsk
}

// countTaskRefined reports how many refinements were actually published.
func countTaskRefined(f *dispatchFixture) int {
	n := 0
	for _, e := range f.bus.Published() {
		if e.Type() == event.TaskRefined {
			n++
		}
	}
	return n
}

// countTaskPlanned is the assertion that matters most in this epic: a Project
// Manager run must never pass the Definition of Ready checkpoint.
func countTaskPlanned(f *dispatchFixture) int {
	n := 0
	for _, e := range f.bus.Published() {
		if e.Type() == event.TaskPlanned {
			n++
		}
	}
	return n
}

func TestPlan_AppliesProposalAndLeavesTaskInBacklog(t *testing.T) {
	ctx := context.Background()
	planner := &fakePlanner{scope: "уточнённый scope", criteria: []string{"первый", "второй"}}
	f, created := planFixture(t, planner)

	if err := f.dispatcher.Handle(ctx, created); err != nil {
		t.Fatalf("Handle(TaskCreated): %v", err)
	}

	tsk := planTask(t, f)
	if tsk.Scope() != "уточнённый scope" {
		t.Errorf("Scope = %q, want the proposal applied", tsk.Scope())
	}
	if len(tsk.AcceptanceCriteria()) != 2 {
		t.Errorf("AcceptanceCriteria = %v, want the proposed list", tsk.AcceptanceCriteria())
	}
	if n := countTaskRefined(f); n != 1 {
		t.Errorf("TaskRefined published %d times, want 1", n)
	}

	// The point of the epic: prepared, not decided.
	if tsk.State() != shared.StateBacklog {
		t.Errorf("task state = %v, want it left in %v for a human", tsk.State(), shared.StateBacklog)
	}
	if n := countTaskPlanned(f); n != 0 {
		t.Errorf("TaskPlanned published %d times, want 0 — automation must not accept Definition of Ready", n)
	}
	if planner.finished != 1 {
		t.Errorf("Finish called %d times, want 1", planner.finished)
	}
}

// A Project Manager produces no commits, so there is no task branch yet — the
// clone is of the base branch, and the agent gets its own role.
func TestPlan_HandsProjectManagerRoleAndBaseBranch(t *testing.T) {
	ctx := context.Background()
	planner := &fakePlanner{scope: "scope"}
	f, created := planFixture(t, planner)
	branchesBefore := len(f.repos.branches)

	if err := f.dispatcher.Handle(ctx, created); err != nil {
		t.Fatalf("Handle(TaskCreated): %v", err)
	}

	accepted := planner.accepted[len(planner.accepted)-1]
	if accepted.Role != string(shared.RoleProjectManager) {
		t.Errorf("ExecutorTask.Role = %q, want %q", accepted.Role, shared.RoleProjectManager)
	}
	if accepted.Branch != baseBranch {
		t.Errorf("ExecutorTask.Branch = %q, want the base branch %q", accepted.Branch, baseBranch)
	}
	if len(f.repos.branches) != branchesBefore {
		t.Errorf("planning created %d branch(es); a Project Manager commits nothing", len(f.repos.branches)-branchesBefore)
	}
}

// Nothing is ever invented: an absent or unusable proposal leaves the task
// exactly as it was. An invented scope would be worse than none — a human
// accepting Definition of Ready would read something nobody prepared.
func TestPlan_NoProposalIsInventedOnFailure(t *testing.T) {
	cases := []struct {
		name  string
		setup func(*fakePlanner)
	}{
		{
			name: "agent reported failure",
			setup: func(p *fakePlanner) {
				p.statuses = []platform.ExecutionStatus{{State: statusFailed, Message: "exit 1"}}
			},
		},
		{
			name:  "proposal missing or unreadable",
			setup: func(p *fakePlanner) { p.proposeErr = errors.New("no proposal was written") },
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			planner := &fakePlanner{}
			f, created := planFixture(t, planner)
			tc.setup(planner)

			// An error may or may not be returned; the state is what matters.
			_ = f.dispatcher.Handle(ctx, created)

			tsk := planTask(t, f)
			if tsk.Scope() != "исходный scope" {
				t.Errorf("Scope = %q, want the original untouched", tsk.Scope())
			}
			if got := tsk.AcceptanceCriteria(); len(got) != 1 || got[0] != "исходный критерий" {
				t.Errorf("AcceptanceCriteria = %v, want the original untouched", got)
			}
			if tsk.State() != shared.StateBacklog {
				t.Errorf("task state = %v, want %v", tsk.State(), shared.StateBacklog)
			}
			// Nothing was applied, so nothing was announced either.
			if n := countTaskRefined(f); n != 0 {
				t.Errorf("TaskRefined published %d times, want 0", n)
			}
			if n := countTaskPlanned(f); n != 0 {
				t.Errorf("TaskPlanned published %d times, want 0", n)
			}
			if planner.finished == 0 {
				t.Error("Finish was never called; the sandbox would outlive the run")
			}
		})
	}
}

func TestPlan_TimeoutLeavesTaskUntouched(t *testing.T) {
	ctx := context.Background()
	planner := &fakePlanner{}
	f, created := planFixture(t, planner)
	// Never terminal — the guard against a hung sandbox.
	planner.statuses = []platform.ExecutionStatus{{State: statusRunning}}
	f.dispatcher.ExecutionTimeout = 20 * time.Millisecond

	if err := f.dispatcher.Handle(ctx, created); !errors.Is(err, ErrExecutionTimedOut) {
		t.Fatalf("Handle() error = %v, want %v", err, ErrExecutionTimedOut)
	}

	if got := planTask(t, f).Scope(); got != "исходный scope" {
		t.Errorf("Scope = %q, want the original untouched", got)
	}
	if planner.finished == 0 {
		t.Error("Finish was never called after the timeout")
	}
}

// Without its own registry entry the run must not fall back to another role's
// executor (TASK-086), and the task must stay untouched.
func TestPlan_MissingProjectManagerEntryLeavesTaskUntouched(t *testing.T) {
	ctx := context.Background()
	planner := &fakePlanner{scope: "scope"}

	// Developer only: no project-manager entry in the registry.
	f := newDispatchFixture(t, []string{"github.com/org/repo"}, ownDeveloper(t))
	planner.fakeBackend = f.backend
	f.dispatcher.NewPlanner = func() (PlanExecutor, error) { return planner, nil }
	planning := &application.TaskPlanningService{
		Projects: mustProjects(f), Tasks: f.tasks, Events: f.bus, Rules: workflow.Machine{},
	}
	f.dispatcher.Planning = planning
	if _, err := planning.CreateTask(ctx, application.CreateTaskParams{
		ID: "TASK-001", ProjectID: "proj-1", Title: "Заголовок", Type: "feature",
		Scope: "исходный scope",
	}); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	var created platform.Event
	for _, e := range f.bus.Published() {
		if e.Type() == event.TaskCreated {
			created = e
		}
	}

	if err := f.dispatcher.Handle(ctx, created); !errors.Is(err, ErrNoExecutorForRole) {
		t.Fatalf("Handle() error = %v, want %v", err, ErrNoExecutorForRole)
	}
	if got := planTask(t, f).Scope(); got != "исходный scope" {
		t.Errorf("Scope = %q, want the original untouched", got)
	}
}
