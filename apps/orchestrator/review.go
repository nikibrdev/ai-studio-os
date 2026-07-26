package main

import (
	"context"
	"fmt"

	"ai-studio-os/internal/domain/event"
	"ai-studio-os/internal/domain/shared"
	"ai-studio-os/internal/platform"
)

// ReviewExecutor is an Executor that also reports a review decision.
//
// Declared with primitive results rather than an adapter's own verdict type
// on purpose: this package must not name a concrete AI backend
// (module-boundaries.md forbids knowledge of specific providers outside the
// composition root). main.go adapts the real adapter to this interface, which
// is exactly the seam the documented narrow exception exists for.
//
// Reviewing needs one thing the Executor contract does not offer — a
// decision. ADR-005 fixes four capabilities and a verdict is none of them: an
// artifact is evidence of work produced, a verdict is a judgement about work.
// So it lives here, alongside the contract, not inside it.
type ReviewExecutor interface {
	platform.Executor

	// Review reports the decision the agent produced. An error means no
	// usable decision — the platform must then leave the call to a human
	// rather than assume one.
	Review(ctx context.Context) (approved bool, comment string, err error)
}

// dispatchReviewRequested runs a review: hand the task's branch to a
// reviewing agent, then apply its verdict through the Application Layer.
//
// Automating this role is permitted where PM and QA are not: ADR-008 allows
// an agent reviewer, Review is not one of ADR-007's human checkpoints, and a
// review verdict does not merge anything — merging happens at the QA step,
// which stays with a human until the confirmation mechanism exists.
//
// No Execution is created. WorkService.StartTask is the only way to create
// one and it performs a Ready -> In Progress transition, which is not what
// happens here (the task is already in Review). The consequence is honest and
// documented: review work leaves no ExecutionQueued/Started/Succeeded trail,
// so it is not visible in the journal the way development work is — the
// ReviewRequested -> ReviewCompleted pair is its only record.
func (d *Dispatcher) dispatchReviewRequested(ctx context.Context, e platform.Event) error {
	projectID, taskID := e.ProjectID(), e.SubjectID()

	view, ok := d.Views.Get(projectID, taskID)
	if !ok {
		return fmt.Errorf("%w: %s/%s", ErrTaskNotInProjection, projectID, taskID)
	}

	executorID, err := d.executorForRole(ctx, shared.RoleReviewer)
	if err != nil {
		return err
	}

	repo, err := d.repositoryOf(ctx, projectID)
	if err != nil {
		return err
	}

	// The branch already exists — this orchestrator created it when it
	// dispatched the Developer, and the name is derived the same way. If a
	// human opened the pull request on a differently named branch, the clone
	// will not find it and Accept fails, leaving the task in Review.
	branch := branchName(taskID, view.Title)

	backend, err := d.NewReviewer()
	if err != nil {
		return fmt.Errorf("orchestrator: build reviewer for %s/%s: %w", projectID, taskID, err)
	}

	task := platform.ExecutorTask{
		TaskID:             taskID,
		ProjectID:          projectID,
		Role:               string(shared.RoleReviewer),
		Title:              view.Title,
		Type:               view.Type,
		Scope:              view.Scope,
		AcceptanceCriteria: view.AcceptanceCriteria,
		Repository:         repo,
		Branch:             branch,
	}
	if err := backend.Accept(ctx, task); err != nil {
		return fmt.Errorf("orchestrator: accept review of %s/%s: %w", projectID, taskID, err)
	}

	d.Log.Printf("dispatched review of %s/%s to %s, branch %s", projectID, taskID, executorID, branch)
	return d.applyReview(ctx, backend, projectID, taskID)
}

// applyReview watches a review to its end and applies the verdict.
//
// Every failure leaves the task in Review without calling CompleteReview.
// That is the whole point: an invented "approved" would advance code toward
// merging with no real check behind it, and an invented "changes requested"
// would send a developer back to work for no stated reason. When the platform
// does not know, the decision belongs to a human — who can act through
// apps/api, exactly as before this role was automated.
func (d *Dispatcher) applyReview(ctx context.Context, backend ReviewExecutor, projectID, taskID string) error {
	label := fmt.Sprintf("review of %s/%s", projectID, taskID)

	defer func() {
		if err := backend.Finish(ctx); err != nil {
			d.Log.Printf("finishing %s failed: %v", label, err)
		}
	}()

	status, err := d.waitForTerminalStatus(ctx, backend, label)
	if err != nil {
		return err
	}
	if status.State == statusFailed {
		// The agent ran and did not get there. Not an error of ours, and not
		// grounds to guess the verdict.
		d.Log.Printf("%s failed (%s); task stays in review for a human", label, status.Message)
		return nil
	}

	approved, comment, err := backend.Review(ctx)
	if err != nil {
		return fmt.Errorf("orchestrator: no usable verdict for %s: %w", label, err)
	}

	if err := d.Completion.CompleteReview(ctx, projectID, taskID, approved, actor); err != nil {
		return fmt.Errorf("orchestrator: complete review of %s/%s: %w", projectID, taskID, err)
	}

	d.Log.Printf("%s complete: approved=%t %s", label, approved, comment)
	return nil
}

// reviewRequested reports whether the event asks for a review.
func reviewRequested(e platform.Event) bool { return e.Type() == event.ReviewRequested }
