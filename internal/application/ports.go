package application

import (
	"context"
	"errors"

	"ai-studio-os/internal/domain/artifact"
	"ai-studio-os/internal/domain/execution"
	"ai-studio-os/internal/domain/executor"
	"ai-studio-os/internal/domain/project"
	"ai-studio-os/internal/domain/task"
	"ai-studio-os/internal/platform"
)

// ErrNotFound is returned by a Store's Get when no aggregate exists for
// the given identifier. Every Store implementation (fakes here, future
// infrastructure adapters in EPIC-005) must return this sentinel so
// use-cases can react uniformly regardless of the backing technology.
var ErrNotFound = errors.New("application: not found")

// ProjectStore persists and retrieves Project aggregates.
type ProjectStore interface {
	Get(ctx context.Context, id string) (*project.Project, error)
	Save(ctx context.Context, p *project.Project) error
	// List returns every Project, ordered by id for a deterministic
	// result — added in EPIC-009 (TASK-072) for apps/dashboard, which has
	// no way to show a list of projects otherwise.
	List(ctx context.Context) ([]*project.Project, error)
}

// TaskStore persists and retrieves Task aggregates. Get takes projectID
// because the public TASK-NNN identifier is unique only within a Project
// (ADR-011) — a bare id cannot disambiguate between two different
// projects' tasks (BUGFIX-003).
type TaskStore interface {
	Get(ctx context.Context, projectID, id string) (*task.Task, error)
	Save(ctx context.Context, t *task.Task) error
}

// ExecutorStore persists and retrieves Executor aggregates.
type ExecutorStore interface {
	Get(ctx context.Context, id string) (*executor.Executor, error)
	Save(ctx context.Context, e *executor.Executor) error
	// List returns every Executor, ordered by id for a deterministic
	// result — added in EPIC-010 (TASK-079) so a caller (apps/orchestrator)
	// can find an available Executor for a role without already knowing
	// its id, the same reason ProjectStore.List was added in EPIC-009.
	List(ctx context.Context) ([]*executor.Executor, error)
}

// ExecutionStore persists and retrieves Execution aggregates.
type ExecutionStore interface {
	Get(ctx context.Context, id string) (*execution.Execution, error)
	Save(ctx context.Context, e *execution.Execution) error
}

// ArtifactStore persists and retrieves Artifact aggregates.
type ArtifactStore interface {
	Get(ctx context.Context, id string) (*artifact.Artifact, error)
	Save(ctx context.Context, a *artifact.Artifact) error
}

// TaskIDGenerator issues the next public TASK-NNN identifier for a
// project (ADR-011) through a single race-free path — added in EPIC-008
// (TASK-065) because an external API caller cannot safely compute the
// next number itself.
type TaskIDGenerator interface {
	NextID(ctx context.Context, projectID string) (string, error)
}

// JournalEntry is one journaled event together with its position in the
// journal. Seq is the cursor value a poller stores to resume after this
// entry — returned explicitly because the caller, not the journal, owns
// the cursor.
type JournalEntry struct {
	Seq   int64
	Event platform.Event
}

// EventJournal reads previously published events back from durable
// storage, in the order they were written. Added in EPIC-010 (TASK-079):
// the production EventBus (internal/infrastructure/eventbus) delivers
// only to subscribers within its own process (ADR-002), so a separate
// process (apps/orchestrator) cannot use Subscribe — it polls Since with
// an advancing cursor instead. This port exists so apps/orchestrator
// depends only on internal/application, never directly on
// internal/infrastructure (module-boundaries.md: "apps/orchestrator" —
// прямой доступ к хранилищам запрещён).
//
// The cursor is a monotonic insertion sequence, not a timestamp
// (TASK-080, migration 0007). An event's OccurredAt is stamped by the
// domain before the row is written, so it does not order rows by the
// moment they became readable: under concurrency a row can appear with an
// OccurredAt earlier than one already read, and a time-based cursor would
// skip it silently. Sequence order is write order.
type EventJournal interface {
	// Since returns every entry with Seq strictly greater than afterSeq,
	// ordered by Seq. Pass 0 to read the journal from the beginning.
	Since(ctx context.Context, afterSeq int64) ([]JournalEntry, error)
}
