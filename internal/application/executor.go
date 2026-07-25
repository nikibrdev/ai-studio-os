package application

import (
	"context"
	"time"

	"ai-studio-os/internal/domain/event"
	"ai-studio-os/internal/domain/executor"
	"ai-studio-os/internal/domain/shared"
	"ai-studio-os/internal/platform"
)

// ExecutorService manages the Executor registry lifecycle up to Active —
// nothing exposed a way to reach that state from Application Layer until
// this service (TASK-079, EPIC-010): internal/domain/executor is fully
// implemented since EPIC-003, but tests and fakes created an Executor
// directly through the domain package, bypassing Application Layer, the
// same gap ProjectService (TASK-064) closed for Project.
type ExecutorService struct {
	Executors ExecutorStore
	Events    platform.EventBus
}

// RegisterExecutorParams are the inputs to Register.
type RegisterExecutorParams struct {
	ID      string
	Backend string
	Roles   []shared.Role
	Actor   string
}

// Register creates an Executor in the Registered state. Publishes
// ExecutorRegistered. Executor is not owned by a Project (spec: no
// project reference on the aggregate, docs/architecture/events.md lists
// no project field for its events either) — the event's ProjectID is
// left empty, the same way Actor may be empty for a system-initiated
// fact (Envelope's own contract).
func (s *ExecutorService) Register(ctx context.Context, p RegisterExecutorParams) (*executor.Executor, error) {
	e, registered, err := executor.New(p.ID, p.Backend, p.Roles)
	if err != nil {
		return nil, err
	}
	if err := s.Executors.Save(ctx, e); err != nil {
		return nil, err
	}
	if err := s.publish(ctx, event.ExecutorRegistered, p.Actor, p.ID, registered.At); err != nil {
		return nil, err
	}
	return e, nil
}

// Activate transitions an Executor Registered -> Active or Disabled ->
// Active (guard: not Retired, enforced entirely by the domain). Publishes
// ExecutorActivated.
func (s *ExecutorService) Activate(ctx context.Context, id, actor string) error {
	e, err := s.Executors.Get(ctx, id)
	if err != nil {
		return err
	}
	activated, err := e.Activate()
	if err != nil {
		return err
	}
	if err := s.Executors.Save(ctx, e); err != nil {
		return err
	}
	return s.publish(ctx, event.ExecutorActivated, actor, id, activated.At)
}

func (s *ExecutorService) publish(ctx context.Context, eventType, actor, executorID string, at time.Time) error {
	return s.Events.Publish(ctx, NewEvent(eventType, "executor", actor, "", executorID, at))
}
