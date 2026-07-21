// Package inmemory provides deterministic, non-persistent fakes of the
// internal/application Store ports, for this epic's tests only
// (docs/roadmap/EPIC-004-application-layer.md). These are test doubles,
// not an infrastructure adapter: nothing survives past process memory, and
// concurrency safety only needs to cover sequential test use. Real,
// persistence-backed implementations arrive in EPIC-005 (v0.5).
package inmemory
