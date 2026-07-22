// Command api runs apps/api: the REST delivery layer over
// internal/application (ADR-003, EPIC-008). This file only wires
// dependencies (via internal/infrastructure/wiring) and starts the HTTP
// server — no routing or business logic lives here (module-boundaries.md;
// that belongs in httpapi).
package main

import (
	"context"
	"errors"
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
)

const defaultPort = "8080"

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
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
			Tasks: sys.Tasks, Repositories: sys.Repository, Events: sys.Events, Rules: rules,
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
