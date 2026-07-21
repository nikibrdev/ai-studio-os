// Package container manages the lifecycle of one Execution's sandbox per
// ADR-006: one ephemeral Docker container per Execution, a working copy
// cloned from the task's branch at start and destroyed at the end,
// network access restricted to an explicit allowlist, and secrets
// (git token, AI-provider key) injected only as environment variables —
// never baked into an image, never written to a file that outlives the
// container.
//
// This package lives under agents/claude-code, not internal/
// infrastructure: docs/architecture/module-boundaries.md forbids agents/
// from importing Core internals (including internal/infrastructure) —
// an adapter may only depend on the published Executor contract, pkg/,
// and its own provider's SDK/CLI. Docker itself is reached by shelling
// out to the docker CLI (os/exec), not a client library — the same
// reasoning as ADR-017 (PostgreSQL migrations) and the GitHub adapter
// (net/http over a client library): a handful of commands does not
// justify a new dependency, and Docker CLI is already a hard
// requirement of ADR-006 regardless.
package container
