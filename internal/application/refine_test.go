package application_test

import (
	"context"
	"errors"
	"testing"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/application/inmemory"
	"ai-studio-os/internal/domain/event"
	"ai-studio-os/internal/domain/shared"
	"ai-studio-os/internal/domain/task"
	"ai-studio-os/internal/domain/workflow"
)

// refineFixture is a task in Backlog with a projection watching the bus —
// exactly the state a Project Manager works in (EPIC-013).
type refineFixture struct {
	planning *application.TaskPlanningService
	tasks    application.TaskStore
	views    *application.TaskProjection
	bus      *inmemory.EventBus
}

func newRefineFixture(t *testing.T) refineFixture {
	t.Helper()
	ctx := context.Background()
	projects := inmemory.NewProjectStore()
	tasks := inmemory.NewTaskStore()
	bus := inmemory.NewEventBus()
	newActiveProject(t, projects)

	views := application.NewTaskProjection()
	if err := views.Subscribe(bus); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	planning := &application.TaskPlanningService{
		Projects: projects, Tasks: tasks, Events: bus, Rules: workflow.Machine{},
	}
	if _, err := planning.CreateTask(ctx, application.CreateTaskParams{
		ID: "TASK-001", ProjectID: "proj-1", Title: "Задача", Type: "feature",
		Scope: "исходный scope", AcceptanceCriteria: []string{"исходный критерий"},
	}); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	return refineFixture{planning: planning, tasks: tasks, views: views, bus: bus}
}

func (f refineFixture) view(t *testing.T) application.TaskView {
	t.Helper()
	v, ok := f.views.Get("proj-1", "TASK-001")
	if !ok {
		t.Fatal("Get() ok = false, want true")
	}
	return v
}

func TestRefineTask_UpdatesScopeAndCriteria(t *testing.T) {
	ctx := context.Background()
	f := newRefineFixture(t)

	err := f.planning.RefineTask(ctx, application.RefineTaskParams{
		ProjectID: "proj-1", TaskID: "TASK-001",
		Scope: "уточнённый scope", AcceptanceCriteria: []string{"первый", "второй"},
	})
	if err != nil {
		t.Fatalf("RefineTask: %v", err)
	}

	v := f.view(t)
	if v.Scope != "уточнённый scope" {
		t.Errorf("Scope = %q, want the refined value", v.Scope)
	}
	if len(v.AcceptanceCriteria) != 2 || v.AcceptanceCriteria[0] != "первый" {
		t.Errorf("AcceptanceCriteria = %v, want the refined list", v.AcceptanceCriteria)
	}
	// Refining is not a transition: accepting Definition of Ready is a human
	// checkpoint (docs/architecture/workflow.md).
	if v.State != shared.StateBacklog {
		t.Errorf("State = %v, want the task left in %v", v.State, shared.StateBacklog)
	}
}

// A refinement of one field must not wipe the other. The event carries only
// what changed, and the projection must read it that way.
func TestRefineTask_PartialRefinementKeepsOtherField(t *testing.T) {
	ctx := context.Background()

	t.Run("scope only", func(t *testing.T) {
		f := newRefineFixture(t)
		if err := f.planning.RefineTask(ctx, application.RefineTaskParams{
			ProjectID: "proj-1", TaskID: "TASK-001", Scope: "только scope",
		}); err != nil {
			t.Fatalf("RefineTask: %v", err)
		}

		v := f.view(t)
		if v.Scope != "только scope" {
			t.Errorf("Scope = %q, want the refined value", v.Scope)
		}
		if len(v.AcceptanceCriteria) != 1 || v.AcceptanceCriteria[0] != "исходный критерий" {
			t.Errorf("AcceptanceCriteria = %v, want the original preserved", v.AcceptanceCriteria)
		}
	})

	t.Run("criteria only", func(t *testing.T) {
		f := newRefineFixture(t)
		if err := f.planning.RefineTask(ctx, application.RefineTaskParams{
			ProjectID: "proj-1", TaskID: "TASK-001", AcceptanceCriteria: []string{"только критерий"},
		}); err != nil {
			t.Fatalf("RefineTask: %v", err)
		}

		v := f.view(t)
		if v.Scope != "исходный scope" {
			t.Errorf("Scope = %q, want the original preserved", v.Scope)
		}
		if len(v.AcceptanceCriteria) != 1 || v.AcceptanceCriteria[0] != "только критерий" {
			t.Errorf("AcceptanceCriteria = %v, want the refined list", v.AcceptanceCriteria)
		}
	})
}

