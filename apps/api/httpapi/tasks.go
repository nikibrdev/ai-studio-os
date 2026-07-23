package httpapi

import (
	"net/http"
	"time"

	"ai-studio-os/internal/application"
)

// registerTaskCreationRoutes wires the Task operations owned by this task
// (TASK-068): create, plan, and read. StartTask/Review/Testing are
// TASK-069.
//
// Task-scoped routes are nested under /projects/{projectId} (BUGFIX-003):
// TASK-NNN is unique only within a Project (ADR-011), so a bare
// /tasks/{id} path cannot disambiguate which project's task is meant —
// ADR-011 anticipated exactly this ("любой межпроектный контекст обязан
// использовать полностью квалифицированную пару (Project, ID)").
func registerTaskCreationRoutes(mux *http.ServeMux, deps Deps) {
	mux.HandleFunc("GET /projects/{projectId}/tasks", handleListTasks(deps))
	mux.HandleFunc("POST /projects/{projectId}/tasks", handleCreateTask(deps))
	mux.HandleFunc("POST /projects/{projectId}/tasks/{id}/plan", handlePlanTask(deps))
	mux.HandleFunc("GET /projects/{projectId}/tasks/{id}", handleGetTask(deps))
}

// createTaskRequest has no ProjectID field: the project is already in the
// URL path (/projects/{projectId}/tasks), so the body does not repeat it.
type createTaskRequest struct {
	EpicID             string   `json:"epicId"`
	Title              string   `json:"title"`
	Type               string   `json:"type"`
	Scope              string   `json:"scope"`
	AcceptanceCriteria []string `json:"acceptanceCriteria"`
	Actor              string   `json:"actor"`
}

type taskResponse struct {
	ID                 string   `json:"id"`
	ProjectID          string   `json:"projectId"`
	EpicID             string   `json:"epicId"`
	Title              string   `json:"title"`
	Type               string   `json:"type"`
	Scope              string   `json:"scope"`
	AcceptanceCriteria []string `json:"acceptanceCriteria"`
	State              string   `json:"state"`
}

func handleCreateTask(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createTaskRequest
		if err := decodeOptionalJSON(r, &req); err != nil {
			writeInvalidBody(w, err)
			return
		}

		// ID is intentionally never read from the request body
		// (docs/api/tasks.md): the platform generates the public TASK-NNN
		// itself (ADR-011, TASK-065) via TaskPlanningService.IDs.
		t, err := deps.Tasks.CreateTask(r.Context(), application.CreateTaskParams{
			ProjectID:          r.PathValue("projectId"),
			EpicID:             req.EpicID,
			Title:              req.Title,
			Type:               req.Type,
			Scope:              req.Scope,
			AcceptanceCriteria: req.AcceptanceCriteria,
			Actor:              req.Actor,
		})
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, taskResponse{
			ID: t.ID(), ProjectID: t.ProjectID(), EpicID: t.EpicID(), Title: t.Title(),
			Type: t.Type(), Scope: t.Scope(), AcceptanceCriteria: t.AcceptanceCriteria(), State: string(t.State()),
		})
	}
}

func handlePlanTask(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req actorRequest
		if err := decodeOptionalJSON(r, &req); err != nil {
			writeInvalidBody(w, err)
			return
		}

		if err := deps.Tasks.PlanTask(r.Context(), r.PathValue("projectId"), r.PathValue("id"), req.Actor); err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusNoContent, nil)
	}
}

// taskViewResponse mirrors application.TaskView, including the
// descriptive fields captured from TaskCreated (EPIC-009, TASK-076):
// TaskProjection is the only read path for Task (ADR-014), so this is the
// entire shape a client can ever see — there is no separate "full task"
// response.
type taskViewResponse struct {
	ID                 string    `json:"id"`
	ProjectID          string    `json:"projectId"`
	State              string    `json:"state"`
	UpdatedAt          time.Time `json:"updatedAt"`
	Title              string    `json:"title"`
	Type               string    `json:"type"`
	Scope              string    `json:"scope"`
	AcceptanceCriteria []string  `json:"acceptanceCriteria"`
}

func taskViewResponseFrom(v application.TaskView) taskViewResponse {
	return taskViewResponse{
		ID: v.ID, ProjectID: v.ProjectID, State: string(v.State), UpdatedAt: v.UpdatedAt,
		Title: v.Title, Type: v.Type, Scope: v.Scope, AcceptanceCriteria: v.AcceptanceCriteria,
	}
}

func handleListTasks(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		views := deps.Views.ListByProject(r.PathValue("projectId"))

		out := make([]taskViewResponse, len(views))
		for i, v := range views {
			out[i] = taskViewResponseFrom(v)
		}
		writeJSON(w, http.StatusOK, out)
	}
}

func handleGetTask(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		view, ok := deps.Views.Get(r.PathValue("projectId"), r.PathValue("id"))
		if !ok {
			writeError(w, application.ErrNotFound)
			return
		}
		writeJSON(w, http.StatusOK, taskViewResponseFrom(view))
	}
}
