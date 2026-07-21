package execution

import (
	"errors"
	"testing"
)

func newQueued(t *testing.T) *Execution {
	t.Helper()
	e, event, err := New("exec-1", "task-1", "executor-1")
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}
	if event.ID != "exec-1" || event.TaskID != "task-1" || event.ExecutorID != "executor-1" {
		t.Fatalf("Queued event data mismatch: %+v", event)
	}
	return e
}

func newRunning(t *testing.T) *Execution {
	t.Helper()
	e := newQueued(t)
	if _, err := e.Accept(); err != nil {
		t.Fatalf("Accept() unexpected error: %v", err)
	}
	return e
}

// --- New / Structural Invariants ---

func TestNew_SetsQueuedStateAndFixedFields(t *testing.T) {
	e := newQueued(t)
	if e.State() != StateQueued {
		t.Errorf("State() = %v, want %v (Structural Invariant 3: status exists from creation)", e.State(), StateQueued)
	}
	if e.TaskID() != "task-1" || e.ExecutorID() != "executor-1" {
		t.Errorf("fixed references not set: taskID=%q executorID=%q", e.TaskID(), e.ExecutorID())
	}
	if e.ArtifactIDs() != nil {
		t.Errorf("ArtifactIDs() = %v, want nil at creation", e.ArtifactIDs())
	}
}

func TestNew_RequiresAllIdentityFields(t *testing.T) {
	cases := []struct {
		name             string
		id, task, execut string
	}{
		{"missing id", "", "task-1", "executor-1"},
		{"missing taskID", "exec-1", "", "executor-1"},
		{"missing executorID", "exec-1", "task-1", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := New(tc.id, tc.task, tc.execut)
			if !errors.Is(err, ErrMissingField) {
				t.Errorf("New() error = %v, want %v", err, ErrMissingField)
			}
		})
	}
}

// --- Accept / Behavioral Invariant 3 ---

func TestAccept_TransitionsQueuedToRunning(t *testing.T) {
	e := newQueued(t)
	event, err := e.Accept()
	if err != nil {
		t.Fatalf("Accept() unexpected error: %v", err)
	}
	if e.State() != StateRunning {
		t.Errorf("State() = %v, want %v", e.State(), StateRunning)
	}
	if event.ID != e.ID() {
		t.Errorf("Started.ID = %q, want %q", event.ID, e.ID())
	}
}

func TestAccept_RejectedWhenAlreadyRunning(t *testing.T) {
	e := newRunning(t)
	if _, err := e.Accept(); !errors.Is(err, ErrNotQueued) {
		t.Errorf("second Accept() error = %v, want %v", err, ErrNotQueued)
	}
}

func TestAccept_RejectedWhenTerminal(t *testing.T) {
	e := newQueued(t)
	if _, err := e.Abort(); err != nil {
		t.Fatalf("Abort() unexpected error: %v", err)
	}
	if _, err := e.Accept(); !errors.Is(err, ErrTerminal) {
		t.Errorf("Accept() after Abort error = %v, want %v", err, ErrTerminal)
	}
}

// --- RecordArtifact / Behavioral Invariant 4 ---

func TestRecordArtifact_OnlyWhileRunning(t *testing.T) {
	e := newQueued(t)
	if err := e.RecordArtifact("art-1"); !errors.Is(err, ErrNotRunning) {
		t.Errorf("RecordArtifact() in Queued error = %v, want %v (Behavioral Invariant 4: not before Running)", err, ErrNotRunning)
	}
	if _, err := e.Accept(); err != nil {
		t.Fatalf("Accept() unexpected error: %v", err)
	}
	if err := e.RecordArtifact("art-1"); err != nil {
		t.Fatalf("RecordArtifact() in Running unexpected error: %v", err)
	}
	if got := e.ArtifactIDs(); len(got) != 1 || got[0] != "art-1" {
		t.Errorf("ArtifactIDs() = %v, want [art-1]", got)
	}
}

func TestRecordArtifact_RequiresID(t *testing.T) {
	e := newRunning(t)
	if err := e.RecordArtifact(""); !errors.Is(err, ErrMissingField) {
		t.Errorf("RecordArtifact(\"\") error = %v, want %v", err, ErrMissingField)
	}
}

func TestRecordArtifact_RejectedAfterTerminal(t *testing.T) {
	e := newRunning(t)
	if err := e.RecordArtifact("art-1"); err != nil {
		t.Fatalf("RecordArtifact() unexpected error: %v", err)
	}
	if _, err := e.Succeed(); err != nil {
		t.Fatalf("Succeed() unexpected error: %v", err)
	}
	if err := e.RecordArtifact("art-2"); !errors.Is(err, ErrTerminal) {
		t.Errorf("RecordArtifact() after Succeed error = %v, want %v (Behavioral Invariant 1: set is final)", err, ErrTerminal)
	}
	if got := e.ArtifactIDs(); len(got) != 1 {
		t.Errorf("ArtifactIDs() = %v, want exactly the pre-terminal set", got)
	}
}

// --- Succeed / Fail ---

