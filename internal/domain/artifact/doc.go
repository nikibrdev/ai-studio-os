// Package artifact implements the Artifact aggregate: a durable engineering
// result the platform can create, store and reuse independently of the
// process that produced it (docs/specifications/domain/artifact.md, status:
// Reference).
//
// Artifact is the first Domain Layer package carrying real,
// invariant-enforcing logic rather than contracts alone (EPIC-003, stage 2).
// It intentionally does not expose Commands/Queries interfaces the way
// internal/domain/{task,project} do: those describe a write path deferred
// to a not-yet-decided persistence mechanism, which Artifact does not have
// a concrete consumer for yet.
package artifact
