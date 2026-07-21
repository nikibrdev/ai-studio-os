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

// Domain Layer entity events (docs/architecture/events.md, "События
// доменных сущностей Domain Layer") — defined in the approved
// specifications of Artifact, Execution, Executor and Project; catalogued
// here at TASK-042 (EPIC-004), which was the first consumer to need them
// as constants rather than string literals.
const (
	ArtifactCreated   = "ArtifactCreated"
	ArtifactPublished = "ArtifactPublished"
	ArtifactArchived  = "ArtifactArchived"

	ExecutionQueued    = "ExecutionQueued"
	ExecutionStarted   = "ExecutionStarted"
	ExecutionSucceeded = "ExecutionSucceeded"
	ExecutionFailed    = "ExecutionFailed"
	ExecutionAborted   = "ExecutionAborted"

	ExecutorRegistered = "ExecutorRegistered"
	ExecutorActivated  = "ExecutorActivated"
	ExecutorDisabled   = "ExecutorDisabled"
	ExecutorRetired    = "ExecutorRetired"

	ProjectCreated      = "ProjectCreated"
	RepositoryConnected = "RepositoryConnected"
	ProjectActivated    = "ProjectActivated"
	ProjectArchived     = "ProjectArchived"
)
