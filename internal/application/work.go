package application

import (
	"context"
	"errors"
	"time"

	"ai-studio-os/internal/domain/event"
	"ai-studio-os/internal/domain/execution"
	"ai-studio-os/internal/domain/shared"
	"ai-studio-os/internal/domain/workflow"
	"ai-studio-os/internal/platform"
)

// ErrExecutorNotAssignable is returned when the chosen Executor cannot
// take on new work: not Active, or lacking the Developer role required
// for this step of the golden path (docs/architecture/golden-path.md).
var ErrExecutorNotAssignable = errors.New("application: executor is not assignable")

// WorkService implements the "Запуск работы" step of the golden path:
// moving a Task Ready -> In Progress and spawning the Execution that
// carries out the work. Executor selection (ADR-007, Decision Required)
// is out of scope — the caller supplies an already-chosen Executor.
type WorkService struct {
	Tasks      TaskStore
	Executors  ExecutorStore
	Executions ExecutionStore
	Events     platform.EventBus
	Rules      workflow.Rules
}

// StartTaskParams are the inputs to StartTask.
type StartTaskParams struct {
	TaskID     string
	ExecutorID string
	Actor      string
}

// StartTask transitions a Task Ready -> In Progress and spawns an
// Execution for the given Executor, immediately confirmed as accepted
// (spec Execution Commands: Accept) — this MVP starts work synchronously;
// a real backend call arrives with the Executor adapter (v0.6, ADR-006).
// Publishes TaskStarted (source task), then ExecutionQueued and
// ExecutionStarted (source execution).
func (s *WorkService) StartTask(ctx context.Context, p StartTaskParams) (*execution.Execution, error) {
	t, err := s.Tasks.Get(ctx, p.TaskID)
	if err != nil {
		return nil, err
	}
	exec, err := s.Executors.Get(ctx, p.ExecutorID)
	if err != nil {
		return nil, err
	}
	if !exec.AvailableForAssignment() || !exec.HasRole(shared.RoleDeveloper) {
		return nil, ErrExecutorNotAssignable
	}

	transitioned, err := t.Transition(shared.StateInProgress, "", s.Rules)
	if err != nil {
		return nil, err
	}
	if err := s.Tasks.Save(ctx, t); err != nil {
		return nil, err
	}
	if err := s.publish(ctx, event.TaskStarted, "task", p.Actor, t.ProjectID(), t.ID(), transitioned.At); err != nil {
		return nil, err
	}

	run, queued, err := execution.New(NewID(), t.ID(), exec.ID())
	if err != nil {
		return nil, err
	}
	if err := s.publish(ctx, event.ExecutionQueued, "execution", p.Actor, t.ProjectID(), run.ID(), queued.At); err != nil {
		return nil, err
	}

	started, err := run.Accept()
	if err != nil {
		return nil, err
	}
	if err := s.Executions.Save(ctx, run); err != nil {
		return nil, err
	}
	if err := s.publish(ctx, event.ExecutionStarted, "execution", p.Actor, t.ProjectID(), run.ID(), started.At); err != nil {
		return nil, err
	}

	return run, nil
}

func (s *WorkService) publish(ctx context.Context, eventType, source, actor, projectID, subjectID string, at time.Time) error {
	return s.Events.Publish(ctx, NewEvent(eventType, source, actor, projectID, subjectID, at))
}
