package httpapi

import (
	"net/http"

	"ai-studio-os/internal/application"
)

// registerWorkRoutes wires the Work operation from docs/api/tasks.md:
// starting work spawns an Execution as a side effect (WorkService.StartTask)
// — Execution has no creation route of its own (docs/api/executions.md).
func registerWorkRoutes(mux *http.ServeMux, deps Deps) {
	mux.HandleFunc("POST /projects/{projectId}/tasks/{id}/start", handleStartTask(deps))
}

type startTaskRequest struct {
	ExecutorID string `json:"executorId"`
	Actor      string `json:"actor"`
}

type executionResponse struct {
	ExecutionID string `json:"executionId"`
	TaskID      string `json:"taskId"`
	ExecutorID  string `json:"executorId"`
	State       string `json:"state"`
}

func handleStartTask(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req startTaskRequest
		if err := decodeOptionalJSON(r, &req); err != nil {
			writeInvalidBody(w, err)
			return
		}

		run, err := deps.Work.StartTask(r.Context(), application.StartTaskParams{
			ProjectID: r.PathValue("projectId"), TaskID: r.PathValue("id"), ExecutorID: req.ExecutorID, Actor: req.Actor,
		})
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, executionResponse{
			ExecutionID: run.ID(), TaskID: run.TaskID(), ExecutorID: run.ExecutorID(), State: string(run.State()),
		})
	}
}
