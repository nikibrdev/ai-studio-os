package application

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"ai-studio-os/internal/platform"
)

// Envelope adapts a domain event's data to the platform.Event contract
// (docs/architecture/events.md, "Общие поля всех событий"). Domain
// packages return their events as plain data values and never depend on
// internal/platform (ADR-015); the Application Layer is the seam where a
// canonical event type name (internal/domain/event) and delivery metadata
// are attached before publishing through platform.EventBus (ADR-002).
type Envelope struct {
	id            string
	eventType     string
	schemaVersion int
	occurredAt    time.Time
	source        string
	actor         string
	projectID     string
	subjectID     string
}

var _ platform.Event = Envelope{}

// NewEvent wraps a domain fact into a platform.Event envelope. eventType
// must be one of the canonical names in internal/domain/event. actor may
// be "" for a system-initiated fact (platform.Event contract).
func NewEvent(eventType, source, actor, projectID, subjectID string, occurredAt time.Time) Envelope {
	return Envelope{
		id:            newEventID(),
		eventType:     eventType,
		schemaVersion: 1,
		occurredAt:    occurredAt,
		source:        source,
		actor:         actor,
		projectID:     projectID,
		subjectID:     subjectID,
	}
}

// ID returns the unique identifier of this event instance.
func (e Envelope) ID() string { return e.id }

// Type returns the canonical event type name (internal/domain/event).
func (e Envelope) Type() string { return e.eventType }

// SchemaVersion returns the version of the event payload schema.
func (e Envelope) SchemaVersion() int { return e.schemaVersion }

// OccurredAt returns the moment the wrapped fact took place.
func (e Envelope) OccurredAt() time.Time { return e.occurredAt }

// Source returns the name of the domain module that produced the fact.
func (e Envelope) Source() string { return e.source }

// Actor returns the initiator of the underlying action, or "" when the
// event is system-initiated.
func (e Envelope) Actor() string { return e.actor }

// ProjectID returns the project the event belongs to.
func (e Envelope) ProjectID() string { return e.projectID }

// SubjectID returns the identifier of the domain entity the event is
// about.
func (e Envelope) SubjectID() string { return e.subjectID }

// newEventID generates a random hex identifier using only the standard
// library: no UUID dependency is listed in .claude/context/stack.md, and
// event identity needs only uniqueness, not a specific structured format.
func newEventID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// crypto/rand.Read does not fail on any platform this project
		// targets; a zero-value ID would silently violate the Event
		// contract, so panic is the honest response to an impossible
		// condition rather than returning a bad ID.
		panic("application: failed to generate event id: " + err.Error())
	}
	return hex.EncodeToString(b[:])
}