func TestSucceed_CarriesProducedArtifacts(t *testing.T) {
	e := newRunning(t)
	for _, id := range []string{"art-1", "art-2"} {
		if err := e.RecordArtifact(id); err != nil {
			t.Fatalf("RecordArtifact(%q) unexpected error: %v", id, err)
		}
	}
	event, err := e.Succeed()
	if err != nil {
		t.Fatalf("Succeed() unexpected error: %v", err)
	}
	if e.State() != StateSucceeded {
		t.Errorf("State() = %v, want %v", e.State(), StateSucceeded)
	}
	if len(event.ArtifactIDs) != 2 {
		t.Errorf("Succeeded.ArtifactIDs = %v, want 2 entries", event.ArtifactIDs)
	}
}

func TestSucceed_RejectedInQueued(t *testing.T) {
	e := newQueued(t)
	if _, err := e.Succeed(); !errors.Is(err, ErrNotRunning) {
		t.Errorf("Succeed() in Queued error = %v, want %v", err, ErrNotRunning)
	}
}

func TestFail_CarriesArtifactsProducedBeforeFailure(t *testing.T) {
	// Spec Examples: a failed test run still produced its TestReport.
	e := newRunning(t)
	if err := e.RecordArtifact("test-report-1"); err != nil {
		t.Fatalf("RecordArtifact() unexpected error: %v", err)
	}
	event, err := e.Fail()
	if err != nil {
		t.Fatalf("Fail() unexpected error: %v", err)
	}
	if e.State() != StateFailed {
		t.Errorf("State() = %v, want %v", e.State(), StateFailed)
	}
	if len(event.ArtifactIDs) != 1 || event.ArtifactIDs[0] != "test-report-1" {
		t.Errorf("Failed.ArtifactIDs = %v, want [test-report-1]", event.ArtifactIDs)
	}
}

func TestFail_RejectedInQueued(t *testing.T) {
	e := newQueued(t)
	if _, err := e.Fail(); !errors.Is(err, ErrNotRunning) {
		t.Errorf("Fail() in Queued error = %v, want %v", err, ErrNotRunning)
	}
}

// --- Abort ---

func TestAbort_DirectlyFromQueued(t *testing.T) {
	e := newQueued(t)
	event, err := e.Abort()
	if err != nil {
		t.Fatalf("Abort() unexpected error: %v", err)
	}
	if e.State() != StateAborted {
		t.Errorf("State() = %v, want %v", e.State(), StateAborted)
	}
	if event.From != StateQueued {
		t.Errorf("Aborted.From = %v, want %v (spec: direct Queued -> Aborted path)", event.From, StateQueued)
	}
}

func TestAbort_FromRunningKeepsArtifacts(t *testing.T) {
	e := newRunning(t)
	if err := e.RecordArtifact("art-1"); err != nil {
		t.Fatalf("RecordArtifact() unexpected error: %v", err)
	}
	event, err := e.Abort()
	if err != nil {
		t.Fatalf("Abort() unexpected error: %v", err)
	}
	if event.From != StateRunning {
		t.Errorf("Aborted.From = %v, want %v", event.From, StateRunning)
	}
	if got := e.ArtifactIDs(); len(got) != 1 {
		t.Errorf("ArtifactIDs() after Abort = %v, want the produced set kept (spec Commands: Abort never removes)", got)
	}
}

// --- Fail/Abort race: Behavioral Invariant 5 ---

func TestRace_FailThenAbort_FirstWins(t *testing.T) {
	e := newRunning(t)
	if _, err := e.Fail(); err != nil {
		t.Fatalf("Fail() unexpected error: %v", err)
	}
	if _, err := e.Abort(); !errors.Is(err, ErrTerminal) {
		t.Errorf("Abort() after Fail error = %v, want %v (Behavioral Invariant 5: first terminal command wins)", err, ErrTerminal)
	}
	if e.State() != StateFailed {
		t.Errorf("State() = %v, want %v (must not be overwritten)", e.State(), StateFailed)
	}
}

func TestRace_AbortThenFail_FirstWins(t *testing.T) {
	e := newRunning(t)
	if _, err := e.Abort(); err != nil {
		t.Fatalf("Abort() unexpected error: %v", err)
	}
	if _, err := e.Fail(); !errors.Is(err, ErrTerminal) {
		t.Errorf("Fail() after Abort error = %v, want %v", err, ErrTerminal)
	}
	if e.State() != StateAborted {
		t.Errorf("State() = %v, want %v", e.State(), StateAborted)
	}
}

// --- Terminal set immutability through returned slices ---

func TestArtifactIDs_ReturnsCopyNotAlias(t *testing.T) {
	e := newRunning(t)
	if err := e.RecordArtifact("art-1"); err != nil {
		t.Fatalf("RecordArtifact() unexpected error: %v", err)
	}
	event, err := e.Succeed()
	if err != nil {
		t.Fatalf("Succeed() unexpected error: %v", err)
	}
	event.ArtifactIDs[0] = "tampered"
	if got := e.ArtifactIDs(); got[0] != "art-1" {
		t.Errorf("mutating the event slice changed the entity: %v (Behavioral Invariant 1)", got)
	}
	external := e.ArtifactIDs()
	external[0] = "tampered-again"
	if got := e.ArtifactIDs(); got[0] != "art-1" {
		t.Errorf("mutating an accessor result changed the entity: %v", got)
	}
}
