package application_test

import (
	"context"
	"errors"
	"testing"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/application/inmemory"
	"ai-studio-os/internal/domain/event"
	"ai-studio-os/internal/domain/project"
)

func newProjectService() (*application.ProjectService, application.ProjectStore, *inmemory.EventBus) {
	projects := inmemory.NewProjectStore()
	bus := inmemory.NewEventBus()
	svc := &application.ProjectService{Projects: projects, Events: bus}
	return svc, projects, bus
}

func TestCreateProject_Success(t *testing.T) {
	ctx := context.Background()
	svc, _, bus := newProjectService()

	proj, err := svc.CreateProject(ctx, application.CreateProjectParams{
		ID: "proj-1", Name: "AI Studio OS", Actor: "human:architect",
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if proj.State() != project.StateCreated {
		t.Errorf("State() = %v, want %v", proj.State(), project.StateCreated)
	}

	published := bus.Published()
	if len(published) != 1 || published[0].Type() != event.ProjectCreated {
		t.Fatalf("Published() = %v, want one %s event", published, event.ProjectCreated)
	}
	if published[0].SubjectID() != "proj-1" {
		t.Errorf("SubjectID() = %q, want proj-1", published[0].SubjectID())
	}
}

func TestCreateProject_MissingFieldPropagatesDomainError(t *testing.T) {
	ctx := context.Background()
	svc, _, bus := newProjectService()

	_, err := svc.CreateProject(ctx, application.CreateProjectParams{ID: "", Name: "AI Studio OS"})
	if !errors.Is(err, project.ErrMissingField) {
		t.Fatalf("CreateProject() error = %v, want %v", err, project.ErrMissingField)
	}
	if len(bus.Published()) != 0 {
		t.Errorf("Published() = %v, want none on failure", bus.Published())
	}
}

func TestConnectRepository_Success(t *testing.T) {
	ctx := context.Background()
	svc, _, bus := newProjectService()
	if _, err := svc.CreateProject(ctx, application.CreateProjectParams{ID: "proj-1", Name: "AI Studio OS"}); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	if err := svc.ConnectRepository(ctx, "proj-1", "github.com/org/repo", "human:architect"); err != nil {
		t.Fatalf("ConnectRepository: %v", err)
	}

	published := bus.Published()
	if len(published) != 2 || published[1].Type() != event.RepositoryConnected {
		t.Fatalf("Published() = %v, want ProjectCreated then %s", published, event.RepositoryConnected)
	}
}

func TestConnectRepository_AlreadyConnectedIsNoopWithoutEvent(t *testing.T) {
	ctx := context.Background()
	svc, _, bus := newProjectService()
	if _, err := svc.CreateProject(ctx, application.CreateProjectParams{ID: "proj-1", Name: "AI Studio OS"}); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if err := svc.ConnectRepository(ctx, "proj-1", "github.com/org/repo", "human:architect"); err != nil {
		t.Fatalf("first ConnectRepository: %v", err)
	}

	if err := svc.ConnectRepository(ctx, "proj-1", "github.com/org/repo", "human:architect"); err != nil {
		t.Fatalf("second ConnectRepository: %v", err)
	}

	if len(bus.Published()) != 2 {
		t.Fatalf("Published() = %v, want no additional event for a repeat connection", bus.Published())
	}
}

func TestConnectRepository_UnknownProjectPropagatesNotFound(t *testing.T) {
	ctx := context.Background()
	svc, _, _ := newProjectService()

	err := svc.ConnectRepository(ctx, "missing", "github.com/org/repo", "human:architect")
	if !errors.Is(err, application.ErrNotFound) {
		t.Fatalf("ConnectRepository() error = %v, want %v", err, application.ErrNotFound)
	}
}

func TestActivate_Success(t *testing.T) {
	ctx := context.Background()
	svc, projects, bus := newProjectService()
	if _, err := svc.CreateProject(ctx, application.CreateProjectParams{ID: "proj-1", Name: "AI Studio OS"}); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if err := svc.ConnectRepository(ctx, "proj-1", "github.com/org/repo", ""); err != nil {
		t.Fatalf("ConnectRepository: %v", err)
	}

	if err := svc.Activate(ctx, "proj-1", "human:architect"); err != nil {
		t.Fatalf("Activate: %v", err)
	}

	proj, err := projects.Get(ctx, "proj-1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if proj.State() != project.StateActive {
		t.Errorf("State() = %v, want %v", proj.State(), project.StateActive)
	}

	published := bus.Published()
	if len(published) != 3 || published[2].Type() != event.ProjectActivated {
		t.Fatalf("Published() = %v, want ProjectActivated last", published)
	}
}

func TestActivate_NoRepositoryPropagatesGuardError(t *testing.T) {
	ctx := context.Background()
	svc, _, bus := newProjectService()
	if _, err := svc.CreateProject(ctx, application.CreateProjectParams{ID: "proj-1", Name: "AI Studio OS"}); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	err := svc.Activate(ctx, "proj-1", "human:architect")
	if !errors.Is(err, project.ErrNoRepository) {
		t.Fatalf("Activate() error = %v, want %v", err, project.ErrNoRepository)
	}
	if len(bus.Published()) != 1 {
		t.Errorf("Published() = %v, want only ProjectCreated (Activate must not publish on guard failure)", bus.Published())
	}
}

func TestActivate_AlreadyActivePropagatesDomainError(t *testing.T) {
	ctx := context.Background()
	svc, _, _ := newProjectService()
	if _, err := svc.CreateProject(ctx, application.CreateProjectParams{ID: "proj-1", Name: "AI Studio OS"}); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if err := svc.ConnectRepository(ctx, "proj-1", "github.com/org/repo", ""); err != nil {
		t.Fatalf("ConnectRepository: %v", err)
	}
	if err := svc.Activate(ctx, "proj-1", ""); err != nil {
		t.Fatalf("first Activate: %v", err)
	}

	err := svc.Activate(ctx, "proj-1", "")
	if !errors.Is(err, project.ErrAlreadyActive) {
		t.Fatalf("Activate() error = %v, want %v", err, project.ErrAlreadyActive)
	}
}
