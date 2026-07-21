package application

import (
	"context"
	"errors"
	"time"

	"ai-studio-os/internal/domain/event"
	"ai-studio-os/internal/domain/shared"
	"ai-studio-os/internal/domain/task"
	"ai-studio-os/internal/domain/workflow"
	"ai-studio-os/internal/platform"
)

// ErrProjectNotActive is returned when a use-case tries to create content
// in a Project that is not Active (spec Project Behavioral Invariant 4).
var ErrProjectNotActive = errors.New("application: project does not accept new content")

// TaskPlanningService implements the "Постановка задачи" step of the
// golden path (docs/architecture/golden-path.md): creating a Task inside
// an active Project's boundary and bringing it from Backlog to Ready.
type TaskPlanningService struct {
	Projects ProjectStore
	Tasks    TaskStore
	Events   platform.EventBus
	Rules    workflow.Rules
}

// CreateTaskParams are the inputs to CreateTask. EpicID, Scope and
// AcceptanceCriteria are optional (spec Task Structural Invariants 2, 4).
type CreateTaskParams struct {
	ID                 string
	ProjectID          string
	EpicID             string
	Title              string
	Type               string
	Scope              string
	AcceptanceCriteria []string
	Actor              string
}

// CreateTask registers a Task inside the given Project (spec Project
// Behavioral Invariant 4: only an Active project accepts new content) and
// records its scope and acceptance criteria. Publishes TaskCreated.
func (s *TaskPlanningService) CreateTask(ctx context.Context, p CreateTaskParams) (*task.Task, error) {
	proj, err := s.Projects.Get(ctx, p.ProjectID)
	if err != nil {
		return nil, err
	}
	if !proj.AcceptsNewContent() {
		return nil, ErrProjectNotActive
	}

	t, created, err := task.New(p.ID, p.ProjectID, p.EpicID, p.Title, p.Type)
	if err != nil {
		return nil, err
	}
	if p.Scope != "" {
		if err := t.SetScope(p.Scope); err != nil {
			return nil, err
		}
	}
	if len(p.AcceptanceCriteria) > 0 {
		if err := t.SetAcceptanceCriteria(p.AcceptanceCriteria); err != nil {
			return nil, err
		}
	}

	if err := s.Tasks.Save(ctx, t); err != nil {
		return nil, err
	}
	if err := s.publish(ctx, event.TaskCreated, p.Actor, p.ProjectID, t.ID(), created.At); err != nil {
		return nil, err
	}
	return t, nil
}

// PlanTask transitions a Task Backlog -> Ready (Definition of Ready met),
// validated exclusively by the configured workflow.Rules (state-machine.md
// invariant 8: the task module never decides legality itself). Publishes
// TaskPlanned.
func (s *TaskPlanningService) PlanTask(ctx context.Context, taskID, actor string) error {
	t, err := s.Tasks.Get(ctx, taskID)
	if err != nil {
		return err
	}
	transitioned, err := t.Transition(shared.StateReady, "", s.Rules)
	if err != nil {
		return err
	}
	if err := s.Tasks.Save(ctx, t); err != nil {
		return err
	}
	return s.publish(ctx, event.TaskPlanned, actor, t.ProjectID(), t.ID(), transitioned.At)
}

func (s *TaskPlanningService) publish(ctx context.Context, eventType, actor, projectID, subjectID string, at time.Time) error {
	return s.Events.Publish(ctx, NewEvent(eventType, "task", actor, projectID, subjectID, at))
}
