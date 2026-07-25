package application_test

import (
	"context"
	"errors"
	"testing"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/application/inmemory"
	"ai-studio-os/internal/domain/event"
	"ai-studio-os/internal/domain/executor"
	"ai-studio-os/internal/domain/shared"
)

func newExecutorService() (*application.ExecutorService, application.ExecutorStore, *inmemory.EventBus) {
	executors := inmemory.NewExecutorStore()
	bus := inmemory.NewEventBus()
	svc := &application.ExecutorService{Executors: executors, Events: bus}
	return svc, executors, bus
}

func TestRegisterExecutor_Success(t *testing.T) {
	ctx := context.Background()
	svc, _, bus := newExecutorService()

	e, err := svc.Register(ctx, application.RegisterExecutorParams{
		ID: "exec-1", Backend: "claude-code", Roles: []shared.Role{shared.RoleDeveloper}, Actor: "human:architect",
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if e.State() != executor.StateRegistered {
		t.Errorf("State() = %v, want %v", e.State(), executor.StateRegistered)
	}

	published := bus.Published()
	if len(published) != 1 || published[0].Type() != event.ExecutorRegistered {
		t.Fatalf("Published() = %v, want one %s event", published, event.ExecutorRegistered)
	}
	if published[0].SubjectID() != "exec-1" {
		t.Errorf("SubjectID() = %q, want exec-1", published[0].SubjectID())
	}
}

func TestRegisterExecutor_MissingFieldPropagatesDomainError(t *testing.T) {
	ctx := context.Background()
	svc, _, bus := newExecutorService()

	_, err := svc.Register(ctx, application.RegisterExecutorParams{ID: "", Backend: "claude-code", Roles: []shared.Role{shared.RoleDeveloper}})
	if !errors.Is(err, executor.ErrMissingField) {
		t.Fatalf("Register() error = %v, want %v", err, executor.ErrMissingField)
	}
	if len(bus.Published()) != 0 {
		t.Errorf("Published() = %v, want none on failure", bus.Published())
	}
}

func TestRegisterExecutor_NoRolesPropagatesDomainError(t *testing.T) {
	ctx := context.Background()
	svc, _, _ := newExecutorService()

	_, err := svc.Register(ctx, application.RegisterExecutorParams{ID: "exec-1", Backend: "claude-code"})
	if !errors.Is(err, executor.ErrNoRoles) {
		t.Fatalf("Register() error = %v, want %v", err, executor.ErrNoRoles)
	}
}

func TestActivateExecutor_Success(t *testing.T) {
	ctx := context.Background()
	svc, _, bus := newExecutorService()
	if _, err := svc.Register(ctx, application.RegisterExecutorParams{
		ID: "exec-1", Backend: "claude-code", Roles: []shared.Role{shared.RoleDeveloper},
	}); err != nil {
		t.Fatalf("Register: %v", err)
	}

	if err := svc.Activate(ctx, "exec-1", "human:architect"); err != nil {
		t.Fatalf("Activate: %v", err)
	}

	published := bus.Published()
	if len(published) != 2 || published[1].Type() != event.ExecutorActivated {
		t.Fatalf("Published() = %v, want second event %s", published, event.ExecutorActivated)
	}
}

func TestActivateExecutor_AlreadyActivePropagatesDomainError(t *testing.T) {
	ctx := context.Background()
	svc, _, _ := newExecutorService()
	if _, err := svc.Register(ctx, application.RegisterExecutorParams{
		ID: "exec-1", Backend: "claude-code", Roles: []shared.Role{shared.RoleDeveloper},
	}); err != nil {
		t.Fatalf("Register: %v", err)
	}
	if err := svc.Activate(ctx, "exec-1", ""); err != nil {
		t.Fatalf("first Activate: %v", err)
	}

	err := svc.Activate(ctx, "exec-1", "")
	if !errors.Is(err, executor.ErrAlreadyActive) {
		t.Fatalf("second Activate() error = %v, want %v", err, executor.ErrAlreadyActive)
	}
}

func TestActivateExecutor_RetiredPropagatesDomainError(t *testing.T) {
	ctx := context.Background()
	svc, executors, _ := newExecutorService()
	e, err := svc.Register(ctx, application.RegisterExecutorParams{
		ID: "exec-1", Backend: "claude-code", Roles: []shared.Role{shared.RoleDeveloper},
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if _, err := e.Retire(); err != nil {
		t.Fatalf("Retire: %v", err)
	}
	if err := executors.Save(ctx, e); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if err := svc.Activate(ctx, "exec-1", ""); !errors.Is(err, executor.ErrRetired) {
		t.Fatalf("Activate() error = %v, want %v", err, executor.ErrRetired)
	}
}

func TestActivateExecutor_NotFound(t *testing.T) {
	ctx := context.Background()
	svc, _, _ := newExecutorService()

	if err := svc.Activate(ctx, "unknown", ""); !errors.Is(err, application.ErrNotFound) {
		t.Fatalf("Activate() error = %v, want %v", err, application.ErrNotFound)
	}
}

func TestExecutorStore_List_ReturnsAllOrderedByID(t *testing.T) {
	ctx := context.Background()
	svc, executors, _ := newExecutorService()
	if _, err := svc.Register(ctx, application.RegisterExecutorParams{
		ID: "exec-b", Backend: "claude-code", Roles: []shared.Role{shared.RoleDeveloper},
	}); err != nil {
		t.Fatalf("Register b: %v", err)
	}
	if _, err := svc.Register(ctx, application.RegisterExecutorParams{
		ID: "exec-a", Backend: "claude-code", Roles: []shared.Role{shared.RoleDeveloper},
	}); err != nil {
		t.Fatalf("Register a: %v", err)
	}

	got, err := executors.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 2 || got[0].ID() != "exec-a" || got[1].ID() != "exec-b" {
		t.Errorf("List() = %v, want [exec-a exec-b]", got)
	}
}

func TestExecutorStore_List_EmptyIsNotError(t *testing.T) {
	_, executors, _ := newExecutorService()
	got, err := executors.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("List() = %v, want empty", got)
	}
}
