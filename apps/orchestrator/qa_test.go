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
	"ai-studio-os/internal/platform"
)

// fakeReporter is a fakeBackend that also reports QA findings.
type fakeReporter struct {
	*fakeBackend
	report   string
	checkErr error
}

func (r *fakeReporter) Check(context.Context) (string, error) {
	if r.checkErr != nil {
		return "", r.checkErr
	}
	return r.report, nil
}

// ownQA is this orchestrator's own registry entry for the QA role.
func ownQA(t *testing.T) *executor.Executor {
	t.Helper()
	e, _, err := executor.New(
		executorIDForRole(shared.RoleQA), "claude-code", []shared.Role{shared.RoleQA},
	)
	if err != nil {
		t.Fatalf("executor.New: %v", err)
	}
	if _, err := e.Activate(); err != nil {
		t.Fatalf("Activate: %v", err)
	}
	return e
}

// reviewCompleted builds the event that moves a task out of Review. The target
// state is what distinguishes an approval from a request for changes — the
// event name alone is ambiguous (EPIC-004).
func reviewCompleted(to shared.TaskState) platform.Event {
	return application.NewEvent(event.ReviewCompleted, "git", "reviewer", "proj-1", "TASK-001", nowForTest()).
		WithData(map[string]string{"to": string(to)})
}

// qaFixture drives a task to Testing and wires a QA backend.
func qaFixture(t *testing.T, reporter *fakeReporter) *dispatchFixture {
	t.Helper()
	ctx := context.Background()

	f := newDispatchFixture(t, []string{"github.com/org/repo"}, ownDeveloper(t), ownQA(t))
	reporter.fakeBackend = f.backend
	f.dispatcher.NewReporter = func() (ReportExecutor, error) { return reporter, nil }

	// The developer run must produce a commit so a pull request is opened and
	// its reference recorded. Without it CompleteTesting could never succeed
	// (ErrPullRequestUnknown), and the assertions that nothing merges would pass
	// for that reason rather than because the checkpoint is respected.
	f.backend.artifacts = []platform.Artifact{commitArtifact("abc123", "feat: коммит")}

	// Through the real services, so Testing is a state the machine allows.
	planned := seedPlannedTask(t, f, "Заголовок задачи")
	if err := f.dispatcher.Handle(ctx, planned); err != nil {
		t.Fatalf("developer dispatch: %v", err)
	}
	if err := f.dispatcher.Completion.CompleteReview(ctx, "proj-1", "TASK-001", true, "reviewer"); err != nil {
		t.Fatalf("CompleteReview: %v", err)
	}

	// Bring the projection up to date the way the Poller does in production
	// (Seed): it must know the pull request reference the developer dispatch
	// recorded on ReviewRequested (BUGFIX-009).
	//
	// Not cosmetic. Without it CompleteTesting could not succeed even if the
	// code wrongly called it — ErrPullRequestUnknown would stop it — and the
	// assertions that nothing merges would pass for that reason instead of
	// because the checkpoint is respected. Verified by breaking the code on
	// purpose: with this replay the violation is caught, without it, it is not.
	for _, e := range f.bus.Published() {
		if err := f.dispatcher.Observe(ctx, e); err != nil {
			t.Fatalf("Observe: %v", err)
		}
	}
	return f
}

func qaTaskState(t *testing.T, f *dispatchFixture) shared.TaskState {
	t.Helper()
	tsk, err := f.tasks.Get(context.Background(), "proj-1", "TASK-001")
	if err != nil {
		t.Fatalf("Get task: %v", err)
	}
	return tsk.State()
}

// countEventType reports how many events of a type were published.
func countEventType(f *dispatchFixture, typ string) int {
	n := 0
	for _, e := range f.bus.Published() {
		if e.Type() == typ {
			n++
		}
	}
	return n
}

