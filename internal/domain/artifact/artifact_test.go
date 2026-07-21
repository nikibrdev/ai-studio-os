package artifact

import (
	"errors"
	"testing"
)

func newDraft(t *testing.T) *Artifact {
	t.Helper()
	a, event, err := New("art-1", "proj-1", Type("PullRequest"), OriginProduced, Author("nikita"), "")
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}
	if event.ID != "art-1" {
		t.Fatalf("Created.ID = %q, want %q", event.ID, "art-1")
	}
	return a
}

// --- New / Structural Invariants ---

func TestNew_SetsDraftStateAndFixedFields(t *testing.T) {
	a, event, err := New("art-1", "proj-1", Type("Specification"), OriginProduced, Author("nikita"), "exec-1")
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}
	if a.State() != StateDraft {
		t.Errorf("State() = %v, want %v", a.State(), StateDraft)
	}
	if a.ID() != "art-1" || a.ProjectID() != "proj-1" {
		t.Errorf("identity fields not set correctly: id=%q projectID=%q", a.ID(), a.ProjectID())
	}
	if a.ArtifactType() != Type("Specification") {
		t.Errorf("ArtifactType() = %v, want Specification", a.ArtifactType())
	}
	if a.Origin() != OriginProduced {
		t.Errorf("Origin() = %v, want %v", a.Origin(), OriginProduced)
	}
	if a.ProducedBy() != "exec-1" {
		t.Errorf("ProducedBy() = %q, want %q", a.ProducedBy(), "exec-1")
	}
	if a.Payload() != nil {
		t.Errorf("Payload() = %v, want nil (Structural Invariant 4: Payload may be absent in Draft)", a.Payload())
	}
	if event.Type != Type("Specification") || event.Origin != OriginProduced || event.ProducedBy != "exec-1" {
		t.Errorf("Created event data mismatch: %+v", event)
	}
}

func TestNew_ProducedByOptional(t *testing.T) {
	a, event, err := New("art-1", "proj-1", Type("Document"), OriginUploaded, AuthorUnknown, "")
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}
	if a.ProducedBy() != "" {
		t.Errorf("ProducedBy() = %q, want empty (spec Behavioral Invariant 3: reference is optional)", a.ProducedBy())
	}
	if event.ProducedBy != "" {
		t.Errorf("Created.ProducedBy = %q, want empty", event.ProducedBy)
	}
	if a.Author() != AuthorUnknown {
		t.Errorf("Author() = %v, want AuthorUnknown (spec Structural Invariant 3: a legitimate Author value)", a.Author())
	}
}

func TestNew_RequiresAllIdentityFields(t *testing.T) {
	cases := []struct {
		name      string
		id        string
		projectID string
		typ       Type
		origin    Origin
		author    Author
	}{
		{"missing id", "", "proj-1", "Type", OriginProduced, "author"},
		{"missing projectID", "art-1", "", "Type", OriginProduced, "author"},
		{"missing type", "art-1", "proj-1", "", OriginProduced, "author"},
		{"missing origin", "art-1", "proj-1", "Type", "", "author"},
		{"missing author", "art-1", "proj-1", "Type", OriginProduced, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := New(tc.id, tc.projectID, tc.typ, tc.origin, tc.author, "")
			if !errors.Is(err, ErrMissingField) {
				t.Errorf("New() error = %v, want %v", err, ErrMissingField)
			}
		})
	}
}

// --- UpdateDraft ---

func TestUpdateDraft_UpdatesPayloadAndAuthor(t *testing.T) {
	a := newDraft(t)
	if err := a.UpdateDraft([]byte("content"), Author("reviewer")); err != nil {
		t.Fatalf("UpdateDraft() unexpected error: %v", err)
	}
	if string(a.Payload()) != "content" {
		t.Errorf("Payload() = %q, want %q", a.Payload(), "content")
	}
	if a.Author() != Author("reviewer") {
		t.Errorf("Author() = %v, want %v", a.Author(), Author("reviewer"))
	}
}

