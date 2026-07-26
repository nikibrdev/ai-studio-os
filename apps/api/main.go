// Command api runs apps/api: the REST delivery layer over
// internal/application (ADR-003, EPIC-008). This file only wires
// dependencies (via internal/infrastructure/wiring) and starts the HTTP
// server — no routing or business logic lives here (module-boundaries.md;
// that belongs in httpapi).
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ai-studio-os/apps/api/httpapi"
	"ai-studio-os/internal/application"
	"ai-studio-os/internal/domain/workflow"
	"ai-studio-os/internal/infrastructure/postgres"
	"ai-studio-os/internal/infrastructure/wiring"
	"ai-studio-os/internal/platform"
)

const defaultPort = "8080"

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// rebuildViews replays the whole event journal into the projection
// (BUGFIX-010). It reads through the EventJournal port rather than the eventbus
// package directly: the port already exists for exactly this and keeps this
// file free of a second infrastructure import.
//
// Since(0) means "everything" — the read model must reflect all history, which
// is a different question from where an event-poll cursor should resume.
func rebuildViews(ctx context.Context, sys *wiring.System, views *application.TaskProjection) error {
	entries, err := sys.EventJournal.Since(ctx, 0)
	if err != nil {
		return fmt.Errorf("api: read event journal to rebuild read model: %w", err)
	}

	events := make([]platform.Event, 0, len(entries))
	for _, e := range entries {
		events = append(events, e.Event)
	}
	if err := views.Rebuild(ctx, events); err != nil {
		return fmt.Errorf("api: rebuild read model from %d events: %w", len(events), err)
	}
	log.Printf("read model rebuilt from %d journaled events", len(events))
	return nil
}

func run() error {
	ctx := context.Background()

	sys, err := wiring.New(ctx, os.Getenv(postgres.DatabaseURLEnv), os.Getenv("QDRANT_URL"))
	if err != nil {
		return err
	}
	defer sys.Close()

	views := application.NewTaskProjection()
	if err := views.Subscribe(sys.Events); err != nil {
		return err
	}

	// Restore the read model from the durable journal before serving anything
	// (BUGFIX-010). Subscribing alone only catches events published by this
	// process, so every restart used to start with an empty projection: tasks
	// safely in PostgreSQL became invisible — GET /decisions answered "nothing
	// awaits a decision" while a task sat waiting for one. A failure here is
	// fatal on purpose: serving requests from an empty read model is worse than
	// not starting, because it looks like an answer.
	if err := rebuildViews(ctx, sys, views); err != nil {
		return err
	}

	rules := workflow.Machine{}
	deps := httpapi.Deps{
		Projects: &application.ProjectService{Projects: sys.Projects, Events: sys.Events},
		Tasks: &application.TaskPlanningService{
			Projects: sys.Projects, Tasks: sys.Tasks, Events: sys.Events, Rules: rules, IDs: sys.Tasks,
		},
		Work: &application.WorkService{
			Tasks: sys.Tasks, Executors: sys.Executors, Executions: sys.Executions, Events: sys.Events, Rules: rules,
		},
		Results: &application.ResultService{
			Projects: sys.Projects, Tasks: sys.Tasks, Executions: sys.Executions, Artifacts: sys.Artifacts, Events: sys.Events,
		},
		Completion: &application.CompletionService{
			Tasks: sys.Tasks, Repositories: sys.Repository, Events: sys.Events, Rules: rules, Views: views,
		},
		Views: views,
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	server := &http.Server{Addr: ":" + port, Handler: httpapi.NewServer(deps)}

	serveErr := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serveErr:
		return err
	case <-stop:
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return server.Shutdown(shutdownCtx)
}
