package application

import (
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
	data          map[string]string
}

var _ platform.Event = Envelope{}

// NewEvent wraps a domain fact into a platform.Event envelope. eventType
// must be one of the canonical names in internal/domain/event. actor may
// be "" for a system-initiated fact (platform.Event contract).
func NewEvent(eventType, source, actor, projectID, subjectID string, occurredAt time.Time) Envelope {
	return Envelope{
		id:            NewID(),
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

// WithData attaches event-type-specific data to a copy of the envelope.
// This is deliberately NOT part of the platform.Event contract (ADR-002,
// EPIC-002): platform.Event carries only the common fields every event
// has; WithData/Data are extra methods on the concrete Envelope type,
// reachable only by code in internal/application that knows about
// Envelope specifically (e.g. a projection type-asserting the
// platform.Event it receives back to Envelope) — not a change to the
// accepted contract itself.
func (e Envelope) WithData(data map[string]string) Envelope {
	e.data = data
	return e
}

// Data returns the event-type-specific data attached via WithData, or nil
// if none was attached.
func (e Envelope) Data() map[string]string { return e.data }
