package application_test

import (
	"context"
	"errors"
	"testing"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/application/inmemory"
	"ai-studio-os/internal/domain/artifact"
	"ai-studio-os/internal/domain/event"
	"ai-studio-os/internal/domain/execution"
	"ai-studio-os/internal/domain/shared"
	"ai-studio-os/internal/domain/workflow"
)

// resultFixture wires a full set of in-memory stores plus the planning
// and work services needed to reach a Running Execution — the starting
// point every ResultService test needs.
type resultFixture struct {
	result     *application.ResultService
	bus        *inmemory.EventBus
	tasks      application.TaskStore
	executions application.ExecutionStore
	artifacts  application.ArtifactStore
}

func newResultFixture(t *testing.T) resultFixture {
	t.Helper()
	ctx := context.Background()

	projects := inmemory.NewProjectStore()
	tasks := inmemory.NewTaskStore()
	executors := inmemory.NewExecutorStore()
	executions := inmemory.NewExecutionStore()
	artifacts := inmemory.NewArtifactStore()
	bus := inmemory.NewEventBus()
	rules := workflow.Machine{}

	newActiveProject(t, projects)
	planning := &application.TaskPlanningService{Projects: projects, Tasks: tasks, Events: bus, Rules: rules}
	work := &application.WorkService{Tasks: tasks, Executors: executors, Executions: executions, Events: bus, Rules: rules}

	if _, err := planning.CreateTask(ctx, application.CreateTaskParams{ID: "task-1", ProjectID: "proj-1", Title: "Задача", Type: "feature"}); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	if err := planning.PlanTask(ctx, "task-1", ""); err != nil {
		t.Fatalf("PlanTask: %v", err)
	}
	saveExecutor(t, executors, true, shared.RoleDeveloper)
	if _, err := work.StartTask(ctx, application.StartTaskParams{TaskID: "task-1", ExecutorID: "executor-1"}); err != nil {
		t.Fatalf("StartTask: %v", err)
	}

	result := &application.ResultService{Projects: projects, Tasks: tasks, Executions: executions, Artifacts: artifacts, Events: bus}
	return resultFixture{result: result, bus: bus, tasks: tasks, executions: executions, artifacts: artifacts}
}

func runningExecutionID(t *testing.T, f resultFixture) string {
	t.Helper()
	// StartTask always spawns exactly one Execution for task-1 in this
	// fixture; its identifier is generated, so recover it from the store
	// via the events already published rather than assuming a value.
	for _, e := range f.bus.Published() {
		if e.Type() == event.ExecutionStarted {
			return e.SubjectID()
		}
	}
	t.Fatal("no ExecutionStarted event found in fixture")
	return ""
}

func TestRecordDraftArtifact_Success(t *testing.T) {
	ctx := context.Background()
	f := newResultFixture(t)
	execID := runningExecutionID(t, f)

	a, err := f.result.RecordDraftArtifact(ctx, application.RecordDraftArtifactParams{
		ID: "art-1", ProjectID: "proj-1", ExecutionID: execID,
		Type: artifact.Type("PullRequest"), Origin: artifact.OriginProduced, Author: artifact.Author("developer"),
		Actor: "developer:executor-1",
	})
	if err != nil {
		t.Fatalf("RecordDraftArtifact: %v", err)
	}
	if a.ProducedBy() != execID {
		t.Errorf("ProducedBy() = %q, want %q", a.ProducedBy(), execID)
	}

	run, err := f.executions.Get(ctx, execID)
	if err != nil {
		t.Fatalf("Executions.Get: %v", err)
	}
	ids := run.ArtifactIDs()
	if len(ids) != 1 || ids[0] != "art-1" {
		t.Errorf("Execution.ArtifactIDs() = %v, want [art-1] (both sides of the link)", ids)
	}

	last := f.bus.Published()[len(f.bus.Published())-1]
	if last.Type() != event.ArtifactCreated || last.ProjectID() != "proj-1" || last.SubjectID() != "art-1" {
		t.Errorf("last event = %+v, want ArtifactCreated for art-1/proj-1", last)
	}
}

func TestRecordDraftArtifact_RejectedWhenExecutionNotRunning(t *testing.T) {
	ctx := context.Background()
	f := newResultFixture(t)
	execID := runningExecutionID(t, f)
	if err := f.result.SucceedExecution(ctx, execID, ""); err != nil {
		t.Fatalf("SucceedExecution: %v", err)
	}

	if _, err := f.result.RecordDraftArtifact(ctx, application.RecordDraftArtifactParams{
		ID: "art-1", ProjectID: "proj-1", ExecutionID: execID,
		Type: artifact.Type("PullRequest"), Origin: artifact.OriginProduced, Author: artifact.Author("developer"),
	}); !errors.Is(err, application.ErrExecutionNotRunning) {
		t.Errorf("RecordDraftArtifact() after Succeed error = %v, want %v", err, application.ErrExecutionNotRunning)
	}
}

