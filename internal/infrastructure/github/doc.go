// Package github implements platform.RepositoryProvider against the
// GitHub REST API directly over net/http — no client library: stack.md
// does not list one, and the six operations of the contract do not
// justify a new dependency (ADR-017 applied the same reasoning to
// migrations).
package github
