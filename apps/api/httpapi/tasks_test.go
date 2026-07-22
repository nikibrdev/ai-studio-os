package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateTask_GeneratesSequentialID(t *testing.T) {
	server := NewServer(testDeps())
	projectID := createActiveProject(t, server)

	var first, second taskResponse
	rec := doRequest(t, server, httptest.NewRequest(http.MethodPost, "/tasks",
		jsonBody(t, createTaskRequest{ProjectID: projectID, Title: "Первая задача", Type: "feature"})), &first)
	if rec.Code != http.StatusCreated {
		t.Fatalf("first create status = %d, body = %s", rec.Code, rec.Body.String())
	}
	rec = doRequest(t, server, httptest.NewRequest(http.MethodPost, "/tasks",
		jsonBody(t, createTaskRequest{ProjectID: projectID, Title: "Вторая задача", Type: "feature"})), &second)
	if rec.Code != http.StatusCreated {
		t.Fatalf("second create status = %d, body = %s", rec.Code, rec.Body.String())
	}

	if first.ID != "TASK-001" || second.ID != "TASK-002" {
		t.Errorf("ids = %q, %q, want TASK-001, TASK-002", first.ID, second.ID)
	}
	if first.State != "backlog" {
		t.Errorf("State = %q, want backlog", first.State)
	}
}

func TestCreateTask_ProjectNotActiveReturns409(t *testing.T) {
	server := NewServer(testDeps())
	doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects",
		jsonBody(t, createProjectRequest{ID: "proj-1", Name: "AI Studio OS"})), nil)

	rec := doRequest(t, server, httptest.NewRequest(http.MethodPost, "/tasks",
		jsonBody(t, createTaskRequest{ProjectID: "proj-1", Title: "Задача", Type: "feature"})), nil)
	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d, body = %s", rec.Code, http.StatusConflict, rec.Body.String())
	}
}

func TestCreateTask_ProjectNotFoundReturns404(t *testing.T) {
	server := NewServer(testDeps())

	rec := doRequest(t, server, httptest.NewRequest(http.MethodPost, "/tasks",
		jsonBody(t, createTaskRequest{ProjectID: "missing", Title: "Задача", Type: "feature"})), nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestGetTask_UnknownReturns404(t *testing.T) {
	server := NewServer(testDeps())

	rec := doRequest(t, server, httptest.NewRequest(http.MethodGet, "/tasks/missing", nil), nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestGetTask_ReflectsCreatedState(t *testing.T) {
	server := NewServer(testDeps())
	projectID := createActiveProject(t, server)

	var created taskResponse
	doRequest(t, server, httptest.NewRequest(http.MethodPost, "/tasks",
		jsonBody(t, createTaskRequest{ProjectID: projectID, Title: "Задача", Type: "feature"})), &created)

	var view taskViewResponse
	rec := doRequest(t, server, httptest.NewRequest(http.MethodGet, "/tasks/"+created.ID, nil), &view)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if view.ID != created.ID || view.ProjectID != projectID || view.State != "backlog" {
		t.Errorf("view = %+v", view)
	}
}

func TestPlanTask_TransitionsToReady(t *testing.T) {
	server := NewServer(testDeps())
	projectID := createActiveProject(t, server)

	var created taskResponse
	doRequest(t, server, httptest.NewRequest(http.MethodPost, "/tasks",
		jsonBody(t, createTaskRequest{ProjectID: projectID, Title: "Задача", Type: "feature"})), &created)

	rec := doRequest(t, server, httptest.NewRequest(http.MethodPost, "/tasks/"+created.ID+"/plan",
		jsonBody(t, actorRequest{Actor: "pm:executor-1"})), nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("plan status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var view taskViewResponse
	doRequest(t, server, httptest.NewRequest(http.MethodGet, "/tasks/"+created.ID, nil), &view)
	if view.State != "ready" {
		t.Errorf("State = %q, want ready", view.State)
	}
}

func TestPlanTask_UnknownTaskReturns404(t *testing.T) {
	server := NewServer(testDeps())

	rec := doRequest(t, server, httptest.NewRequest(http.MethodPost, "/tasks/missing/plan", nil), nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
