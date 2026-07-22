package httpapi

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"ai-studio-os/internal/application/inmemory"
)

// taskThroughReview drives a task from creation through StartTask and
// RequestReview via real HTTP requests, returning its (projectID, taskID).
func taskThroughReview(t *testing.T, server http.Handler, deps Deps) (string, string) {
	t.Helper()
	projectID := createActiveProject(t, server)
	taskID := createReadyTask(t, server, projectID)
	startExecution(t, server, deps, projectID, taskID)

	rec := doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/"+projectID+"/tasks/"+taskID+"/request-review", nil), nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("request-review status = %d, body = %s", rec.Code, rec.Body.String())
	}
	return projectID, taskID
}

func TestRequestReview_UnknownTaskReturns404(t *testing.T) {
	server := NewServer(testDeps())
	projectID := createActiveProject(t, server)

	rec := doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/"+projectID+"/tasks/missing/request-review", nil), nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestCompleteReview_ApprovedMovesToTesting(t *testing.T) {
	deps := testDeps()
	server := NewServer(deps)
	projectID, taskID := taskThroughReview(t, server, deps)

	rec := doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/"+projectID+"/tasks/"+taskID+"/complete-review",
		jsonBody(t, completeReviewRequest{Approved: true})), nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var view taskViewResponse
	doRequest(t, server, httptest.NewRequest(http.MethodGet, "/projects/"+projectID+"/tasks/"+taskID, nil), &view)
	if view.State != "testing" {
		t.Errorf("State = %q, want testing", view.State)
	}
}

func TestCompleteReview_RejectedReturnsToInProgress(t *testing.T) {
	deps := testDeps()
	server := NewServer(deps)
	projectID, taskID := taskThroughReview(t, server, deps)

	rec := doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/"+projectID+"/tasks/"+taskID+"/complete-review",
		jsonBody(t, completeReviewRequest{Approved: false})), nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var view taskViewResponse
	doRequest(t, server, httptest.NewRequest(http.MethodGet, "/projects/"+projectID+"/tasks/"+taskID, nil), &view)
	if view.State != "in-progress" {
		t.Errorf("State = %q, want in-progress", view.State)
	}
}

func TestCompleteTesting_PassedMergesAndReachesDone(t *testing.T) {
	deps := testDeps()
	server := NewServer(deps)
	projectID, taskID := taskThroughReview(t, server, deps)
	doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/"+projectID+"/tasks/"+taskID+"/complete-review",
		jsonBody(t, completeReviewRequest{Approved: true})), nil)

	rec := doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/"+projectID+"/tasks/"+taskID+"/complete-testing",
		jsonBody(t, completeTestingRequest{Passed: true, Repository: "github.com/org/repo", PullRequestID: "42"})), nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var view taskViewResponse
	doRequest(t, server, httptest.NewRequest(http.MethodGet, "/projects/"+projectID+"/tasks/"+taskID, nil), &view)
	if view.State != "done" {
		t.Errorf("State = %q, want done", view.State)
	}
}

func TestCompleteTesting_FailedReturnsToInProgress(t *testing.T) {
	deps := testDeps()
	server := NewServer(deps)
	projectID, taskID := taskThroughReview(t, server, deps)
	doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/"+projectID+"/tasks/"+taskID+"/complete-review",
		jsonBody(t, completeReviewRequest{Approved: true})), nil)

	rec := doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/"+projectID+"/tasks/"+taskID+"/complete-testing",
		jsonBody(t, completeTestingRequest{Passed: false})), nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var view taskViewResponse
	doRequest(t, server, httptest.NewRequest(http.MethodGet, "/projects/"+projectID+"/tasks/"+taskID, nil), &view)
	if view.State != "in-progress" {
		t.Errorf("State = %q, want in-progress", view.State)
	}
}

// TestCompleteTesting_MergeFailureKeepsTaskInTesting reproduces ADR-008's
// guard through HTTP end-to-end: if the merge itself fails, Done is
// unreachable — the task stays in Testing rather than advancing on a
// merge that never actually happened.
func TestCompleteTesting_MergeFailureKeepsTaskInTesting(t *testing.T) {
	deps := testDeps()
	server := NewServer(deps)
	projectID, taskID := taskThroughReview(t, server, deps)
	doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/"+projectID+"/tasks/"+taskID+"/complete-review",
		jsonBody(t, completeReviewRequest{Approved: true})), nil)

	repos, ok := deps.Completion.Repositories.(*inmemory.RepositoryProvider)
	if !ok {
		t.Fatalf("deps.Completion.Repositories = %T, want *inmemory.RepositoryProvider", deps.Completion.Repositories)
	}
	repos.MergeErr = errors.New("merge conflict")

	rec := doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/"+projectID+"/tasks/"+taskID+"/complete-testing",
		jsonBody(t, completeTestingRequest{Passed: true, Repository: "github.com/org/repo", PullRequestID: "42"})), nil)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}

	var view taskViewResponse
	doRequest(t, server, httptest.NewRequest(http.MethodGet, "/projects/"+projectID+"/tasks/"+taskID, nil), &view)
	if view.State != "testing" {
		t.Errorf("State = %q, want testing (Done must be unreachable without a successful merge)", view.State)
	}
}
