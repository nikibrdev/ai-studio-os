package application

import (
	"sort"

	"ai-studio-os/internal/domain/shared"
)

// DecisionKind names a decision a human has to make about a task.
//
// A checkpoint is a state the platform does not leave on its own
// (docs/architecture/workflow.md, "Контрольные точки человека") — not a
// separate state and not an extension of the state machine. These two are
// the ones ADR-007 names.
type DecisionKind string

const (
	// DecisionDefinitionOfReady is the acceptance of a task's Definition of
	// Ready: Backlog -> Ready.
	DecisionDefinitionOfReady DecisionKind = "definition-of-ready"

	// DecisionAcceptance is the final acceptance decision: Testing -> Done.
	// Saying yes merges the pull request (ADR-008), so it is not reversible
	// within the task.
	DecisionAcceptance DecisionKind = "acceptance"
)

// AwaitingDecision is one task waiting on a human, and which decision it
// waits for.
type AwaitingDecision struct {
	Task     TaskView
	Decision DecisionKind
}

// decisionForState maps a task state to the decision a human owes it, or
// false when the state is not a checkpoint.
//
// Deliberately an explicit mapping of the two states ADR-007 names, not
// something derived from workflow.Rules.NextRole ("states owned by PM and
// QA"): checkpoints are defined by naming them, so deriving the list from
// "whatever is not automated yet" would silently empty it once PM and QA are
// automated. Mirrors the normative table in docs/architecture/workflow.md.
//
// Lives here rather than in the domain because it answers a read-model
// question — who owes a decision — and because changing workflow.Rules would
// mean amending an approved specification.
func decisionForState(state shared.TaskState) (DecisionKind, bool) {
	switch state {
	case shared.StateBacklog:
		return DecisionDefinitionOfReady, true
	case shared.StateTesting:
		return DecisionAcceptance, true
	default:
		return "", false
	}
}

// DecisionFor reports the decision a human owes this task, or an empty kind
// when none is owed. Exported so a delivery layer can annotate a single task
// without re-deriving the rule — duplicating it in the UI is forbidden
// (docs/architecture/module-boundaries.md).
func DecisionFor(state shared.TaskState) DecisionKind {
	kind, _ := decisionForState(state)
	return kind
}

// ListAwaitingDecision returns every task waiting on a human, across all
// projects, ordered by project then task id for a deterministic result.
//
// Across projects on purpose: the question it answers is "what is waiting for
// me", not "what is waiting for me in this project". A per-project operation
// is not added — a project's own page already shows the same flag on each
// task through TaskView's decision annotation.
func (p *TaskProjection) ListAwaitingDecision() []AwaitingDecision {
	p.mu.Lock()
	defer p.mu.Unlock()

	var out []AwaitingDecision
	for _, v := range p.views {
		if kind, ok := decisionForState(v.State); ok {
			out = append(out, AwaitingDecision{Task: v, Decision: kind})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Task.ProjectID != out[j].Task.ProjectID {
			return out[i].Task.ProjectID < out[j].Task.ProjectID
		}
		return out[i].Task.ID < out[j].Task.ID
	})
	return out
}
