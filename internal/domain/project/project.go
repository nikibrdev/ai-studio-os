package project

import (
	"errors"
	"time"
)

// State is the Project lifecycle state (spec Lifecycle:
// Created -> Active -> Archived).
type State string

// The three Lifecycle states (spec Lifecycle).
const (
	StateCreated  State = "created"
	StateActive   State = "active"
	StateArchived State = "archived"
)

// Sentinel errors returned by Project commands.
var (
	// ErrMissingField is returned when a required field is empty.
	ErrMissingField = errors.New("project: required field is missing")

	// ErrArchived is returned when a command reaches an archived Project:
	// Archived is terminal, none of the Project's own attributes change
	// afterwards (spec Behavioral Invariant 1).
	ErrArchived = errors.New("project: archived projects cannot change")

	// ErrAlreadyActive is returned by Activate when the Project is
	// already Active.
	ErrAlreadyActive = errors.New("project: already active")

	// ErrNoRepository is returned by Activate when no repository is
	// connected yet: Activate's guard requires at least one (spec
	// Structural Invariant 1, Lifecycle).
	ErrNoRepository = errors.New("project: at least one repository must be connected before activation")

	// ErrNotActive is returned by Archive when the Project is not Active
	// (spec Lifecycle: only Active -> Archived is allowed).
	ErrNotActive = errors.New("project: only an active project can be archived")
)

// Project is a software-development initiative managed by the platform:
// the boundary within which Epic, Task and Artifact exist
// (docs/specifications/domain/project.md). Archiving the Project freezes
// its own attributes but never rewrites the lifecycle states of the
// content inside (spec Behavioral Invariant 2) — content lives in its own
// modules and is referenced, not owned structurally, here.
type Project struct {
	id           string
	name         string
	repositories []string
	createdAt    time.Time
	state        State
}

// New registers a Project in the Created state (spec Commands: Create).
func New(id, name string) (*Project, Created, error) {
	if id == "" || name == "" {
		return nil, Created{}, ErrMissingField
	}

	now := time.Now()
	p := &Project{id: id, name: name, createdAt: now, state: StateCreated}
	return p, Created{ID: id, Name: name, At: now}, nil
}

// Restore reconstructs a Project from previously persisted state, without
// re-running business rules or producing an event. It exists for storage
// adapters (internal/infrastructure) loading an aggregate that was already
// validated by New/ConnectRepository/Activate/Archive at the time it was
// saved — callers outside a Store implementation should not use it.
func Restore(id, name string, repositories []string, createdAt time.Time, state State) *Project {
	return &Project{
		id:           id,
		name:         name,
		repositories: append([]string(nil), repositories...),
		createdAt:    createdAt,
		state:        state,
	}
}

// ConnectRepository attaches a repository reference; allowed in Created
// and Active, never in Archived (spec Commands). It never transitions
// state by itself — it only satisfies Activate's guard (spec: no hidden
// state change inside a command with a different name). Connecting an
// already-connected repository is a no-op: the resulting set is identical
// either way. A connected repository cannot be detached — a deliberate v1
// restriction (spec Behavioral Invariant 3).
func (p *Project) ConnectRepository(repo string) (RepositoryConnected, bool, error) {
	if repo == "" {
		return RepositoryConnected{}, false, ErrMissingField
	}
	if p.state == StateArchived {
		return RepositoryConnected{}, false, ErrArchived
	}
	for _, r := range p.repositories {
		if r == repo {
			return RepositoryConnected{}, false, nil
		}
	}
	p.repositories = append(p.repositories, repo)
	return RepositoryConnected{ProjectID: p.id, Repository: repo, At: time.Now()}, true, nil
}

// Activate transitions Created -> Active (spec Commands: Activate — the
// explicit command decided by the final architecture review). Guard: at
// least one repository is connected (spec Structural Invariant 1).
func (p *Project) Activate() (Activated, error) {
	if p.state == StateArchived {
		return Activated{}, ErrArchived
	}
	if p.state == StateActive {
		return Activated{}, ErrAlreadyActive
	}
	if len(p.repositories) == 0 {
		return Activated{}, ErrNoRepository
	}

	p.state = StateActive
	return Activated{ID: p.id, At: time.Now()}, nil
}

// Archive transitions Active -> Archived (spec Commands: Archive). After
// it, none of the Project's own attributes or links change (spec
// Behavioral Invariant 1); content inside keeps its independently reached
// states (Behavioral Invariant 2).
func (p *Project) Archive() (Archived, error) {
	if p.state == StateArchived {
		return Archived{}, ErrArchived
	}
	if p.state != StateActive {
		return Archived{}, ErrNotActive
	}

	p.state = StateArchived
	return Archived{ID: p.id, At: time.Now()}, nil
}

// AcceptsNewContent reports whether new Epic, Task or Artifact may be
// created within this Project: only while Active — not Created (not ready
// yet), not Archived (immutable) (spec Behavioral Invariant 4). The
// creation itself happens in the owning modules; the domain only answers
// the question.
func (p *Project) AcceptsNewContent() bool { return p.state == StateActive }

// ID returns the Project's identifier.
func (p *Project) ID() string { return p.id }

// Name returns the Project's name.
func (p *Project) Name() string { return p.name }

// Repositories returns a copy of the connected repository references —
// the set only grows (spec Behavioral Invariant 3).
func (p *Project) Repositories() []string {
	if len(p.repositories) == 0 {
		return nil
	}
	return append([]string(nil), p.repositories...)
}

// CreatedAt returns the moment this Project was registered.
func (p *Project) CreatedAt() time.Time { return p.createdAt }

// State returns the Project's current Lifecycle state (spec Structural
// Invariant 3).
func (p *Project) State() State { return p.state }
