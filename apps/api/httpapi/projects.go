package httpapi

import (
	"net/http"
	"time"

	"ai-studio-os/internal/application"
)

// registerProjectRoutes wires the three Project operations from
// docs/api/projects.md — required order: create -> connect a repository
// (at least once) -> activate.
func registerProjectRoutes(mux *http.ServeMux, deps Deps) {
	mux.HandleFunc("GET /projects", handleListProjects(deps))
	mux.HandleFunc("POST /projects", handleCreateProject(deps))
	mux.HandleFunc("POST /projects/{id}/repositories", handleConnectRepository(deps))
	mux.HandleFunc("POST /projects/{id}/activate", handleActivateProject(deps))
}

type createProjectRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type projectResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	State     string    `json:"state"`
	CreatedAt time.Time `json:"createdAt"`
}

func handleListProjects(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projects, err := deps.Projects.ListProjects(r.Context())
		if err != nil {
			writeError(w, err)
			return
		}

		out := make([]projectResponse, len(projects))
		for i, proj := range projects {
			out[i] = projectResponse{
				ID: proj.ID(), Name: proj.Name(), State: string(proj.State()), CreatedAt: proj.CreatedAt(),
			}
		}
		writeJSON(w, http.StatusOK, out)
	}
}

func handleCreateProject(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createProjectRequest
		if err := decodeOptionalJSON(r, &req); err != nil {
			writeInvalidBody(w, err)
			return
		}

		proj, err := deps.Projects.CreateProject(r.Context(), application.CreateProjectParams{ID: req.ID, Name: req.Name})
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, projectResponse{
			ID: proj.ID(), Name: proj.Name(), State: string(proj.State()), CreatedAt: proj.CreatedAt(),
		})
	}
}

type connectRepositoryRequest struct {
	Repository string `json:"repository"`
	Actor      string `json:"actor"`
}

func handleConnectRepository(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req connectRepositoryRequest
		if err := decodeOptionalJSON(r, &req); err != nil {
			writeInvalidBody(w, err)
			return
		}

		if err := deps.Projects.ConnectRepository(r.Context(), r.PathValue("id"), req.Repository, req.Actor); err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusNoContent, nil)
	}
}

type actorRequest struct {
	Actor string `json:"actor"`
}

func handleActivateProject(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req actorRequest
		if err := decodeOptionalJSON(r, &req); err != nil {
			writeInvalidBody(w, err)
			return
		}

		if err := deps.Projects.Activate(r.Context(), r.PathValue("id"), req.Actor); err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusNoContent, nil)
	}
}
