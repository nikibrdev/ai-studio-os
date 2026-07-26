// Command orchestrator runs apps/orchestrator: the coordinator that reacts
// to task lifecycle events and starts Executors through the Executor
// contract (EPIC-010). It holds no domain rules and no durable state of
// its own (module-boundaries.md) — decisions about which transitions are
// legal belong to the task/workflow modules, reached only through
// internal/application.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	claudecode "ai-studio-os/agents/claude-code"
	"ai-studio-os/internal/application"
	"ai-studio-os/internal/domain/shared"
	"ai-studio-os/internal/domain/workflow"
	"ai-studio-os/internal/infrastructure/github"
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
	executorID, err := EnsureExecutor(ctx, executors, sys.Executors, shared.RoleDeveloper)
	if err != nil {
		return err
	}
	logger.Printf("developer executor ready: %s (image %s)", executorID, cfg.ExecutionImage)
	if cfg.ProviderAPIKey == "" {
		logger.Printf("warning: %s is empty — sandboxes will start but the AI provider will not be called", providerAPIKeyEnv)
	}

	if sys.Repository == nil {
		return fmt.Errorf("orchestrator: %s is not set — a repository provider is required to create task branches", github.TokenEnv)
	}

	dispatcher := &Dispatcher{
		Projects:  sys.Projects,
		Executors: sys.Executors,
		Work: &application.WorkService{
			Tasks: sys.Tasks, Executors: sys.Executors, Executions: sys.Executions,
			Events: sys.Events, Rules: workflow.Machine{},
		},
		Results: &application.ResultService{
			Projects: sys.Projects, Tasks: sys.Tasks, Executions: sys.Executions,
			Artifacts: sys.Artifacts, Events: sys.Events,
		},
		Completion: &application.CompletionService{
			Tasks: sys.Tasks, Repositories: sys.Repository, Events: sys.Events, Rules: workflow.Machine{},
		},
		Views: application.NewTaskProjection(),
		Repos: sys.Repository,
		// One adapter serves every role; the role reaches the agent through
		// ExecutorTask.Role, which shapes its prompt (ADR-007).
		NewExecutor: func(shared.Role) (platform.Executor, error) {
			return claudecode.New(cfg.ExecutionImage, cfg.GitToken, cfg.ProviderAPIKey)
		},
		Log: logger,
	}

	poller := &Poller{
		Journal:  sys.EventJournal,
		Interval: cfg.PollInterval,
		Seed:     dispatcher.Observe,
		Handle:   dispatcher.Handle,
		Log:      logger,
	}
	if err := poller.Start(ctx); err != nil {
		return err
	}

	logger.Printf("polling event journal every %s", cfg.PollInterval)
	return poller.Run(ctx)
}
