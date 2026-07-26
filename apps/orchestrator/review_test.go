package main

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/domain/event"
	"ai-studio-os/internal/domain/executor"
	"ai-studio-os/internal/domain/shared"
	"ai-studio-os/internal/domain/workflow"
	"ai-studio-os/internal/platform"
)

// fakeReviewer is a fakeBackend that also reports a verdict.
type fakeReviewer struct {
	*fakeBackend
	approved  bool
	comment   string
	reviewErr error
}

func (r *fakeReviewer) Review(context.Context) (bool, string, error) {
	if r.reviewErr != nil {
		return false, "", r.reviewErr
	}
	return r.approved, r.comment, nil
}

// activeReviewer is this orchestrator's own registry entry for the Reviewer
// role — looked up by derived identifier, same as Developer (TASK-086).
func activeReviewer(t *testing.T) *executor.Executor {
	t.Helper()
	e, _, err := executor.New(executorIDForRole(shared.RoleReviewer), "claude-code", []shared.Role{shared.RoleReviewer})
	if err != nil {
		t.Fatalf("executor.New: %v", err)
	}
	if _, err := e.Activate(); err != nil {
		t.Fatalf("Activate: %v", err)
	}
	return e
}

// reviewFixture drives a task to Review and wires a reviewing backend.
func reviewFixture(t *testing.T, rev *fakeReviewer) (*dispatchFixture, platform.Event) {
	t.Helper()
	ctx := context.Background()

	f := newDispatchFixture(t, []string{"github.com/org/repo"}, ownDeveloper(t), activeReviewer(t))
	rev.fakeBackend = f.backend
	f.dispatcher.NewReviewer = func() (ReviewExecutor, error) { return rev, nil }

	// Take the task to Review through the real services, so the state the
	// dispatcher reads is one the state machine actually allows.
	planned := seedPlannedTask(t, f, "Заголовок")
	if err := f.dispatcher.Handle(ctx, planned); err != nil {
		t.Fatalf("developer dispatch: %v", err)
	}

	var requested platform.Event
	for _, e := range f.bus.Published() {
		if e.Type() == event.ReviewRequested {
			requested = e
		}
	}
	if requested == nil {
		t.Fatal("no ReviewRequested event was published by the developer dispatch")
	}
	return f, requested
}

func taskState(t *testing.T, f *dispatchFixture) shared.TaskState {
	t.Helper()
	task, err := f.tasks.Get(context.Background(), "proj-1", "TASK-001")
	if err != nil {
		t.Fatalf("Get task: %v", err)
	}
	return task.State()
}

func TestReview_ApprovedMovesTaskToTesting(t *testing.T) {
	rev := &fakeReviewer{approved: true, comment: "замечаний нет"}
	f, requested := reviewFixture(t, rev)

	if err := f.dispatcher.Handle(context.Background(), requested); err != nil {
		t.Fatalf("Handle(ReviewRequested): %v", err)
	}

	if got := taskState(t, f); got != shared.StateTesting {
		t.Errorf("task state = %v, want %v", got, shared.StateTesting)
	}
}

func TestReview_ChangesRequestedReturnsTaskToInProgress(t *testing.T) {
	rev := &fakeReviewer{approved: false, comment: "нет тестов"}
	f, requested := reviewFixture(t, rev)

	if err := f.dispatcher.Handle(context.Background(), requested); err != nil {
		t.Fatalf("Handle(ReviewRequested): %v", err)
	}

	if got := taskState(t, f); got != shared.StateInProgress {
		t.Errorf("task state = %v, want %v", got, shared.StateInProgress)
	}
}

// The reviewing agent gets the reviewer role and the same branch — not a new
// one, since the branch already exists.
func TestReview_HandsReviewerRoleAndExistingBranch(t *testing.T) {
	rev := &fakeReviewer{approved: true}
	f, requested := reviewFixture(t, rev)
	branchesBefore := len(f.repos.branches)

	if err := f.dispatcher.Handle(context.Background(), requested); err != nil {
		t.Fatalf("Handle(ReviewRequested): %v", err)
	}

	accepted := rev.accepted
	last := accepted[len(accepted)-1]
	if last.Role != string(shared.RoleReviewer) {
		t.Errorf("ExecutorTask.Role = %q, want %q", last.Role, shared.RoleReviewer)
	}
	if last.Branch != "feature/TASK-001" {
		t.Errorf("ExecutorTask.Branch = %q, want the task's existing branch", last.Branch)
	}
	if len(f.repos.branches) != branchesBefore {
		t.Errorf("review created %d extra branch(es); the branch already exists", len(f.repos.branches)-branchesBefore)
	}
}

