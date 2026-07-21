package platform

import "context"

// ExecutorTask is the work assignment handed to an Executor's Accept
// method — enough identity, planning content and git coordinates for a
// backend to start work.
//
// A concrete struct with primitive fields, not a domain type:
// internal/platform stays domain-agnostic (ADR-015) and never imports
// internal/domain. internal/application constructs values of this type
// from the real Task/Project aggregates; agents/ adapters receive it
// as-is and never need to import internal/domain themselves (ADR-005:
// the contract's data shape follows Domain Layer's shape once it exists,
// without depending on the domain packages).
type ExecutorTask struct {
	// TaskID is the identifier of the task being assigned.
	TaskID string

	// ProjectID is the identifier of the owning project.
	ProjectID string

	// Role is the platform role this assignment executes (e.g.
	// "developer" — docs/domain/ubiquitous-language.md).
	Role string

	// Title, Type, Scope and AcceptanceCriteria are the task's planning
	// content (state-machine.md: Definition of Ready fields).
	Title              string
	Type               string
	Scope              string
	AcceptanceCriteria []string

	// Repository and Branch are the git coordinates of the work: the
	// Executor clones Branch from Repository into its own working copy
	// and destroys it when the Execution ends (ADR-006 — one ephemeral
	// working copy per Execution).
	Repository string
	Branch     string
}

// Artifact is what an Executor reports having produced: a commit, a Pull
// Request, a document, a test run — anything the system produces as
// evidence of work done (docs/domain/ubiquitous-language.md).
//
// A concrete struct with primitive fields, not a domain type — see
// ExecutorTask's doc comment for why. Type/Origin/Author mirror the
// vocabulary domain/artifact.Artifact settled on (ADR-016) without this
// package importing that package.
type Artifact struct {
	ID      string
	Type    string
	Origin  string
	Author  string
	Payload []byte
}

// ExecutionStatus is a point-in-time status report from a running
// execution.
type ExecutionStatus struct {
	// State is the Executor's own read of its progress (e.g. "running",
	// "succeeded", "failed") — informational only: the authoritative
	// Execution state lives in domain/execution and is set by whichever
	// application service calls Finish, not derived automatically from
	// this field.
	State string

	// Message is a human-readable detail, empty when there is nothing to
	// report beyond State.
	Message string
}

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
