package httpapi

import "net/http"

// registerExecutionRoutes wires the two Execution operations from
// docs/api/executions.md — Execution creation is not a route of its own
// (a side effect of POST /projects/{projectId}/tasks/{id}/start, work.go).
// Execution ids are globally unique (crypto/rand, internal/application.NewID),
// unlike Task's public TASK-NNN, so these routes are not nested under a
// project — but ResultService still needs projectID (BUGFIX-003: it
// validates the Execution's owning Task via (projectID, TaskID)), supplied
// in the request body instead of the path.
func registerExecutionRoutes(mux *http.ServeMux, deps Deps) {
	mux.HandleFunc("POST /executions/{id}/succeed", handleSucceedExecution(deps))
	mux.HandleFunc("POST /executions/{id}/fail", handleFailExecution(deps))
}

type executionActionRequest struct {
	ProjectID string `json:"projectId"`
	Actor     string `json:"actor"`
}

func handleSucceedExecution(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req executionActionRequest
		if err := decodeOptionalJSON(r, &req); err != nil {
			writeInvalidBody(w, err)
			return
		}

		if err := deps.Results.SucceedExecution(r.Context(), req.ProjectID, r.PathValue("id"), req.Actor); err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusNoContent, nil)
	}
}

func handleFailExecution(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req executionActionRequest
		if err := decodeOptionalJSON(r, &req); err != nil {
			writeInvalidBody(w, err)
			return
		}

		if err := deps.Results.FailExecution(r.Context(), req.ProjectID, r.PathValue("id"), req.Actor); err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusNoContent, nil)
	}
}
