// Package executor implements the Executor aggregate: a registry entry for
// a technical backend (a human or a system) registered with the platform
// and capable of performing work on behalf of one or more Roles
// (docs/specifications/domain/executor.md, status: Утверждена).
//
// This domain entity is deliberately distinct from the platform adapter
// contract internal/platform.Executor (ADR-005): the contract describes
// HOW the core technically calls a backend (Accept/Artifacts/Status/
// Finish); this package describes WHO is registered, WHAT they can do and
// whether they are available right now. The two are linked by backend
// identity, not by one owning the other (ADR-015).
package executor
