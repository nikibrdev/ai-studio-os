package application_test

import (
	"context"
	"errors"
	"testing"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/application/inmemory"
	"ai-studio-os/internal/domain/event"
	"ai-studio-os/internal/domain/project"
	"ai-studio-os/internal/domain/shared"
	"ai-studio-os/internal/domain/workflow"
)

func newActiveProject(t *testing.T, store application.ProjectStore) *project.Project {
	t.Helper()
	ctx := context.Background()
	p, _, err := project.New("proj-1", "AI Studio OS")
	if err != nil {
		t.Fatalf("project.New: %v", err)
	}
	if _, _, err := p.ConnectRepository("github.com/org/repo"); err != nil {
		t.Fatalf("ConnectRepository: %v", err)
	}
	if _, err := p.Activate(); err != nil {
		t.Fatalf("Activate: %v", err)
	}
	if err := store.Save(ctx, p); err != nil {
		t.Fatalf("Save: %v", err)
	}
	return p
}

func newService() (*application.TaskPlanningService, application.ProjectStore, *inmemory.EventBus) {
	projects := inmemory.NewProjectStore()
	tasks := inmemory.NewTaskStore()
	bus := inmemory.NewEventBus()
	svc := &application.TaskPlanningService{
		Projects: projects,
		Tasks:    tasks,
		Events:   bus,
		Rules:    workflow.Machine{},
	}
	return svc, projects, bus
}

func TestCreateTask_Success(t *testing.T) {
	ctx := context.Background()
	svc, projects, bus := newService()
	newActiveProject(t, projects)

	tsk, err := svc.CreateTask(ctx, application.CreateTaskParams{
		ID:                 "task-1",
		ProjectID:          "proj-1",
		Title:              "Реализовать use-case",
		Type:               "feature",
		Scope:              "Постановка задачи",
		AcceptanceCriteria: []string{"задача создана в Active-проекте"},
		Actor:              "developer:executor-1",
	})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	if tsk.State() != shared.StateBacklog {
		t.Errorf("State() = %v, want %v", tsk.State(), shared.StateBacklog)
	}
	if tsk.Scope() != "Постановка задачи" {
		t.Errorf("Scope() = %q", tsk.Scope())
	}

	published := bus.Published()
	if len(published) != 1 {
		t.Fatalf("published = %d events, want 1", len(published))
	}
	e := published[0]
	if e.Type() != event.TaskCreated {
		t.Errorf("Type() = %q, want %q", e.Type(), event.TaskCreated)
	}
	if e.Source() != "task" || e.ProjectID() != "proj-1" || e.SubjectID() != "task-1" {
		t.Errorf("envelope fields = source=%q project=%q subject=%q", e.Source(), e.ProjectID(), e.SubjectID())
	}
	if e.Actor() != "developer:executor-1" {
		t.Errorf("Actor() = %q", e.Actor())
	}
}

func TestCreateTask_RejectedWhenProjectNotActive(t *testing.T) {
	ctx := context.Background()
	svc, projects, bus := newService()
	p, _, err := project.New("proj-1", "AI Studio OS")
	if err != nil {
		t.Fatalf("project.New: %v", err)
	}
	if err := projects.Save(ctx, p); err != nil { // still Created, never activated
		t.Fatalf("Save: %v", err)
	}

	if _, err := svc.CreateTask(ctx, application.CreateTaskParams{
		ID: "task-1", ProjectID: "proj-1", Title: "Задача", Type: "feature",
	}); !errors.Is(err, application.ErrProjectNotActive) {
		t.Errorf("CreateTask() error = %v, want %v", err, application.ErrProjectNotActive)
	}
	if len(bus.Published()) != 0 {
		t.Errorf("published = %v, want no events on rejection", bus.Published())
	}
}

func TestCreateTask_PropagatesDomainValidationError(t *testing.T) {
	ctx := context.Background()
	svc, projects, _ := newService()
	newActiveProject(t, projects)

	if _, err := svc.CreateTask(ctx, application.CreateTaskParams{
		ID: "task-1", ProjectID: "proj-1", Title: "", Type: "feature",
	}); err == nil {
		t.Error("CreateTask() with empty Title error = nil, want the domain's own validation error")
	}
}

func TestCreateTask_ProjectNotFound(t *testing.T) {
	svc, _, _ := newService()
	if _, err := svc.CreateTask(context.Background(), application.CreateTaskParams{
		ID: "task-1", ProjectID: "missing", Title: "Задача", Type: "feature",
	}); !errors.Is(err, application.ErrNotFound) {
		t.Errorf("CreateTask() error = %v, want %v", err, application.ErrNotFound)
	}
}

func TestPlanTask_Success(t *testing.T) {
	ctx := context.Background()
	svc, projects, bus := newService()
	newActiveProject(t, projects)
	if _, err := svc.CreateTask(ctx, application.CreateTaskParams{
		ID: "task-1", ProjectID: "proj-1", Title: "Задача", Type: "feature",
	}); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	if err := svc.PlanTask(ctx, "task-1", "pm:executor-2"); err != nil {
		t.Fatalf("PlanTask: %v", err)
	}

	got, err := svc.Tasks.Get(ctx, "task-1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.State() != shared.StateReady {
		t.Errorf("State() = %v, want %v", got.State(), shared.StateReady)
	}

	published := bus.Published()
	if len(published) != 2 {
		t.Fatalf("published = %d events, want 2 (Created, Planned)", len(published))
	}
	if published[1].Type() != event.TaskPlanned {
		t.Errorf("second event Type() = %q, want %q", published[1].Type(), event.TaskPlanned)
	}
}

