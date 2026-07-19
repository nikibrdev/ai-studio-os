// Package workflow defines the contracts of process rules and definitions:
// which steps a task goes through, which role owns each step and whether a
// state transition is allowed
// (docs/architecture/domain-model.md, "Workflow" and "Workflow Step").
//
// The transition table itself is implemented in the Domain Layer epic
// strictly per docs/architecture/state-machine.md.
package workflow

import "ai-studio-os/internal/domain/shared"

// Definition is a published, versioned description of a process. A published
// version is immutable: any change is a new version
// (Lifecycle: Draft -> Published -> Deprecated).
type Definition interface {
	// Name returns the workflow name (the MVP has one standard workflow).
	Name() string

	// Version returns the version of this definition.
	Version() int

	// Steps returns the ordered steps of the process.
	Steps() []Step
}

// Step is one step of a workflow with a single responsible role and a
// verifiable entry condition expressed as the task state the step acts on.
type Step interface {
	// Name returns the step name (e.g. "planning", "implementation").
	Name() string

	// Role returns the single role responsible for this step.
	Role() shared.Role

	// EntryState returns the task state in which this step becomes active.
	EntryState() shared.TaskState
}