// The refinement must survive replay from the journal, not only the live bus —
// BUGFIX-004's lesson, and the reason apps/api can restore its read model
// (BUGFIX-010).
func TestRefineTask_SurvivesProjectionRebuild(t *testing.T) {
	ctx := context.Background()
	f := newRefineFixture(t)

	if err := f.planning.RefineTask(ctx, application.RefineTaskParams{
		ProjectID: "proj-1", TaskID: "TASK-001",
		Scope: "уточнённый scope", AcceptanceCriteria: []string{"после пересборки"},
	}); err != nil {
		t.Fatalf("RefineTask: %v", err)
	}

	rebuilt := application.NewTaskProjection()
	if err := rebuilt.Rebuild(ctx, f.bus.Published()); err != nil {
		t.Fatalf("Rebuild: %v", err)
	}

	v, ok := rebuilt.Get("proj-1", "TASK-001")
	if !ok {
		t.Fatal("rebuilt.Get() ok = false, want true")
	}
	if v.Scope != "уточнённый scope" {
		t.Errorf("Scope after rebuild = %q, want the refined value", v.Scope)
	}
	if len(v.AcceptanceCriteria) != 1 || v.AcceptanceCriteria[0] != "после пересборки" {
		t.Errorf("AcceptanceCriteria after rebuild = %v, want the refined list", v.AcceptanceCriteria)
	}
}

// Past Backlog the domain refuses: the task has been accepted, and its scope is
// no longer a proposal. The service must propagate that, not re-decide it.
func TestRefineTask_RejectedOnceTaskLeftBacklog(t *testing.T) {
	ctx := context.Background()
	f := newRefineFixture(t)
	if err := f.planning.PlanTask(ctx, "proj-1", "TASK-001", ""); err != nil {
		t.Fatalf("PlanTask: %v", err)
	}

	err := f.planning.RefineTask(ctx, application.RefineTaskParams{
		ProjectID: "proj-1", TaskID: "TASK-001", Scope: "поздно",
	})
	if !errors.Is(err, task.ErrNotBacklog) {
		t.Fatalf("RefineTask() error = %v, want %v", err, task.ErrNotBacklog)
	}

	if v := f.view(t); v.Scope != "исходный scope" {
		t.Errorf("Scope = %q, want it unchanged after a rejected refinement", v.Scope)
	}
}

// An empty refinement is a caller mistake, not a no-op: publishing an event
// that changes nothing is journal noise and a false signal that a Project
// Manager did something.
func TestRefineTask_EmptyRefinementIsAnError(t *testing.T) {
	ctx := context.Background()
	f := newRefineFixture(t)
	before := len(f.bus.Published())

	err := f.planning.RefineTask(ctx, application.RefineTaskParams{
		ProjectID: "proj-1", TaskID: "TASK-001",
	})
	if !errors.Is(err, application.ErrNothingToRefine) {
		t.Fatalf("RefineTask() error = %v, want %v", err, application.ErrNothingToRefine)
	}
	if got := len(f.bus.Published()) - before; got != 0 {
		t.Errorf("published %d events, want none", got)
	}
}

func TestRefineTask_UnknownTaskNotFound(t *testing.T) {
	ctx := context.Background()
	f := newRefineFixture(t)

	err := f.planning.RefineTask(ctx, application.RefineTaskParams{
		ProjectID: "proj-1", TaskID: "missing", Scope: "x",
	})
	if !errors.Is(err, application.ErrNotFound) {
		t.Fatalf("RefineTask() error = %v, want %v", err, application.ErrNotFound)
	}
}

