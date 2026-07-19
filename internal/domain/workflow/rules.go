package workflow

import "ai-studio-os/internal/domain/shared"

// Rules is the "Workflow" contract of docs/architecture/interfaces.md: it
// applies the canonical task state machine
// (docs/architecture/state-machine.md), deciding whether a transition is
// allowed and which role owns the next step. It decides — it never acts:
// state is changed by the task module, events are published by their owners
// (ADR-014).
//
// Contract constraints:
//   - decisions are deterministic: identical input yields identical output;
//   - only transitions listed in state-machine.md may be allowed;
//   - the implementation holds no mutable state and performs no I/O;
//     persistence, if any, goes through ports (Workflow -> SQL is
//     forbidden, ADR-014).
type Rules interface {
	// CanTransition reports whether the transition from one state to
	// another is allowed. A nil return means the transition is allowed;
	// otherwise the error explains the violated rule.
	CanTransition(from, to shared.TaskState) error

	// NextRole returns the role responsible for acting on a task in the
	// given state (docs/architecture/workflow.md, participation table).
	NextRole(state shared.TaskState) (shared.Role, error)
}
