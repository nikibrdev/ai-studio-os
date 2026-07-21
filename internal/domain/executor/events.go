package executor

import (
	"time"

	"ai-studio-os/internal/domain/shared"
)

// Registered is the data of the event published when an Executor is added
// to the registry (enters Registered) (spec Domain Events:
// ExecutorRegistered).
type Registered struct {
	ID      string
	Backend string // the backend identity, e.g. a concrete Claude Code instance
	Roles   []shared.Role
	At      time.Time
}

// Activated is the data of the event published when an Executor enters
// Active, from either Registered or Disabled — one event for both paths,
// since the resulting state is the same (spec Domain Events:
// ExecutorActivated).
type Activated struct {
	ID   string
	From State // Registered or Disabled
	At   time.Time
}

// Disabled is the data of the event published on Active -> Disabled (spec
// Domain Events: ExecutorDisabled).
type Disabled struct {
	ID string
	At time.Time
}

// Retired is the data of the event published when an Executor enters
// Retired, from Registered, Active or Disabled — one event for all paths
// (spec Domain Events: ExecutorRetired).
type Retired struct {
	ID   string
	From State // Registered, Active or Disabled
	At   time.Time
}
