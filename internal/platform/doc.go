// Package platform defines the infrastructure-agnostic contracts of the
// AI Studio OS runtime: event delivery, agent execution, tools, memory and
// git hosting.
//
// These are platform abstractions, not domain language — the domain
// vocabulary lives in internal/domain/shared. The package is deliberately
// domain-agnostic: it never imports domain packages (ADR-015).
//
// Everything outside internal/ (applications, agent adapters, tools,
// infrastructure adapters) depends on these contracts and never on module
// internals (docs/architecture/module-boundaries.md). The package contains
// interfaces, types and constants only — no logic and no implementations;
// implementations live in internal/infrastructure and the extension
// directories (agents/, tools/).
//
// Frozen architecture references:
//   - docs/architecture/interfaces.md — conceptual contracts (source of truth)
//   - docs/adr/ADR-002 — In-Memory Event Bus (interface stays stable)
//   - docs/adr/ADR-014 — everything goes through Core
//   - docs/adr/ADR-015 — internal layering (domain / application / platform /
//     infrastructure)
package platform
