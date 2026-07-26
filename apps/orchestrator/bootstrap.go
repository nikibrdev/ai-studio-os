package main

import (
	"context"
	"fmt"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/domain/executor"
	"ai-studio-os/internal/domain/shared"
)

// executorIDPrefix begins the identifier of every registry entry this
// orchestrator owns. The identifier is derived from the role rather than
// generated, for two reasons:
//
//   - it IS the idempotency mechanism — restarting must find the same entry,
//     not add another one;
//   - it makes the entry and the backend structurally consistent (TASK-086):
//     dispatch looks the entry up by the role it is dispatching, and the
//     backend factory is handed that same role, so "one entry selected, a
//     different backend run" is no longer expressible.
//
// For Developer this yields exactly the identifier EPIC-010 already used
// (executor-claude-code-developer), so existing installations need no
// migration.
const executorIDPrefix = "executor-claude-code-"

// executorBackend names the technical backend behind these entries (spec
// Executor Structural Invariant 1: fixed for the entry's lifetime). One
// adapter serves every role, differing only by prompt (ADR-007).
const executorBackend = "claude-code"

// executorIDForRole returns the identifier of the registry entry this
// orchestrator owns for the given role.
func executorIDForRole(role shared.Role) string {
	return executorIDPrefix + string(role)
}

// EnsureExecutor brings the registry to the state the dispatcher needs for
// one role — an Active Executor holding it, under this orchestrator's own
// identifier — and returns that identifier. Safe to call on every start:
//
//   - absent      -> Register + Activate;
//   - not Active  -> Activate (Registered or Disabled from a previous run);
//   - Active      -> nothing;
//   - Retired     -> an error, since Retired is terminal (domain rule) and
//     silently registering a replacement under a new id would leave two
//     entries and hide an operator decision to decommission this backend.
//
// One Executor per role is the accepted v1.0 limitation (EPIC-010 "Риски") —
// a trusted single-user installation, the same premise as ADR-012.
func EnsureExecutor(ctx context.Context, svc *application.ExecutorService, store application.ExecutorStore, role shared.Role) (string, error) {
	id := executorIDForRole(role)

	existing, err := findExecutor(ctx, store, id)
	if err != nil {
		return "", err
	}

	if existing == nil {
		if _, err := svc.Register(ctx, application.RegisterExecutorParams{
			ID:      id,
			Backend: executorBackend,
			Roles:   []shared.Role{role},
			Actor:   actor,
		}); err != nil {
			return "", fmt.Errorf("orchestrator: register %s executor: %w", role, err)
		}
		if err := svc.Activate(ctx, id, actor); err != nil {
			return "", fmt.Errorf("orchestrator: activate %s executor: %w", role, err)
		}
		return id, nil
	}

	switch existing.State() {
	case executor.StateActive:
		return id, nil
	case executor.StateRetired:
		return "", fmt.Errorf("orchestrator: %s executor %s is retired and cannot be reused", role, id)
	default:
		if err := svc.Activate(ctx, id, actor); err != nil {
			return "", fmt.Errorf("orchestrator: activate %s executor: %w", role, err)
		}
		return id, nil
	}
}

// findExecutor returns the registry entry under the given identifier, or nil
// when it does not exist yet. It reads through List rather than Get so a
// missing entry is an ordinary empty result instead of an error to classify.
func findExecutor(ctx context.Context, store application.ExecutorStore, id string) (*executor.Executor, error) {
	all, err := store.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("orchestrator: list executors: %w", err)
	}
	for _, e := range all {
		if e.ID() == id {
			return e, nil
		}
	}
	return nil, nil
}
