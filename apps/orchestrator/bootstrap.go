package main

import (
	"context"
	"fmt"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/domain/executor"
	"ai-studio-os/internal/domain/shared"
)

// developerExecutorID is the fixed identifier of the single Developer
// Executor this orchestrator drives. Fixed rather than generated because
// the identifier IS the idempotency mechanism: restarting the process must
// find the same registry entry, not add another one. One Executor per role
// is the accepted v1.0 limitation (EPIC-010 "Риски") — a trusted
// single-user installation, the same premise as ADR-012.
const developerExecutorID = "executor-claude-code-developer"

// developerExecutorBackend names the technical backend behind that entry
// (spec Executor Structural Invariant 1: fixed for the entry's lifetime).
const developerExecutorBackend = "claude-code"

// EnsureDeveloperExecutor brings the registry to the state the dispatcher
// needs — exactly one Active Executor holding the Developer role — and
// returns its id. Safe to call on every start:
//
//   - absent      -> Register + Activate;
//   - not Active  -> Activate (Registered or Disabled from a previous run);
//   - Active      -> nothing;
//   - Retired     -> an error, since Retired is terminal (domain rule) and
//     silently registering a replacement under a new id would leave two
//     entries and hide an operator decision to decommission this backend.
func EnsureDeveloperExecutor(ctx context.Context, svc *application.ExecutorService, store application.ExecutorStore) (string, error) {
	existing, err := findDeveloperExecutor(ctx, store)
	if err != nil {
		return "", err
	}

	if existing == nil {
		if _, err := svc.Register(ctx, application.RegisterExecutorParams{
			ID:      developerExecutorID,
			Backend: developerExecutorBackend,
			Roles:   []shared.Role{shared.RoleDeveloper},
			Actor:   actor,
		}); err != nil {
			return "", fmt.Errorf("orchestrator: register developer executor: %w", err)
		}
		if err := svc.Activate(ctx, developerExecutorID, actor); err != nil {
			return "", fmt.Errorf("orchestrator: activate developer executor: %w", err)
		}
		return developerExecutorID, nil
	}

	switch existing.State() {
	case executor.StateActive:
		return developerExecutorID, nil
	case executor.StateRetired:
		return "", fmt.Errorf("orchestrator: developer executor %s is retired and cannot be reused", developerExecutorID)
	default:
		if err := svc.Activate(ctx, developerExecutorID, actor); err != nil {
			return "", fmt.Errorf("orchestrator: activate developer executor: %w", err)
		}
		return developerExecutorID, nil
	}
}

// findDeveloperExecutor returns the registry entry under
// developerExecutorID, or nil when it does not exist yet. It reads through
// List rather than Get so a missing entry is an ordinary empty result
// instead of an error to classify.
func findDeveloperExecutor(ctx context.Context, store application.ExecutorStore) (*executor.Executor, error) {
	all, err := store.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("orchestrator: list executors: %w", err)
	}
	for _, e := range all {
		if e.ID() == developerExecutorID {
			return e, nil
		}
	}
	return nil, nil
}
