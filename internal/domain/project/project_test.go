package project

import (
	"errors"
	"testing"
)

func newCreated(t *testing.T) *Project {
	t.Helper()
	p, event, err := New("proj-1", "AI Studio OS")
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}
	if event.ID != "proj-1" || event.Name != "AI Studio OS" {
		t.Fatalf("Created event data mismatch: %+v", event)
	}
	return p
}

func newActive(t *testing.T) *Project {
	t.Helper()
	p := newCreated(t)
	if _, _, err := p.ConnectRepository("github.com/org/repo"); err != nil {
		t.Fatalf("ConnectRepository() unexpected error: %v", err)
	}
	if _, err := p.Activate(); err != nil {
		t.Fatalf("Activate() unexpected error: %v", err)
	}
	return p
}

// --- New ---

func TestNew_SetsCreatedState(t *testing.T) {
	p := newCreated(t)
	if p.State() != StateCreated {
		t.Errorf("State() = %v, want %v", p.State(), StateCreated)
	}
	if p.Repositories() != nil {
		t.Errorf("Repositories() = %v, want nil at creation (Structural Invariant 1 applies to Active only)", p.Repositories())
	}
	if p.AcceptsNewContent() {
		t.Error("AcceptsNewContent() in Created = true, want false (Behavioral Invariant 4)")
	}
}

func TestNew_RequiresFields(t *testing.T) {
	if _, _, err := New("", "name"); !errors.Is(err, ErrMissingField) {
		t.Errorf("New() without id error = %v, want %v", err, ErrMissingField)
	}
	if _, _, err := New("proj-1", ""); !errors.Is(err, ErrMissingField) {
		t.Errorf("New() without name error = %v, want %v", err, ErrMissingField)
	}
}

// --- ConnectRepository ---

func TestConnectRepository_AllowedInCreatedAndActive(t *testing.T) {
	p := newCreated(t)
	event, added, err := p.ConnectRepository("github.com/org/repo")
	if err != nil || !added {
		t.Fatalf("ConnectRepository() in Created = (%v, %v), want added", err, added)
	}
	if event.Repository != "github.com/org/repo" {
		t.Errorf("RepositoryConnected.Repository = %q", event.Repository)
	}
	if _, err := p.Activate(); err != nil {
		t.Fatalf("Activate() unexpected error: %v", err)
	}
	if _, added, err := p.ConnectRepository("github.com/org/infra"); err != nil || !added {
		t.Fatalf("ConnectRepository() in Active = (%v, %v), want added (spec: several repositories)", err, added)
	}
	if got := p.Repositories(); len(got) != 2 {
		t.Errorf("Repositories() = %v, want 2", got)
	}
}

func TestConnectRepository_DuplicateIsNoOp(t *testing.T) {
	p := newCreated(t)
	if _, _, err := p.ConnectRepository("github.com/org/repo"); err != nil {
		t.Fatalf("ConnectRepository() unexpected error: %v", err)
	}
	_, added, err := p.ConnectRepository("github.com/org/repo")
	if err != nil {
		t.Fatalf("duplicate ConnectRepository() error = %v, want nil (idempotent)", err)
	}
	if added {
		t.Error("duplicate ConnectRepository() reported added = true, want false")
	}
	if got := p.Repositories(); len(got) != 1 {
		t.Errorf("Repositories() = %v, want no duplicates", got)
	}
}

func TestConnectRepository_DoesNotTransitionState(t *testing.T) {
	// The final architecture review replaced the implicit-trigger
	// hypothesis with the explicit Activate command: connecting the first
	// repository must NOT flip the state by itself.
	p := newCreated(t)
	if _, _, err := p.ConnectRepository("github.com/org/repo"); err != nil {
		t.Fatalf("ConnectRepository() unexpected error: %v", err)
	}
	if p.State() != StateCreated {
		t.Errorf("State() after first ConnectRepository = %v, want still %v", p.State(), StateCreated)
	}
}

func TestConnectRepository_RejectedWhenArchived(t *testing.T) {
	p := newActive(t)
	if _, err := p.Archive(); err != nil {
		t.Fatalf("Archive() unexpected error: %v", err)
	}
	if _, _, err := p.ConnectRepository("github.com/org/late"); !errors.Is(err, ErrArchived) {
		t.Errorf("ConnectRepository() in Archived error = %v, want %v", err, ErrArchived)
	}
}

// --- Activate: explicit command with guard ---

func TestActivate_GuardRequiresRepository(t *testing.T) {
	p := newCreated(t)
	if _, err := p.Activate(); !errors.Is(err, ErrNoRepository) {
		t.Errorf("Activate() without repository error = %v, want %v (guard: Structural Invariant 1)", err, ErrNoRepository)
	}
	if p.State() != StateCreated {
		t.Errorf("State() = %v, want unchanged %v", p.State(), StateCreated)
	}
}

func TestActivate_WithRepository(t *testing.T) {
	p := newCreated(t)
	if _, _, err := p.ConnectRepository("github.com/org/repo"); err != nil {
		t.Fatalf("ConnectRepository() unexpected error: %v", err)
	}
	event, err := p.Activate()
	if err != nil {
		t.Fatalf("Activate() unexpected error: %v", err)
	}
	if p.State() != StateActive {
		t.Errorf("State() = %v, want %v", p.State(), StateActive)
	}
	if event.ID != p.ID() {
		t.Errorf("Activated.ID = %q, want %q", event.ID, p.ID())
	}
	if !p.AcceptsNewContent() {
		t.Error("AcceptsNewContent() in Active = false, want true")
	}
}

func TestActivate_RejectedWhenAlreadyActive(t *testing.T) {
	p := newActive(t)
	if _, err := p.Activate(); !errors.Is(err, ErrAlreadyActive) {
		t.Errorf("second Activate() error = %v, want %v", err, ErrAlreadyActive)
	}
}

// --- Archive: terminal ---

func TestArchive_OnlyFromActive(t *testing.T) {
	p := newCreated(t)
	if _, err := p.Archive(); !errors.Is(err, ErrNotActive) {
		t.Errorf("Archive() in Created error = %v, want %v (spec Lifecycle: only Active -> Archived)", err, ErrNotActive)
	}
}

func TestArchive_IsTerminal(t *testing.T) {
	p := newActive(t)
	if _, err := p.Archive(); err != nil {
		t.Fatalf("Archive() unexpected error: %v", err)
	}
	if p.State() != StateArchived {
		t.Errorf("State() = %v, want %v", p.State(), StateArchived)
	}
	if p.AcceptsNewContent() {
		t.Error("AcceptsNewContent() in Archived = true, want false (Behavioral Invariant 4)")
	}
	if _, err := p.Archive(); !errors.Is(err, ErrArchived) {
		t.Errorf("second Archive() error = %v, want %v (Behavioral Invariant 1)", err, ErrArchived)
	}
	if _, err := p.Activate(); !errors.Is(err, ErrArchived) {
		t.Errorf("Activate() after Archive error = %v, want %v", err, ErrArchived)
	}
	if got := p.Repositories(); len(got) != 1 {
		t.Errorf("Repositories() after Archive = %v, want kept (archived != deleted)", got)
	}
}

// --- Accessor isolation ---

func TestRepositories_ReturnsCopyNotAlias(t *testing.T) {
	p := newActive(t)
	out := p.Repositories()
	out[0] = "tampered"
	if got := p.Repositories(); got[0] != "github.com/org/repo" {
		t.Errorf("mutating accessor result changed the entity: %v", got)
	}
}
