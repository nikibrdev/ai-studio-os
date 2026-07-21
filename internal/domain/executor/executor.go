package executor

import (
	"errors"
	"time"

	"ai-studio-os/internal/domain/shared"
)

// Sentinel errors returned by Executor commands.
var (
	// ErrMissingField is returned by New when the identifier or backend
	// identity is empty (spec Structural Invariant 1).
	ErrMissingField = errors.New("executor: required field is missing")

	// ErrNoRoles is returned by New when the initial role set is empty:
	// an Executor without at least one capability makes no sense as an
	// entity (spec Structural Invariant 2).
	ErrNoRoles = errors.New("executor: at least one role is required")

	// ErrRetired is returned when any command reaches a retired Executor:
	// Retired is terminal (spec Behavioral Invariant 1).
	ErrRetired = errors.New("executor: retired executors cannot change")

	// ErrAlreadyActive is returned by Activate when the Executor is
	// already Active.
	ErrAlreadyActive = errors.New("executor: already active")

	// ErrNotActive is returned by Disable when the Executor is not Active
	// (spec Lifecycle: only Active -> Disabled is allowed).
	ErrNotActive = errors.New("executor: only an active executor can be disabled")

	// ErrLastRole is returned by RevokeRole when the role is the last one
	// remaining: the role set can never become empty — full
	// decommissioning is Retire, not emptying the set (spec Behavioral
	// Invariant 3, Structural Invariant 2).
	ErrLastRole = errors.New("executor: the last remaining role cannot be revoked")

	// ErrRoleNotGranted is returned by RevokeRole when the Executor does
	// not have the role.
	ErrRoleNotGranted = errors.New("executor: role is not granted")
)

// Executor is a registry entry for a technical backend capable of
// performing work on behalf of one or more Roles
// (docs/specifications/domain/executor.md). Replacing the backend means
// registering a new Executor, never mutating this one (spec Structural
// Invariant 1).
type Executor struct {
	id           string
	backend      string
	roles        []shared.Role
	registeredAt time.Time
	state        State
}

// New registers an Executor in the Registered state (spec Commands:
// Register). The backend identity is fixed for the Executor's lifetime
// (Structural Invariant 1); the initial role set must not be empty
// (Structural Invariant 2).
func New(id, backend string, roles []shared.Role) (*Executor, Registered, error) {
	if id == "" || backend == "" {
		return nil, Registered{}, ErrMissingField
	}
	if len(roles) == 0 {
		return nil, Registered{}, ErrNoRoles
	}

	now := time.Now()
	e := &Executor{
		id:           id,
		backend:      backend,
		roles:        dedupe(roles),
		registeredAt: now,
		state:        StateRegistered,
	}
	event := Registered{ID: id, Backend: backend, Roles: e.Roles(), At: now}
	return e, event, nil
}

// Activate transitions Registered -> Active or Disabled -> Active (spec
// Commands: Activate). What exactly is verified before activation is an
// open question of the specification (Application Layer) — the domain
// only enforces the transition itself.
func (e *Executor) Activate() (Activated, error) {
	if e.state.Terminal() {
		return Activated{}, ErrRetired
	}
	if e.state == StateActive {
		return Activated{}, ErrAlreadyActive
	}

	from := e.state
	e.state = StateActive
	return Activated{ID: e.id, From: from, At: time.Now()}, nil
}

// Disable transitions Active -> Disabled: temporarily unavailable for NEW
// assignments; already-assigned executions are not affected (spec
// Behavioral Invariant 2). Reversible via Activate.
func (e *Executor) Disable() (Disabled, error) {
	if e.state.Terminal() {
		return Disabled{}, ErrRetired
	}
	if e.state != StateActive {
		return Disabled{}, ErrNotActive
	}

	e.state = StateDisabled
	return Disabled{ID: e.id, At: time.Now()}, nil
}

// Retire transitions Registered, Active or Disabled -> Retired: the
// Executor is permanently decommissioned (spec Commands: Retire; the
// direct Registered -> Retired path was added by the final architecture
// review — forcing activation just to decommission an unused backend
// protects no invariant).
func (e *Executor) Retire() (Retired, error) {
	if e.state.Terminal() {
		return Retired{}, ErrRetired
	}

	from := e.state
	e.state = StateRetired
	return Retired{ID: e.id, From: from, At: time.Now()}, nil
}

// GrantRole adds a Role to the set the Executor can perform; the set may
// only grow (spec Behavioral Invariant 3). Granting an already-granted
// role is a no-op: the resulting set is identical either way.
func (e *Executor) GrantRole(role shared.Role) error {
	if role == "" {
		return ErrMissingField
	}
	if e.state.Terminal() {
		return ErrRetired
	}
	for _, r := range e.roles {
		if r == role {
			return nil
		}
	}
	e.roles = append(e.roles, role)
	return nil
}

// RevokeRole removes a Role from the set. The last remaining role cannot
// be revoked: full decommissioning is Retire, never an empty role set
// (spec Behavioral Invariant 3, Structural Invariant 2).
func (e *Executor) RevokeRole(role shared.Role) error {
	if role == "" {
		return ErrMissingField
	}
	if e.state.Terminal() {
		return ErrRetired
	}

	idx := -1
	for i, r := range e.roles {
		if r == role {
			idx = i
			break
		}
	}
	if idx == -1 {
		return ErrRoleNotGranted
	}
	if len(e.roles) == 1 {
		return ErrLastRole
	}
	e.roles = append(e.roles[:idx], e.roles[idx+1:]...)
	return nil
}

// AvailableForAssignment reports whether a new Execution may be assigned
// to this Executor: only while Active (spec Behavioral Invariant 4). The
// assignment itself is performed by the caller (Application Layer), the
// domain only answers the question.
func (e *Executor) AvailableForAssignment() bool { return e.state == StateActive }

// HasRole reports whether the Executor can perform the given Role.
func (e *Executor) HasRole(role shared.Role) bool {
	for _, r := range e.roles {
		if r == role {
			return true
		}
	}
	return false
}

// ID returns the Executor's identifier.
func (e *Executor) ID() string { return e.id }

// Backend returns the backend identity, fixed at registration (spec
// Structural Invariant 1).
func (e *Executor) Backend() string { return e.backend }

// Roles returns a copy of the set of Roles the Executor can perform —
// always at least one (spec Structural Invariant 2).
func (e *Executor) Roles() []shared.Role {
	out := make([]shared.Role, len(e.roles))
	copy(out, e.roles)
	return out
}

// RegisteredAt returns the moment this Executor was registered.
func (e *Executor) RegisteredAt() time.Time { return e.registeredAt }

// State returns the Executor's current Lifecycle state.
func (e *Executor) State() State { return e.state }

// dedupe returns the roles with duplicates removed, preserving order.
func dedupe(roles []shared.Role) []shared.Role {
	out := make([]shared.Role, 0, len(roles))
	for _, r := range roles {
		seen := false
		for _, existing := range out {
			if existing == r {
				seen = true
				break
			}
		}
		if !seen {
			out = append(out, r)
		}
	}
	return out
}
