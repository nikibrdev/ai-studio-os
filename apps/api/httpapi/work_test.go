package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStartTask_Success(t *testing.T) {
	deps := testDeps()
	server := NewServer(deps)
	projectID := createActiveProject(t, server)
	taskID := createReadyTask(t, server, projectID)
	seedActiveDeveloperExecutor(t, deps, "executor-1")

	var got executionResponse
	rec := doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/"+projectID+"/tasks/"+taskID+"/start",
		jsonBody(t, startTaskRequest{ExecutorID: "executor-1", Actor: "pm:executor-2"})), &got)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if got.TaskID != taskID || got.ExecutorID != "executor-1" || got.State != "running" {
		t.Errorf("response = %+v", got)
	}
}

func TestStartTask_ExecutorNotAssignableReturns409(t *testing.T) {
	deps := testDeps()
	server := NewServer(deps)
	projectID := createActiveProject(t, server)
	taskID := createReadyTask(t, server, projectID)
	// No executor seeded: ExecutorID references nothing.

	rec := doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/"+projectID+"/tasks/"+taskID+"/start",
		jsonBody(t, startTaskRequest{ExecutorID: "missing"})), nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d, body = %s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestStartTask_UnknownTaskReturns404(t *testing.T) {
	deps := testDeps()
	server := NewServer(deps)
	projectID := createActiveProject(t, server)
	seedActiveDeveloperExecutor(t, deps, "executor-1")

	rec := doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/"+projectID+"/tasks/missing/start",
		jsonBody(t, startTaskRequest{ExecutorID: "executor-1"})), nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