func TestQA_RecordsReportAndLeavesTaskInTesting(t *testing.T) {
	ctx := context.Background()
	reporter := &fakeReporter{report: "Проверил критерии: два из трёх сходятся."}
	f := qaFixture(t, reporter)
	// The fixture shares one fake backend with the developer dispatch that got
	// the task to Testing, and that dispatch published its own artifact — so
	// every count here is a delta around the QA run, never a total.
	finishedBefore := reporter.finished
	artifactsBefore := countEventType(f, event.ArtifactPublished)

	if err := f.dispatcher.Handle(ctx, reviewCompleted(shared.StateTesting)); err != nil {
		t.Fatalf("Handle(ReviewCompleted→testing): %v", err)
	}

	// The report is stored as a published artifact a human can read.
	if n := countEventType(f, event.ArtifactPublished) - artifactsBefore; n != 1 {
		t.Errorf("ArtifactPublished published %d times for the QA run, want 1", n)
	}

	// The point of the epic: checked, not decided.
	if got := qaTaskState(t, f); got != shared.StateTesting {
		t.Errorf("task state = %v, want it left in %v for a human", got, shared.StateTesting)
	}
	if n := countEventType(f, event.TestsPassed); n != 0 {
		t.Errorf("TestsPassed published %d times, want 0 — automation must not make the acceptance decision", n)
	}
	if n := countEventType(f, event.MergeCompleted); n != 0 {
		t.Errorf("MergeCompleted published %d times, want 0 — automation must not merge", n)
	}
	if n := len(f.repos.MergeCalls); n != 0 {
		t.Errorf("MergePullRequest called %d times, want 0", n)
	}
	if got := reporter.finished - finishedBefore; got != 1 {
		t.Errorf("Finish called %d times for the QA run, want 1", got)
	}
}

// QA checks the branch that was just reviewed — it already exists, and QA
// produces no commits.
func TestQA_HandsQARoleAndExistingBranch(t *testing.T) {
	ctx := context.Background()
	reporter := &fakeReporter{report: "ок"}
	f := qaFixture(t, reporter)
	branchesBefore := len(f.repos.branches)

	if err := f.dispatcher.Handle(ctx, reviewCompleted(shared.StateTesting)); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	accepted := reporter.accepted[len(reporter.accepted)-1]
	if accepted.Role != string(shared.RoleQA) {
		t.Errorf("ExecutorTask.Role = %q, want %q", accepted.Role, shared.RoleQA)
	}
	if accepted.Branch != "feature/TASK-001" {
		t.Errorf("ExecutorTask.Branch = %q, want the task's existing branch", accepted.Branch)
	}
	if len(f.repos.branches) != branchesBefore {
		t.Errorf("QA created %d branch(es); it commits nothing", len(f.repos.branches)-branchesBefore)
	}
}

// A request for changes sends the task back to the developer — there is nothing
// for QA to check, and starting a run would waste a sandbox on a task that is
// not in Testing.
func TestQA_NotDispatchedWhenChangesRequested(t *testing.T) {
	ctx := context.Background()
	reporter := &fakeReporter{report: "не должно быть вызвано"}
	f := qaFixture(t, reporter)
	acceptedBefore := len(reporter.accepted)
	artifactsBefore := countEventType(f, event.ArtifactPublished)

	if err := f.dispatcher.Handle(ctx, reviewCompleted(shared.StateInProgress)); err != nil {
		t.Fatalf("Handle(ReviewCompleted→in-progress): %v", err)
	}

	if len(reporter.accepted) != acceptedBefore {
		t.Errorf("QA was dispatched for a request for changes: %d new Accept calls", len(reporter.accepted)-acceptedBefore)
	}
	if n := countEventType(f, event.ArtifactPublished) - artifactsBefore; n != 0 {
		t.Errorf("ArtifactPublished published %d times, want 0", n)
	}
}

