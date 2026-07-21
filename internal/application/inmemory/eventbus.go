package inmemory

import (
	"context"
	"sync"

	"ai-studio-os/internal/platform"
)

// EventBus is a deterministic, synchronous fake of platform.EventBus:
// Publish delivers to current subscribers immediately, in registration
// order, and records every published event for test assertions. Not an
// infrastructure adapter — the real in-memory bus (ADR-002) is EPIC-005
// scope.
type EventBus struct {
	mu          sync.Mutex
	subscribers map[string][]*subEntry
	published   []platform.Event
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

// NewEventBus creates an empty EventBus.
func NewEventBus() *EventBus {
	return &EventBus{subscribers: make(map[string][]*subEntry)}
}

// Publish implements platform.EventBus.
func (b *EventBus) Publish(ctx context.Context, e platform.Event) error {
	b.mu.Lock()
	b.published = append(b.published, e)
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
func (b *EventBus) Subscribe(eventType string, h platform.EventHandler) (platform.Subscription, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	entry := &subEntry{handler: h}
	b.subscribers[eventType] = append(b.subscribers[eventType], entry)
	return &subscriptionHandle{entry: entry}, nil
}

// Published returns a copy of every event published so far, in order.
func (b *EventBus) Published() []platform.Event {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]platform.Event(nil), b.published...)
}
