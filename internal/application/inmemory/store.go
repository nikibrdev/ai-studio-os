package inmemory

import (
	"context"
	"sort"
	"sync"

	"ai-studio-os/internal/application"
)

// Store is a generic Get/Save fake backing every application.*Store port:
// the five aggregates (Project, Task, Executor, Execution, Artifact) need
// an identical shape of fake, differing only in the Go type and how its
// identifier is read — a single generic implementation replaces five
// near-duplicate ones.
type Store[T any] struct {
	mu    sync.Mutex
	items map[string]*T
	idOf  func(*T) string
}

// NewStore creates an empty Store. idOf extracts the aggregate's own
// identifier — the store never invents or reassigns identity.
func NewStore[T any](idOf func(*T) string) *Store[T] {
	return &Store[T]{items: make(map[string]*T), idOf: idOf}
}

// Get returns the aggregate for id, or application.ErrNotFound.
func (s *Store[T]) Get(_ context.Context, id string) (*T, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.items[id]
	if !ok {
		return nil, application.ErrNotFound
	}
	return v, nil
}

// Save stores the aggregate under its own identifier, overwriting any
// previous value for the same id.
func (s *Store[T]) Save(_ context.Context, v *T) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[s.idOf(v)] = v
	return nil
}

// List returns every stored aggregate, ordered by id for a deterministic
// result (EPIC-009, TASK-072).
func (s *Store[T]) List(_ context.Context) ([]*T, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*T, 0, len(s.items))
	for _, v := range s.items {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return s.idOf(out[i]) < s.idOf(out[j]) })
	return out, nil
}
