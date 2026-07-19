// Package task defines the contracts of the task write and read paths.
//
// ADR-004 (accepted): PostgreSQL is the source of truth, all state changes
// flow through a single validated write path, and the tasks/ directory is a
// generated markdown export. The internal design of the write path (plain
// persistence or Command -> Event -> Projection) is intentionally NOT fixed
// by these contracts — they name operations, not the mechanism.
package task

import (
	"context"

	"ai-studio-os/internal/domain/shared"
)

// Commands is the single write path for task state (ADR-004). Every
// transition is validated against the canonical state machine
// (docs/architecture/state-machine.md) via the workflow rules contract
// (internal/domain/workflow) and published as a lifecycle event
// (docs/architecture/events.md).
type Commands interface {
	// Create registers a new task in the Backlog state and returns its
	// identifier (format — ADR-011, Decision Required; strings until then).
	Create(ctx context.Context, projectID, title, taskType string) (string, error)

	// Transition moves the task to the target state. The reason is
	// mandatory for transitions that require one (Blocked, Cancelled).
	// Illegal transitions are rejected with an error.
	Transition(ctx context.Context, taskID string, to shared.TaskState, reason string) error
}

// Queries provides task state for the delivery layer (apps/api,
// orchestrator). It is a read contract for applications, not a
// cross-module contract: domain modules read foreign data only through
// events and their own projections (ADR-014).
type Queries interface {
	// State returns the current lifecycle state of the task.
	State(ctx context.Context, taskID string) (shared.TaskState, error)
}

// Exporter renders tasks from the source of truth into human-readable
// markdown files under tasks/ (ADR-004). Export output is a view: editing
// exported files does not change system state.
type Exporter interface {
	// Export renders one task to its markdown representation.
	Export(ctx context.Context, taskID string) error

	// ExportAll renders every task of the platform.
	ExportAll(ctx context.Context) error
}
