package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateProject_ReturnsCreatedProject(t *testing.T) {
	server := NewServer(testDeps())

	var got projectResponse
	rec := doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects",
		jsonBody(t, createProjectRequest{ID: "proj-1", Name: "AI Studio OS"})), &got)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if got.ID != "proj-1" || got.Name != "AI Studio OS" || got.State != "created" {
		t.Errorf("response = %+v", got)
	}
}

func TestCreateProject_InvalidBodyReturns400(t *testing.T) {
	server := NewServer(testDeps())

	rec := doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects", strings.NewReader("not json")), nil)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestCreateProject_MissingNameReturns400(t *testing.T) {
	server := NewServer(testDeps())

	rec := doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects",
		jsonBody(t, createProjectRequest{ID: "proj-1"})), nil)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d, body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestConnectRepository_UnknownProjectReturns404(t *testing.T) {
	server := NewServer(testDeps())

	rec := doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/missing/repositories",
		jsonBody(t, connectRepositoryRequest{Repository: "github.com/org/repo"})), nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestActivateProject_NoRepositoryReturns409(t *testing.T) {
	server := NewServer(testDeps())
	doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects",
		jsonBody(t, createProjectRequest{ID: "proj-1", Name: "AI Studio OS"})), nil)

	rec := doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects/proj-1/activate", nil), nil)
	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d, body = %s", rec.Code, http.StatusConflict, rec.Body.String())
	}
}

func TestActivateProject_FullSequenceSucceeds(t *testing.T) {
	server := NewServer(testDeps())
	// createActiveProject itself asserts every step (create ->
	// connect-repository -> activate) succeeds — this test exists to name
	// that full sequence explicitly as a scenario.
	createActiveProject(t, server)
}

func TestListProjects_ReturnsCreatedProjects(t *testing.T) {
	server := NewServer(testDeps())
	doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects",
		jsonBody(t, createProjectRequest{ID: "proj-b", Name: "B"})), nil)
	doRequest(t, server, httptest.NewRequest(http.MethodPost, "/projects",
		jsonBody(t, createProjectRequest{ID: "proj-a", Name: "A"})), nil)

	var got []projectResponse
	rec := doRequest(t, server, httptest.NewRequest(http.MethodGet, "/projects", nil), &got)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if len(got) != 2 || got[0].ID != "proj-a" || got[1].ID != "proj-b" {
		t.Fatalf("response = %+v, want [proj-a, proj-b] ordered by id", got)
	}
}

func TestListProjects_EmptyReturnsEmptyArrayNotNull(t *testing.T) {
	server := NewServer(testDeps())

	rec := doRequest(t, server, httptest.NewRequest(http.MethodGet, "/projects", nil), nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if rec.Body.String() != "[]\n" {
		t.Errorf("body = %q, want an empty JSON array, not null", rec.Body.String())
	}
}
