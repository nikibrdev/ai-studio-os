package main

import (
	"context"
	"fmt"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/domain/event"
	"ai-studio-os/internal/domain/shared"
	"ai-studio-os/internal/platform"
)

// PlanExecutor is an Executor that also reports a Definition of Ready proposal.
//
// Declared with primitive results for the same reason as ReviewExecutor: this
// package must not name a concrete AI backend (module-boundaries.md), so the
// adapter's own proposal type stops in main.go.
type PlanExecutor interface {
	platform.Executor

	// Propose reports the scope and acceptance criteria the agent prepared.
	// An error means no usable proposal — the platform must then leave the task
	// as it is rather than invent one.
	Propose(ctx context.Context) (scope string, criteria []string, err error)
}

// dispatchTaskCreated has a Project Manager agent prepare a task for a human to
// accept: it proposes a sharper scope and acceptance criteria, and the platform
// applies the proposal through the Application Layer.
//
// The task stays in Backlog. Accepting Definition of Ready is a human
// checkpoint (docs/architecture/workflow.md), and PlanTask is never called from
// here — nor can the agent call it itself, since no role reaches apps/api
// (EPIC-013).
func (d *Dispatcher) dispatchTaskCreated(ctx context.Context, e platform.Event) error {
	projectID, taskID := e.ProjectID(), e.SubjectID()

	view, ok := d.Views.Get(projectID, taskID)
	if !ok {
		return fmt.Errorf("%w: %s/%s", ErrTaskNotInProjection, projectID, taskID)
	}

	executorID, err := d.executorForRole(ctx, shared.RoleProjectManager)
	if err != nil {
		return err
	}

	repo, err := d.repositoryOf(ctx, projectID)
	if err != nil {
		return err
	}

	backend, err := d.NewPlanner()
	if err != nil {
		return fmt.Errorf("orchestrator: build planner for %s/%s: %w", projectID, taskID, err)
	}

	task := platform.ExecutorTask{
		TaskID:             taskID,
		ProjectID:          projectID,
		Role:               string(shared.RoleProjectManager),
		Title:              view.Title,
		Type:               view.Type,
		Scope:              view.Scope,
		AcceptanceCriteria: view.AcceptanceCriteria,
		Repository:         repo,
		// The base branch, not a task branch: a task branch does not exist yet
		// when the task is created, and a Project Manager produces no commits.
		// The clone still matters — a proposal should rest on the real code
		// rather than on guesses.
		Branch: baseBranch,
	}
	if err := backend.Accept(ctx, task); err != nil {
		return fmt.Errorf("orchestrator: accept planning of %s/%s: %w", projectID, taskID, err)
	}

	d.Log.Printf("dispatched planning of %s/%s to %s", projectID, taskID, executorID)
	return d.applyProposal(ctx, backend, projectID, taskID)
}

// applyProposal watches the planning run and applies what the agent prepared.
//
// Every failure leaves the task exactly as it was. An invented scope would be
// worse than none: a human accepting Definition of Ready would be reading
// something no one prepared, and would have no way to tell.
func (d *Dispatcher) applyProposal(ctx context.Context, backend PlanExecutor, projectID, taskID string) error {
	label := fmt.Sprintf("planning of %s/%s", projectID, taskID)

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
		// The agent ran and did not produce a proposal. Not an error of ours,
		// and not grounds to invent one.
		d.Log.Printf("%s failed (%s); task stays as it is for a human", label, status.Message)
		return nil
	}

	scope, criteria, err := backend.Propose(ctx)
	if err != nil {
		return fmt.Errorf("orchestrator: no usable proposal for %s: %w", label, err)
	}

	err = d.Planning.RefineTask(ctx, application.RefineTaskParams{
		ProjectID: projectID, TaskID: taskID,
		Scope: scope, AcceptanceCriteria: criteria, Actor: actor,
	})
	if err != nil {
		return fmt.Errorf("orchestrator: apply proposal for %s/%s: %w", projectID, taskID, err)
	}

	d.Log.Printf("%s applied: task refined and awaiting a human's Definition of Ready", label)
	return nil
}

// taskCreated reports whether the event announces a new task.
func taskCreated(e platform.Event) bool { return e.Type() == event.TaskCreated }