func TestPlanTask_RulesRejectionKeepsState(t *testing.T) {
	ctx := context.Background()
	svc, projects, bus := newService()
	newActiveProject(t, projects)
	if _, err := svc.CreateTask(ctx, application.CreateTaskParams{
		ID: "task-1", ProjectID: "proj-1", Title: "Задача", Type: "feature",
	}); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	// Plan once (Backlog -> Ready), then attempting to plan again is an
	// illegal transition under the real Machine (Ready -> Ready).
	if err := svc.PlanTask(ctx, "task-1", ""); err != nil {
		t.Fatalf("first PlanTask: %v", err)
	}
	before := len(bus.Published())

	if err := svc.PlanTask(ctx, "task-1", ""); err == nil {
		t.Fatal("second PlanTask() error = nil, want the workflow.Machine's rejection")
	}
	if len(bus.Published()) != before {
		t.Errorf("published grew after a rejected transition: %d -> %d", before, len(bus.Published()))
	}
}

func TestPlanTask_TaskNotFound(t *testing.T) {
	svc, _, _ := newService()
	if err := svc.PlanTask(context.Background(), "missing", ""); !errors.Is(err, application.ErrNotFound) {
		t.Errorf("PlanTask() error = %v, want %v", err, application.ErrNotFound)
	}
}

// fakeTaskIDGenerator is a deterministic application.TaskIDGenerator fake
// for TASK-065's tests (EPIC-008): CreateTask must generate an ID via
// this port only when CreateTaskParams.ID is left empty.
type fakeTaskIDGenerator struct {
	id        string
	err       error
	calls     int
	projectID string
}

func (f *fakeTaskIDGenerator) NextID(_ context.Context, projectID string) (string, error) {
	f.calls++
	f.projectID = projectID
	if f.err != nil {
		return "", f.err
	}
	return f.id, nil
}

func TestCreateTask_GeneratesIDWhenEmpty(t *testing.T) {
	ctx := context.Background()
	svc, projects, _ := newService()
	newActiveProject(t, projects)
	ids := &fakeTaskIDGenerator{id: "TASK-042"}
	svc.IDs = ids

	tsk, err := svc.CreateTask(ctx, application.CreateTaskParams{
		ProjectID: "proj-1", Title: "Задача", Type: "feature",
	})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	if tsk.ID() != "TASK-042" {
		t.Errorf("ID() = %q, want TASK-042", tsk.ID())
	}
	if ids.calls != 1 || ids.projectID != "proj-1" {
		t.Errorf("IDs.NextID called %d times with project %q, want 1 call with proj-1", ids.calls, ids.projectID)
	}
}

func TestCreateTask_ExplicitIDIgnoresGenerator(t *testing.T) {
	ctx := context.Background()
	svc, projects, _ := newService()
	newActiveProject(t, projects)
	ids := &fakeTaskIDGenerator{id: "TASK-042"}
	svc.IDs = ids

	tsk, err := svc.CreateTask(ctx, application.CreateTaskParams{
		ID: "task-explicit", ProjectID: "proj-1", Title: "Задача", Type: "feature",
	})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	if tsk.ID() != "task-explicit" {
		t.Errorf("ID() = %q, want task-explicit (caller-supplied ID must win)", tsk.ID())
	}
	if ids.calls != 0 {
		t.Errorf("IDs.NextID called %d times, want 0 (explicit ID must skip the generator)", ids.calls)
	}
}

func TestCreateTask_IDGeneratorErrorPropagates(t *testing.T) {
	ctx := context.Background()
	svc, projects, bus := newService()
	newActiveProject(t, projects)
	wantErr := errors.New("sequence unavailable")
	svc.IDs = &fakeTaskIDGenerator{err: wantErr}

	_, err := svc.CreateTask(ctx, application.CreateTaskParams{
		ProjectID: "proj-1", Title: "Задача", Type: "feature",
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("CreateTask() error = %v, want %v", err, wantErr)
	}
	if len(bus.Published()) != 0 {
		t.Errorf("published = %v, want no events when ID generation fails", bus.Published())
	}
}

func TestCreateTask_EmptyIDWithoutGeneratorPropagatesDomainError(t *testing.T) {
	ctx := context.Background()
	svc, projects, _ := newService()
	newActiveProject(t, projects)
	// svc.IDs left nil: pre-EPIC-008 behavior — the caller must supply
	// CreateTaskParams.ID, and task.New's own validation fires otherwise.
	if _, err := svc.CreateTask(ctx, application.CreateTaskParams{
		ProjectID: "proj-1", Title: "Задача", Type: "feature",
	}); err == nil {
		t.Error("CreateTask() with empty ID and nil IDs error = nil, want the domain's own validation error")
	}
}
