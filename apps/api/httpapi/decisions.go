package httpapi

import "net/http"

// registerDecisionRoutes registers the read operation for what awaits a
// human's decision (docs/api/decisions.md).
//
// Not nested under a project, unlike task operations: the question is "what is
// waiting for a decision at all", not "in this project". A project's own page
// gets the same flag per task through taskViewResponse.AwaitingDecision, so a
// per-project variant would add a second way to ask the same thing.
func registerDecisionRoutes(mux *http.ServeMux, deps Deps) {
	mux.HandleFunc("GET /decisions", handleListDecisions(deps))
}

type decisionResponse struct {
	Decision string           `json:"decision"`
	Task     taskViewResponse `json:"task"`
}

type listDecisionsResponse struct {
	Decisions []decisionResponse `json:"decisions"`
}

func handleListDecisions(deps Deps) http.HandlerFunc {
	// No request parameters: the operation takes neither path values nor
	// filters — it answers "what awaits a decision at all".
	return func(w http.ResponseWriter, _ *http.Request) {
		awaiting := deps.Views.ListAwaitingDecision()

		// Built with a non-nil slice so an empty result serialises as [] rather
		// than null: nothing awaiting a decision is a normal state, and a client
		// should not have to treat it as a special case.
		out := make([]decisionResponse, 0, len(awaiting))
		for _, a := range awaiting {
			out = append(out, decisionResponse{
				Decision: string(a.Decision),
				Task:     taskViewResponseFrom(a.Task),
			})
		}

		writeJSON(w, http.StatusOK, listDecisionsResponse{Decisions: out})
	}
}
