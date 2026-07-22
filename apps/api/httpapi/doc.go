// Package httpapi implements the REST delivery layer described in
// docs/api/*.md (ADR-003, EPIC-008). apps/api's only dependency is
// internal/application (module-boundaries.md: no direct access to
// internal/domain business logic or internal/infrastructure storage) —
// every handler parses a request, calls exactly one use-case method, and
// serializes the result or error. The narrow exception is referencing
// already-public domain sentinel error values (errors.go) purely for
// identity comparison in the HTTP status mapping: internal/application's
// use-case methods return these values unwrapped to their own callers by
// design (EPIC-004 chose not to re-wrap them), so any caller that needs
// to distinguish error kinds already has to know them — this is not an
// invocation of domain business logic.
package httpapi
