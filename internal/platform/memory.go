package platform

import (
	"context"
	"time"
)

// MemoryEntry is one piece of knowledge accumulated for a project
// (docs/architecture/domain-model.md, "Memory"). Entries are human-readable
// and auditable.
type MemoryEntry interface {
	// ID returns the unique identifier of the entry.
	ID() string

	// ProjectID returns the project the knowledge belongs to. Knowledge of
	// different projects is never mixed.
	ProjectID() string

	// Kind returns the knowledge category (project fact, decision,
	// execution experience). The exact taxonomy is designed in v0.7.
	Kind() string

	// Content returns the knowledge itself in human-readable form.
	Content() string

	// Source returns where the knowledge came from (execution, document,
	// human input).
	Source() string

	// RecordedAt returns when the knowledge was recorded.
	RecordedAt() time.Time
}

// MemoryProvider stores project knowledge and finds it again — from v0.7
// semantically, via the Qdrant adapter
// (docs/architecture/interfaces.md, "Memory Provider").
//
// Contract constraints:
//   - memory is NOT a source of truth: documentation and code always win on
//     conflict (docs/architecture/memory.md);
//   - project isolation is mandatory: a search never returns entries of
//     another project;
//   - implementations are swappable (file-based before v0.7, Qdrant after)
//     without changes to this interface.
type MemoryProvider interface {
	// Record stores one knowledge entry.
	Record(ctx context.Context, entry MemoryEntry) error

	// Search returns up to limit entries of the project relevant to the
	// query, most relevant first.
	Search(ctx context.Context, projectID, query string, limit int) ([]MemoryEntry, error)
}
