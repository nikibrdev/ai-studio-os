package eventbus

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"ai-studio-os/internal/platform"
)

// testEvent is a minimal platform.Event fake — no dependency on
// internal/application, matching the package's own layering (this
// package only knows the platform.Event interface).
type testEvent struct {
	id, typ, source, actor, projectID, subjectID string
	schemaVersion                                int
	occurredAt                                   time.Time
}

func (e testEvent) ID() string            { return e.id }
func (e testEvent) Type() string          { return e.typ }
func (e testEvent) SchemaVersion() int    { return e.schemaVersion }
func (e testEvent) OccurredAt() time.Time { return e.occurredAt }
func (e testEvent) Source() string        { return e.source }
func (e testEvent) Actor() string         { return e.actor }
func (e testEvent) ProjectID() string     { return e.projectID }
func (e testEvent) SubjectID() string     { return e.subjectID }

// testEventWithData additionally satisfies dataCarrier, mirroring
// application.Envelope.WithData/Data.
type testEventWithData struct {
	testEvent
	data map[string]string
}

func (e testEventWithData) Data() map[string]string { return e.data }

func newTestEvent(id, typ string) testEvent {
	return testEvent{id: id, typ: typ, source: "test", occurredAt: time.Now(), schemaVersion: 1}
}

// fakeExecer records every Exec call and can be made to fail.
type fakeExecer struct {
	execErr error
	calls   []fakeExecCall
}

type fakeExecCall struct {
	sql  string
	args []any
}

func (f *fakeExecer) Exec(_ context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	f.calls = append(f.calls, fakeExecCall{sql: sql, args: args})
	if f.execErr != nil {
		return pgconn.CommandTag{}, f.execErr
	}
	return pgconn.CommandTag{}, nil
}

func newBus(exec *fakeExecer) *Bus {
	return &Bus{pool: exec, subscribers: make(map[string][]*subEntry)}
}

func TestPublish_JournalsBeforeDelivering(t *testing.T) {
	exec := &fakeExecer{}
	b := newBus(exec)

	var delivered platform.Event
	if _, err := b.Subscribe("task.created", func(_ context.Context, e platform.Event) error {
		delivered = e
		return nil
	}); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	e := newTestEvent("evt-1", "task.created")
	if err := b.Publish(context.Background(), e); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	if len(exec.calls) != 1 {
		t.Fatalf("journal Exec calls = %d, want 1", len(exec.calls))
	}
	if delivered == nil || delivered.ID() != "evt-1" {
		t.Fatalf("handler was not delivered the event: %+v", delivered)
	}
}

func TestPublish_DeliversToMultipleSubscribersInOrder(t *testing.T) {
	b := newBus(&fakeExecer{})

	var order []string
	sub := func(name string) platform.EventHandler {
		return func(_ context.Context, _ platform.Event) error {
			order = append(order, name)
			return nil
		}
	}
	if _, err := b.Subscribe("task.created", sub("first")); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	if _, err := b.Subscribe("task.created", sub("second")); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	if err := b.Publish(context.Background(), newTestEvent("evt-1", "task.created")); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	if len(order) != 2 || order[0] != "first" || order[1] != "second" {
		t.Errorf("delivery order = %v, want [first second]", order)
	}
}

func TestPublish_OnlyDeliversToSubscribersOfMatchingType(t *testing.T) {
	b := newBus(&fakeExecer{})

	called := false
	if _, err := b.Subscribe("task.completed", func(_ context.Context, _ platform.Event) error {
		called = true
		return nil
	}); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	if err := b.Publish(context.Background(), newTestEvent("evt-1", "task.created")); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if called {
		t.Error("handler subscribed to a different event type was called")
	}
}

func TestCancel_StopsDelivery(t *testing.T) {
	b := newBus(&fakeExecer{})

	called := false
	sub, err := b.Subscribe("task.created", func(_ context.Context, _ platform.Event) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	if err := sub.Cancel(); err != nil {
		t.Fatalf("Cancel: %v", err)
	}
	if err := sub.Cancel(); err != nil {
		t.Fatalf("second Cancel (no-op) unexpected error: %v", err)
	}

	if err := b.Publish(context.Background(), newTestEvent("evt-1", "task.created")); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if called {
		t.Error("cancelled subscription still received the event")
	}
}

func TestPublish_JournalFailureStopsDeliveryAndReturnsError(t *testing.T) {
	wantErr := errors.New("connection refused")
	b := newBus(&fakeExecer{execErr: wantErr})

	called := false
	if _, err := b.Subscribe("task.created", func(_ context.Context, _ platform.Event) error {
		called = true
		return nil
	}); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	err := b.Publish(context.Background(), newTestEvent("evt-1", "task.created"))
	if !errors.Is(err, wantErr) {
		t.Fatalf("Publish() error = %v, want wrapping %v", err, wantErr)
	}
	if called {
		t.Error("handler was called despite the journal write failing")
	}
}

func TestPublish_HandlerErrorPropagates(t *testing.T) {
	b := newBus(&fakeExecer{})
	wantErr := errors.New("handler failed")

	if _, err := b.Subscribe("task.created", func(_ context.Context, _ platform.Event) error {
		return wantErr
	}); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	err := b.Publish(context.Background(), newTestEvent("evt-1", "task.created"))
	if !errors.Is(err, wantErr) {
		t.Fatalf("Publish() error = %v, want %v", err, wantErr)
	}
}

func TestPublish_JournalsDataCarrierPayload(t *testing.T) {
	exec := &fakeExecer{}
	b := newBus(exec)

	e := testEventWithData{
		testEvent: newTestEvent("evt-1", "task.review-completed"),
		data:      map[string]string{"to": "testing"},
	}
	if err := b.Publish(context.Background(), e); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	if len(exec.calls) != 1 {
		t.Fatalf("journal Exec calls = %d, want 1", len(exec.calls))
	}
	payload, ok := exec.calls[0].args[8].([]byte)
	if !ok {
		t.Fatalf("journal payload arg has type %T, want []byte", exec.calls[0].args[8])
	}
	if got := string(payload); got != `{"to":"testing"}` {
		t.Errorf("journal payload = %s, want {\"to\":\"testing\"}", got)
	}
}

func TestPublish_EventWithoutDataCarrierJournalsEmptyObject(t *testing.T) {
	exec := &fakeExecer{}
	b := newBus(exec)

	if err := b.Publish(context.Background(), newTestEvent("evt-1", "task.created")); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	payload, ok := exec.calls[0].args[8].([]byte)
	if !ok {
		t.Fatalf("journal payload arg has type %T, want []byte", exec.calls[0].args[8])
	}
	if got := string(payload); got != `{}` {
		t.Errorf("journal payload = %s, want {}", got)
	}
}
