package application

import (
	"context"
	"errors"
	"time"

	"ai-studio-os/internal/domain/artifact"
	"ai-studio-os/internal/domain/event"
	"ai-studio-os/internal/domain/execution"
	"ai-studio-os/internal/platform"
)

// ErrExecutionNotRunning is returned when an operation that requires an
// in-progress Execution (recording a produced Artifact) is attempted
// while the Execution is not Running (spec Execution Behavioral
// Invariant 4).
var ErrExecutionNotRunning = errors.New("application: execution is not running")

// ResultService implements the "Производство результата" step of the
// golden path: recording and publishing the Artifacts an Execution
// produces, and concluding the Execution itself. Artifact and Execution
// reference each other from both sides (spec Artifact Behavioral
// Invariant 3, spec Execution Behavioral Invariant 4) — RecordDraftArtifact
// keeps both sides consistent in one use-case, as EPIC-004's decomposition
// intends.
type ResultService struct {
	Projects   ProjectStore
	Tasks      TaskStore
	Executions ExecutionStore
	Artifacts  ArtifactStore
	Events     platform.EventBus
}

// RecordDraftArtifactParams are the inputs to RecordDraftArtifact. Payload
// is optional — a Draft may be filled in later via UpdateArtifactDraft.
type RecordDraftArtifactParams struct {
	ID          string
	ProjectID   string
	ExecutionID string
	Type        artifact.Type
	Origin      artifact.Origin
	Author      artifact.Author
	Payload     []byte
	Actor       string
}

// RecordDraftArtifact creates an Artifact in Draft within the given
// Project's boundary (spec Project Behavioral Invariant 4) and links it to
// the producing Execution from both sides: the Artifact's ProducedBy and
// the Execution's own produced-Artifact set (spec Execution Behavioral
// Invariant 4: only while Running). Publishes ArtifactCreated.
func (s *ResultService) RecordDraftArtifact(ctx context.Context, p RecordDraftArtifactParams) (*artifact.Artifact, error) {
	proj, err := s.Projects.Get(ctx, p.ProjectID)
	if err != nil {
		return nil, err
	}
	if !proj.AcceptsNewContent() {
		return nil, ErrProjectNotActive
	}

	run, err := s.Executions.Get(ctx, p.ExecutionID)
	if err != nil {
		return nil, err
	}
	if run.State() != execution.StateRunning {
		return nil, ErrExecutionNotRunning
	}

	a, created, err := artifact.New(p.ID, p.ProjectID, p.Type, p.Origin, p.Author, p.ExecutionID)
	if err != nil {
		return nil, err
	}
	if len(p.Payload) > 0 {
		if err := a.UpdateDraft(p.Payload, ""); err != nil {
			return nil, err
		}
	}

	if err := run.RecordArtifact(a.ID()); err != nil {
		return nil, err
	}
	if err := s.Artifacts.Save(ctx, a); err != nil {
		return nil, err
	}
	if err := s.Executions.Save(ctx, run); err != nil {
		return nil, err
	}
	return a, s.publish(ctx, event.ArtifactCreated, "artifact", p.Actor, p.ProjectID, a.ID(), created.At)
}

// RecordTestReportParams are the inputs to RecordTestReport.
type RecordTestReportParams struct {
	ID        string
	ProjectID string
	Report    []byte
	Author    artifact.Author
	Actor     string
}

// RecordTestReport stores what a QA agent found as a published TestReport
// artifact (EPIC-013, TASK-099).
//
// Deliberately not tied to an Execution, unlike RecordDraftArtifact: a QA run
// has no Execution to attach to — StartTask is the only way to create one and it
// performs a Ready -> In Progress transition, which does not happen when a task
// enters Testing. The specification allows exactly this: the reference to a
// producing Execution is "необязательная… у Artifact не обязательно есть такое
// Execution", and its absence "не делает Artifact менее полноценным"
// (docs/specifications/domain/artifact.md, Behavioral Invariant 3).
//
// Created and published in one step, because a report "либо есть целиком, либо
// его ещё нет" — the specification names TestReport as the type whose lifecycle
// passes Draft -> Published almost instantly.
//
// Recording a report is not a verdict: the acceptance decision stays with a
// human (docs/architecture/workflow.md), and nothing here moves the task.
func (s *ResultService) RecordTestReport(ctx context.Context, p RecordTestReportParams) (*artifact.Artifact, error) {
	proj, err := s.Projects.Get(ctx, p.ProjectID)
	if err != nil {
		return nil, err
	}
	if !proj.AcceptsNewContent() {
		return nil, ErrProjectNotActive
	}

	a, created, err := artifact.New(
		p.ID, p.ProjectID, artifact.Type("TestReport"), artifact.OriginProduced, p.Author, "",
	)
	if err != nil {
		return nil, err
	}
	if err := a.UpdateDraft(p.Report, ""); err != nil {
		return nil, err
	}
	published, err := a.Publish()
	if err != nil {
		return nil, err
	}
	if err := s.Artifacts.Save(ctx, a); err != nil {
		return nil, err
	}

	if err := s.publish(ctx, event.ArtifactCreated, "artifact", p.Actor, p.ProjectID, a.ID(), created.At); err != nil {
		return nil, err
	}
	return a, s.publish(ctx, event.ArtifactPublished, "artifact", p.Actor, p.ProjectID, a.ID(), published.At)
}

