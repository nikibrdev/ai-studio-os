package main

import (
	"context"
	"errors"
	"io"
	"log"
	"testing"
	"time"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/platform"
)

// fakeEvent is a minimal platform.Event — the poller only reads metadata.
type fakeEvent struct {
	id, typ string
}

func (e fakeEvent) ID() string            { return e.id }
func (e fakeEvent) Type() string          { return e.typ }
func (e fakeEvent) SchemaVersion() int    { return 1 }
func (e fakeEvent) OccurredAt() time.Time { return time.Now() }
func (e fakeEvent) Source() string        { return "test" }
func (e fakeEvent) Actor() string         { return "" }
func (e fakeEvent) ProjectID() string     { return "proj-1" }
func (e fakeEvent) SubjectID() string     { return "TASK-001" }

// fakeJournal records the cursors it was asked for and returns whatever the
// test queued for each call.
type fakeJournal struct {
	calls     []int64
	responses [][]application.JournalEntry
	err       error
}

func (j *fakeJournal) Since(_ context.Context, afterSeq int64) ([]application.JournalEntry, error) {
	j.calls = append(j.calls, afterSeq)
	if j.err != nil {
		return nil, j.err
	}
	if len(j.responses) == 0 {
		return nil, nil
	}
	next := j.responses[0]
	j.responses = j.responses[1:]
	return next, nil
}

func entry(seq int64, typ string) application.JournalEntry {
	return application.JournalEntry{Seq: seq, Event: fakeEvent{id: "evt-" + typ, typ: typ}}
}

func newPoller(j application.EventJournal, h Handler) *Poller {
	return &Poller{
		Journal:  j,
		Interval: time.Millisecond,
		Handle:   h,
		Log:      log.New(io.Discard, "", 0),
	}
}

// Start must skip existing history: replaying it would re-dispatch
// TaskPlanned events whose work already ran.
func TestPollerStart_PositionsCursorAtJournalTip(t *testing.T) {
	journal := &fakeJournal{responses: [][]application.JournalEntry{
		{entry(7, "TaskCreated"), entry(11, "TaskPlanned")},
	}}
	var handled []string
	p := newPoller(journal, func(_ context.Context, e platform.Event) error {
		handled = append(handled, e.Type())
		return nil
	})

	if err := p.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}

	if p.cursor != 11 {
		t.Errorf("cursor = %d, want 11 (the tip)", p.cursor)
	}
	if len(handled) != 0 {
		t.Errorf("Start handled %v, want nothing — existing history must not be dispatched", handled)
	}
	if len(journal.calls) != 1 || journal.calls[0] != 0 {
		t.Errorf("Since calls = %v, want a single call with cursor 0", journal.calls)
	}
}

func TestPollerStart_EmptyJournalLeavesCursorAtZero(t *testing.T) {
	p := newPoller(&fakeJournal{}, func(context.Context, platform.Event) error { return nil })

	if err := p.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if p.cursor != 0 {
		t.Errorf("cursor = %d, want 0", p.cursor)
	}
}

func TestPollerStart_PropagatesJournalError(t *testing.T) {
	wantErr := errors.New("connection refused")
	p := newPoller(&fakeJournal{err: wantErr}, func(context.Context, platform.Event) error { return nil })

	if err := p.Start(context.Background()); !errors.Is(err, wantErr) {
		t.Fatalf("Start() error = %v, want wrapping %v", err, wantErr)
	}
}

func TestPollerPoll_HandlesEveryEntryAndAdvancesCursor(t *testing.T) {
	journal := &fakeJournal{responses: [][]application.JournalEntry{
		{entry(4, "TaskPlanned"), entry(5, "TaskStarted")},
	}}
	var handled []string
	p := newPoller(journal, func(_ context.Context, e platform.Event) error {
		handled = append(handled, e.Type())
		return nil
	})

	if err := p.poll(context.Background()); err != nil {
		t.Fatalf("poll: %v", err)
	}

	if len(handled) != 2 || handled[0] != "TaskPlanned" || handled[1] != "TaskStarted" {
		t.Errorf("handled = %v, want [TaskPlanned TaskStarted]", handled)
	}
	if p.cursor != 5 {
		t.Errorf("cursor = %d, want 5", p.cursor)
	}
}

func TestPollerPoll_QueriesFromCurrentCursor(t *testing.T) {
	journal := &fakeJournal{responses: [][]application.JournalEntry{
		{entry(9, "TaskPlanned")},
		{entry(12, "TaskStarted")},
	}}
	p := newPoller(journal, func(context.Context, platform.Event) error { return nil })
	ctx := context.Background()

	if err := p.poll(ctx); err != nil {
		t.Fatalf("first poll: %v", err)
	}
	if err := p.poll(ctx); err != nil {
		t.Fatalf("second poll: %v", err)
	}

	if len(journal.calls) != 2 || journal.calls[0] != 0 || journal.calls[1] != 9 {
		t.Errorf("Since calls = %v, want [0 9] — the second poll must resume after the first", journal.calls)
	}
	if p.cursor != 12 {
		t.Errorf("cursor = %d, want 12", p.cursor)
	}
}

// A handler error must not stall the loop: the cursor still advances, so one
// unhandleable event cannot block every later event behind it.
func TestPollerPoll_HandlerErrorStillAdvancesCursorAndContinues(t *testing.T) {
	journal := &fakeJournal{responses: [][]application.JournalEntry{
		{entry(1, "TaskPlanned"), entry(2, "TaskStarted")},
	}}
	var handled []string
	p := newPoller(journal, func(_ context.Context, e platform.Event) error {
		handled = append(handled, e.Type())
		if e.Type() == "TaskPlanned" {
			return errors.New("dispatch failed")
		}
		return nil
	})

	if err := p.poll(context.Background()); err != nil {
		t.Fatalf("poll() error = %v, want nil — a handler failure is logged, not returned", err)
	}

	if len(handled) != 2 {
		t.Errorf("handled = %v, want both entries — a failure must not skip the rest", handled)
	}
	if p.cursor != 2 {
		t.Errorf("cursor = %d, want 2 — the failed entry must not be retried forever", p.cursor)
	}
}

// A journal read failure must leave the cursor untouched so the next tick
// retries the same window rather than skipping it.
func TestPollerPoll_JournalErrorLeavesCursorUnchanged(t *testing.T) {
	wantErr := errors.New("connection refused")
	journal := &fakeJournal{err: wantErr}
	p := newPoller(journal, func(context.Context, platform.Event) error { return nil })
	p.cursor = 42

	if err := p.poll(context.Background()); !errors.Is(err, wantErr) {
		t.Fatalf("poll() error = %v, want %v", err, wantErr)
	}
	if p.cursor != 42 {
		t.Errorf("cursor = %d, want it unchanged at 42", p.cursor)
	}
}

func TestPollerRun_StopsOnContextCancel(t *testing.T) {
	journal := &fakeJournal{}
	p := newPoller(journal, func(context.Context, platform.Event) error { return nil })

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- p.Run(ctx) }()

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run() error = %v, want nil on cancellation", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after the context was cancelled")
	}
}

// A transient journal failure must not kill the process — Run logs it and
// keeps ticking.
func TestPollerRun_SurvivesJournalErrors(t *testing.T) {
	journal := &fakeJournal{err: errors.New("connection refused")}
	p := newPoller(journal, func(context.Context, platform.Event) error { return nil })

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	if err := p.Run(ctx); err != nil {
		t.Fatalf("Run() error = %v, want nil — poll failures are retried, not fatal", err)
	}
	if len(journal.calls) == 0 {
		t.Error("Run never polled the journal")
	}
}
