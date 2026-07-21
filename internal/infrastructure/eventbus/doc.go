// Package eventbus provides the production implementation of
// platform.EventBus (ADR-002): a synchronous, in-process bus — the same
// delivery semantics as internal/application/inmemory.EventBus used in
// EPIC-004 tests — plus a durable journal of every published event in
// PostgreSQL (event_journal table, internal/infrastructure/postgres
// migration 0004).
package eventbus