// UpdateArtifactDraft updates the Payload and/or Author of an Artifact
// still in Draft (spec Commands: UpdateDraft).
func (s *ResultService) UpdateArtifactDraft(ctx context.Context, artifactID string, payload []byte, author artifact.Author) error {
	a, err := s.Artifacts.Get(ctx, artifactID)
	if err != nil {
		return err
	}
	if err := a.UpdateDraft(payload, author); err != nil {
		return err
	}
	return s.Artifacts.Save(ctx, a)
}

// PublishArtifact transitions an Artifact Draft -> Published (spec
// Commands: Publish — requires a non-empty Payload). Publishes
// ArtifactPublished.
func (s *ResultService) PublishArtifact(ctx context.Context, artifactID, actor string) error {
	a, err := s.Artifacts.Get(ctx, artifactID)
	if err != nil {
		return err
	}
	published, err := a.Publish()
	if err != nil {
		return err
	}
	if err := s.Artifacts.Save(ctx, a); err != nil {
		return err
	}
	return s.publish(ctx, event.ArtifactPublished, "artifact", actor, a.ProjectID(), a.ID(), published.At)
}

// SucceedExecution transitions an Execution Running -> Succeeded (spec
// Commands: Succeed), finalizing its produced-Artifact set. Publishes
// ExecutionSucceeded. If Fail or Abort already concluded the Execution,
// the domain's own Behavioral Invariant 5 rejects this call — the
// use-case does not re-decide the race. projectID is required (BUGFIX-003:
// looking the owning Task up by its bare TaskID alone is no longer
// possible, since TASK-NNN is unique only within a Project, ADR-011) —
// also validated against the owning Task, not merely trusted.
func (s *ResultService) SucceedExecution(ctx context.Context, projectID, executionID, actor string) error {
	run, err := s.Executions.Get(ctx, executionID)
	if err != nil {
		return err
	}
	succeeded, err := run.Succeed()
	if err != nil {
		return err
	}
	if err := s.Executions.Save(ctx, run); err != nil {
		return err
	}
	return s.publishExecutionEvent(ctx, event.ExecutionSucceeded, projectID, actor, run, succeeded.At)
}

// FailExecution transitions an Execution Running -> Failed (spec Commands:
// Fail), keeping any Artifacts already produced. Publishes ExecutionFailed.
// See SucceedExecution on projectID.
func (s *ResultService) FailExecution(ctx context.Context, projectID, executionID, actor string) error {
	run, err := s.Executions.Get(ctx, executionID)
	if err != nil {
		return err
	}
	failed, err := run.Fail()
	if err != nil {
		return err
	}
	if err := s.Executions.Save(ctx, run); err != nil {
		return err
	}
	return s.publishExecutionEvent(ctx, event.ExecutionFailed, projectID, actor, run, failed.At)
}

// publishExecutionEvent looks up the owning Task by (projectID, TaskID) —
// both to confirm projectID is genuinely the Execution's own project (not
// merely trusted from the caller) and for the event's ProjectID field
// (Execution itself carries only a TaskID, per ADR-015: no cross-module
// reference beyond an id).
func (s *ResultService) publishExecutionEvent(ctx context.Context, eventType, projectID, actor string, run *execution.Execution, at time.Time) error {
	t, err := s.Tasks.Get(ctx, projectID, run.TaskID())
	if err != nil {
		return err
	}
	return s.publish(ctx, eventType, "execution", actor, t.ProjectID(), run.ID(), at)
}

func (s *ResultService) publish(ctx context.Context, eventType, source, actor, projectID, subjectID string, at time.Time) error {
	return s.Events.Publish(ctx, NewEvent(eventType, source, actor, projectID, subjectID, at))
}
