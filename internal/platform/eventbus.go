package platform

import (
	"context"
	"time"
)

// Event is an immutable fact that happened in the system. Once published it
// is never modified or deleted; a change in meaning is a new event and a
// change in structure is a new schema version
// (docs/architecture/event-model.md).
//
// The accessors mirror the common event fields catalogued in
// docs/architecture/events.md. Concrete event types are defined by their
// source modules in later epics.
type Event interface {
	// ID returns the unique identifier of this event instance.
	ID() string

	// Type returns the event type identifier as catalogued in
	// docs/architecture/events.md (see internal/domain/event for the
	// constants).
	Type() string

	// SchemaVersion returns the version of the event payload schema.
	SchemaVersion() int

	// OccurredAt returns the moment the fact took place.
	OccurredAt() time.Time

	// Source returns the name of the Core module that published the event.
	Source() string

	// Actor returns the initiator of the underlying action: the role and
	// executor (human or agent). Empty when the event is system-initiated.
	Actor() string

	// ProjectID returns the project the event belongs to. Empty for
	// platform-wide events.
	ProjectID() string

	// SubjectID returns the identifier of the domain entity the event is
	// about (usually a task). Empty when not applicable.
	SubjectID() string
}

// EventHandler processes one delivered event. Handlers must be idempotent:
// the bus contract permits repeated delivery of the same event (ADR-002).
// A handler must not block for long and must not assume any delivery order
// beyond what the bus implementation documents.
type EventHandler func(ctx context.Context, e Event) error

// Subscription represents an active subscription of a handler to an event
// type. Lifecycle: Registered -> Active -> Cancelled.
type Subscription interface {
	// Cancel stops delivery to the subscribed handler. Cancelling an
	// already cancelled subscription is a no-op.
	Cancel() error
}

// EventBus delivers published events to all subscribers of the matching
// event type and feeds the event journal
// (docs/architecture/interfaces.md, "Event Bus").
//
// Contract constraints:
//   - events are immutable; the bus never alters them;
//   - publishers do not know their subscribers; adding a subscriber never
//     requires changes on the publishing side;
//   - the MVP implementation is an in-process (in-memory) bus (ADR-002);
//     this interface stays unchanged when the implementation is replaced
//     with Redis Streams or NATS.
type EventBus interface {
	// Publish delivers the event to all current subscribers of its type
	// and returns after the bus has accepted the event.
	Publish(ctx context.Context, e Event) error

	// Subscribe registers a handler for the given event type (a constant
	// from internal/events) and returns the active subscription.
	Subscribe(eventType string, h EventHandler) (Subscription, error)
}
