package workflow

import (
	"errors"
	"fmt"

	"ai-studio-os/internal/domain/shared"
)

// Sentinel errors returned by Machine.
var (
	// ErrTransitionNotAllowed is returned by CanTransition for any pair
	// of states not listed in the canonical transition table
	// (docs/architecture/state-machine.md, invariant 1: only the listed
	// transitions are legal, anything else is an error).
	ErrTransitionNotAllowed = errors.New("workflow: transition is not allowed by the canonical state machine")

	// ErrUnknownState is returned when a state is not one of the nine
	// canonical states.
	ErrUnknownState = errors.New("workflow: unknown task state")
)

// Machine is the stateless implementation of the Rules contract: the
// canonical task state machine of docs/architecture/state-machine.md and
// the per-stage role table of docs/architecture/workflow.md, expressed as
// immutable package-level tables. It holds no mutable state and performs
// no I/O; identical input always yields identical output (contract
// constraints — rules.go). Implemented per the architect decision
// engineering/decisions/2026-07-21-workflow-rules-canonical-source.md:
// state-machine.md is itself the specification-grade single source of
// truth for this module.
type Machine struct{}

// Compile-time check: Machine satisfies the Rules contract.
var _ Rules = Machine{}

// allowedTransitions is the verbatim transition table of
// docs/architecture/state-machine.md (20 transitions; creation into
// Backlog is not a state-to-state transition and is not listed).
var allowedTransitions = map[shared.TaskState]map[shared.TaskState]bool{
	shared.StateBacklog: {
		shared.StateReady:     true, // TaskPlanned
		shared.StateCancelled: true, // TaskCancelled
	},
	shared.StateReady: {
		shared.StateBacklog:    true, // TaskReturnedToBacklog
		shared.StateInProgress: true, // TaskStarted
		shared.StateBlocked:    true, // TaskBlocked
		shared.StateCancelled:  true, // TaskCancelled
	},
	shared.StateInProgress: {
		shared.StateReview:    true, // ReviewRequested
		shared.StateBlocked:   true, // TaskBlocked
		shared.StateCancelled: true, // TaskCancelled
	},
	shared.StateReview: {
		shared.StateInProgress: true, // ReviewCompleted (changes requested)
		shared.StateTesting:    true, // ReviewCompleted (approved)
		shared.StateBlocked:    true, // TaskBlocked
	},
	shared.StateTesting: {
		shared.StateInProgress: true, // TestsFailed
		shared.StateDone:       true, // TestsPassed -> TaskCompleted
		shared.StateBlocked:    true, // TaskBlocked
	},
	shared.StateDone: {
		shared.StateArchived: true, // TaskArchived
	},
	shared.StateBlocked: {
		shared.StateReady:      true, // TaskUnblocked
		shared.StateInProgress: true, // TaskUnblocked
		shared.StateCancelled:  true, // TaskCancelled
	},
	shared.StateCancelled: {
		shared.StateArchived: true, // TaskArchived
	},
	shared.StateArchived: {}, // terminal (invariant 7)
}

// stateRoles is the verbatim per-stage responsibility table of
// docs/architecture/workflow.md ("Участие ролей по стадиям").
var stateRoles = map[shared.TaskState]shared.Role{
	shared.StateBacklog:    shared.RoleProjectManager,
	shared.StateReady:      shared.RoleProjectManager,
	shared.StateInProgress: shared.RoleDeveloper,
	shared.StateReview:     shared.RoleReviewer,
	shared.StateTesting:    shared.RoleQA,
	shared.StateDone:       shared.RoleQA,
	shared.StateBlocked:    shared.RoleProjectManager,
	shared.StateCancelled:  shared.RoleProjectManager,
	shared.StateArchived:   shared.RoleProjectManager,
}

// CanTransition reports whether the transition is allowed by the
// canonical state machine. A nil return means allowed; otherwise the
// error names the violated rule.
func (Machine) CanTransition(from, to shared.TaskState) error {
	targets, known := allowedTransitions[from]
	if !known {
		return fmt.Errorf("%w: %q", ErrUnknownState, from)
	}
	if _, known := allowedTransitions[to]; !known {
		return fmt.Errorf("%w: %q", ErrUnknownState, to)
	}
	if !targets[to] {
		return fmt.Errorf("%w: %s -> %s (docs/architecture/state-machine.md)", ErrTransitionNotAllowed, from, to)
	}
	return nil
}

// NextRole returns the role responsible for acting on a task in the given
// state (docs/architecture/workflow.md, participation table).
func (Machine) NextRole(state shared.TaskState) (shared.Role, error) {
	role, known := stateRoles[state]
	if !known {
		return "", fmt.Errorf("%w: %q", ErrUnknownState, state)
	}
	return role, nil
}
