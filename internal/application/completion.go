package application

import (
	"context"
	"time"

	"ai-studio-os/internal/domain/event"
	"ai-studio-os/internal/domain/shared"
	"ai-studio-os/internal/domain/workflow"
	"ai-studio-os/internal/platform"
)

// CompletionService implements the "Завершение задачи" step of the golden
// path: Review, Testing, and the Done transition — including the merge
// order fixed by ADR-008 (git policies): TestsPassed -> MergeCompleted ->
// TaskCompleted, with the merge itself gating the Testing -> Done
// transition in code, not only in the ADR's text.
type CompletionService struct {
	Tasks        TaskStore
	Repositories platform.RepositoryProvider
	Events       platform.EventBus
	Rules        workflow.Rules
}

// RequestReview transitions a Task In Progress -> Review. Publishes
// ReviewRequested (source: task, per docs/architecture/events.md).
func (s *CompletionService) RequestReview(ctx context.Context, taskID, actor string) error {
	t, err := s.Tasks.Get(ctx, taskID)
	if err != nil {
		return err
	}
	transitioned, err := t.Transition(shared.StateReview, "", s.Rules)
	if err != nil {
		return err
	}
	if err := s.Tasks.Save(ctx, t); err != nil {
		return err
	}
	return s.publish(ctx, event.ReviewRequested, "task", actor, t.ProjectID(), t.ID(), transitioned.At)
}

// CompleteReview transitions a Task out of Review: to Testing if approved,
// back to In Progress if changes were requested. Publishes ReviewCompleted
// (source: git, per docs/architecture/events.md — the verdict originates
// from the pull request review, even though the task module performs the
// transition) with the target state attached via Envelope.WithData, so a
// subscriber (internal/application/projection.go) can tell the two
// outcomes apart without re-deriving them from anywhere else.
func (s *CompletionService) CompleteReview(ctx context.Context, taskID string, approved bool, actor string) error {
	t, err := s.Tasks.Get(ctx, taskID)
	if err != nil {
		return err
	}
	to := shared.StateInProgress
	if approved {
		to = shared.StateTesting
	}
	transitioned, err := t.Transition(to, "", s.Rules)
	if err != nil {
		return err
	}
	if err := s.Tasks.Save(ctx, t); err != nil {
		return err
	}
	e := NewEvent(event.ReviewCompleted, "git", actor, t.ProjectID(), t.ID(), transitioned.At).
		WithData(map[string]string{"to": string(to)})
	return s.Events.Publish(ctx, e)
}

// CompleteTestingParams are the inputs to CompleteTesting. Repository and
// PullRequestID are required only when Passed is true (the merge call
// needs them); a real Task does not yet carry a git reference of its own
// (the domain git module is outside EPIC-003/004 scope) — the caller
// (Application-adjacent orchestration, later a real Executor adapter)
// supplies them explicitly, the same way TASK-042 accepts an
// already-chosen Executor.
type CompleteTestingParams struct {
	TaskID        string
	Passed        bool
	Repository    string
	PullRequestID string
	Actor         string
}

// CompleteTesting concludes the Testing stage. On failure: Testing -> In
// Progress, publishes TestsFailed. On success: publishes TestsPassed,
// merges the pull request, publishes MergeCompleted, and only then
// transitions Testing -> Done and publishes TaskCompleted — the exact
// order ADR-008 fixes. If the merge fails, the Task stays in Testing and
// TaskCompleted is never published: the merge is a code-level guard on
// Done, not just a documented expectation.
func (s *CompletionService) CompleteTesting(ctx context.Context, p CompleteTestingParams) error {
	t, err := s.Tasks.Get(ctx, p.TaskID)
	if err != nil {
		return err
	}

	if !p.Passed {
		transitioned, err := t.Transition(shared.StateInProgress, "", s.Rules)
		if err != nil {
			return err
		}
		if err := s.Tasks.Save(ctx, t); err != nil {
			return err
		}
		return s.publish(ctx, event.TestsFailed, "execution", p.Actor, t.ProjectID(), t.ID(), transitioned.At)
	}

	if err := s.publish(ctx, event.TestsPassed, "execution", p.Actor, t.ProjectID(), t.ID(), time.Now()); err != nil {
		return err
	}
	if err := s.Repositories.MergePullRequest(ctx, p.Repository, p.PullRequestID); err != nil {
		return err
	}
	if err := s.publish(ctx, event.MergeCompleted, "git", p.Actor, t.ProjectID(), t.ID(), time.Now()); err != nil {
		return err
	}

	transitioned, err := t.Transition(shared.StateDone, "", s.Rules)
	if err != nil {
		return err
	}
	if err := s.Tasks.Save(ctx, t); err != nil {
		return err
	}
	return s.publish(ctx, event.TaskCompleted, "task", p.Actor, t.ProjectID(), t.ID(), transitioned.At)
}

func (s *CompletionService) publish(ctx context.Context, eventType, source, actor, projectID, subjectID string, at time.Time) error {
	return s.Events.Publish(ctx, NewEvent(eventType, source, actor, projectID, subjectID, at))
}
