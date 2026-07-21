package artifact

import (
	"errors"
	"time"
)

// Sentinel errors returned by Artifact commands.
var (
	// ErrMissingField is returned by New when a required identity field
	// (Identifier, ProjectID, Type, Origin or Author) is empty (spec
	// Structural Invariants 1-3).
	ErrMissingField = errors.New("artifact: required field is missing")

	// ErrArchived is returned when a command is attempted against an
	// Archived Artifact: Archived is terminal (spec Lifecycle).
	ErrArchived = errors.New("artifact: archived artifacts cannot change")

	// ErrPublished is returned when UpdateDraft or Publish is attempted
	// against an already-Published Artifact: it is immutable after
	// Publish (spec Behavioral Invariant 1).
	ErrPublished = errors.New("artifact: published artifacts cannot change except by archiving")

	// ErrPayloadRequired is returned by Publish when the Artifact has no
	// Payload yet: only something that already exists can be fixed as
	// immutable (spec Commands: Publish).
	ErrPayloadRequired = errors.New("artifact: payload is required to publish")
)

// Artifact is a durable engineering result the platform can create, store
// and reuse independently of the process that produced it
// (docs/specifications/domain/artifact.md).
//
// An Artifact is immutable once Published (Behavioral Invariant 1): a
// correction is a new Artifact, never a mutation of this one.
type Artifact struct {
	id         string
	projectID  string
	typ        Type
	origin     Origin
	author     Author
	createdAt  time.Time
	producedBy string // identifier of the Execution that first introduced this Artifact; empty if none
	payload    []byte
	state      State
}

// New creates an Artifact in the Draft state (spec Commands: Create).
// Identifier, ProjectID, Type and Origin are fixed for the Artifact's
// lifetime (Structural Invariants 1 and 3) — no later command changes
// them. producedBy is optional: pass "" if this Artifact was not
// introduced by an Execution (spec Behavioral Invariant 3).
func New(id, projectID string, typ Type, origin Origin, author Author, producedBy string) (*Artifact, Created, error) {
	if id == "" || projectID == "" || typ == "" || origin == "" || author == "" {
		return nil, Created{}, ErrMissingField
	}

	now := time.Now()
	a := &Artifact{
		id:         id,
		projectID:  projectID,
		typ:        typ,
		origin:     origin,
		author:     author,
		createdAt:  now,
		producedBy: producedBy,
		state:      StateDraft,
	}
	event := Created{
		ID:         id,
		ProjectID:  projectID,
		Type:       typ,
		Origin:     origin,
		Author:     author,
		ProducedBy: producedBy,
		At:         now,
	}
	return a, event, nil
}

// UpdateDraft updates the Payload and/or refines the Author while the
// Artifact is in Draft (spec Commands: UpdateDraft). Type and Origin are
// not parameters here: Type is fixed at Create (Structural Invariant 1),
// Origin is a historical fact of how the Artifact entered the system and is
// not redefined afterwards either (Structural Invariant 3). Passing an
// empty Author or a nil payload leaves that field unchanged.
func (a *Artifact) UpdateDraft(payload []byte, author Author) error {
	switch a.state {
	case StateArchived:
		return ErrArchived
	case StatePublished:
		return ErrPublished
	}
	if author != "" {
		a.author = author
	}
	if payload != nil {
		a.payload = payload
	}
	return nil
}

// Publish transitions Draft -> Published (spec Commands: Publish). It
// requires a non-empty Payload: only something that already exists can be
// fixed as immutable. After Publish, UpdateDraft is no longer a valid
// command for this Artifact (Behavioral Invariant 1).
func (a *Artifact) Publish() (Published, error) {
	switch a.state {
	case StateArchived:
		return Published{}, ErrArchived
	case StatePublished:
		return Published{}, ErrPublished
	}
	if len(a.payload) == 0 {
		return Published{}, ErrPayloadRequired
	}

	a.state = StatePublished
	return Published{ID: a.id, ProducedBy: a.producedBy, At: time.Now()}, nil
}

// Archive transitions into Archived, from either Draft or Published (spec
// Commands: Archive; spec Lifecycle allows both paths). It never changes
// or removes content: Archived does not mean deleted (Behavioral
// Invariant 2).
func (a *Artifact) Archive() (Archived, error) {
	if a.state == StateArchived {
		return Archived{}, ErrArchived
	}

	from := a.state
	a.state = StateArchived
	return Archived{ID: a.id, From: from, At: time.Now()}, nil
}

// ID returns the Artifact's identifier.
func (a *Artifact) ID() string { return a.id }

// ProjectID returns the identifier of the Project that owns this Artifact
// (spec Relationships: Project владеет Artifact).
func (a *Artifact) ProjectID() string { return a.projectID }

// ArtifactType returns the Artifact's Type, fixed at creation.
func (a *Artifact) ArtifactType() Type { return a.typ }

// Origin returns how the Artifact entered the system, fixed at creation.
func (a *Artifact) Origin() Origin { return a.origin }

// Author returns the subject currently responsible for the Artifact.
func (a *Artifact) Author() Author { return a.author }

// CreatedAt returns the moment this Artifact was created.
func (a *Artifact) CreatedAt() time.Time { return a.createdAt }

// ProducedBy returns the identifier of the Execution that first introduced
// this Artifact, or "" if none (spec Behavioral Invariant 3).
func (a *Artifact) ProducedBy() string { return a.producedBy }

// Payload returns the Artifact's content, or nil if none has been set yet
// (spec Structural Invariant 4).
func (a *Artifact) Payload() []byte { return a.payload }

// State returns the Artifact's current Lifecycle state.
func (a *Artifact) State() State { return a.state }
