package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/platform"
)

// Handler reacts to one journaled event. Dispatching Developer work is
// TASK-082; this task wires a handler that only logs, so the loop itself
// can be proven first.
type Handler func(ctx context.Context, e platform.Event) error

// Poller reads the event journal by cursor and hands each new event to
// Handler. It exists because the production EventBus delivers only within
// its own process (ADR-002), so this process cannot subscribe to the one
// apps/api publishes through — see EPIC-010 "Контекст".
type Poller struct {
	Journal  application.EventJournal
	Interval time.Duration
	Handle   Handler
	Log      *log.Logger

	// cursor is the seq of the last entry handled. Held in memory only:
	// a restart resumes from the journal's tip, skipping whatever happened
	// while the process was down — the accepted v1.0 limitation
	// (EPIC-010 "Риски"), and operationally visible, unlike a silent skip.
	cursor int64
}

// Start positions the cursor at the journal's current tip so a restart
// does not replay history — replaying would re-dispatch TaskPlanned events
// whose work already ran.
//
// The tip is found by reading the journal once, because the EventJournal
// port offers no cheaper way to ask for it. At this scale the journal is
// small; if it grows enough to matter, the port gains a LatestSeq method
// rather than this loop keeping a second cursor.
func (p *Poller) Start(ctx context.Context) error {
	entries, err := p.Journal.Since(ctx, 0)
	if err != nil {
		return fmt.Errorf("orchestrator: read journal tip: %w", err)
	}
	if len(entries) > 0 {
		p.cursor = entries[len(entries)-1].Seq
	}
	p.Log.Printf("journal cursor starts at seq %d (%d events already recorded)", p.cursor, len(entries))
	return nil
}

// Run polls until ctx is cancelled. Poll failures are logged and retried
// on the next tick rather than returning: a transient database outage must
// not kill the process.
func (p *Poller) Run(ctx context.Context) error {
	ticker := time.NewTicker(p.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			p.Log.Printf("stopping at seq %d", p.cursor)
			return nil
		case <-ticker.C:
			if err := p.poll(ctx); err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					continue
				}
				p.Log.Printf("poll failed at seq %d, retrying next tick: %v", p.cursor, err)
			}
		}
	}
}

// poll performs one pass: read everything after the cursor and handle it.
// Separate from Run so tests exercise the cursor logic without timers.
//
// A Handler error is logged and the cursor still advances. Retrying
// forever on one unhandleable event would stall every later event behind
// it — the task whose event failed stays in its current state and can be
// driven manually through apps/api, which is strictly better than the
// whole loop wedging. Retries are deliberately out of EPIC-010's scope.
func (p *Poller) poll(ctx context.Context) error {
	entries, err := p.Journal.Since(ctx, p.cursor)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if err := p.Handle(ctx, entry.Event); err != nil {
			p.Log.Printf("handling %s (%s, seq %d) failed: %v", entry.Event.Type(), entry.Event.ID(), entry.Seq, err)
		}
		p.cursor = entry.Seq
	}
	return nil
}
