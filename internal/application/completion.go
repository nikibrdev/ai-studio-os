package application

import (
	"context"
	"errors"
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

	// Views, when set, lets CompleteTesting find the pull request the platform
	// itself opened instead of requiring the caller to supply it (BUGFIX-009).
	// Optional so existing callers that pass the reference explicitly — and
	// tests built before this — keep working unchanged.
	Views *TaskProjection
}

// RequestReviewParams are the inputs to RequestReview. ProjectID is required
// because TASK-NNN is unique only within a Project (ADR-011, BUGFIX-003).
//
// Repository and PullRequestID carry the pull request being reviewed —
// docs/architecture/events.md requires ReviewRequested to carry a reference to
// it ("Данные: идентификатор задачи, ссылка на PR, автор изменений"), and
// nothing did until BUGFIX-009. Losing it made the acceptance decision
// impossible to perform: CompleteTesting needs it to merge, and no other part
// of the platform remembered it.
//
// Both may be empty — a review can be requested before a pull request exists
// (a human driving the task by hand), and the domain does not require one.
type RequestReviewParams struct {
	ProjectID     string
	TaskID        string
	Repository    string
	PullRequestID string
	Actor         string
}

// RequestReview transitions a Task In Progress -> Review. Publishes
// ReviewRequested (source: task, per docs/architecture/events.md) carrying the
// pull request reference, so whoever later makes the acceptance decision does
// not have to supply it.
func (s *CompletionService) RequestReview(ctx context.Context, p RequestReviewParams) error {
	t, err := s.Tasks.Get(ctx, p.ProjectID, p.TaskID)
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

	e := NewEvent(event.ReviewRequested, "task", p.Actor, t.ProjectID(), t.ID(), transitioned.At)
	if p.Repository != "" || p.PullRequestID != "" {
		e = e.WithData(map[string]string{
			dataKeyRepository:    p.Repository,
			dataKeyPullRequestID: p.PullRequestID,
		})
	}
	return s.Events.Publish(ctx, e)
}

// CompleteReview transitions a Task out of Review: to Testing if approved,
// back to In Progress if changes were requested. Publishes ReviewCompleted
// (source: git, per docs/architecture/events.md — the verdict originates
// from the pull request review, even though the task module performs the
// transition) with the target state attached via Envelope.WithData, so a
// subscriber (internal/application/projection.go) can tell the two
// outcomes apart without re-deriving them from anywhere else.
func (s *CompletionService) CompleteReview(ctx context.Context, projectID, taskID string, approved bool, actor string) error {
	t, err := s.Tasks.Get(ctx, projectID, taskID)
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
		WithData(map[string]string{dataKeyTo: string(to)})
	return s.Events.Publish(ctx, e)
}

// CompleteTestingParams are the inputs to CompleteTesting. ProjectID is
// required because TASK-NNN is unique only within a Project (ADR-011,
// BUGFIX-003).
//
// Repository and PullRequestID are optional (BUGFIX-009): when empty and
// Passed is true, the service looks the reference up in Views — the platform
// opened the pull request itself and recorded the reference on
// ReviewRequested, so a caller should not have to know it. Explicitly passed
// values win, which keeps existing callers and tests working and allows a
// human to merge a pull request the platform never opened.
type CompleteTestingParams struct {
	ProjectID     string
	TaskID        string
	Passed        bool
	Repository    string
	PullRequestID string
	Actor         string
}

// ErrPullRequestUnknown is returned when a positive acceptance decision needs
// a pull request to merge and neither the caller nor the projection knows one.
var ErrPullRequestUnknown = errors.New("application: no pull request is known for this task")

// CompleteTesting concludes the Testing stage. On failure: Testing -> In
// Progress, publishes TestsFailed. On success: publishes TestsPassed,
// merges the pull request, publishes MergeCompleted, and only then
// transitions Testing -> Done and publishes TaskCompleted — the exact
// order ADR-008 fixes. If the merge fails, the Task stays in Testing and
// TaskCompleted is never published: the merge is a code-level guard on
// Done, not just a documented expectation.
func (s *CompletionService) CompleteTesting(ctx context.Context, p CompleteTestingParams) error {
	t, err := s.Tasks.Get(ctx, p.ProjectID, p.TaskID)
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

	repo, prID, err := s.pullRequestFor(p)
	if err != nil {
		return err
	}

	if err := s.publish(ctx, event.TestsPassed, "execution", p.Actor, t.ProjectID(), t.ID(), time.Now()); err != nil {
		return err
	}
	if err := s.Repositories.MergePullRequest(ctx, repo, prID); err != nil {
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

// pullRequestFor resolves which pull request to merge: what the caller passed,
// otherwise what the platform recorded when review was requested (BUGFIX-009).
//
// Resolved before TestsPassed is published, not at the merge call: an
// unresolvable reference must abort the whole sequence, and publishing
// TestsPassed first would announce a passing test run for a task that then
// stays in Testing — ADR-008's order exists precisely so events describe what
// actually happened.
func (s *CompletionService) pullRequestFor(p CompleteTestingParams) (string, string, error) {
	repo, prID := p.Repository, p.PullRequestID
	if repo != "" && prID != "" {
		return repo, prID, nil
	}

	if s.Views != nil {
		if v, ok := s.Views.Get(p.ProjectID, p.TaskID); ok {
			if repo == "" {
				repo = v.Repository
			}
			if prID == "" {
				prID = v.PullRequestID
			}
		}
	}

	if repo == "" || prID == "" {
		return "", "", ErrPullRequestUnknown
	}
	return repo, prID, nil
}

func (s *CompletionService) publish(ctx context.Context, eventType, source, actor, projectID, subjectID string, at time.Time) error {
	return s.Events.Publish(ctx, NewEvent(eventType, source, actor, projectID, subjectID, at))
}