func TestRecordDraftArtifact_RejectedWhenProjectNotActive(t *testing.T) {
	ctx := context.Background()
	f := newResultFixture(t)
	execID := runningExecutionID(t, f)

	otherProjects := inmemory.NewProjectStore()
	f.result.Projects = otherProjects // proj-2 was never saved/activated

	if _, err := f.result.RecordDraftArtifact(ctx, application.RecordDraftArtifactParams{
		ID: "art-1", ProjectID: "proj-2", ExecutionID: execID,
		Type: artifact.Type("PullRequest"), Origin: artifact.OriginProduced, Author: artifact.Author("developer"),
	}); !errors.Is(err, application.ErrNotFound) {
		t.Errorf("RecordDraftArtifact() with unknown project error = %v, want %v", err, application.ErrNotFound)
	}
}

func TestPublishArtifact_RequiresPayload(t *testing.T) {
	ctx := context.Background()
	f := newResultFixture(t)
	execID := runningExecutionID(t, f)
	if _, err := f.result.RecordDraftArtifact(ctx, application.RecordDraftArtifactParams{
		ID: "art-1", ProjectID: "proj-1", ExecutionID: execID,
		Type: artifact.Type("PullRequest"), Origin: artifact.OriginProduced, Author: artifact.Author("developer"),
	}); err != nil {
		t.Fatalf("RecordDraftArtifact: %v", err)
	}

	if err := f.result.PublishArtifact(ctx, "art-1", ""); err == nil {
		t.Error("PublishArtifact() without payload error = nil, want the domain's ErrPayloadRequired")
	}
}

func TestUpdateDraftThenPublishArtifact_Success(t *testing.T) {
	ctx := context.Background()
	f := newResultFixture(t)
	execID := runningExecutionID(t, f)
	if _, err := f.result.RecordDraftArtifact(ctx, application.RecordDraftArtifactParams{
		ID: "art-1", ProjectID: "proj-1", ExecutionID: execID,
		Type: artifact.Type("PullRequest"), Origin: artifact.OriginProduced, Author: artifact.Author("developer"),
	}); err != nil {
		t.Fatalf("RecordDraftArtifact: %v", err)
	}
	if err := f.result.UpdateArtifactDraft(ctx, "art-1", []byte("diff --git ..."), ""); err != nil {
		t.Fatalf("UpdateArtifactDraft: %v", err)
	}
	if err := f.result.PublishArtifact(ctx, "art-1", "developer:executor-1"); err != nil {
		t.Fatalf("PublishArtifact: %v", err)
	}

	a, err := f.artifacts.Get(ctx, "art-1")
	if err != nil {
		t.Fatalf("Artifacts.Get: %v", err)
	}
	if a.State() != artifact.StatePublished {
		t.Errorf("State() = %v, want %v", a.State(), artifact.StatePublished)
	}

	last := f.bus.Published()[len(f.bus.Published())-1]
	if last.Type() != event.ArtifactPublished {
		t.Errorf("last event Type() = %q, want %q", last.Type(), event.ArtifactPublished)
	}
}

func TestSucceedExecution_Success(t *testing.T) {
	ctx := context.Background()
	f := newResultFixture(t)
	execID := runningExecutionID(t, f)
	if _, err := f.result.RecordDraftArtifact(ctx, application.RecordDraftArtifactParams{
		ID: "art-1", ProjectID: "proj-1", ExecutionID: execID,
		Type: artifact.Type("PullRequest"), Origin: artifact.OriginProduced, Author: artifact.Author("developer"),
		Payload: []byte("content"),
	}); err != nil {
		t.Fatalf("RecordDraftArtifact: %v", err)
	}

	if err := f.result.SucceedExecution(ctx, execID, "developer:executor-1"); err != nil {
		t.Fatalf("SucceedExecution: %v", err)
	}
	run, err := f.executions.Get(ctx, execID)
	if err != nil {
		t.Fatalf("Executions.Get: %v", err)
	}
	if run.State() != execution.StateSucceeded {
		t.Errorf("State() = %v, want %v", run.State(), execution.StateSucceeded)
	}

	last := f.bus.Published()[len(f.bus.Published())-1]
	if last.Type() != event.ExecutionSucceeded || last.ProjectID() != "proj-1" {
		t.Errorf("last event = %+v, want ExecutionSucceeded with ProjectID proj-1 (looked up via Task)", last)
	}
}

