// Command orchestrator runs apps/orchestrator: the coordinator that reacts
// to task lifecycle events and starts Executors through the Executor
// contract (EPIC-010). It holds no domain rules and no durable state of
// its own (module-boundaries.md) — decisions about which transitions are
// legal belong to the task/workflow modules, reached only through
// internal/application.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/infrastructure/wiring"
	"ai-studio-os/internal/platform"
)

// actor identifies this process as the initiator of the commands it issues
// (platform.Event's Actor field, docs/architecture/events.md: role plus
// executor identity).
const actor = "orchestrator"

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	logger := log.New(os.Stdout, "orchestrator: ", log.LstdFlags)

	cfg, err := FromEnv()
	if err != nil {
		return err
	}

	// signal.NotifyContext cancels ctx on SIGINT/SIGTERM, which unwinds the
	// poll loop — apps/api uses the equivalent select on a signal channel
	// because it also has an HTTP server to shut down.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	sys, err := wiring.New(ctx, cfg.DatabaseURL, cfg.QdrantURL)
	if err != nil {
		return err
	}
	defer sys.Close()

	executors := &application.ExecutorService{Executors: sys.Executors, Events: sys.Events}
	executorID, err := EnsureDeveloperExecutor(ctx, executors, sys.Executors)
	if err != nil {
		return err
	}
	logger.Printf("developer executor ready: %s (image %s)", executorID, cfg.ExecutionImage)
	if cfg.ProviderAPIKey == "" {
		logger.Printf("warning: %s is empty — sandboxes will start but the AI provider will not be called", providerAPIKeyEnv)
	}

	poller := &Poller{
		Journal:  sys.EventJournal,
		Interval: cfg.PollInterval,
		Handle:   logHandler(logger),
		Log:      logger,
	}
	if err := poller.Start(ctx); err != nil {
		return err
	}

	logger.Printf("polling event journal every %s", cfg.PollInterval)
	return poller.Run(ctx)
}

// logHandler is the placeholder Handler for TASK-081: it proves the loop
// observes real events. Dispatching Developer work on TaskPlanned is
// TASK-082, which replaces this.
func logHandler(logger *log.Logger) Handler {
	return func(_ context.Context, e platform.Event) error {
		logger.Printf("event %s (%s) project=%q subject=%q", e.Type(), e.ID(), e.ProjectID(), e.SubjectID())
		return nil
	}
}
