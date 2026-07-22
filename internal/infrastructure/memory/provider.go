package memory

import (
	"context"
	"fmt"
	"time"

	"ai-studio-os/internal/platform"
)

// fileStore is the subset of *FileStore Provider needs — narrowed so
// tests can inject a fake without touching the filesystem (same pattern
// as agents/claude-code's sandbox interface).
type fileStore interface {
	Write(ctx context.Context, entry Entry) error
	List(ctx context.Context, projectID string) ([]Entry, error)
}

// vectorIndex is the subset of *QdrantClient Provider needs.
type vectorIndex interface {
	Upsert(ctx context.Context, id string, vector []float32, payload map[string]any) error
	Search(ctx context.Context, projectID string, vector []float32, limit int) ([]Point, error)
}

// Provider implements platform.MemoryProvider (ADR-018) over a FileStore
// (durable source of truth) and a QdrantClient (derived, rebuildable
// search index) — the same "durable store + rebuildable index" principle
// as event_journal/ReadJournal (EPIC-005).
type Provider struct {
	files fileStore
	index vectorIndex
}

var _ platform.MemoryProvider = (*Provider)(nil)

// NewProvider creates a Provider backed by files (memory/<projectID>/<id>.md)
// and index (Qdrant). Callers are responsible for the collection already
// existing (QdrantClient.EnsureCollection) before use.
func NewProvider(files *FileStore, index *QdrantClient) *Provider {
	return &Provider{files: files, index: index}
}

// Record implements platform.MemoryProvider: it writes the entry to its
// file first, then indexes it in Qdrant. If indexing fails, the file is
// already durably saved and can be recovered with Reindex — the known
// limitation that the two writes are not atomic with each other
// (docs/roadmap/EPIC-007-memory-system.md, "Риски").
func (p *Provider) Record(ctx context.Context, entry platform.MemoryEntry) error {
	e := toEntry(entry)
	if err := p.files.Write(ctx, e); err != nil {
		return fmt.Errorf("memory: record: write file: %w", err)
	}
	if err := p.index.Upsert(ctx, e.ID(), embed(e.Content()), payloadOf(e)); err != nil {
		return fmt.Errorf("memory: record: index: %w", err)
	}
	return nil
}

// Search implements platform.MemoryProvider: it embeds the query, searches
// Qdrant filtered by projectID, and reconstructs entries from the
// self-contained payload without touching files.
func (p *Provider) Search(ctx context.Context, projectID, query string, limit int) ([]platform.MemoryEntry, error) {
	points, err := p.index.Search(ctx, projectID, embed(query), limit)
	if err != nil {
		return nil, fmt.Errorf("memory: search: %w", err)
	}

	entries := make([]platform.MemoryEntry, 0, len(points))
	for _, pt := range points {
		e, err := entryFromPayload(pt.ID, pt.Payload)
		if err != nil {
			return nil, fmt.Errorf("memory: search: reconstruct entry %s: %w", pt.ID, err)
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// Reindex rebuilds projectID's presence in the Qdrant index from its
// files from scratch — recovery from the file/index divergence Record's
// two-step write can leave behind (same principle as
// eventbus.ReadJournal/TaskProjection.Rebuild, EPIC-005).
func (p *Provider) Reindex(ctx context.Context, projectID string) error {
	entries, err := p.files.List(ctx, projectID)
	if err != nil {
		return fmt.Errorf("memory: reindex: list files: %w", err)
	}
	for _, e := range entries {
		if err := p.index.Upsert(ctx, e.ID(), embed(e.Content()), payloadOf(e)); err != nil {
			return fmt.Errorf("memory: reindex: upsert %s: %w", e.ID(), err)
		}
	}
	return nil
}

// toEntry converts any platform.MemoryEntry implementation into the
// concrete Entry FileStore/embed operate on.
func toEntry(entry platform.MemoryEntry) Entry {
	return restoreEntry(entry.ID(), entry.ProjectID(), entry.Kind(), entry.Source(), entry.Content(), entry.RecordedAt())
}

// payloadOf renders entry as the self-contained Qdrant payload
// (ADR-018): project_id, kind, content, source, recorded_at.
func payloadOf(e Entry) map[string]any {
	return map[string]any{
		"project_id":  e.ProjectID(),
		"kind":        e.Kind(),
		"content":     e.Content(),
		"source":      e.Source(),
		"recorded_at": e.RecordedAt().UTC().Format(time.RFC3339),
	}
}

// entryFromPayload reconstructs an Entry from a Qdrant search result's
// payload (ADR-018: the payload is self-contained, no file access needed).
func entryFromPayload(id string, payload map[string]any) (Entry, error) {
	projectID, _ := payload["project_id"].(string)
	kind, _ := payload["kind"].(string)
	content, _ := payload["content"].(string)
	source, _ := payload["source"].(string)
	recordedAtStr, _ := payload["recorded_at"].(string)

	recordedAt, err := time.Parse(time.RFC3339, recordedAtStr)
	if err != nil {
		return Entry{}, fmt.Errorf("invalid recorded_at %q: %w", recordedAtStr, err)
	}
	return restoreEntry(id, projectID, kind, source, content, recordedAt), nil
}
