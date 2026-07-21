package workflow

import (
	"errors"
	"testing"

	"ai-studio-os/internal/domain/shared"
)

// allStates are the nine canonical states (state-machine.md).
var allStates = []shared.TaskState{
	shared.StateBacklog,
	shared.StateReady,
	shared.StateInProgress,
	shared.StateReview,
	shared.StateTesting,
	shared.StateDone,
	shared.StateBlocked,
	shared.StateCancelled,
	shared.StateArchived,
}

// expected is the full transition table of state-machine.md, written out
// explicitly and independently of the implementation table so the test
// cross-checks the code against the document, not against itself.
var expected = map[shared.TaskState][]shared.TaskState{
	shared.StateBacklog:    {shared.StateReady, shared.StateCancelled},
	shared.StateReady:      {shared.StateBacklog, shared.StateInProgress, shared.StateBlocked, shared.StateCancelled},
	shared.StateInProgress: {shared.StateReview, shared.StateBlocked, shared.StateCancelled},
	shared.StateReview:     {shared.StateInProgress, shared.StateTesting, shared.StateBlocked},
	shared.StateTesting:    {shared.StateInProgress, shared.StateDone, shared.StateBlocked},
	shared.StateDone:       {shared.StateArchived},
	shared.StateBlocked:    {shared.StateReady, shared.StateInProgress, shared.StateCancelled},
	shared.StateCancelled:  {shared.StateArchived},
	shared.StateArchived:   {},
}

func isExpected(from, to shared.TaskState) bool {
	for _, t := range expected[from] {
		if t == to {
			return true
		}
	}
	return false
}

// TestCanTransition_Exhaustive checks every one of the 81 state pairs:
// exactly the 20 transitions of state-machine.md are allowed, everything
// else is rejected (invariant 1).
func TestCanTransition_Exhaustive(t *testing.T) {
	m := Machine{}
	allowed := 0
	for _, from := range allStates {
		for _, to := range allStates {
			err := m.CanTransition(from, to)
			if isExpected(from, to) {
				allowed++
				if err != nil {
					t.Errorf("CanTransition(%s, %s) = %v, want allowed", from, to, err)
				}
			} else if !errors.Is(err, ErrTransitionNotAllowed) {
				t.Errorf("CanTransition(%s, %s) = %v, want %v", from, to, err, ErrTransitionNotAllowed)
			}
		}
	}
	if allowed != 20 {
		t.Errorf("expected table lists %d allowed transitions, want 20 (state-machine.md)", allowed)
	}
}

func TestCanTransition_ArchivedIsTerminal(t *testing.T) {
	m := Machine{}
	for _, to := range allStates {
		if err := m.CanTransition(shared.StateArchived, to); !errors.Is(err, ErrTransitionNotAllowed) {
			t.Errorf("CanTransition(Archived, %s) = %v, want %v (invariant 7)", to, err, ErrTransitionNotAllowed)
		}
	}
}

func TestCanTransition_UnknownStates(t *testing.T) {
	m := Machine{}
	if err := m.CanTransition("nonsense", shared.StateReady); !errors.Is(err, ErrUnknownState) {
		t.Errorf("CanTransition(unknown, Ready) = %v, want %v", err, ErrUnknownState)
	}
	if err := m.CanTransition(shared.StateReady, "nonsense"); !errors.Is(err, ErrUnknownState) {
		t.Errorf("CanTransition(Ready, unknown) = %v, want %v", err, ErrUnknownState)
	}
}

// TestNextRole_ParticipationTable checks the full role table of
// docs/architecture/workflow.md.
func TestNextRole_ParticipationTable(t *testing.T) {
	m := Machine{}
	want := map[shared.TaskState]shared.Role{
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
	for state, role := range want {
		got, err := m.NextRole(state)
		if err != nil {
			t.Errorf("NextRole(%s) unexpected error: %v", state, err)
			continue
		}
		if got != role {
			t.Errorf("NextRole(%s) = %v, want %v (workflow.md participation table)", state, got, role)
		}
	}
}

func TestNextRole_UnknownState(t *testing.T) {
	m := Machine{}
	if _, err := m.NextRole("nonsense"); !errors.Is(err, ErrUnknownState) {
		t.Errorf("NextRole(unknown) = %v, want %v", err, ErrUnknownState)
	}
}

// TestMachine_SatisfiesTaskEntity closes the loop with the Task entity
// (TASK-037): the real Machine, not a stub, drives the golden path of the
// canonical lifecycle end to end.
func TestMachine_GoldenPathStates(t *testing.T) {
	m := Machine{}
	path := []shared.TaskState{
		shared.StateBacklog,
		shared.StateReady,
		shared.StateInProgress,
		shared.StateReview,
		shared.StateTesting,
		shared.StateDone,
		shared.StateArchived,
	}
	for i := 0; i < len(path)-1; i++ {
		if err := m.CanTransition(path[i], path[i+1]); err != nil {
			t.Errorf("golden path step %s -> %s rejected: %v", path[i], path[i+1], err)
		}
	}
}
