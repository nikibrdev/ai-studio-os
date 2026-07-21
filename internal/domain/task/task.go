package task

import (
	"errors"
	"time"

	"ai-studio-os/internal/domain/shared"
	"ai-studio-os/internal/domain/workflow"
)

// Sentinel errors returned by Task commands.
var (
	// ErrMissingField is returned by New when a required identity field
	// (Identifier, ProjectID, Title or Type) is empty (spec Structural
	// Invariants 1, 3-4). EpicID is NOT required: the Task<->Epic link is
	// optional (Structural Invariant 2).
	ErrMissingField = errors.New("task: required field is missing")

	// ErrNilRules is returned by Transition when no workflow rules are
	// supplied: the task module never decides transition legality itself
	// (state-machine.md invariant 8, ADR-014).
	ErrNilRules = errors.New("task: workflow rules are required for a transition")

	// ErrNotBacklog is returned by SetScope, SetAcceptanceCriteria and
	// AttachToEpic outside Backlog: Ready means Definition of Ready is
	// met; changed requirements go through Ready -> Backlog first
	// (state-machine.md: TaskReturnedToBacklog).
	ErrNotBacklog = errors.New("task: scope, criteria and epic can change only in backlog")

	// ErrReasonRequired is returned by Transition into Blocked or
	// Cancelled without a recorded reason (state-machine.md invariant 3).
	ErrReasonRequired = errors.New("task: blocked and cancelled require a reason")
)

// Task is a unit of work with the canonical formalized lifecycle
// (docs/specifications/domain/task.md; states and transitions —
// docs/architecture/state-machine.md). The entity holds state; whether a
// transition is allowed is decided exclusively by the workflow rules
// contract passed into Transition (ADR-014).
type Task struct {
	id                 string
	projectID          string
	epicID             string
	title              string
	taskType           string
	scope              string
	acceptanceCriteria []string
	createdAt          time.Time
	state              shared.TaskState
}

// New creates a Task in the Backlog state (spec Commands: Create;
// state-machine.md: TaskCreated). epicID is optional — "" means the task
// belongs to no Epic (spec Structural Invariant 2). Identifiers are
// strings until ADR-011.
func New(id, projectID, epicID, title, taskType string) (*Task, Created, error) {
	if id == "" || projectID == "" || title == "" || taskType == "" {
		return nil, Created{}, ErrMissingField
	}

	now := time.Now()
	t := &Task{
		id:        id,
		projectID: projectID,
		epicID:    epicID,
		title:     title,
		taskType:  taskType,
		createdAt: now,
		state:     shared.StateBacklog,
	}
	event := Created{ID: id, ProjectID: projectID, EpicID: epicID, Title: title, Type: taskType, At: now}
	return t, event, nil
}

// AttachToEpic sets or changes the owning Epic. Valid only in Backlog:
// once the task is planned, its organizational placement is part of what
// Definition of Ready approved.
func (t *Task) AttachToEpic(epicID string) error {
	if epicID == "" {
		return ErrMissingField
	}
	if t.state != shared.StateBacklog {
		return ErrNotBacklog
	}
	t.epicID = epicID
	return nil
}

// SetScope records the task's goal and scope (spec Commands: SetScope).
// Backlog-only — see ErrNotBacklog.
func (t *Task) SetScope(scope string) error {
	if scope == "" {
		return ErrMissingField
	}
	if t.state != shared.StateBacklog {
		return ErrNotBacklog
	}
	t.scope = scope
	return nil
}

// SetAcceptanceCriteria records the task's acceptance criteria (spec
// Commands: SetAcceptanceCriteria). Backlog-only — see ErrNotBacklog.
func (t *Task) SetAcceptanceCriteria(criteria []string) error {
	if len(criteria) == 0 {
		return ErrMissingField
	}
	if t.state != shared.StateBacklog {
		return ErrNotBacklog
	}
	t.acceptanceCriteria = append([]string(nil), criteria...)
	return nil
}

// Transition moves the task to the target state. Legality is decided by
// the supplied workflow rules (state-machine.md invariant 8: the workflow
// module decides, the task module changes state); the entity itself only
// enforces what it alone can know — that Blocked and Cancelled carry a
// recorded reason (invariant 3).
func (t *Task) Transition(to shared.TaskState, reason string, rules workflow.Rules) (Transitioned, error) {
	if rules == nil {
		return Transitioned{}, ErrNilRules
	}
	if (to == shared.StateBlocked || to == shared.StateCancelled) && reason == "" {
		return Transitioned{}, ErrReasonRequired
	}
	if err := rules.CanTransition(t.state, to); err != nil {
		return Transitioned{}, err
	}

	from := t.state
	t.state = to
	return Transitioned{ID: t.id, From: from, To: to, Reason: reason, At: time.Now()}, nil
}

// ID returns the task's identifier.
func (t *Task) ID() string { return t.id }

// ProjectID returns the identifier of the Project that owns this task
// (spec Structural Invariant 1: a Task outside a Project does not exist).
func (t *Task) ProjectID() string { return t.projectID }

// EpicID returns the identifier of the owning Epic, or "" if none (spec
// Structural Invariant 2: the link is optional).
func (t *Task) EpicID() string { return t.epicID }

// Title returns the task's title.
func (t *Task) Title() string { return t.title }

// Type returns the task's type (feature, bugfix, docs, ...).
func (t *Task) Type() string { return t.taskType }

// Scope returns the recorded goal and scope, or "" if not set yet.
func (t *Task) Scope() string { return t.scope }

// AcceptanceCriteria returns a copy of the recorded acceptance criteria.
func (t *Task) AcceptanceCriteria() []string {
	if len(t.acceptanceCriteria) == 0 {
		return nil
	}
	return append([]string(nil), t.acceptanceCriteria...)
}

// CreatedAt returns the moment this task was created.
func (t *Task) CreatedAt() time.Time { return t.createdAt }

// State returns the task's current canonical lifecycle state (spec
// Structural Invariant 3: always exactly one of the nine states).
func (t *Task) State() shared.TaskState { return t.state }