func TestUpdateDraft_LeavesFieldsUnchangedWhenOmitted(t *testing.T) {
	a := newDraft(t)
	originalAuthor := a.Author()
	if err := a.UpdateDraft(nil, ""); err != nil {
		t.Fatalf("UpdateDraft() unexpected error: %v", err)
	}
	if a.Author() != originalAuthor {
		t.Errorf("Author() changed to %v despite empty argument", a.Author())
	}
	if a.Payload() != nil {
		t.Errorf("Payload() = %v, want nil (nil argument must not set an empty payload)", a.Payload())
	}
}

func TestUpdateDraft_DoesNotExposeTypeOrOriginChange(t *testing.T) {
	// Structural Invariant 1 and 3: Type and Origin are fixed at Create and
	// UpdateDraft's signature has no parameters for them at all — the
	// strongest guarantee available (a caller cannot even attempt it).
	a := newDraft(t)
	originalType, originalOrigin := a.ArtifactType(), a.Origin()
	if err := a.UpdateDraft([]byte("x"), "someone"); err != nil {
		t.Fatalf("UpdateDraft() unexpected error: %v", err)
	}
	if a.ArtifactType() != originalType || a.Origin() != originalOrigin {
		t.Errorf("Type/Origin changed: got (%v, %v), want (%v, %v)", a.ArtifactType(), a.Origin(), originalType, originalOrigin)
	}
}

func TestUpdateDraft_RejectedAfterPublish(t *testing.T) {
	a := newDraft(t)
	if err := a.UpdateDraft([]byte("content"), ""); err != nil {
		t.Fatalf("UpdateDraft() unexpected error: %v", err)
	}
	if _, err := a.Publish(); err != nil {
		t.Fatalf("Publish() unexpected error: %v", err)
	}
	if err := a.UpdateDraft([]byte("changed"), ""); !errors.Is(err, ErrPublished) {
		t.Errorf("UpdateDraft() after Publish error = %v, want %v (Behavioral Invariant 1: immutable after Publish)", err, ErrPublished)
	}
}

func TestUpdateDraft_RejectedAfterArchive(t *testing.T) {
	a := newDraft(t)
	if _, err := a.Archive(); err != nil {
		t.Fatalf("Archive() unexpected error: %v", err)
	}
	if err := a.UpdateDraft([]byte("changed"), ""); !errors.Is(err, ErrArchived) {
		t.Errorf("UpdateDraft() after Archive error = %v, want %v", err, ErrArchived)
	}
}

// --- Publish ---

func TestPublish_RequiresNonEmptyPayload(t *testing.T) {
	a := newDraft(t)
	if _, err := a.Publish(); !errors.Is(err, ErrPayloadRequired) {
		t.Errorf("Publish() with no payload error = %v, want %v (spec Commands: Publish requires non-empty Payload)", err, ErrPayloadRequired)
	}
}

func TestPublish_SucceedsWithPayloadAndEmitsEvent(t *testing.T) {
	a := newDraft(t)
	if err := a.UpdateDraft([]byte("content"), ""); err != nil {
		t.Fatalf("UpdateDraft() unexpected error: %v", err)
	}
	event, err := a.Publish()
	if err != nil {
		t.Fatalf("Publish() unexpected error: %v", err)
	}
	if a.State() != StatePublished {
		t.Errorf("State() = %v, want %v", a.State(), StatePublished)
	}
	if event.ID != a.ID() {
		t.Errorf("Published.ID = %q, want %q", event.ID, a.ID())
	}
}

func TestPublish_RejectedWhenAlreadyPublished(t *testing.T) {
	a := newDraft(t)
	if err := a.UpdateDraft([]byte("content"), ""); err != nil {
		t.Fatalf("UpdateDraft() unexpected error: %v", err)
	}
	if _, err := a.Publish(); err != nil {
		t.Fatalf("Publish() unexpected error: %v", err)
	}
	if _, err := a.Publish(); !errors.Is(err, ErrPublished) {
		t.Errorf("second Publish() error = %v, want %v", err, ErrPublished)
	}
}

