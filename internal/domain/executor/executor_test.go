package executor

import (
	"errors"
	"testing"

	"ai-studio-os/internal/domain/shared"
)

func newRegistered(t *testing.T) *Executor {
	t.Helper()
	e, event, err := New("exec-1", "claude-code-instance-1", []shared.Role{shared.RoleDeveloper})
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}
	if event.ID != "exec-1" || event.Backend != "claude-code-instance-1" {
		t.Fatalf("Registered event data mismatch: %+v", event)
	}
	return e
}

func newActive(t *testing.T) *Executor {
	t.Helper()
	e := newRegistered(t)
	if _, err := e.Activate(); err != nil {
		t.Fatalf("Activate() unexpected error: %v", err)
	}
	return e
}

// --- New / Structural Invariants ---

func TestNew_SetsRegisteredStateAndFixedIdentity(t *testing.T) {
	e := newRegistered(t)
	if e.State() != StateRegistered {
		t.Errorf("State() = %v, want %v (Structural Invariant 3)", e.State(), StateRegistered)
	}
	if e.Backend() != "claude-code-instance-1" {
		t.Errorf("Backend() = %q, want fixed identity (Structural Invariant 1)", e.Backend())
	}
	if got := e.Roles(); len(got) != 1 || got[0] != shared.RoleDeveloper {
		t.Errorf("Roles() = %v, want [developer] (Structural Invariant 2)", got)
	}
	if e.AvailableForAssignment() {
		t.Error("AvailableForAssignment() in Registered = true, want false (Behavioral Invariant 4)")
	}
}

func TestNew_RequiresIdentityFields(t *testing.T) {
	if _, _, err := New("", "backend", []shared.Role{shared.RoleDeveloper}); !errors.Is(err, ErrMissingField) {
		t.Errorf("New() without id error = %v, want %v", err, ErrMissingField)
	}
	if _, _, err := New("exec-1", "", []shared.Role{shared.RoleDeveloper}); !errors.Is(err, ErrMissingField) {
		t.Errorf("New() without backend error = %v, want %v", err, ErrMissingField)
	}
}

func TestNew_RejectsEmptyRoleSet(t *testing.T) {
	if _, _, err := New("exec-1", "backend", nil); !errors.Is(err, ErrNoRoles) {
		t.Errorf("New() with no roles error = %v, want %v (Structural Invariant 2)", err, ErrNoRoles)
	}
}

func TestNew_DeduplicatesInitialRoles(t *testing.T) {
	e, _, err := New("exec-1", "backend", []shared.Role{shared.RoleDeveloper, shared.RoleDeveloper, shared.RoleReviewer})
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}
	if got := e.Roles(); len(got) != 2 {
		t.Errorf("Roles() = %v, want 2 unique roles", got)
	}
}

// --- Activate ---

func TestActivate_FromRegistered(t *testing.T) {
	e := newRegistered(t)
	event, err := e.Activate()
	if err != nil {
		t.Fatalf("Activate() unexpected error: %v", err)
	}
	if e.State() != StateActive {
		t.Errorf("State() = %v, want %v", e.State(), StateActive)
	}
	if event.From != StateRegistered {
		t.Errorf("Activated.From = %v, want %v", event.From, StateRegistered)
	}
	if !e.AvailableForAssignment() {
		t.Error("AvailableForAssignment() in Active = false, want true")
	}
}

func TestActivate_FromDisabled(t *testing.T) {
	e := newActive(t)
	if _, err := e.Disable(); err != nil {
		t.Fatalf("Disable() unexpected error: %v", err)
	}
	event, err := e.Activate()
	if err != nil {
		t.Fatalf("Activate() from Disabled unexpected error: %v", err)
	}
	if event.From != StateDisabled {
		t.Errorf("Activated.From = %v, want %v (single event for both paths)", event.From, StateDisabled)
	}
}

func TestActivate_RejectedWhenAlreadyActive(t *testing.T) {
	e := newActive(t)
	if _, err := e.Activate(); !errors.Is(err, ErrAlreadyActive) {
		t.Errorf("second Activate() error = %v, want %v", err, ErrAlreadyActive)
	}
}

// --- Disable ---

func TestDisable_OnlyFromActive(t *testing.T) {
	e := newRegistered(t)
	if _, err := e.Disable(); !errors.Is(err, ErrNotActive) {
		t.Errorf("Disable() in Registered error = %v, want %v", err, ErrNotActive)
	}
	if _, err := e.Activate(); err != nil {
		t.Fatalf("Activate() unexpected error: %v", err)
	}
	if _, err := e.Disable(); err != nil {
		t.Fatalf("Disable() from Active unexpected error: %v", err)
	}
	if e.State() != StateDisabled {
		t.Errorf("State() = %v, want %v", e.State(), StateDisabled)
	}
	if e.AvailableForAssignment() {
		t.Error("AvailableForAssignment() in Disabled = true, want false (Behavioral Invariant 4)")
	}
}

// --- Retire / Behavioral Invariant 1 ---

