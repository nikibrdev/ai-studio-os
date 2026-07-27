package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

	// IDs generates the public TASK-NNN identifier (ADR-011) when
	// CreateTaskParams.ID is left empty — added in EPIC-008 (TASK-065).
	// Optional: nil preserves the original EPIC-004 behavior of requiring
	// the caller to supply ID (task.New's own ErrMissingField fires if
	// both are absent).
	IDs TaskIDGenerator
}

// CreateTaskParams are the inputs to CreateTask. ID may be left empty to
// have TaskPlanningService generate the next TASK-NNN for ProjectID via
// IDs (TASK-065) — required when IDs is nil. EpicID, Scope and
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
// records its scope and acceptance criteria. Publishes TaskCreated with
// title/type/scope/acceptanceCriteria attached via Envelope.WithData
// (EPIC-009, TASK-076) — TaskProjection is the only read path for Task
// (ADR-014), so the task detail page needs these fields from the event,
// not a direct TaskStore read.
func (s *TaskPlanningService) CreateTask(ctx context.Context, p CreateTaskParams) (*task.Task, error) {
	proj, err := s.Projects.Get(ctx, p.ProjectID)
	if err != nil {
		return nil, err
	}
	if !proj.AcceptsNewContent() {
		return nil, ErrProjectNotActive
	}

	id := p.ID
	if id == "" && s.IDs != nil {
		id, err = s.IDs.NextID(ctx, p.ProjectID)
		if err != nil {
			return nil, err
		}
	}

	t, created, err := task.New(id, p.ProjectID, p.EpicID, p.Title, p.Type)
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

	acceptanceCriteriaJSON, err := json.Marshal(t.AcceptanceCriteria())
	if err != nil {
		return nil, fmt.Errorf("application: encode acceptance criteria: %w", err)
	}
	e := NewEvent(event.TaskCreated, "task", p.Actor, p.ProjectID, t.ID(), created.At).
		WithData(map[string]string{
			dataKeyTitle:              t.Title(),
			dataKeyType:               t.Type(),
			dataKeyScope:              t.Scope(),
			dataKeyAcceptanceCriteria: string(acceptanceCriteriaJSON),
		})
	if err := s.Events.Publish(ctx, e); err != nil {
		return nil, err
	}
	return t, nil
}

// PlanTask transitions a Task Backlog -> Ready (Definition of Ready met),
// validated exclusively by the configured workflow.Rules (state-machine.md
// invariant 8: the task module never decides legality itself). Publishes
// TaskPlanned. projectID is required because TASK-NNN is unique only
// within a Project (ADR-011, BUGFIX-003).
func (s *TaskPlanningService) PlanTask(ctx context.Context, projectID, taskID, actor string) error {
	t, err := s.Tasks.Get(ctx, projectID, taskID)
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

// ErrNothingToRefine is returned by RefineTask when neither scope nor
// acceptance criteria were supplied: publishing an event that changes nothing
// would be journal noise and a false signal that a Project Manager did
// something.
var ErrNothingToRefine = errors.New("application: refinement supplied neither scope nor acceptance criteria")

// RefineTaskParams are the inputs to RefineTask. Scope and AcceptanceCriteria
// are each optional — a refinement may touch one without the other — but at
// least one is required.
type RefineTaskParams struct {
	ProjectID          string
	TaskID             string
	Scope              string
	AcceptanceCriteria []string
	Actor              string
}

// RefineTask records a better scope and/or acceptance criteria on a task still
// in Backlog — what a Project Manager prepares before a human accepts
// Definition of Ready (EPIC-013).
//
// The Backlog-only rule is the domain's: SetScope/SetAcceptanceCriteria return
// ErrNotBacklog past that point, and this service propagates it rather than
// duplicating the check — the same division as ProjectService.Activate and its
// "at least one repository" guard.
//
// The task deliberately stays in Backlog: refining is not a transition, and
// accepting Definition of Ready is a human checkpoint
// (docs/architecture/workflow.md).
//
// Publishes TaskRefined carrying only the fields that actually changed, so a
// projection can tell "not touched" from "cleared". That distinction is safe
// because the domain forbids clearing at all (SetScope("") is
// ErrMissingField), so an absent key can only mean "unchanged".
func (s *TaskPlanningService) RefineTask(ctx context.Context, p RefineTaskParams) error {
	if p.Scope == "" && len(p.AcceptanceCriteria) == 0 {
		return ErrNothingToRefine
	}

	t, err := s.Tasks.Get(ctx, p.ProjectID, p.TaskID)
	if err != nil {
		return err
	}

	data := map[string]string{}
	if p.Scope != "" {
		if err := t.SetScope(p.Scope); err != nil {
			return err
		}
		data[dataKeyScope] = t.Scope()
	}
	if len(p.AcceptanceCriteria) > 0 {
		if err := t.SetAcceptanceCriteria(p.AcceptanceCriteria); err != nil {
			return err
		}
		encoded, err := json.Marshal(t.AcceptanceCriteria())
		if err != nil {
			return fmt.Errorf("application: encode acceptance criteria: %w", err)
		}
		data[dataKeyAcceptanceCriteria] = string(encoded)
	}

	if err := s.Tasks.Save(ctx, t); err != nil {
		return err
	}

	e := NewEvent(event.TaskRefined, "task", p.Actor, t.ProjectID(), t.ID(), time.Now()).WithData(data)
	return s.Events.Publish(ctx, e)
}

func (s *TaskPlanningService) publish(ctx context.Context, eventType, actor, projectID, subjectID string, at time.Time) error {
	return s.Events.Publish(ctx, NewEvent(eventType, "task", actor, projectID, subjectID, at))
}
