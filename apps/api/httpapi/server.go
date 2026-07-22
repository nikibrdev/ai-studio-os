package httpapi

import (
	"net/http"

	"ai-studio-os/internal/application"
)

// Deps is every use-case service and read projection the HTTP layer
// calls into — assembled by apps/api/main.go from a wiring.System
// (internal/infrastructure/wiring).
type Deps struct {
	Projects   *application.ProjectService
	Tasks      *application.TaskPlanningService
	Work       *application.WorkService
	Results    *application.ResultService
	Completion *application.CompletionService
	Views      *application.TaskProjection
}

// NewServer builds the HTTP router (docs/api/README.md). Resource routes
// are added here as their handlers are implemented — Projects/Tasks
// creation (TASK-068), Work/Result/Completion (TASK-069).
func NewServer(deps Deps) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealthz)
	registerProjectRoutes(mux, deps)
	registerTaskCreationRoutes(mux, deps)
	return mux
}

// handleHealthz reports liveness without touching any dependency —
// useful to confirm the process itself is up before checking Postgres/
// Qdrant connectivity separately.
func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
