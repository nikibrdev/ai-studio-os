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

// reviewAdapter presents a claudecode.Executor as a ReviewExecutor.
//
// This wrapper is the reason dispatch logic never names a concrete AI backend:
// the adapter's verdict type stops here, and the rest of the process sees only
// primitives (module-boundaries.md — naming a provider is permitted in this
// composition root and nowhere else).
type reviewAdapter struct{ *claudecode.Executor }

// Review reports the agent's decision, flattening the adapter's own verdict
// type so ReviewExecutor stays free of it.
func (a reviewAdapter) Review(ctx context.Context) (bool, string, error) {
	v, err := a.Verdict(ctx)
	if err != nil {
		return false, "", err
	}
	return v.Approved, v.Comment, nil
}

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

	reviewerID, err := EnsureExecutor(ctx, executors, sys.Executors, shared.RoleReviewer)
	if err != nil {
		return err
	}
	logger.Printf("reviewer executor ready: %s", reviewerID)
	if cfg.ProviderAPIKey == "" {
		logger.Printf("warning: %s is empty — sandboxes will start but the AI provider will not be called", providerAPIKeyEnv)
	}

	if sys.Repository == nil {
		return fmt.Errorf("orchestrator: %s is not set — a repository provider is required to create task branches", github.TokenEnv)
	}

	// One projection instance, shared: the dispatcher fills it from the journal
	// and CompletionService reads the pull request reference back out of it
	// (BUGFIX-009). Two instances would leave the reference invisible to the
	// service that needs it.
	views := application.NewTaskProjection()

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
			Tasks: sys.Tasks, Repositories: sys.Repository, Events: sys.Events,
			Rules: workflow.Machine{}, Views: views,
		},
		Views: views,
		Repos: sys.Repository,
		// One adapter serves every role; the role reaches the agent through
		// ExecutorTask.Role, which shapes its prompt (ADR-007).
		NewExecutor: func(shared.Role) (platform.Executor, error) {
			return claudecode.New(cfg.ExecutionImage, cfg.GitToken, cfg.ProviderAPIKey)
		},
		NewReviewer: func() (ReviewExecutor, error) {
			e, err := claudecode.New(cfg.ExecutionImage, cfg.GitToken, cfg.ProviderAPIKey)
			if err != nil {
				return nil, err
			}
			return reviewAdapter{e}, nil
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
