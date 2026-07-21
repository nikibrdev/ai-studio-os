package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"ai-studio-os/internal/platform"
)

// execer is the subset of *pgxpool.Pool this package needs — narrowed so
// tests can inject a fake without a real database.
type execer interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

// dataCarrier is satisfied by event envelopes (e.g. application.Envelope,
// via its WithData/Data methods) that attach event-type-specific data
// beyond the fixed platform.Event fields. Matched structurally so this
// package does not need to import internal/application.
type dataCarrier interface {
	Data() map[string]string
}

// Bus is the production implementation of platform.EventBus: synchronous,
// in-process delivery to every current subscriber of an event's type, in
// registration order — plus a durable journal in PostgreSQL. Every
// Publish journals the event before delivering it: if the journal write
// fails, Publish returns an error and no handler runs, so the journal is
// never behind what subscribers have seen.
type Bus struct {
	pool execer

	mu          sync.Mutex
	subscribers map[string][]*subEntry
}

type subEntry struct {
	handler   platform.EventHandler
	cancelled bool
}

type subscriptionHandle struct{ entry *subEntry }

// Cancel implements platform.Subscription. Cancelling an already
// cancelled subscription is a no-op.
func (h *subscriptionHandle) Cancel() error {
	h.entry.cancelled = true
	return nil
}

var _ platform.EventBus = (*Bus)(nil)

// New creates a Bus that journals every published event through the given
// pool (see internal/infrastructure/postgres migration 0004_event_journal.sql).
func New(pool *pgxpool.Pool) *Bus {
	return &Bus{pool: pool, subscribers: make(map[string][]*subEntry)}
}

// Publish implements platform.EventBus.
func (b *Bus) Publish(ctx context.Context, e platform.Event) error {
	if err := b.journal(ctx, e); err != nil {
		return err
	}

	b.mu.Lock()
	entries := append([]*subEntry(nil), b.subscribers[e.Type()]...)
	b.mu.Unlock()

	for _, entry := range entries {
		if entry.cancelled {
			continue
		}
		if err := entry.handler(ctx, e); err != nil {
			return err
		}
	}
	return nil
}

// Subscribe implements platform.EventBus.
func (b *Bus) Subscribe(eventType string, h platform.EventHandler) (platform.Subscription, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	entry := &subEntry{handler: h}
	b.subscribers[eventType] = append(b.subscribers[eventType], entry)
	return &subscriptionHandle{entry: entry}, nil
}

func (b *Bus) journal(ctx context.Context, e platform.Event) error {
	data := map[string]string{}
	if dc, ok := e.(dataCarrier); ok && dc.Data() != nil {
		data = dc.Data()
	}
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("eventbus: marshal event data %s: %w", e.ID(), err)
	}

	const q = `
INSERT INTO event_journal (id, type, schema_version, occurred_at, source, actor, project_id, subject_id, data)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err = b.pool.Exec(ctx, q,
		e.ID(), e.Type(), e.SchemaVersion(), e.OccurredAt(), e.Source(), e.Actor(), e.ProjectID(), e.SubjectID(), payload,
	)
	if err != nil {
		return fmt.Errorf("eventbus: journal event %s: %w", e.ID(), err)
	}
	return nil
}
