package application

import (
	"context"
	"errors"

	"ai-studio-os/internal/domain/artifact"
	"ai-studio-os/internal/domain/execution"
	"ai-studio-os/internal/domain/executor"
	"ai-studio-os/internal/domain/project"
	"ai-studio-os/internal/domain/task"
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
}

// TaskStore persists and retrieves Task aggregates.
type TaskStore interface {
	Get(ctx context.Context, id string) (*task.Task, error)
	Save(ctx context.Context, t *task.Task) error
}

// ExecutorStore persists and retrieves Executor aggregates.
type ExecutorStore interface {
	Get(ctx context.Context, id string) (*executor.Executor, error)
	Save(ctx context.Context, e *executor.Executor) error
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
