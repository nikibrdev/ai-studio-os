package inmemory

import (
	"ai-studio-os/internal/domain/artifact"
	"ai-studio-os/internal/domain/execution"
	"ai-studio-os/internal/domain/executor"
	"ai-studio-os/internal/domain/project"
	"ai-studio-os/internal/domain/task"
)

// NewProjectStore returns an application.ProjectStore fake.
func NewProjectStore() *Store[project.Project] {
	return NewStore(func(p *project.Project) string { return p.ID() })
}

// NewTaskStore returns an application.TaskStore fake.
func NewTaskStore() *Store[task.Task] {
	return NewStore(func(t *task.Task) string { return t.ID() })
}

// NewExecutorStore returns an application.ExecutorStore fake.
func NewExecutorStore() *Store[executor.Executor] {
	return NewStore(func(e *executor.Executor) string { return e.ID() })
}

// NewExecutionStore returns an application.ExecutionStore fake.
func NewExecutionStore() *Store[execution.Execution] {
	return NewStore(func(e *execution.Execution) string { return e.ID() })
}

// NewArtifactStore returns an application.ArtifactStore fake.
func NewArtifactStore() *Store[artifact.Artifact] {
	return NewStore(func(a *artifact.Artifact) string { return a.ID() })
}
