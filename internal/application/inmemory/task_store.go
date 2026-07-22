package inmemory

import (
	"context"
	"sync"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/domain/task"
)

// TaskStore is a dedicated application.TaskStore fake, keyed by the pair
// (ProjectID, ID) — unlike the other four aggregates (Store[T], a single
// string key), Task's public identifier is unique only within a Project
// (ADR-011), not globally, so a bare-id key would let two different
// projects' tasks collide (BUGFIX-003).
type TaskStore struct {
	mu    sync.Mutex
	items map[string]*task.Task // key: taskKey(projectID, id)
}

// NewTaskStore creates an empty TaskStore fake.
func NewTaskStore() *TaskStore {
	return &TaskStore{items: make(map[string]*task.Task)}
}

func taskKey(projectID, id string) string { return projectID + "\x00" + id }

// Get returns the Task for (projectID, id), or application.ErrNotFound.
func (s *TaskStore) Get(_ context.Context, projectID, id string) (*task.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.items[taskKey(projectID, id)]
	if !ok {
		return nil, application.ErrNotFound
	}
	return t, nil
}

// Save stores the Task under its own (ProjectID, ID) pair, overwriting
// any previous value for the same pair.
func (s *TaskStore) Save(_ context.Context, t *task.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[taskKey(t.ProjectID(), t.ID())] = t
	return nil
}
