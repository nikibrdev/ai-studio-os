package execution

import "time"

// Queued is the data of the event published when an Execution is created
// (enters Queued) (spec Domain Events: ExecutionQueued).
type Queued struct {
	ID         string
	TaskID     string
	ExecutorID string
	At         time.Time
}

// Started is the data of the event published on Queued -> Running, after
// the Executor confirmed accepting the work (spec Domain Events:
// ExecutionStarted).
type Started struct {
	ID string
	At time.Time
}

// Succeeded is the data of the event published on Running -> Succeeded
// (spec Domain Events: ExecutionSucceeded).
type Succeeded struct {
	ID          string
	ArtifactIDs []string // artifacts produced during this execution, if any
	At          time.Time
}

// Failed is the data of the event published on Running -> Failed. Artifacts
// produced before the failure are still carried: a failure report is a
// result of work too (spec Domain Events: ExecutionFailed).
type Failed struct {
	ID          string
	ArtifactIDs []string
	At          time.Time
}

// Aborted is the data of the event published when an Execution enters
// Aborted, from either Queued or Running — one event for both paths, since
// the resulting state is the same (spec Domain Events: ExecutionAborted).
type Aborted struct {
	ID   string
	From State // Queued or Running
	At   time.Time
}
