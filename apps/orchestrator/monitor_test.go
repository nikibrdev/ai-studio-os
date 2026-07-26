package main

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"ai-studio-os/internal/domain/execution"
	"ai-studio-os/internal/domain/shared"
	"ai-studio-os/internal/platform"
)

// commitArtifact mirrors what agents/claude-code reports for a produced
// commit: the adapter's vocabulary already matches the domain's.
func commitArtifact(hash, subject string) platform.Artifact {
	return platform.Artifact{
		ID: hash, Type: "Commit", Origin: "produced", Author: "claude-code",
		Payload: []byte(subject),
	}
}

// runDispatch drives one TaskPlanned through dispatch and monitoring.
func runDispatch(t *testing.T, f *dispatchFixture) error {
	t.Helper()
	planned := seedPlannedTask(t, f, "Заголовок задачи")
	return f.dispatcher.Handle(context.Background(), planned)
}

func TestMonitor_SuccessRecordsArtifactsOpensPRAndRequestsReview(t *testing.T) {
	ctx := context.Background()
	f := newDispatchFixture(t, []string{"github.com/org/repo"}, ownDeveloper(t))
	f.backend.statuses = []platform.ExecutionStatus{
		{State: statusRunning},
		{State: statusSucceeded},
	}
	f.backend.artifacts = []platform.Artifact{
		commitArtifact("abc123", "feat: первый коммит"),
		commitArtifact("def456", "test: покрытие"),
	}

	if err := runDispatch(t, f); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	// Both commits recorded and published.
	for _, hash := range []string{"abc123", "def456"} {
		a, err := f.artifacts.Get(ctx, hash)
		if err != nil {
			t.Fatalf("artifact %s not stored: %v", hash, err)
		}
		if a.State() != "published" {
			t.Errorf("artifact %s state = %v, want published", hash, a.State())
		}
	}

	// Pull request opened on the task branch, with the task's content in it.
	if len(f.repos.pullRequests) != 1 {
		t.Fatalf("OpenPullRequest called %d times, want 1", len(f.repos.pullRequests))
	}
	pr := f.repos.pullRequests[0]
	if pr.repo != "github.com/org/repo" || pr.branch != "feature/TASK-001" {
		t.Errorf("pull request = %s on %s, want feature/TASK-001 on github.com/org/repo", pr.branch, pr.repo)
	}
	if pr.title != "Заголовок задачи" {
		t.Errorf("pull request title = %q, want the task title", pr.title)
	}
	if !strings.Contains(pr.body, "TASK-001") || !strings.Contains(pr.body, "критерий один") {
		t.Errorf("pull request body = %q, want the task id and acceptance criteria", pr.body)
	}

	// Execution succeeded and the task reached Review.
	run := f.lastExecution(t)
	if run.State() != execution.StateSucceeded {
		t.Errorf("execution state = %v, want %v", run.State(), execution.StateSucceeded)
	}
	task, err := f.tasks.Get(ctx, "proj-1", "TASK-001")
	if err != nil {
		t.Fatalf("Get task: %v", err)
	}
	if task.State() != shared.StateReview {
		t.Errorf("task state = %v, want %v", task.State(), shared.StateReview)
	}
	if f.backend.finished != 1 {
		t.Errorf("Finish called %d times, want 1", f.backend.finished)
	}
}

// A failing agent is a real outcome, not an orchestrator error: the
// Execution records the failure and the sandbox is torn down.
func TestMonitor_ExecutorReportedFailureFailsExecutionWithoutPullRequest(t *testing.T) {
	ctx := context.Background()
	f := newDispatchFixture(t, []string{"github.com/org/repo"}, ownDeveloper(t))
	f.backend.statuses = []platform.ExecutionStatus{{State: statusFailed, Message: "exit code 1"}}

	if err := runDispatch(t, f); err != nil {
		t.Fatalf("Handle() error = %v, want nil — a failed agent is an outcome, not an error", err)
	}

	run := f.lastExecution(t)
	if run.State() != execution.StateFailed {
		t.Errorf("execution state = %v, want %v", run.State(), execution.StateFailed)
	}
	if len(f.repos.pullRequests) != 0 {
		t.Error("pull request opened for a failed execution")
	}
	if f.backend.finished != 1 {
		t.Errorf("Finish called %d times, want 1", f.backend.finished)
	}

	// ExecutionFailed drives no task transition anywhere in the platform, so
	// the task stays In Progress and a human decides what happens next.
	task, err := f.tasks.Get(ctx, "proj-1", "TASK-001")
	if err != nil {
		t.Fatalf("Get task: %v", err)
	}
	if task.State() != shared.StateInProgress {
		t.Errorf("task state = %v, want %v (ExecutionFailed moves no task)", task.State(), shared.StateInProgress)
	}
}

func TestMonitor_TimeoutFailsExecution(t *testing.T) {
	f := newDispatchFixture(t, []string{"github.com/org/repo"}, ownDeveloper(t))
	// Never terminal — the guard against a hung sandbox.
	f.backend.statuses = []platform.ExecutionStatus{{State: statusRunning}}
	f.dispatcher.ExecutionTimeout = 20 * time.Millisecond

	err := runDispatch(t, f)
	if !errors.Is(err, ErrExecutionTimedOut) {
		t.Fatalf("Handle() error = %v, want %v", err, ErrExecutionTimedOut)
	}
	if run := f.lastExecution(t); run.State() != execution.StateFailed {
		t.Errorf("execution state = %v, want %v", run.State(), execution.StateFailed)
	}
	if f.backend.finished != 1 {
		t.Errorf("Finish called %d times, want 1", f.backend.finished)
	}
}

