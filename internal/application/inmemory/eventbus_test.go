package inmemory_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"ai-studio-os/internal/application/inmemory"
	"ai-studio-os/internal/platform"
)

type fakeEvent struct{ typ string }

func (f fakeEvent) ID() string            { return "evt-1" }
func (f fakeEvent) Type() string          { return f.typ }
func (f fakeEvent) SchemaVersion() int    { return 1 }
func (f fakeEvent) OccurredAt() time.Time { return time.Time{} }
func (f fakeEvent) Source() string        { return "test" }
func (f fakeEvent) Actor() string         { return "" }
func (f fakeEvent) ProjectID() string     { return "" }
func (f fakeEvent) SubjectID() string     { return "" }

var _ platform.Event = fakeEvent{}

func TestEventBus_PublishRecordsEvent(t *testing.T) {
	bus := inmemory.NewEventBus()
	if err := bus.Publish(context.Background(), fakeEvent{typ: "Something"}); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if got := bus.Published(); len(got) != 1 || got[0].Type() != "Something" {
		t.Errorf("Published() = %v, want one Something event", got)
	}
}

func TestEventBus_DeliversToSubscribers(t *testing.T) {
	bus := inmemory.NewEventBus()
	var received []string
	if _, err := bus.Subscribe("Something", func(_ context.Context, e platform.Event) error {
		received = append(received, e.Type())
		return nil
	}); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	if err := bus.Publish(context.Background(), fakeEvent{typ: "Something"}); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if err := bus.Publish(context.Background(), fakeEvent{typ: "Other"}); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if len(received) != 1 || received[0] != "Something" {
		t.Errorf("received = %v, want exactly one Something delivery", received)
	}
}

func TestEventBus_CancelStopsDelivery(t *testing.T) {
	bus := inmemory.NewEventBus()
	calls := 0
	sub, err := bus.Subscribe("Something", func(_ context.Context, _ platform.Event) error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	if err := bus.Publish(context.Background(), fakeEvent{typ: "Something"}); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if err := sub.Cancel(); err != nil {
		t.Fatalf("Cancel: %v", err)
	}
	if err := sub.Cancel(); err != nil { // cancelling twice is a no-op
		t.Fatalf("second Cancel: %v", err)
	}
	if err := bus.Publish(context.Background(), fakeEvent{typ: "Something"}); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1 (no delivery after Cancel)", calls)
	}
}

func TestEventBus_HandlerErrorPropagates(t *testing.T) {
	bus := inmemory.NewEventBus()
	wantErr := errors.New("handler failed")
	if _, err := bus.Subscribe("Something", func(_ context.Context, _ platform.Event) error {
		return wantErr
	}); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	if err := bus.Publish(context.Background(), fakeEvent{typ: "Something"}); !errors.Is(err, wantErr) {
		t.Errorf("Publish() error = %v, want %v", err, wantErr)
	}
}
