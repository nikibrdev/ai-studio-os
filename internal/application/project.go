package application

import (
	"context"
	"time"

	"ai-studio-os/internal/domain/event"
	"ai-studio-os/internal/domain/project"
	"ai-studio-os/internal/platform"
)

// ProjectService manages the Project lifecycle up to the point the
// golden path's existing four services (EPIC-004) take over: they all act
// inside an already Active Project, but nothing exposed a way to reach
// that state from Application Layer until this service (TASK-064,
// EPIC-008) — tests and the golden path previously created a Project
// directly through internal/domain/project, bypassing Application Layer.
type ProjectService struct {
	Projects ProjectStore
	Events   platform.EventBus
}

// CreateProjectParams are the inputs to CreateProject.
type CreateProjectParams struct {
	ID    string
	Name  string
	Actor string
}

// CreateProject registers a Project in the Created state. Publishes
// ProjectCreated.
func (s *ProjectService) CreateProject(ctx context.Context, p CreateProjectParams) (*project.Project, error) {
	proj, created, err := project.New(p.ID, p.Name)
	if err != nil {
		return nil, err
	}
	if err := s.Projects.Save(ctx, proj); err != nil {
		return nil, err
	}
	if err := s.publish(ctx, event.ProjectCreated, p.Actor, p.ID, p.ID, created.At); err != nil {
		return nil, err
	}
	return proj, nil
}

// ListProjects returns every Project, ordered by id (EPIC-009, TASK-072 —
// apps/dashboard has no other way to show a list of projects).
func (s *ProjectService) ListProjects(ctx context.Context) ([]*project.Project, error) {
	return s.Projects.List(ctx)
}

// ConnectRepository attaches a repository reference to the Project —
// required before Activate can succeed (spec Structural Invariant 1: at
// least one repository connected). Publishes RepositoryConnected, unless
// the repository was already connected: the domain treats that as a
// no-op (spec Behavioral Invariant 3), and no event is published for a
// fact that did not change.
func (s *ProjectService) ConnectRepository(ctx context.Context, projectID, repo, actor string) error {
	proj, err := s.Projects.Get(ctx, projectID)
	if err != nil {
		return err
	}
	connected, changed, err := proj.ConnectRepository(repo)
	if err != nil {
		return err
	}
	if !changed {
		return nil
	}
	if err := s.Projects.Save(ctx, proj); err != nil {
		return err
	}
	return s.publish(ctx, event.RepositoryConnected, actor, projectID, projectID, connected.At)
}

// Activate transitions a Project Created -> Active (guard: at least one
// repository connected, spec Structural Invariant 1, enforced entirely by
// the domain — this service does not duplicate the check). Publishes
// ProjectActivated.
func (s *ProjectService) Activate(ctx context.Context, projectID, actor string) error {
	proj, err := s.Projects.Get(ctx, projectID)
	if err != nil {
		return err
	}
	activated, err := proj.Activate()
	if err != nil {
		return err
	}
	if err := s.Projects.Save(ctx, proj); err != nil {
		return err
	}
	return s.publish(ctx, event.ProjectActivated, actor, projectID, projectID, activated.At)
}

func (s *ProjectService) publish(ctx context.Context, eventType, actor, projectID, subjectID string, at time.Time) error {
	return s.Events.Publish(ctx, NewEvent(eventType, "project", actor, projectID, subjectID, at))
}
