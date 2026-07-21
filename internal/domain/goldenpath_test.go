// Package domain_test verifies the Domain Layer as a whole: the golden
// path of docs/architecture/golden-path.md walked end to end through the
// real entities and the real canonical state machine, composed the way the
// Application Layer (v0.4) will compose them. Domain modules never import
// each other (ADR-015); this cross-module composition lives in a
// layer-level external test package, not inside any module.
package domain_test

import (
	"testing"

	"ai-studio-os/internal/domain/artifact"
	"ai-studio-os/internal/domain/execution"
	"ai-studio-os/internal/domain/executor"
	"ai-studio-os/internal/domain/project"
	"ai-studio-os/internal/domain/shared"
	"ai-studio-os/internal/domain/task"
	"ai-studio-os/internal/domain/workflow"
)

// TestGoldenPath drives one task through the full canonical lifecycle
// (Backlog -> ... -> Archived, including a Blocked detour), spawning an
// Execution and producing a published Artifact along the way — the v0.3
// result stated in ROADMAP.md: "доменная логика (в т.ч. state machine
// задачи) работает и покрыта тестами, без внешних зависимостей".
func TestGoldenPath(t *testing.T) {
	rules := workflow.Machine{}

	// Project: created -> repository connected -> explicitly activated.
	proj, _, err := project.New("proj-1", "AI Studio OS")
	if err != nil {
		t.Fatalf("project.New: %v", err)
	}
	if _, _, err := proj.ConnectRepository("github.com/nikibrdev/ai-studio-os"); err != nil {
		t.Fatalf("ConnectRepository: %v", err)
	}
	if _, err := proj.Activate(); err != nil {
		t.Fatalf("project.Activate: %v", err)
	}
	if !proj.AcceptsNewContent() {
		t.Fatal("active project must accept new content (Behavioral Invariant 4)")
	}

	// Executor: registered with the Developer role, activated.
	exec, _, err := executor.New("executor-1", "claude-code-instance", []shared.Role{shared.RoleDeveloper})
	if err != nil {
		t.Fatalf("executor.New: %v", err)
	}
	if _, err := exec.Activate(); err != nil {
		t.Fatalf("executor.Activate: %v", err)
	}
	if !exec.AvailableForAssignment() || !exec.HasRole(shared.RoleDeveloper) {
		t.Fatal("active executor with the developer role must be assignable")
	}

	// Task: created in the active project's boundary, prepared in Backlog.
	tsk, _, err := task.New("task-1", proj.ID(), "", "Реализовать golden path", "feature")
	if err != nil {
		t.Fatalf("task.New: %v", err)
	}
	if err := tsk.SetScope("Сквозной сценарий доменного слоя"); err != nil {
		t.Fatalf("SetScope: %v", err)
	}
	if err := tsk.SetAcceptanceCriteria([]string{"задача проходит все девять состояний"}); err != nil {
		t.Fatalf("SetAcceptanceCriteria: %v", err)
	}

	mustTransition := func(to shared.TaskState, reason string) {
		t.Helper()
		if _, err := tsk.Transition(to, reason, rules); err != nil {
			t.Fatalf("Transition(%s -> %s): %v", tsk.State(), to, err)
		}
	}

	mustTransition(shared.StateReady, "")
	mustTransition(shared.StateInProgress, "")

	// The developer hits an external blocker and returns (Blocked detour).
	mustTransition(shared.StateBlocked, "ожидает внешнего решения")
	mustTransition(shared.StateInProgress, "")

	// Starting work spawns an Execution for the assigned Executor.
	run, _, err := execution.New("exec-run-1", tsk.ID(), exec.ID())
	if err != nil {
		t.Fatalf("execution.New: %v", err)
	}
	if _, err := run.Accept(); err != nil {
		t.Fatalf("execution.Accept: %v", err)
	}

	// The work produces an Artifact: drafted, filled, published.
	art, _, err := artifact.New("art-1", proj.ID(), artifact.Type("PullRequest"), artifact.OriginProduced, artifact.Author("developer"), run.ID())
	if err != nil {
		t.Fatalf("artifact.New: %v", err)
	}
	if err := art.UpdateDraft([]byte("diff --git ..."), ""); err != nil {
		t.Fatalf("UpdateDraft: %v", err)
	}
	if err := run.RecordArtifact(art.ID()); err != nil {
		t.Fatalf("RecordArtifact: %v", err)
	}
	published, err := art.Publish()
	if err != nil {
		t.Fatalf("artifact.Publish: %v", err)
	}
	if published.ProducedBy != run.ID() {
		t.Errorf("ArtifactPublished.ProducedBy = %q, want %q", published.ProducedBy, run.ID())
	}

	// The execution finishes successfully, carrying the produced artifact.
	succeeded, err := run.Succeed()
	if err != nil {
		t.Fatalf("execution.Succeed: %v", err)
	}
	if len(succeeded.ArtifactIDs) != 1 || succeeded.ArtifactIDs[0] != art.ID() {
		t.Errorf("ExecutionSucceeded.ArtifactIDs = %v, want [%s]", succeeded.ArtifactIDs, art.ID())
	}

	// The task completes its canonical path.
	mustTransition(shared.StateReview, "")
	mustTransition(shared.StateTesting, "")
	mustTransition(shared.StateDone, "")
	mustTransition(shared.StateArchived, "")

	if tsk.State() != shared.StateArchived {
		t.Errorf("task final state = %v, want %v", tsk.State(), shared.StateArchived)
	}
}

// TestGoldenPath_CancelledBranch covers the two canonical states the
// golden path itself never visits: Cancelled and its archival.
func TestGoldenPath_CancelledBranch(t *testing.T) {
	rules := workflow.Machine{}
	tsk, _, err := task.New("task-2", "proj-1", "", "Устаревшая задача", "chore")
	if err != nil {
		t.Fatalf("task.New: %v", err)
	}
	if _, err := tsk.Transition(shared.StateCancelled, "требования устарели", rules); err != nil {
		t.Fatalf("Transition(Cancelled): %v", err)
	}
	if _, err := tsk.Transition(shared.StateArchived, "", rules); err != nil {
		t.Fatalf("Transition(Archived): %v", err)
	}
	if tsk.State() != shared.StateArchived {
		t.Errorf("final state = %v, want %v", tsk.State(), shared.StateArchived)
	}
}
