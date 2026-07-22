package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSucceedExecution_Success(t *testing.T) {
	deps := testDeps()
	server := NewServer(deps)
	projectID := createActiveProject(t, server)
	taskID := createReadyTask(t, server, projectID)
	executionID := startExecution(t, server, deps, projectID, taskID)

	rec := doRequest(t, server, httptest.NewRequest(http.MethodPost, "/executions/"+executionID+"/succeed",
		jsonBody(t, executionActionRequest{ProjectID: projectID})), nil)
	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d, body = %s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
}

func TestSucceedExecution_AlreadyTerminalReturns409(t *testing.T) {
	deps := testDeps()
	server := NewServer(deps)
	projectID := createActiveProject(t, server)
	taskID := createReadyTask(t, server, projectID)
	executionID := startExecution(t, server, deps, projectID, taskID)
	doRequest(t, server, httptest.NewRequest(http.MethodPost, "/executions/"+executionID+"/succeed",
		jsonBody(t, executionActionRequest{ProjectID: projectID})), nil)

	rec := doRequest(t, server, httptest.NewRequest(http.MethodPost, "/executions/"+executionID+"/succeed",
		jsonBody(t, executionActionRequest{ProjectID: projectID})), nil)
	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d, body = %s", rec.Code, http.StatusConflict, rec.Body.String())
	}
}

func TestFailExecution_Success(t *testing.T) {
	deps := testDeps()
	server := NewServer(deps)
	projectID := createActiveProject(t, server)
	taskID := createReadyTask(t, server, projectID)
	executionID := startExecution(t, server, deps, projectID, taskID)

	rec := doRequest(t, server, httptest.NewRequest(http.MethodPost, "/executions/"+executionID+"/fail",
		jsonBody(t, executionActionRequest{ProjectID: projectID})), nil)
	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d, body = %s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
}

func TestFailExecution_UnknownReturns404(t *testing.T) {
	deps := testDeps()
	server := NewServer(deps)
	projectID := createActiveProject(t, server)

	rec := doRequest(t, server, httptest.NewRequest(http.MethodPost, "/executions/missing/fail",
		jsonBody(t, executionActionRequest{ProjectID: projectID})), nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
