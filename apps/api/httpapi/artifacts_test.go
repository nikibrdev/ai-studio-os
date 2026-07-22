package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecordDraftArtifact_Success(t *testing.T) {
	deps := testDeps()
	server := NewServer(deps)
	projectID := createActiveProject(t, server)
	taskID := createReadyTask(t, server, projectID)
	executionID := startExecution(t, server, deps, projectID, taskID)

	var got artifactResponse
	rec := doRequest(t, server, httptest.NewRequest(http.MethodPost, "/artifacts",
		jsonBody(t, recordDraftArtifactRequest{
			ID: "artifact-1", ProjectID: projectID, ExecutionID: executionID,
			Type: "code", Origin: "agent", Author: "claude-code",
		})), &got)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if got.ID != "artifact-1" || got.State != "draft" {
		t.Errorf("response = %+v", got)
	}
}

func TestRecordDraftArtifact_ExecutionNotRunningReturns409(t *testing.T) {
	deps := testDeps()
	server := NewServer(deps)
	projectID := createActiveProject(t, server)
	taskID := createReadyTask(t, server, projectID)
	executionID := startExecution(t, server, deps, projectID, taskID)
	doRequest(t, server, httptest.NewRequest(http.MethodPost, "/executions/"+executionID+"/succeed",
		jsonBody(t, executionActionRequest{ProjectID: projectID})), nil)

	rec := doRequest(t, server, httptest.NewRequest(http.MethodPost, "/artifacts",
		jsonBody(t, recordDraftArtifactRequest{
			ID: "artifact-1", ProjectID: projectID, ExecutionID: executionID,
			Type: "code", Origin: "agent", Author: "claude-code",
		})), nil)
	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d, body = %s", rec.Code, http.StatusConflict, rec.Body.String())
	}
}

func recordDraftArtifact(t *testing.T, server http.Handler, projectID, executionID string) string {
	t.Helper()
	var got artifactResponse
	rec := doRequest(t, server, httptest.NewRequest(http.MethodPost, "/artifacts",
		jsonBody(t, recordDraftArtifactRequest{
			ID: "artifact-1", ProjectID: projectID, ExecutionID: executionID,
			Type: "code", Origin: "agent", Author: "claude-code",
		})), &got)
	if rec.Code != http.StatusCreated {
		t.Fatalf("record draft artifact status = %d, body = %s", rec.Code, rec.Body.String())
	}
	return got.ID
}

func TestUpdateArtifactDraft_Success(t *testing.T) {
	deps := testDeps()
	server := NewServer(deps)
	projectID := createActiveProject(t, server)
	taskID := createReadyTask(t, server, projectID)
	executionID := startExecution(t, server, deps, projectID, taskID)
	artifactID := recordDraftArtifact(t, server, projectID, executionID)

	rec := doRequest(t, server, httptest.NewRequest(http.MethodPatch, "/artifacts/"+artifactID,
		jsonBody(t, updateArtifactDraftRequest{Payload: []byte("результат работы")})), nil)
	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d, body = %s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
}

func TestUpdateArtifactDraft_UnknownReturns404(t *testing.T) {
	server := NewServer(testDeps())

	rec := doRequest(t, server, httptest.NewRequest(http.MethodPatch, "/artifacts/missing",
		jsonBody(t, updateArtifactDraftRequest{Payload: []byte("x")})), nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestPublishArtifact_RequiresNonEmptyPayload(t *testing.T) {
	deps := testDeps()
	server := NewServer(deps)
	projectID := createActiveProject(t, server)
	taskID := createReadyTask(t, server, projectID)
	executionID := startExecution(t, server, deps, projectID, taskID)
	artifactID := recordDraftArtifact(t, server, projectID, executionID)

	rec := doRequest(t, server, httptest.NewRequest(http.MethodPost, "/artifacts/"+artifactID+"/publish", nil), nil)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("publish without payload status = %d, want %d, body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}

	doRequest(t, server, httptest.NewRequest(http.MethodPatch, "/artifacts/"+artifactID,
		jsonBody(t, updateArtifactDraftRequest{Payload: []byte("результат работы")})), nil)

	rec = doRequest(t, server, httptest.NewRequest(http.MethodPost, "/artifacts/"+artifactID+"/publish", nil), nil)
	if rec.Code != http.StatusNoContent {
		t.Errorf("publish with payload status = %d, want %d, body = %s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
}
