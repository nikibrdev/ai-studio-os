package task

import (
	"errors"
	"testing"

	"ai-studio-os/internal/domain/shared"
)

// allowAll is a deterministic workflow.Rules stub permitting every
// transition; the real table lives in the workflow module (TASK-039) —
// the entity must work against the contract alone.
type allowAll struct{}

func (allowAll) CanTransition(_, _ shared.TaskState) error { return nil }
func (allowAll) NextRole(shared.TaskState) (shared.Role, error) {
	return shared.RoleDeveloper, nil
}

// denyAll is a stub rejecting every transition.
type denyAll struct{ err error }

func (d denyAll) CanTransition(_, _ shared.TaskState) error { return d.err }
func (denyAll) NextRole(shared.TaskState) (shared.Role, error) {
	return "", errors.New("no role")
}

func newBacklog(t *testing.T) *Task {
	t.Helper()
	task, event, err := New("task-1", "proj-1", "", "Название", "feature")
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}
	if event.ID != "task-1" || event.EpicID != "" {
		t.Fatalf("Created event data mismatch: %+v", event)
	}
	return task
}

// --- New / Structural Invariants ---

func TestNew_SetsBacklogStateAndFields(t *testing.T) {
	task := newBacklog(t)
	if task.State() != shared.StateBacklog {
		t.Errorf("State() = %v, want %v", task.State(), shared.StateBacklog)
	}
	if task.ProjectID() != "proj-1" {
		t.Errorf("ProjectID() = %q, want proj-1 (Structural Invariant 1)", task.ProjectID())
	}
	if task.EpicID() != "" {
		t.Errorf("EpicID() = %q, want empty (Structural Invariant 2: optional)", task.EpicID())
	}
}

func TestNew_EpicIsOptionalButOtherFieldsAreNot(t *testing.T) {
	if _, _, err := New("task-1", "proj-1", "epic-1", "Название", "feature"); err != nil {
		t.Errorf("New() with epic error = %v, want nil", err)
	}
	cases := []struct {
		name                          string
		id, project, epic, title, typ string
	}{
		{"missing id", "", "proj-1", "", "Название", "feature"},
		{"missing projectID", "task-1", "", "", "Название", "feature"},
		{"missing title", "task-1", "proj-1", "", "", "feature"},
		{"missing type", "task-1", "proj-1", "", "Название", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := New(tc.id, tc.project, tc.epic, tc.title, tc.typ); !errors.Is(err, ErrMissingField) {
				t.Errorf("New() error = %v, want %v", err, ErrMissingField)
			}
		})
	}
}

// --- Backlog-only edits ---

func TestSetScopeAndCriteria_InBacklog(t *testing.T) {
	task := newBacklog(t)
	if err := task.SetScope("Цель и объём"); err != nil {
		t.Fatalf("SetScope() unexpected error: %v", err)
	}
	if err := task.SetAcceptanceCriteria([]string{"критерий 1", "критерий 2"}); err != nil {
		t.Fatalf("SetAcceptanceCriteria() unexpected error: %v", err)
	}
	if task.Scope() != "Цель и объём" {
		t.Errorf("Scope() = %q", task.Scope())
	}
	if got := task.AcceptanceCriteria(); len(got) != 2 {
		t.Errorf("AcceptanceCriteria() = %v, want 2 entries", got)
	}
}

func TestEdits_RejectedOutsideBacklog(t *testing.T) {
	task := newBacklog(t)
	if _, err := task.Transition(shared.StateReady, "", allowAll{}); err != nil {
		t.Fatalf("Transition() unexpected error: %v", err)
	}
	if err := task.SetScope("поздно"); !errors.Is(err, ErrNotBacklog) {
		t.Errorf("SetScope() in Ready error = %v, want %v", err, ErrNotBacklog)
	}
	if err := task.SetAcceptanceCriteria([]string{"поздно"}); !errors.Is(err, ErrNotBacklog) {
		t.Errorf("SetAcceptanceCriteria() in Ready error = %v, want %v", err, ErrNotBacklog)
	}
	if err := task.AttachToEpic("epic-1"); !errors.Is(err, ErrNotBacklog) {
		t.Errorf("AttachToEpic() in Ready error = %v, want %v", err, ErrNotBacklog)
	}
}

