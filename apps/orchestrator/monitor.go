package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/domain/artifact"
	"ai-studio-os/internal/platform"
)

// Executor status values reported by an adapter (agents/claude-code's
// Status). Only these three are meaningful to the orchestrator; anything
// else is treated as still running, since an adapter reporting an unknown
// state is not grounds to abandon a live sandbox.
const (
	statusRunning   = "running"
	statusSucceeded = "succeeded"
	statusFailed    = "failed"
)

// Defaults for watching an execution. Fields on Dispatcher override them so
// tests do not wait minutes.
const (
	defaultStatusPollInterval = 10 * time.Second
	defaultExecutionTimeout   = 30 * time.Minute
)

// ErrExecutionTimedOut is returned when an execution never reaches a
// terminal state within the timeout — the guard against a hung sandbox.
var ErrExecutionTimedOut = errors.New("orchestrator: execution did not finish within the timeout")

// executionContext is everything monitorExecution needs about the work it
// is watching, gathered once by the dispatcher.
type executionContext struct {
	executionID string
	projectID   string
	taskID      string
	repository  string
	branch      string
	view        application.TaskView
}

// monitorExecution watches an accepted execution to its end and records the
// outcome through the Application Layer: artifacts and a pull request on
// success, a failed Execution otherwise.
//
// Finish is always called (deferred), so the sandbox never outlives the
// execution regardless of how this returns — the ephemeral working copy
// dies with the container (ADR-006).
//
// Runs synchronously inside the event handler. That blocks the poll loop
// for the duration of the execution, which is accepted for v1.0: the domain
// does not track whether an Executor is busy (AvailableForAssignment is
// only "is it Active"), so serialising work is what keeps two tasks from
// being handed to the same backend at once. Concurrency here needs a
// busy-tracking decision in the Executor specification, not a goroutine
// invented in the orchestrator.
func (d *Dispatcher) monitorExecution(ctx context.Context, backend platform.Executor, ec executionContext) error {
	defer func() {
		// Logged, not returned: Finish failing must not mask the outcome of
		// the work itself, but a sandbox that would not stop is worth seeing.
		if err := backend.Finish(ctx); err != nil {
			d.Log.Printf("finishing execution %s failed: %v", ec.executionID, err)
		}
	}()

	status, waitErr := d.waitForTerminalStatus(ctx, backend, ec.executionID)
	if waitErr != nil {
		return d.failExecution(ctx, ec, waitErr)
	}
	if status.State == statusFailed {
		// Not an error of ours: the agent ran and did not succeed. The
		// Execution records that outcome; the Task is left as it is.
		d.Log.Printf("execution %s reported failure: %s", ec.executionID, status.Message)
		return d.failExecution(ctx, ec, nil)
	}

	if err := d.recordOutcome(ctx, backend, ec); err != nil {
		return d.failExecution(ctx, ec, err)
	}
	return nil
}

// waitForTerminalStatus polls until the executor reports success or failure,
// the context ends, or the timeout expires. A failed poll is logged and
// retried: one transient docker error must not abandon a live execution.
func (d *Dispatcher) waitForTerminalStatus(ctx context.Context, backend platform.Executor, label string) (platform.ExecutionStatus, error) {
	deadline := time.Now().Add(d.executionTimeout())
	ticker := time.NewTicker(d.statusPollInterval())
	defer ticker.Stop()

	for {
		status, err := backend.Status(ctx)
		switch {
		case err != nil:
			d.Log.Printf("status of execution %s unavailable, retrying: %v", label, err)
		case status.State == statusSucceeded, status.State == statusFailed:
			return status, nil
		case status.State != statusRunning:
			// An adapter reporting something unexpected is not a reason to
			// tear down work that may still be progressing.
			d.Log.Printf("execution %s reported unknown state %q, treating as running", label, status.State)
		}

		if time.Now().After(deadline) {
			return platform.ExecutionStatus{}, fmt.Errorf("%w: %s", ErrExecutionTimedOut, label)
		}
		select {
		case <-ctx.Done():
			return platform.ExecutionStatus{}, ctx.Err()
		case <-ticker.C:
		}
	}
}

// recordOutcome records a successful execution: its artifacts, a pull
// request for review, the Execution's success, and the Task's move into
// Review.
//
// Order matters. Artifacts are recorded while the Execution is still
// Running, because that is the domain's requirement for attaching one
// (Execution Behavioral Invariant 4) — after Succeed it is no longer
// Running and RecordDraftArtifact would be rejected.
func (d *Dispatcher) recordOutcome(ctx context.Context, backend platform.Executor, ec executionContext) error {
	produced, err := backend.Artifacts(ctx)
	if err != nil {
		return fmt.Errorf("orchestrator: read artifacts of execution %s: %w", ec.executionID, err)
	}

	for _, a := range produced {
		if err := d.recordArtifact(ctx, ec, a); err != nil {
			return err
		}
	}

	if len(produced) == 0 {
		// A succeeded execution that produced no commits is a real outcome,
		// not a failure: there is simply nothing to open a pull request for.
		// The Execution is still concluded as succeeded and the Task still
		// goes to Review, where a human sees that nothing was produced.
		d.Log.Printf("execution %s succeeded without producing commits; no pull request to open", ec.executionID)
	} else if err := d.openPullRequest(ctx, ec); err != nil {
		return err
	}

	if err := d.Results.SucceedExecution(ctx, ec.projectID, ec.executionID, actor); err != nil {
		return fmt.Errorf("orchestrator: succeed execution %s: %w", ec.executionID, err)
	}
	if err := d.Completion.RequestReview(ctx, ec.projectID, ec.taskID, actor); err != nil {
		return fmt.Errorf("orchestrator: request review for %s/%s: %w", ec.projectID, ec.taskID, err)
	}

	d.Log.Printf("execution %s succeeded: %d artifact(s), task %s/%s in review",
		ec.executionID, len(produced), ec.projectID, ec.taskID)
	return nil
}

