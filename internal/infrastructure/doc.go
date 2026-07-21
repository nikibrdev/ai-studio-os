// Package infrastructure implements the Infrastructure Layer (v0.5,
// docs/roadmap/EPIC-005-infrastructure-layer.md): adapters connecting the
// ports declared by internal/application and internal/platform to real
// technologies — PostgreSQL (internal/infrastructure/postgres), the
// production Event Bus (internal/infrastructure/eventbus, TASK-049) and the
// GitHub Repository Provider (internal/infrastructure/github, TASK-050).
//
// None of the adapters here change the contracts they implement; this
// layer only supplies real implementations for ports that, through
// EPIC-004, were satisfied by the in-memory fakes in
// internal/application/inmemory.
package infrastructure
