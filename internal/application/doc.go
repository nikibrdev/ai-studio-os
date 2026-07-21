// Package application implements the use-cases of the Application Layer
// (v0.4, docs/roadmap/EPIC-004-application-layer.md): command scenarios
// composed over the completed Domain Layer, and the ports (ProjectStore,
// TaskStore, ExecutorStore, ExecutionStore, ArtifactStore) each use-case
// depends on instead of a concrete storage technology.
//
// Ports are declared here rather than in internal/platform because they
// carry domain types, and internal/platform is deliberately domain-agnostic
// (ADR-015) — see
// engineering/decisions/2026-07-21-application-ports-placement.md.
// Infrastructure implementations arrive in EPIC-005 (v0.5); this epic's
// tests use the deterministic fakes in internal/application/inmemory.
package application
