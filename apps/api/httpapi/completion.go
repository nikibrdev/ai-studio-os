package httpapi

import (
	"net/http"

	"ai-studio-os/internal/application"
)

// registerCompletionRoutes wires the three final golden-path operations
// from docs/api/tasks.md (CompletionService, ADR-008).
func registerCompletionRoutes(mux *http.ServeMux, deps Deps) {
	mux.HandleFunc("POST /projects/{projectId}/tasks/{id}/request-review", handleRequestReview(deps))
	mux.HandleFunc("POST /projects/{projectId}/tasks/{id}/complete-review", handleCompleteReview(deps))
	mux.HandleFunc("POST /projects/{projectId}/tasks/{id}/complete-testing", handleCompleteTesting(deps))
}

// requestReviewRequest optionally carries the pull request under review —
// docs/architecture/events.md requires ReviewRequested to reference it, and
// the acceptance decision later needs it to merge (BUGFIX-009). Both fields
// may be omitted: a review can be requested before a pull request exists.
type requestReviewRequest struct {
	Repository    string `json:"repository"`
	PullRequestID string `json:"pullRequestId"`
	Actor         string `json:"actor"`
}

func handleRequestReview(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req requestReviewRequest
		if err := decodeOptionalJSON(r, &req); err != nil {
			writeInvalidBody(w, err)
			return
		}

		err := deps.Completion.RequestReview(r.Context(), application.RequestReviewParams{
			ProjectID:     r.PathValue("projectId"),
			TaskID:        r.PathValue("id"),
			Repository:    req.Repository,
			PullRequestID: req.PullRequestID,
			Actor:         req.Actor,
		})
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusNoContent, nil)
	}
}

type completeReviewRequest struct {
	Approved bool   `json:"approved"`
	Actor    string `json:"actor"`
}

func handleCompleteReview(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req completeReviewRequest
		if err := decodeOptionalJSON(r, &req); err != nil {
			writeInvalidBody(w, err)
			return
		}

		if err := deps.Completion.CompleteReview(r.Context(), r.PathValue("projectId"), r.PathValue("id"), req.Approved, req.Actor); err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusNoContent, nil)
	}
}

type completeTestingRequest struct {
	Passed        bool   `json:"passed"`
	Repository    string `json:"repository"`
	PullRequestID string `json:"pullRequestId"`
	Actor         string `json:"actor"`
}

func handleCompleteTesting(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req completeTestingRequest
		if err := decodeOptionalJSON(r, &req); err != nil {
			writeInvalidBody(w, err)
			return
		}

		err := deps.Completion.CompleteTesting(r.Context(), application.CompleteTestingParams{
			ProjectID: r.PathValue("projectId"), TaskID: r.PathValue("id"), Passed: req.Passed,
			Repository: req.Repository, PullRequestID: req.PullRequestID, Actor: req.Actor,
		})
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusNoContent, nil)
	}
}