func TestRetire_DirectlyFromRegistered(t *testing.T) {
	// The direct Registered -> Retired path added by the final
	// architecture review: no forced activation just to decommission.
	e := newRegistered(t)
	event, err := e.Retire()
	if err != nil {
		t.Fatalf("Retire() unexpected error: %v", err)
	}
	if event.From != StateRegistered {
		t.Errorf("Retired.From = %v, want %v", event.From, StateRegistered)
	}
	if e.State() != StateRetired {
		t.Errorf("State() = %v, want %v", e.State(), StateRetired)
	}
}

func TestRetire_FromActiveAndDisabled(t *testing.T) {
	active := newActive(t)
	event, err := active.Retire()
	if err != nil {
		t.Fatalf("Retire() from Active unexpected error: %v", err)
	}
	if event.From != StateActive {
		t.Errorf("Retired.From = %v, want %v", event.From, StateActive)
	}

	disabled := newActive(t)
	if _, err := disabled.Disable(); err != nil {
		t.Fatalf("Disable() unexpected error: %v", err)
	}
	event, err = disabled.Retire()
	if err != nil {
		t.Fatalf("Retire() from Disabled unexpected error: %v", err)
	}
	if event.From != StateDisabled {
		t.Errorf("Retired.From = %v, want %v", event.From, StateDisabled)
	}
}

func TestRetired_IsTerminalForAllCommands(t *testing.T) {
	e := newActive(t)
	if _, err := e.Retire(); err != nil {
		t.Fatalf("Retire() unexpected error: %v", err)
	}
	if _, err := e.Activate(); !errors.Is(err, ErrRetired) {
		t.Errorf("Activate() after Retire error = %v, want %v (Behavioral Invariant 1)", err, ErrRetired)
	}
	if _, err := e.Disable(); !errors.Is(err, ErrRetired) {
		t.Errorf("Disable() after Retire error = %v, want %v", err, ErrRetired)
	}
	if _, err := e.Retire(); !errors.Is(err, ErrRetired) {
		t.Errorf("second Retire() error = %v, want %v", err, ErrRetired)
	}
	if err := e.GrantRole(shared.RoleQA); !errors.Is(err, ErrRetired) {
		t.Errorf("GrantRole() after Retire error = %v, want %v", err, ErrRetired)
	}
	if err := e.RevokeRole(shared.RoleDeveloper); !errors.Is(err, ErrRetired) {
		t.Errorf("RevokeRole() after Retire error = %v, want %v", err, ErrRetired)
	}
}

// --- GrantRole / RevokeRole / Behavioral Invariant 3 ---

func TestGrantRole_ExpandsSet(t *testing.T) {
	e := newRegistered(t)
	if err := e.GrantRole(shared.RoleReviewer); err != nil {
		t.Fatalf("GrantRole() unexpected error: %v", err)
	}
	if !e.HasRole(shared.RoleReviewer) || !e.HasRole(shared.RoleDeveloper) {
		t.Errorf("Roles() = %v, want developer+reviewer", e.Roles())
	}
}

func TestGrantRole_AlreadyGrantedIsNoOp(t *testing.T) {
	e := newRegistered(t)
	if err := e.GrantRole(shared.RoleDeveloper); err != nil {
		t.Fatalf("GrantRole() of existing role error = %v, want nil (idempotent)", err)
	}
	if got := e.Roles(); len(got) != 1 {
		t.Errorf("Roles() = %v, want no duplicates", got)
	}
}

func TestRevokeRole_LastRoleForbidden(t *testing.T) {
	e := newRegistered(t)
	if err := e.RevokeRole(shared.RoleDeveloper); !errors.Is(err, ErrLastRole) {
		t.Errorf("RevokeRole() of last role error = %v, want %v (Structural Invariant 2)", err, ErrLastRole)
	}
	if !e.HasRole(shared.RoleDeveloper) {
		t.Error("last role was removed despite the error")
	}
}

func TestRevokeRole_RemovesWhenNotLast(t *testing.T) {
	e := newRegistered(t)
	if err := e.GrantRole(shared.RoleReviewer); err != nil {
		t.Fatalf("GrantRole() unexpected error: %v", err)
	}
	if err := e.RevokeRole(shared.RoleDeveloper); err != nil {
		t.Fatalf("RevokeRole() unexpected error: %v", err)
	}
	if e.HasRole(shared.RoleDeveloper) {
		t.Error("revoked role still present")
	}
	if !e.HasRole(shared.RoleReviewer) {
		t.Error("remaining role lost")
	}
}

func TestRevokeRole_NotGranted(t *testing.T) {
	e := newRegistered(t)
	if err := e.RevokeRole(shared.RoleQA); !errors.Is(err, ErrRoleNotGranted) {
		t.Errorf("RevokeRole() of absent role error = %v, want %v", err, ErrRoleNotGranted)
	}
}

// --- Accessor isolation ---

func TestRoles_ReturnsCopyNotAlias(t *testing.T) {
	e := newRegistered(t)
	out := e.Roles()
	out[0] = shared.RoleArchitect
	if got := e.Roles(); got[0] != shared.RoleDeveloper {
		t.Errorf("mutating accessor result changed the entity: %v", got)
	}
}