// TestRecordTestReport_AttachesReportToTask is TASK-100: a report recorded
// without a task reference is unreachable from the task whose acceptance
// decision it informs — Artifact carries no task field of its own.
func TestRecordTestReport_AttachesReportToTask(t *testing.T) {
	ctx := context.Background()
	projects := inmemory.NewProjectStore()
	tasks := inmemory.NewTaskStore()
	artifacts := inmemory.NewArtifactStore()
	bus := inmemory.NewEventBus()
	newActiveProject(t, projects)

	views := application.NewTaskProjection()
	if err := views.Subscribe(bus); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	planning := &application.TaskPlanningService{
		Projects: projects, Tasks: tasks, Events: bus, Rules: workflow.Machine{},
	}
	if _, err := planning.CreateTask(ctx, application.CreateTaskParams{
		ID: "TASK-001", ProjectID: "proj-1", Title: "Задача", Type: "feature",
	}); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	results := &application.ResultService{
		Projects: projects, Tasks: tasks, Artifacts: artifacts, Events: bus,
	}
	stored, err := results.RecordTestReport(ctx, application.RecordTestReportParams{
		ID: "report-1", ProjectID: "proj-1", TaskID: "TASK-001",
		Report: []byte("Проверил: критерии сходятся."), Author: "claude-code",
	})
	if err != nil {
		t.Fatalf("RecordTestReport: %v", err)
	}

	view, ok := views.Get("proj-1", "TASK-001")
	if !ok {
		t.Fatal("Get() ok = false, want true")
	}
	if view.QAReportID != stored.ID() {
		t.Errorf("QAReportID = %q, want the recorded report %q", view.QAReportID, stored.ID())
	}
	// The report itself stays in the artifact — the view holds only the id.
	if got, err := results.Artifact(ctx, view.QAReportID); err != nil {
		t.Errorf("Artifact: %v", err)
	} else if string(got.Payload()) != "Проверил: критерии сходятся." {
		t.Errorf("payload = %q", got.Payload())
	}
}

// An artifact published for something other than a task — a commit produced by a
// developer run — must not be mistaken for a QA report.
func TestPublishedArtifact_OnlyTestReportsAttachToTasks(t *testing.T) {
	ctx := context.Background()
	views := application.NewTaskProjection()

	created := journalLikeEvent{
		typ: event.TaskCreated, projectID: "proj-1", subjectID: "TASK-001",
		data: map[string]string{"title": "Задача"},
	}
	if err := views.Handle(ctx, created); err != nil {
		t.Fatalf("Handle(TaskCreated): %v", err)
	}

	commit := journalLikeEvent{
		typ: event.ArtifactPublished, projectID: "proj-1", subjectID: "abc123",
		data: map[string]string{"taskId": "TASK-001", "artifactType": "Commit"},
	}
	if err := views.Handle(ctx, commit); err != nil {
		t.Fatalf("Handle(ArtifactPublished): %v", err)
	}

	if view, _ := views.Get("proj-1", "TASK-001"); view.QAReportID != "" {
		t.Errorf("QAReportID = %q, want empty — a commit is not a QA report", view.QAReportID)
	}
}

// The artifact is the event's subject, not the task. Without special handling
// the projection would key a view by the artifact's identifier and invent a task
// that does not exist.
func TestPublishedArtifact_DoesNotInventTasks(t *testing.T) {
	ctx := context.Background()
	views := application.NewTaskProjection()

	report := journalLikeEvent{
		typ: event.ArtifactPublished, projectID: "proj-1", subjectID: "report-1",
		data: map[string]string{"taskId": "TASK-404", "artifactType": "TestReport"},
	}
	if err := views.Handle(ctx, report); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	if _, ok := views.Get("proj-1", "report-1"); ok {
		t.Error("a view was created keyed by the artifact's id")
	}
	if _, ok := views.Get("proj-1", "TASK-404"); ok {
		t.Error("a view was invented for a task the projection has never seen")
	}
}
