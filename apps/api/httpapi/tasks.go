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

	// AwaitingDecision names the decision a human owes this task, or is empty
	// when none is owed (docs/api/decisions.md). Served here so a client does
	// not map states to decisions itself — that rule belongs to the platform,
	// and duplicating it in a UI is forbidden (module-boundaries.md).
	AwaitingDecision string `json:"awaitingDecision"`

	// Repository and PullRequestID identify the pull request under review,
	// empty until one is known (BUGFIX-009). Exposed so a human about to make
	// the acceptance decision can see what they are merging — the decision is
	// irreversible within the task (ADR-008).
	Repository    string `json:"repository"`
	PullRequestID string `json:"pullRequestId"`

	// QAReportID identifies the QA report to read before the acceptance
	// decision, empty when none exists (TASK-100). Only the identifier — the
	// report itself comes from GET /artifacts/{id}.
	QAReportID string `json:"qaReportId"`
}

func taskViewResponseFrom(v application.TaskView) taskViewResponse {
	return taskViewResponse{
		ID: v.ID, ProjectID: v.ProjectID, State: string(v.State), UpdatedAt: v.UpdatedAt,
		Title: v.Title, Type: v.Type, Scope: v.Scope, AcceptanceCriteria: v.AcceptanceCriteria,
		AwaitingDecision: string(application.DecisionFor(v.State)),
		Repository:       v.Repository,
		PullRequestID:    v.PullRequestID,
		QAReportID:       v.QAReportID,
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