// Nothing is invented: an absent or unusable report records nothing. An empty
// report that looks like a check is more dangerous than no report — a human
// would decide believing a check had happened.
func TestQA_NoReportIsInventedOnFailure(t *testing.T) {
	cases := []struct {
		name  string
		setup func(*fakeReporter)
	}{
		{
			name: "agent reported failure",
			setup: func(r *fakeReporter) {
				r.statuses = []platform.ExecutionStatus{{State: statusFailed, Message: "exit 1"}}
			},
		},
		{
			name:  "report missing or unreadable",
			setup: func(r *fakeReporter) { r.checkErr = errors.New("no report was written") },
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			reporter := &fakeReporter{}
			f := qaFixture(t, reporter)
			tc.setup(reporter)
			artifactsBefore := countEventType(f, event.ArtifactPublished)

			// An error may or may not be returned; the state is what matters.
			_ = f.dispatcher.Handle(ctx, reviewCompleted(shared.StateTesting))

			if n := countEventType(f, event.ArtifactPublished) - artifactsBefore; n != 0 {
				t.Errorf("ArtifactPublished published %d times, want 0 — no report to record", n)
			}
			if got := qaTaskState(t, f); got != shared.StateTesting {
				t.Errorf("task state = %v, want %v", got, shared.StateTesting)
			}
			if n := countEventType(f, event.TestsPassed); n != 0 {
				t.Errorf("TestsPassed published %d times, want 0", n)
			}
			if reporter.finished == 0 {
				t.Error("Finish was never called; the sandbox would outlive the run")
			}
		})
	}
}

func TestQA_TimeoutRecordsNothing(t *testing.T) {
	ctx := context.Background()
	reporter := &fakeReporter{}
	f := qaFixture(t, reporter)
	// Never terminal — the guard against a hung sandbox.
	reporter.statuses = []platform.ExecutionStatus{{State: statusRunning}}
	f.dispatcher.ExecutionTimeout = 20 * time.Millisecond
	artifactsBefore := countEventType(f, event.ArtifactPublished)

	if err := f.dispatcher.Handle(ctx, reviewCompleted(shared.StateTesting)); !errors.Is(err, ErrExecutionTimedOut) {
		t.Fatalf("Handle() error = %v, want %v", err, ErrExecutionTimedOut)
	}
	if n := countEventType(f, event.ArtifactPublished) - artifactsBefore; n != 0 {
		t.Errorf("ArtifactPublished published %d times, want 0", n)
	}
	if reporter.finished == 0 {
		t.Error("Finish was never called after the timeout")
	}
}

// Without its own registry entry the run must not fall back to another role's
// executor (TASK-086).
func TestQA_MissingQAEntryRecordsNothing(t *testing.T) {
	ctx := context.Background()
	reporter := &fakeReporter{report: "ок"}

	// Developer only: no qa entry in the registry.
	f := newDispatchFixture(t, []string{"github.com/org/repo"}, ownDeveloper(t))
	reporter.fakeBackend = f.backend
	f.dispatcher.NewReporter = func() (ReportExecutor, error) { return reporter, nil }
	planned := seedPlannedTask(t, f, "Заголовок задачи")
	if err := f.dispatcher.Handle(ctx, planned); err != nil {
		t.Fatalf("developer dispatch: %v", err)
	}
	if err := f.dispatcher.Completion.CompleteReview(ctx, "proj-1", "TASK-001", true, "reviewer"); err != nil {
		t.Fatalf("CompleteReview: %v", err)
	}

	artifactsBefore := countEventType(f, event.ArtifactPublished)

	err := f.dispatcher.Handle(ctx, reviewCompleted(shared.StateTesting))
	if !errors.Is(err, ErrNoExecutorForRole) {
		t.Fatalf("Handle() error = %v, want %v", err, ErrNoExecutorForRole)
	}
	if n := countEventType(f, event.ArtifactPublished) - artifactsBefore; n != 0 {
		t.Errorf("ArtifactPublished published %d times, want 0", n)
	}
}
