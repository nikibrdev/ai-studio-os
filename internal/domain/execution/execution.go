package execution

import (
	"errors"
	"time"
)

// Sentinel errors returned by Execution commands.
var (
	// ErrMissingField is returned by New when a required identity field
	// (Identifier, TaskID or ExecutorID) is empty (spec Structural
	// Invariants 1-2: exactly one Task and exactly one Executor, fixed at
	// creation).
	ErrMissingField = errors.New("execution: required field is missing")

	// ErrNotQueued is returned by Accept when the Execution has already
	// left Queued (spec Behavioral Invariant 3: Running is entered only
	// through the Executor's confirmed Accept, and only once).
	ErrNotQueued = errors.New("execution: only a queued execution can be accepted")

	// ErrNotRunning is returned by RecordArtifact, Succeed and Fail when
	// the Execution is not in Running (spec Behavioral Invariant 4 for
	// RecordArtifact; spec Commands for Succeed/Fail).
	ErrNotRunning = errors.New("execution: command requires the running state")

	// ErrTerminal is returned when any state-changing command reaches an
	// Execution already in Succeeded, Failed or Aborted: the first
	// terminal transition wins, later ones are simply invalid (spec
	// Behavioral Invariants 1-2 and 5).
	ErrTerminal = errors.New("execution: terminal executions cannot change")
)

// Execution is a single, bounded run of one Executor performing work for
// one Task (docs/specifications/domain/execution.md). A retry after
// failure is a new Execution, never a reuse of this one (spec Behavioral
// Invariant 1).
type Execution struct {
	id          string
	taskID      string
	executorID  string
	createdAt   time.Time
	artifactIDs []string
	state       State
}

// New creates an Execution in the Queued state (spec Commands: Create).
// TaskID and ExecutorID are fixed for the Execution's lifetime (Structural
// Invariants 1-2). Task, Executor and Artifact are referenced by
// identifier only (ADR-015: domain modules never import each other).
func New(id, taskID, executorID string) (*Execution, Queued, error) {
	if id == "" || taskID == "" || executorID == "" {
		return nil, Queued{}, ErrMissingField
	}

	now := time.Now()
	e := &Execution{
		id:         id,
		taskID:     taskID,
		executorID: executorID,
		createdAt:  now,
		state:      StateQueued,
	}
	event := Queued{ID: id, TaskID: taskID, ExecutorID: executorID, At: now}
	return e, event, nil
}

// Accept transitions Queued -> Running: the Executor confirmed taking the
// work (spec Commands: Accept; ADR-005 capability Accept). Invalid once
// the Execution has left Queued (Behavioral Invariant 3).
func (e *Execution) Accept() (Started, error) {
	if e.state.Terminal() {
		return Started{}, ErrTerminal
	}
	if e.state != StateQueued {
		return Started{}, ErrNotQueued
	}

	e.state = StateRunning
	return Started{ID: e.id, At: time.Now()}, nil
}

// RecordArtifact adds a produced Artifact reference to the Execution's
// set. Valid only while Running: an Artifact counts as produced by this
// Execution only if it appeared during Running — not before, not after
// (spec Behavioral Invariant 4; ADR-005 capability Artifacts).
func (e *Execution) RecordArtifact(artifactID string) error {
	if artifactID == "" {
		return ErrMissingField
	}
	if e.state.Terminal() {
		return ErrTerminal
	}
	if e.state != StateRunning {
		return ErrNotRunning
	}
	e.artifactIDs = append(e.artifactIDs, artifactID)
	return nil
}

// Succeed transitions Running -> Succeeded: the Executor finished the work
// as intended (spec Commands: Succeed; ADR-005 capabilities Status/Finish).
// It finalizes the produced-Artifact set (Behavioral Invariant 1).
func (e *Execution) Succeed() (Succeeded, error) {
	if e.state.Terminal() {
		return Succeeded{}, ErrTerminal
	}
	if e.state != StateRunning {
		return Succeeded{}, ErrNotRunning
	}

	e.state = StateSucceeded
	return Succeeded{ID: e.id, ArtifactIDs: e.artifacts(), At: time.Now()}, nil
}

// Fail transitions Running -> Failed: the Executor itself reported it
// cannot finish the work (spec Commands: Fail). Artifacts produced before
// the failure stay recorded — a failure report is a result of work too.
// If Fail and Abort race, whichever command executes first wins; the
// other receives ErrTerminal (spec Behavioral Invariant 5).
func (e *Execution) Fail() (Failed, error) {
	if e.state.Terminal() {
		return Failed{}, ErrTerminal
	}
	if e.state != StateRunning {
		return Failed{}, ErrNotRunning
	}

	e.state = StateFailed
	return Failed{ID: e.id, ArtifactIDs: e.artifacts(), At: time.Now()}, nil
}

// Abort transitions Queued -> Aborted or Running -> Aborted: the execution
// is stopped by an external decision, not by the Executor itself (spec
// Commands: Abort). It never changes or removes already-produced Artifact
// references.
func (e *Execution) Abort() (Aborted, error) {
	if e.state.Terminal() {
		return Aborted{}, ErrTerminal
	}

	from := e.state
	e.state = StateAborted
	return Aborted{ID: e.id, From: from, At: time.Now()}, nil
}

// artifacts returns a copy of the produced-Artifact reference set, so the
// finalized set inside a terminal event cannot be mutated through the
// original slice (Behavioral Invariant 1).
func (e *Execution) artifacts() []string {
	if len(e.artifactIDs) == 0 {
		return nil
	}
	out := make([]string, len(e.artifactIDs))
	copy(out, e.artifactIDs)
	return out
}

// ID returns the Execution's identifier.
func (e *Execution) ID() string { return e.id }

// TaskID returns the identifier of the Task that owns this Execution
// (spec Relationships: Task создаёт и владеет Execution).
func (e *Execution) TaskID() string { return e.taskID }

// ExecutorID returns the identifier of the Executor used by this
// Execution, fixed at creation (spec Structural Invariant 1).
func (e *Execution) ExecutorID() string { return e.executorID }

// CreatedAt returns the moment this Execution was created.
func (e *Execution) CreatedAt() time.Time { return e.createdAt }

// ArtifactIDs returns a copy of the identifiers of Artifacts produced by
// this Execution so far (references, not ownership — ADR-016).
func (e *Execution) ArtifactIDs() []string { return e.artifacts() }

// State returns the Execution's current Lifecycle state.
func (e *Execution) State() State { return e.state }