// recordArtifact stores one produced artifact and publishes it. The
// adapter's vocabulary already matches the domain's (Type "Commit", Origin
// "produced" — artifact.OriginProduced, Author "claude-code"), so the values
// pass through rather than being remapped; the identifier is the commit hash
// the adapter reported, which keeps the artifact traceable to real history.
func (d *Dispatcher) recordArtifact(ctx context.Context, ec executionContext, a platform.Artifact) error {
	stored, err := d.Results.RecordDraftArtifact(ctx, application.RecordDraftArtifactParams{
		ID:          a.ID,
		ProjectID:   ec.projectID,
		ExecutionID: ec.executionID,
		Type:        artifact.Type(a.Type),
		Origin:      artifact.Origin(a.Origin),
		Author:      artifact.Author(a.Author),
		Payload:     a.Payload,
		Actor:       actor,
	})
	if err != nil {
		return fmt.Errorf("orchestrator: record artifact %s: %w", a.ID, err)
	}
	if err := d.Results.PublishArtifact(ctx, stored.ID(), actor); err != nil {
		return fmt.Errorf("orchestrator: publish artifact %s: %w", stored.ID(), err)
	}
	return nil
}

// openPullRequest opens the pull request for the task's branch. Opening it
// is the caller's job, not the adapter's: the Executor contract reports
// produced commits and nothing more (agents/claude-code/README.md).
func (d *Dispatcher) openPullRequest(ctx context.Context, ec executionContext) error {
	title := ec.view.Title
	if title == "" {
		title = ec.taskID
	}
	prID, err := d.Repos.OpenPullRequest(ctx, ec.repository, ec.branch, title, pullRequestBody(ec))
	if err != nil {
		return fmt.Errorf("orchestrator: open pull request for %s: %w", ec.branch, err)
	}

	// RequestReview on the provider marks the pull request itself; the Task's
	// own move into Review happens through CompletionService.
	if err := d.Repos.RequestReview(ctx, ec.repository, prID); err != nil {
		// The pull request exists, which is the part that matters; failing to
		// annotate it must not fail the execution.
		d.Log.Printf("marking review requested on PR %s failed: %v", prID, err)
	}
	return nil
}

// pullRequestBody builds the description from the task's planning content,
// so a human reviewing the pull request sees what was asked for.
func pullRequestBody(ec executionContext) string {
	body := fmt.Sprintf("Задача: %s (проект %s)\n", ec.taskID, ec.projectID)
	if ec.view.Scope != "" {
		body += "\n## Scope\n\n" + ec.view.Scope + "\n"
	}
	if len(ec.view.AcceptanceCriteria) > 0 {
		body += "\n## Критерии приёмки\n\n"
		for _, c := range ec.view.AcceptanceCriteria {
			body += "- " + c + "\n"
		}
	}
	body += "\nPull request открыт автоматически (apps/orchestrator)."
	return body
}

// failExecution concludes the Execution as failed, carrying the cause (if
// any) alongside any error from the transition itself.
//
// The Task is deliberately not moved: ExecutionFailed does not drive a task
// transition anywhere in the platform (state-machine.md, and the projection
// does not derive one from it), so the task stays In Progress with a failed
// Execution and a human decides what happens next through apps/api. That is
// the existing domain design, not an omission here.
func (d *Dispatcher) failExecution(ctx context.Context, ec executionContext, cause error) error {
	if err := d.Results.FailExecution(ctx, ec.projectID, ec.executionID, actor); err != nil {
		if cause != nil {
			return fmt.Errorf("orchestrator: fail execution %s (%v): %w", ec.executionID, cause, err)
		}
		return fmt.Errorf("orchestrator: fail execution %s: %w", ec.executionID, err)
	}
	return cause
}

// statusPollInterval and executionTimeout fall back to the defaults unless a
// test (or a future configuration) sets them.
func (d *Dispatcher) statusPollInterval() time.Duration {
	if d.StatusPollInterval > 0 {
		return d.StatusPollInterval
	}
	return defaultStatusPollInterval
}

func (d *Dispatcher) executionTimeout() time.Duration {
	if d.ExecutionTimeout > 0 {
		return d.ExecutionTimeout
	}
	return defaultExecutionTimeout
}