func TestPublish_RejectedWhenArchived(t *testing.T) {
	a := newDraft(t)
	if _, err := a.Archive(); err != nil {
		t.Fatalf("Archive() unexpected error: %v", err)
	}
	if _, err := a.Publish(); !errors.Is(err, ErrArchived) {
		t.Errorf("Publish() after Archive error = %v, want %v", err, ErrArchived)
	}
}

// --- Archive ---

func TestArchive_DirectlyFromDraft(t *testing.T) {
	a := newDraft(t)
	event, err := a.Archive()
	if err != nil {
		t.Fatalf("Archive() unexpected error: %v", err)
	}
	if a.State() != StateArchived {
		t.Errorf("State() = %v, want %v", a.State(), StateArchived)
	}
	if event.From != StateDraft {
		t.Errorf("Archived.From = %v, want %v (spec: subscribers must be able to tell an archived draft apart)", event.From, StateDraft)
	}
}

func TestArchive_FromPublished(t *testing.T) {
	a := newDraft(t)
	if err := a.UpdateDraft([]byte("content"), ""); err != nil {
		t.Fatalf("UpdateDraft() unexpected error: %v", err)
	}
	if _, err := a.Publish(); err != nil {
		t.Fatalf("Publish() unexpected error: %v", err)
	}
	event, err := a.Archive()
	if err != nil {
		t.Fatalf("Archive() unexpected error: %v", err)
	}
	if event.From != StatePublished {
		t.Errorf("Archived.From = %v, want %v", event.From, StatePublished)
	}
}

func TestArchive_DoesNotClearContent(t *testing.T) {
	// Behavioral Invariant 2: "Archived does not mean deleted."
	a := newDraft(t)
	if err := a.UpdateDraft([]byte("content"), Author("nikita")); err != nil {
		t.Fatalf("UpdateDraft() unexpected error: %v", err)
	}
	if _, err := a.Archive(); err != nil {
		t.Fatalf("Archive() unexpected error: %v", err)
	}
	if string(a.Payload()) != "content" {
		t.Errorf("Payload() after Archive = %q, want %q (Behavioral Invariant 2: Archived != deleted)", a.Payload(), "content")
	}
	if a.Author() != Author("nikita") {
		t.Errorf("Author() after Archive = %v, want unchanged", a.Author())
	}
}

func TestArchive_RejectedWhenAlreadyArchived(t *testing.T) {
	a := newDraft(t)
	if _, err := a.Archive(); err != nil {
		t.Fatalf("Archive() unexpected error: %v", err)
	}
	if _, err := a.Archive(); !errors.Is(err, ErrArchived) {
		t.Errorf("second Archive() error = %v, want %v (Lifecycle: Archived is terminal)", err, ErrArchived)
	}
}

// --- Lifecycle irreversibility (Behavioral Invariant 4) ---

func TestLifecycle_TransitionsAreIrreversible(t *testing.T) {
	a := newDraft(t)
	if err := a.UpdateDraft([]byte("content"), ""); err != nil {
		t.Fatalf("UpdateDraft() unexpected error: %v", err)
	}
	if _, err := a.Publish(); err != nil {
		t.Fatalf("Publish() unexpected error: %v", err)
	}
	if _, err := a.Archive(); err != nil {
		t.Fatalf("Archive() unexpected error: %v", err)
	}
	// No method exists to move the Artifact back from Archived to
	// Published or Draft, or from Published back to Draft — the type's
	// API surface itself enforces one-way progression (Draft -> Published
	// -> Archived); this test documents that guarantee at the state level.
	if a.State() != StateArchived {
		t.Fatalf("State() = %v, want %v", a.State(), StateArchived)
	}
	if _, err := a.Publish(); !errors.Is(err, ErrArchived) {
		t.Errorf("Publish() from Archived error = %v, want %v", err, ErrArchived)
	}
}