func TestAttachToEpic_InBacklog(t *testing.T) {
	task := newBacklog(t)
	if err := task.AttachToEpic("epic-1"); err != nil {
		t.Fatalf("AttachToEpic() unexpected error: %v", err)
	}
	if task.EpicID() != "epic-1" {
		t.Errorf("EpicID() = %q, want epic-1", task.EpicID())
	}
}

// --- Transition: delegation to workflow.Rules (invariant 8) ---

func TestTransition_DelegatesLegalityToRules(t *testing.T) {
	task := newBacklog(t)
	event, err := task.Transition(shared.StateReady, "", allowAll{})
	if err != nil {
		t.Fatalf("Transition() unexpected error: %v", err)
	}
	if task.State() != shared.StateReady {
		t.Errorf("State() = %v, want %v", task.State(), shared.StateReady)
	}
	if event.From != shared.StateBacklog || event.To != shared.StateReady {
		t.Errorf("Transitioned = %+v, want Backlog->Ready", event)
	}
}

func TestTransition_RulesRejectionKeepsState(t *testing.T) {
	task := newBacklog(t)
	denied := errors.New("state machine: transition not allowed")
	if _, err := task.Transition(shared.StateDone, "", denyAll{err: denied}); !errors.Is(err, denied) {
		t.Errorf("Transition() error = %v, want the rules' own error passed through", err)
	}
	if task.State() != shared.StateBacklog {
		t.Errorf("State() = %v, want unchanged %v", task.State(), shared.StateBacklog)
	}
}

func TestTransition_NilRulesRejected(t *testing.T) {
	task := newBacklog(t)
	if _, err := task.Transition(shared.StateReady, "", nil); !errors.Is(err, ErrNilRules) {
		t.Errorf("Transition() with nil rules error = %v, want %v (invariant 8: task never decides itself)", err, ErrNilRules)
	}
}

// --- Reason requirement: invariant 3 ---

func TestTransition_BlockedAndCancelledRequireReason(t *testing.T) {
	for _, to := range []shared.TaskState{shared.StateBlocked, shared.StateCancelled} {
		task := newBacklog(t)
		if _, err := task.Transition(to, "", allowAll{}); !errors.Is(err, ErrReasonRequired) {
			t.Errorf("Transition(%v) without reason error = %v, want %v", to, err, ErrReasonRequired)
		}
		event, err := task.Transition(to, "причина зафиксирована", allowAll{})
		if err != nil {
			t.Fatalf("Transition(%v) with reason unexpected error: %v", to, err)
		}
		if event.Reason == "" {
			t.Errorf("Transitioned.Reason empty, want recorded reason")
		}
	}
}

func TestTransition_ReasonOptionalElsewhere(t *testing.T) {
	task := newBacklog(t)
	if _, err := task.Transition(shared.StateReady, "", allowAll{}); err != nil {
		t.Errorf("Transition(Ready) without reason error = %v, want nil", err)
	}
}

// --- Accessor isolation ---

func TestAcceptanceCriteria_ReturnsCopyNotAlias(t *testing.T) {
	task := newBacklog(t)
	src := []string{"критерий"}
	if err := task.SetAcceptanceCriteria(src); err != nil {
		t.Fatalf("SetAcceptanceCriteria() unexpected error: %v", err)
	}
	src[0] = "подменён снаружи"
	if got := task.AcceptanceCriteria(); got[0] != "критерий" {
		t.Errorf("mutating the source slice changed the entity: %v", got)
	}
	out := task.AcceptanceCriteria()
	out[0] = "подменён через аксессор"
	if got := task.AcceptanceCriteria(); got[0] != "критерий" {
		t.Errorf("mutating an accessor result changed the entity: %v", got)
	}
}
