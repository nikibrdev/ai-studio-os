package main

import (
	"context"
	"testing"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/application/inmemory"
	"ai-studio-os/internal/domain/event"
	"ai-studio-os/internal/domain/executor"
	"ai-studio-os/internal/domain/shared"
)

// developerID is the identifier this orchestrator owns for the Developer
// role — derived from the role, not a separate constant, which is what keeps
// the registry entry and the launched backend consistent (TASK-086).
var developerID = executorIDForRole(shared.RoleDeveloper)

func newBootstrapDeps() (*application.ExecutorService, application.ExecutorStore, *inmemory.EventBus) {
	store := inmemory.NewExecutorStore()
	bus := inmemory.NewEventBus()
	return &application.ExecutorService{Executors: store, Events: bus}, store, bus
}

func TestEnsureExecutor_RegistersAndActivatesWhenAbsent(t *testing.T) {
	ctx := context.Background()
	svc, store, bus := newBootstrapDeps()

	id, err := EnsureExecutor(ctx, svc, store, shared.RoleDeveloper)
	if err != nil {
		t.Fatalf("EnsureExecutor: %v", err)
	}
	if id != developerID {
		t.Errorf("id = %q, want %q", id, developerID)
	}

	e, err := store.Get(ctx, developerID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if e.State() != executor.StateActive {
		t.Errorf("State() = %v, want %v", e.State(), executor.StateActive)
	}
	if !e.HasRole(shared.RoleDeveloper) {
		t.Error("HasRole(Developer) = false, want true")
	}
	if e.Backend() != executorBackend {
		t.Errorf("Backend() = %q, want %q", e.Backend(), executorBackend)
	}

	published := bus.Published()
	if len(published) != 2 || published[0].Type() != event.ExecutorRegistered || published[1].Type() != event.ExecutorActivated {
		t.Errorf("Published() = %v, want ExecutorRegistered then ExecutorActivated", published)
	}
}

// Restarting the process must not add a second registry entry — the fixed
// identifier is the whole idempotency mechanism.
func TestEnsureExecutor_IsIdempotentAcrossRestarts(t *testing.T) {
	ctx := context.Background()
	svc, store, bus := newBootstrapDeps()

	if _, err := EnsureExecutor(ctx, svc, store, shared.RoleDeveloper); err != nil {
		t.Fatalf("first call: %v", err)
	}
	publishedAfterFirst := len(bus.Published())

	if _, err := EnsureExecutor(ctx, svc, store, shared.RoleDeveloper); err != nil {
		t.Fatalf("second call: %v", err)
	}

	all, err := store.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(all) != 1 {
		t.Errorf("List() returned %d executors, want exactly 1", len(all))
	}
	if got := len(bus.Published()); got != publishedAfterFirst {
		t.Errorf("Published() grew to %d on the second call, want it to stay at %d", got, publishedAfterFirst)
	}
}

func TestEnsureExecutor_ActivatesExistingDisabledEntry(t *testing.T) {
	ctx := context.Background()
	svc, store, _ := newBootstrapDeps()

	e, _, err := executor.New(developerID, executorBackend, []shared.Role{shared.RoleDeveloper})
	if err != nil {
		t.Fatalf("executor.New: %v", err)
	}
	if _, err := e.Activate(); err != nil {
		t.Fatalf("Activate: %v", err)
	}
	if _, err := e.Disable(); err != nil {
		t.Fatalf("Disable: %v", err)
	}
	if err := store.Save(ctx, e); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if _, err := EnsureExecutor(ctx, svc, store, shared.RoleDeveloper); err != nil {
		t.Fatalf("EnsureExecutor: %v", err)
	}

	got, err := store.Get(ctx, developerID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.State() != executor.StateActive {
		t.Errorf("State() = %v, want %v", got.State(), executor.StateActive)
	}
}

// Retired is terminal in the domain. Quietly registering a replacement
// under a different id would leave two entries and override an operator's
// decision to decommission this backend — so this reports an error instead.
func TestEnsureExecutor_RefusesRetiredEntry(t *testing.T) {
	ctx := context.Background()
	svc, store, _ := newBootstrapDeps()

	e, _, err := executor.New(developerID, executorBackend, []shared.Role{shared.RoleDeveloper})
	if err != nil {
		t.Fatalf("executor.New: %v", err)
	}
	if _, err := e.Retire(); err != nil {
		t.Fatalf("Retire: %v", err)
	}
	if err := store.Save(ctx, e); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if _, err := EnsureExecutor(ctx, svc, store, shared.RoleDeveloper); err == nil {
		t.Fatal("EnsureDeveloperExecutor() error = nil, want an error for a retired executor")
	}
}

// Another project's executors must not be mistaken for this one: the lookup
// matches on the fixed id, not on role or backend.
func TestEnsureExecutor_IgnoresUnrelatedExecutors(t *testing.T) {
	ctx := context.Background()
	svc, store, _ := newBootstrapDeps()

	other, _, err := executor.New("some-other-executor", "codex", []shared.Role{shared.RoleDeveloper})
	if err != nil {
		t.Fatalf("executor.New: %v", err)
	}
	if err := store.Save(ctx, other); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if _, err := EnsureExecutor(ctx, svc, store, shared.RoleDeveloper); err != nil {
		t.Fatalf("EnsureExecutor: %v", err)
	}

	all, err := store.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("List() returned %d executors, want 2 (the unrelated one plus ours)", len(all))
	}
	ours, err := store.Get(ctx, developerID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if ours.State() != executor.StateActive {
		t.Errorf("State() = %v, want %v", ours.State(), executor.StateActive)
	}
}
