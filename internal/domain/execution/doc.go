// Package execution implements the Execution aggregate: a single, bounded
// run of one Executor performing work for one Task, producing Artifacts and
// carrying the execution status
// (docs/specifications/domain/execution.md, status: Утверждена).
//
// Per ADR-015 domain modules never depend on each other: Task, Executor and
// Artifact are referenced by string identifiers, never by importing their
// packages. Like internal/domain/artifact, this package exposes the entity
// with command methods returning domain events as values — no
// Commands/Queries interfaces until a concrete consumer exists.
package execution