// No verdict is ever invented. Each of these leaves the task in Review so a
// human decides — an invented approval would advance code toward merging with
// no real check behind it.
func TestReview_NoVerdictIsInventedOnFailure(t *testing.T) {
	cases := []struct {
		name  string
		setup func(*fakeReviewer)
	}{
		{
			name: "agent reported failure",
			setup: func(r *fakeReviewer) {
				r.statuses = []platform.ExecutionStatus{{State: statusFailed, Message: "exit 1"}}
			},
		},
		{
			name:  "verdict missing or unreadable",
			setup: func(r *fakeReviewer) { r.reviewErr = errors.New("no verdict was written") },
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rev := &fakeReviewer{}
			f, requested := reviewFixture(t, rev)
			tc.setup(rev)

			// An error may or may not be returned; what matters is the state.
			_ = f.dispatcher.Handle(context.Background(), requested)

			if got := taskState(t, f); got != shared.StateReview {
				t.Errorf("task state = %v, want it left at %v for a human", got, shared.StateReview)
			}
			if rev.finished == 0 {
				t.Error("Finish was never called; the sandbox would outlive the review")
			}
		})
	}
}

func TestReview_TimeoutLeavesTaskInReview(t *testing.T) {
	rev := &fakeReviewer{}
	f, requested := reviewFixture(t, rev)
	// Never terminal — the guard against a hung sandbox.
	rev.statuses = []platform.ExecutionStatus{{State: statusRunning}}
	f.dispatcher.ExecutionTimeout = 20 * time.Millisecond

	err := f.dispatcher.Handle(context.Background(), requested)
	if !errors.Is(err, ErrExecutionTimedOut) {
		t.Fatalf("Handle() error = %v, want %v", err, ErrExecutionTimedOut)
	}
	if got := taskState(t, f); got != shared.StateReview {
		t.Errorf("task state = %v, want %v", got, shared.StateReview)
	}
	if rev.finished == 0 {
		t.Error("Finish was never called after the timeout")
	}
}

// Without its own registry entry the review must not fall back to some other
// executor (TASK-086), and the task must stay in Review.
func TestReview_MissingReviewerEntryLeavesTaskInReview(t *testing.T) {
	ctx := context.Background()
	rev := &fakeReviewer{approved: true}

	// Developer only: no reviewer entry in the registry.
	f := newDispatchFixture(t, []string{"github.com/org/repo"}, ownDeveloper(t))
	rev.fakeBackend = f.backend
	f.dispatcher.NewReviewer = func() (ReviewExecutor, error) { return rev, nil }

	planned := seedPlannedTask(t, f, "Заголовок")
	if err := f.dispatcher.Handle(ctx, planned); err != nil {
		t.Fatalf("developer dispatch: %v", err)
	}
	var requested platform.Event
	for _, e := range f.bus.Published() {
		if e.Type() == event.ReviewRequested {
			requested = e
		}
	}

	err := f.dispatcher.Handle(ctx, requested)
	if !errors.Is(err, ErrNoExecutorForRole) {
		t.Fatalf("Handle() error = %v, want %v", err, ErrNoExecutorForRole)
	}
	if got := taskState(t, f); got != shared.StateReview {
		t.Errorf("task state = %v, want %v", got, shared.StateReview)
	}
}

// Reviewing does not create an Execution: StartTask is the only way to create
// one and it performs a Ready -> In Progress transition, which does not happen
// here. Asserted so the documented limitation cannot drift silently.
func TestReview_CreatesNoAdditionalExecution(t *testing.T) {
	rev := &fakeReviewer{approved: true}
	f, requested := reviewFixture(t, rev)

	queuedBefore := countEvents(f, event.ExecutionQueued)
	if err := f.dispatcher.Handle(context.Background(), requested); err != nil {
		t.Fatalf("Handle(ReviewRequested): %v", err)
	}
	if got := countEvents(f, event.ExecutionQueued); got != queuedBefore {
		t.Errorf("ExecutionQueued events = %d, want unchanged at %d", got, queuedBefore)
	}
}

func countEvents(f *dispatchFixture, typ string) int {
	n := 0
	for _, e := range f.bus.Published() {
		if e.Type() == typ {
			n++
		}
	}
	return n
}

// Reviewing must be wired through the real CompletionService, not by writing
// the task state directly — the state machine decides what Review may go to.
func TestReview_UsesCompletionServiceRules(t *testing.T) {
	rev := &fakeReviewer{approved: true}
	f, requested := reviewFixture(t, rev)

	// A dispatcher whose Completion service refuses everything must not leave
	// the task moved.
	f.dispatcher.Completion = &application.CompletionService{
		Tasks: f.tasks, Repositories: f.repos, Events: f.bus, Rules: refusingRules{},
	}

	if err := f.dispatcher.Handle(context.Background(), requested); err == nil {
		t.Fatal("Handle() error = nil, want the domain's refusal to propagate")
	}
	if got := taskState(t, f); got != shared.StateReview {
		t.Errorf("task state = %v, want %v", got, shared.StateReview)
	}
}

// refusingRules rejects every transition, standing in for a domain rule the
// orchestrator must not override.
type refusingRules struct{}

func (refusingRules) CanTransition(from, to shared.TaskState) error {
	return fmt.Errorf("refusingRules: %s -> %s is not allowed", from, to)
}

func (refusingRules) NextRole(shared.TaskState) (shared.Role, error) {
	return shared.RoleDeveloper, nil
}

var _ workflow.Rules = refusingRules{}
