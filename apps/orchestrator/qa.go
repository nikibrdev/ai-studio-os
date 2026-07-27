package main

import (
	"context"
	"fmt"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/domain/artifact"
	"ai-studio-os/internal/domain/event"
	"ai-studio-os/internal/domain/shared"
	"ai-studio-os/internal/platform"
)

// ReportExecutor is an Executor that also reports what a QA check found.
//
// Primitive result for the same reason as the other role interfaces: this
// package must not name a concrete AI backend (module-boundaries.md).
type ReportExecutor interface {
	platform.Executor

	// Check reports the QA agent's findings. An error means no usable report —
	// the platform must then record nothing rather than invent reassurance.
	Check(ctx context.Context) (report string, err error)
}

// reportAuthor is who the platform records as responsible for the report.
//
// A concrete author rather than the "unknown" the specification also permits
// for fully automatic runs: the run is automatic, but it is attributable — a
// human reading the report should know it came from an agent, not a person.
const reportAuthor = "claude-code"

// enteredTesting reports whether the event moved a task into Testing.
//
// Derived from ReviewCompleted's attached target state, which is the only thing
// that distinguishes an approval from a request for changes — the event name
// alone is ambiguous (EPIC-004). A request for changes must not start a QA run.
func enteredTesting(e platform.Event) bool {
	if e.Type() != event.ReviewCompleted {
		return false
	}
	carrier, ok := e.(interface{ Data() map[string]string })
	if !ok {
		return false
	}
	return carrier.Data()["to"] == string(shared.StateTesting)
}

// dispatchEnteredTesting has a QA agent check a task that just reached Testing
// and records its report for the human who will make the acceptance decision.
//
// The task stays in Testing. CompleteTesting is never called from here — it
// would merge the pull request and finish the task, which is a human checkpoint
// (docs/architecture/workflow.md); nor can the agent call it itself, since no
// role reaches apps/api (EPIC-013).
func (d *Dispatcher) dispatchEnteredTesting(ctx context.Context, e platform.Event) error {
	projectID, taskID := e.ProjectID(), e.SubjectID()

	view, ok := d.Views.Get(projectID, taskID)
	if !ok {
		return fmt.Errorf("%w: %s/%s", ErrTaskNotInProjection, projectID, taskID)
	}

	executorID, err := d.executorForRole(ctx, shared.RoleQA)
	if err != nil {
		return err
	}

	repo, err := d.repositoryOf(ctx, projectID)
	if err != nil {
		return err
	}

	backend, err := d.NewReporter()
	if err != nil {
		return fmt.Errorf("orchestrator: build QA executor for %s/%s: %w", projectID, taskID, err)
	}

	task := platform.ExecutorTask{
		TaskID:             taskID,
		ProjectID:          projectID,
		Role:               string(shared.RoleQA),
		Title:              view.Title,
		Type:               view.Type,
		Scope:              view.Scope,
		AcceptanceCriteria: view.AcceptanceCriteria,
		Repository:         repo,
		// The task branch — the same one that was just reviewed. It already
		// exists; QA produces no commits and creates no branch.
		Branch: branchName(taskID, view.Title),
	}
	if err := backend.Accept(ctx, task); err != nil {
		return fmt.Errorf("orchestrator: accept QA of %s/%s: %w", projectID, taskID, err)
	}

	d.Log.Printf("dispatched QA of %s/%s to %s", projectID, taskID, executorID)
	return d.recordReport(ctx, backend, projectID, taskID)
}

// recordReport watches the QA run and stores its report.
//
// Every failure records nothing. An empty report that looks like a check is
// more dangerous than no report at all: a human would make the acceptance
// decision believing a check had happened.
func (d *Dispatcher) recordReport(ctx context.Context, backend ReportExecutor, projectID, taskID string) error {
	label := fmt.Sprintf("QA of %s/%s", projectID, taskID)

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
		d.Log.Printf("%s failed (%s); no report recorded, task stays in testing", label, status.Message)
		return nil
	}

	report, err := backend.Check(ctx)
	if err != nil {
		return fmt.Errorf("orchestrator: no usable report for %s: %w", label, err)
	}

	stored, err := d.Results.RecordTestReport(ctx, application.RecordTestReportParams{
		ID:        application.NewID(),
		ProjectID: projectID,
		Report:    []byte(report),
		Author:    artifact.Author(reportAuthor),
		Actor:     actor,
	})
	if err != nil {
		return fmt.Errorf("orchestrator: record QA report for %s/%s: %w", projectID, taskID, err)
	}

	d.Log.Printf("%s complete: report %s recorded, task awaiting a human's acceptance decision", label, stored.ID())
	return nil
}
