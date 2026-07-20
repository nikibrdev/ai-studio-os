package platform

import "context"

// ExecutorTask is what an Executor is asked to do.
//
// Deliberately abstract: the full shape (what a task assignment carries, how
// context is delivered) is Domain Layer's job (v0.3, EPIC-003) — ADR-005
// fixes the four capabilities of the Executor contract, not the data shape.
type ExecutorTask any

// Artifact is the tangible output of work: a commit, a Pull Request, a
// Markdown document, an ADR, a test run, a diagram — anything the system
// produces as evidence of work done (docs/domain/ubiquitous-language.md).
//
// Deliberately abstract here: internal/platform is domain-agnostic
// (ADR-015) and cannot depend on internal/domain, while a fully-specified
// Artifact belongs to the domain's ubiquitous language. Where the concrete
// Artifact entity ultimately lives relative to this placeholder is an open
// question — ADR-005.
type Artifact any

// ExecutionStatus is a point-in-time status report from a running
// execution.
//
// Deliberately abstract until Domain Layer fixes the shape (ADR-005).
type ExecutionStatus any

// Executor is the contract every execution backend implements: a human,
// Claude Code, Codex, OpenHands, or any future backend. The platform core
// knows only this contract and never a concrete backend
// (docs/architecture/interfaces.md, VISION.md).
//
// ADR-005 fixes exactly four capabilities — nothing else is part of the
// contract:
//
//	Accept Task -> Produce Artifact -> Report Status -> Finish Execution
//
// Contract constraints (ADR-014):
//   - an Executor never accesses platform storage (Agent -> Database is
//     forbidden, using the pre-ADR-005 wording of the rule); all effects go
//     through tools and the platform process;
//   - an Executor must stay within the scope of its accepted task;
//   - an adapter contains no platform domain logic.
type Executor interface {
	// Accept begins work on the given task. Returns an error if the
	// Executor cannot take it on.
	Accept(ctx context.Context, task ExecutorTask) error

	// Artifacts returns the artifacts produced by the execution so far.
	Artifacts(ctx context.Context) ([]Artifact, error)

	// Status reports the current status of the execution.
	Status(ctx context.Context) (ExecutionStatus, error)

	// Finish concludes the execution.
	Finish(ctx context.Context) error
}