func TestFailExecution_KeepsProducedArtifacts(t *testing.T) {
	ctx := context.Background()
	f := newResultFixture(t)
	execID := runningExecutionID(t, f)
	if _, err := f.result.RecordDraftArtifact(ctx, application.RecordDraftArtifactParams{
		ID: "art-1", ProjectID: "proj-1", ExecutionID: execID,
		Type: artifact.Type("TestReport"), Origin: artifact.OriginProduced, Author: artifact.AuthorUnknown,
		Payload: []byte("failure log"),
	}); err != nil {
		t.Fatalf("RecordDraftArtifact: %v", err)
	}

	if err := f.result.FailExecution(ctx, execID, ""); err != nil {
		t.Fatalf("FailExecution: %v", err)
	}
	run, err := f.executions.Get(ctx, execID)
	if err != nil {
		t.Fatalf("Executions.Get: %v", err)
	}
	if run.State() != execution.StateFailed {
		t.Errorf("State() = %v, want %v", run.State(), execution.StateFailed)
	}
	if ids := run.ArtifactIDs(); len(ids) != 1 {
		t.Errorf("ArtifactIDs() = %v, want the TestReport kept despite failure", ids)
	}
}

func TestSucceedExecution_RejectedAfterFail_RaceAlreadyResolvedByDomain(t *testing.T) {
	ctx := context.Background()
	f := newResultFixture(t)
	execID := runningExecutionID(t, f)
	if err := f.result.FailExecution(ctx, execID, ""); err != nil {
		t.Fatalf("FailExecution: %v", err)
	}
	if err := f.result.SucceedExecution(ctx, execID, ""); err == nil {
		t.Error("SucceedExecution() after Fail error = nil, want the domain's ErrTerminal (Behavioral Invariant 5)")
	}
}

func TestRecordDraftArtifact_ExecutionNotFound(t *testing.T) {
	f := newResultFixture(t)
	if _, err := f.result.RecordDraftArtifact(context.Background(), application.RecordDraftArtifactParams{
		ID: "art-1", ProjectID: "proj-1", ExecutionID: "missing",
		Type: artifact.Type("PullRequest"), Origin: artifact.OriginProduced, Author: artifact.Author("developer"),
	}); !errors.Is(err, application.ErrNotFound) {
		t.Errorf("RecordDraftArtifact() error = %v, want %v", err, application.ErrNotFound)
	}
}

func TestUpdateArtifactDraft_NotFound(t *testing.T) {
	f := newResultFixture(t)
	if err := f.result.UpdateArtifactDraft(context.Background(), "missing", []byte("x"), ""); !errors.Is(err, application.ErrNotFound) {
		t.Errorf("UpdateArtifactDraft() error = %v, want %v", err, application.ErrNotFound)
	}
}

func TestUpdateArtifactDraft_RejectedAfterPublish(t *testing.T) {
	ctx := context.Background()
	f := newResultFixture(t)
	execID := runningExecutionID(t, f)
	if _, err := f.result.RecordDraftArtifact(ctx, application.RecordDraftArtifactParams{
		ID: "art-1", ProjectID: "proj-1", ExecutionID: execID,
		Type: artifact.Type("PullRequest"), Origin: artifact.OriginProduced, Author: artifact.Author("developer"),
		Payload: []byte("content"),
	}); err != nil {
		t.Fatalf("RecordDraftArtifact: %v", err)
	}
	if err := f.result.PublishArtifact(ctx, "art-1", ""); err != nil {
		t.Fatalf("PublishArtifact: %v", err)
	}
	if err := f.result.UpdateArtifactDraft(ctx, "art-1", []byte("changed"), ""); err == nil {
		t.Error("UpdateArtifactDraft() after Publish error = nil, want the domain's ErrPublished")
	}
}

func TestPublishArtifact_NotFound(t *testing.T) {
	f := newResultFixture(t)
	if err := f.result.PublishArtifact(context.Background(), "missing", ""); !errors.Is(err, application.ErrNotFound) {
		t.Errorf("PublishArtifact() error = %v, want %v", err, application.ErrNotFound)
	}
}

func TestFailExecution_NotFound(t *testing.T) {
	f := newResultFixture(t)
	if err := f.result.FailExecution(context.Background(), "missing", ""); !errors.Is(err, application.ErrNotFound) {
		t.Errorf("FailExecution() error = %v, want %v", err, application.ErrNotFound)
	}
}

func TestSucceedExecution_NotFound(t *testing.T) {
	f := newResultFixture(t)
	if err := f.result.SucceedExecution(context.Background(), "missing", ""); !errors.Is(err, application.ErrNotFound) {
		t.Errorf("SucceedExecution() error = %v, want %v", err, application.ErrNotFound)
	}
}
