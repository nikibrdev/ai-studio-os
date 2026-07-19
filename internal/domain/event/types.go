// Package event defines the event type identifiers of the platform.
//
// The catalogue — sources, recipients, data and consequences of every
// event — is docs/architecture/events.md; the names below match it verbatim.
// Event payload schemas belong to their source modules and are defined in
// the Domain Layer epic. Subscriptions must use these constants, never
// string literals.
package event

// Primary lifecycle events (docs/architecture/events.md).
const (
	TaskCreated     = "TaskCreated"
	TaskPlanned     = "TaskPlanned"
	TaskStarted     = "TaskStarted"
	TaskCompleted   = "TaskCompleted"
	ReviewRequested = "ReviewRequested"
	ReviewCompleted = "ReviewCompleted"
	TestsPassed     = "TestsPassed"
	TestsFailed     = "TestsFailed"
	MergeRequested  = "MergeRequested"
	MergeCompleted  = "MergeCompleted"
)

// Supplementary lifecycle events covering the remaining state machine
// transitions (docs/architecture/state-machine.md).
const (
	TaskReturnedToBacklog = "TaskReturnedToBacklog"
	TaskBlocked           = "TaskBlocked"
	TaskUnblocked         = "TaskUnblocked"
	TaskCancelled         = "TaskCancelled"
	TaskArchived          = "TaskArchived"
)
