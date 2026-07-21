// Package wiring is the Infrastructure Layer's composition root
// (TASK-051, EPIC-005): it assembles the real adapters — PostgreSQL Store
// implementations, the production EventBus, and (when a GitHub token is
// available) the RepositoryProvider — into one System, applying pending
// migrations along the way. It does not start an HTTP server or any other
// delivery mechanism (that is v0.9, API); it exists so the same
// Application Layer services exercised on in-memory fakes in EPIC-004 can
// run on real infrastructure, in tests today and behind a future API.
package wiring
