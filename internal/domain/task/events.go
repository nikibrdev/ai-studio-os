package task

import (
	"time"

	"ai-studio-os/internal/domain/shared"
)

// Created is the data of the event published when a Task is registered
// (enters Backlog). Mapping to the canonical event name TaskCreated
// (internal/domain/event, docs/architecture/events.md) is the publisher's
// job, not the entity's.
type Created struct {
	ID        string
	ProjectID string
	EpicID    string // empty for a task outside any Epic
	Title     string
	Type      string
	At        time.Time
}

// Transitioned is the data of the event published on every state
// transition. One transition — one event (state-machine.md invariant 2;
// the Testing -> Done double event TestsPassed+TaskCompleted is the
// publisher's mapping concern). Mapping From/To pairs to the 15 canonical
// event names of docs/architecture/events.md belongs to the publisher
// (Application Layer), keeping the entity free of the event catalogue.
type Transitioned struct {
	ID     string
	From   shared.TaskState
	To     shared.TaskState
	Reason string // non-empty for Blocked and Cancelled (state-machine.md invariant 3)
	At     time.Time
}