// A transient Status error must be retried, not treated as failure.
func TestMonitor_TransientStatusErrorIsRetried(t *testing.T) {
	f := newDispatchFixture(t, []string{"github.com/org/repo"}, ownDeveloper(t))
	f.backend.statusErr = errors.New("docker daemon busy")
	f.dispatcher.ExecutionTimeout = 20 * time.Millisecond

	// With Status permanently failing the only possible end is the timeout —
	// which proves the error itself did not end the wait early.
	if err := runDispatch(t, f); !errors.Is(err, ErrExecutionTimedOut) {
		t.Fatalf("Handle() error = %v, want %v", err, ErrExecutionTimedOut)
	}
	if f.backend.statusN == 0 && f.backend.statusErr == nil {
		t.Error("Status was never polled")
	}
}

func TestMonitor_ArtifactsErrorFailsExecution(t *testing.T) {
	f := newDispatchFixture(t, []string{"github.com/org/repo"}, ownDeveloper(t))
	wantErr := errors.New("git log failed")
	f.backend.artifactsErr = wantErr

	if err := runDispatch(t, f); !errors.Is(err, wantErr) {
		t.Fatalf("Handle() error = %v, want wrapping %v", err, wantErr)
	}
	if run := f.lastExecution(t); run.State() != execution.StateFailed {
		t.Errorf("execution state = %v, want %v", run.State(), execution.StateFailed)
	}
	if len(f.repos.pullRequests) != 0 {
		t.Error("pull request opened despite artifacts being unreadable")
	}
	if f.backend.finished != 1 {
		t.Errorf("Finish called %d times, want 1", f.backend.finished)
	}
}

func TestMonitor_PullRequestErrorFailsExecution(t *testing.T) {
	ctx := context.Background()
	f := newDispatchFixture(t, []string{"github.com/org/repo"}, ownDeveloper(t))
	wantErr := errors.New("github unavailable")
	f.repos.openPRErr = wantErr
	f.backend.artifacts = []platform.Artifact{commitArtifact("abc123", "feat: коммит")}

	if err := runDispatch(t, f); !errors.Is(err, wantErr) {
		t.Fatalf("Handle() error = %v, want wrapping %v", err, wantErr)
	}
	if run := f.lastExecution(t); run.State() != execution.StateFailed {
		t.Errorf("execution state = %v, want %v", run.State(), execution.StateFailed)
	}
	// The artifact was recorded before the pull request was attempted, and
	// that record stands: the work happened even though the PR did not.
	if _, err := f.artifacts.Get(ctx, "abc123"); err != nil {
		t.Errorf("artifact recorded before the failure should still exist: %v", err)
	}
	if f.backend.finished != 1 {
		t.Errorf("Finish called %d times, want 1", f.backend.finished)
	}
}

// Success with no commits is a real outcome: nothing to open a pull request
// for, but the Execution still succeeds and the task still goes to Review,
// where a human can see nothing was produced.
func TestMonitor_SuccessWithoutArtifactsStillCompletes(t *testing.T) {
	ctx := context.Background()
	f := newDispatchFixture(t, []string{"github.com/org/repo"}, ownDeveloper(t))
	f.backend.artifacts = nil

	if err := runDispatch(t, f); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	if len(f.repos.pullRequests) != 0 {
		t.Error("pull request opened with no commits to review")
	}
	if run := f.lastExecution(t); run.State() != execution.StateSucceeded {
		t.Errorf("execution state = %v, want %v", run.State(), execution.StateSucceeded)
	}
	task, err := f.tasks.Get(ctx, "proj-1", "TASK-001")
	if err != nil {
		t.Fatalf("Get task: %v", err)
	}
	if task.State() != shared.StateReview {
		t.Errorf("task state = %v, want %v", task.State(), shared.StateReview)
	}
}

// Finish failing must not change the outcome of the work itself.
func TestMonitor_FinishErrorDoesNotFailSuccessfulExecution(t *testing.T) {
	f := newDispatchFixture(t, []string{"github.com/org/repo"}, ownDeveloper(t))
	f.backend.finishErr = errors.New("container already gone")
	f.backend.artifacts = []platform.Artifact{commitArtifact("abc123", "feat: коммит")}

	if err := runDispatch(t, f); err != nil {
		t.Fatalf("Handle() error = %v, want nil — Finish failing is logged, not returned", err)
	}
	if run := f.lastExecution(t); run.State() != execution.StateSucceeded {
		t.Errorf("execution state = %v, want %v", run.State(), execution.StateSucceeded)
	}
}

// Marking the pull request is cosmetic; failing it must not fail the work.
func TestMonitor_RequestReviewOnProviderFailureDoesNotFailExecution(t *testing.T) {
	f := newDispatchFixture(t, []string{"github.com/org/repo"}, ownDeveloper(t))
	f.repos.requestReviewErr = errors.New("cannot comment")
	f.backend.artifacts = []platform.Artifact{commitArtifact("abc123", "feat: коммит")}

	if err := runDispatch(t, f); err != nil {
		t.Fatalf("Handle() error = %v, want nil — annotating the PR is not essential", err)
	}
	if run := f.lastExecution(t); run.State() != execution.StateSucceeded {
		t.Errorf("execution state = %v, want %v", run.State(), execution.StateSucceeded)
	}
}
