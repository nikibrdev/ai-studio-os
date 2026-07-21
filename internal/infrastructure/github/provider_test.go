package github

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ai-studio-os/internal/platform"
)

func newTestProvider(t *testing.T, handler http.HandlerFunc) *Provider {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return &Provider{baseURL: server.URL, token: "test-token", client: server.Client()}
}

func writeJSON(t *testing.T, w http.ResponseWriter, status int, v any) {
	t.Helper()
	w.WriteHeader(status)
	if v == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Fatalf("encode response: %v", err)
	}
}

func TestNew_MissingToken(t *testing.T) {
	t.Setenv(TokenEnv, "")

	_, err := New()
	if !errors.Is(err, ErrTokenNotSet) {
		t.Fatalf("New() error = %v, want ErrTokenNotSet", err)
	}
}

func TestNew_UsesEnvToken(t *testing.T) {
	t.Setenv(TokenEnv, "env-token")

	p, err := New()
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}
	if p.token != "env-token" {
		t.Errorf("token = %q, want env-token", p.token)
	}
}

func TestDo_SendsAuthAndVersionHeaders(t *testing.T) {
	p := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization header = %q, want Bearer test-token", got)
		}
		if got := r.Header.Get("Accept"); got != "application/vnd.github+json" {
			t.Errorf("Accept header = %q", got)
		}
		if got := r.Header.Get("X-GitHub-Api-Version"); got != "2022-11-28" {
			t.Errorf("X-GitHub-Api-Version header = %q", got)
		}
		writeJSON(t, w, http.StatusOK, map[string]any{"merged": false, "state": "open"})
	})

	if _, err := p.PullRequestState(context.Background(), "org/repo", "1"); err != nil {
		t.Fatalf("PullRequestState: %v", err)
	}
}

func TestCreateBranch_Success(t *testing.T) {
	var gotRefBody map[string]string
	p := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/repos/org/repo/git/ref/heads/main":
			writeJSON(t, w, http.StatusOK, map[string]any{
				"object": map[string]string{"sha": "abc123"},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/repos/org/repo/git/refs":
			if err := json.NewDecoder(r.Body).Decode(&gotRefBody); err != nil {
				t.Fatalf("decode request body: %v", err)
			}
			writeJSON(t, w, http.StatusCreated, nil)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})

	if err := p.CreateBranch(context.Background(), "org/repo", "feature/x", "main"); err != nil {
		t.Fatalf("CreateBranch: %v", err)
	}
	if gotRefBody["ref"] != "refs/heads/feature/x" || gotRefBody["sha"] != "abc123" {
		t.Errorf("create ref request body = %+v", gotRefBody)
	}
}

func TestCreateBranch_BaseNotFound(t *testing.T) {
	p := newTestProvider(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, http.StatusNotFound, map[string]string{"message": "Not Found"})
	})

	err := p.CreateBranch(context.Background(), "org/repo", "feature/x", "does-not-exist")
	if err == nil {
		t.Fatal("CreateBranch() with a missing base branch: expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("CreateBranch() error = %v, want wrapping *APIError with status 404", err)
	}
}

func TestOpenPullRequest_Success(t *testing.T) {
	var gotBody map[string]string
	p := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/repos/org/repo/pulls" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		writeJSON(t, w, http.StatusCreated, map[string]int{"number": 42})
	})

	id, err := p.OpenPullRequest(context.Background(), "org/repo", "feature/x", "Title", "Body")
	if err != nil {
		t.Fatalf("OpenPullRequest: %v", err)
	}
	if id != "42" {
		t.Errorf("OpenPullRequest() = %q, want 42", id)
	}
	if gotBody["base"] != "main" || gotBody["head"] != "feature/x" {
		t.Errorf("open PR request body = %+v, want base=main head=feature/x", gotBody)
	}
}

func TestOpenPullRequest_ValidationFailure(t *testing.T) {
	p := newTestProvider(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, http.StatusUnprocessableEntity, map[string]string{"message": "Validation Failed"})
	})

	_, err := p.OpenPullRequest(context.Background(), "org/repo", "feature/x", "Title", "Body")
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("OpenPullRequest() error = %v, want wrapping *APIError with status 422", err)
	}
}

func TestRequestReview_Success(t *testing.T) {
	p := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/repos/org/repo/issues/7/comments" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		writeJSON(t, w, http.StatusCreated, nil)
	})

	if err := p.RequestReview(context.Background(), "org/repo", "7"); err != nil {
		t.Fatalf("RequestReview: %v", err)
	}
}

func TestRequestReview_PRNotFound(t *testing.T) {
	p := newTestProvider(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, http.StatusNotFound, map[string]string{"message": "Not Found"})
	})

	err := p.RequestReview(context.Background(), "org/repo", "999")
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("RequestReview() error = %v, want wrapping *APIError with status 404", err)
	}
}

func TestMergePullRequest_Success(t *testing.T) {
	var gotBody map[string]string
	p := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/repos/org/repo/pulls/7/merge" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		writeJSON(t, w, http.StatusOK, map[string]bool{"merged": true})
	})

	if err := p.MergePullRequest(context.Background(), "org/repo", "7"); err != nil {
		t.Fatalf("MergePullRequest: %v", err)
	}
	if gotBody["merge_method"] != "merge" {
		t.Errorf("merge request body = %+v, want merge_method=merge (ADR-008)", gotBody)
	}
}

func TestMergePullRequest_NotMergeable(t *testing.T) {
	p := newTestProvider(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, http.StatusMethodNotAllowed, map[string]string{"message": "Pull Request is not mergeable"})
	})

	err := p.MergePullRequest(context.Background(), "org/repo", "7")
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("MergePullRequest() error = %v, want wrapping *APIError with status 405", err)
	}
}

func TestClosePullRequest_Success(t *testing.T) {
	var gotBody map[string]string
	p := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/repos/org/repo/pulls/7" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		writeJSON(t, w, http.StatusOK, nil)
	})

	if err := p.ClosePullRequest(context.Background(), "org/repo", "7"); err != nil {
		t.Fatalf("ClosePullRequest: %v", err)
	}
	if gotBody["state"] != "closed" {
		t.Errorf("close request body = %+v, want state=closed", gotBody)
	}
}

func TestPullRequestState(t *testing.T) {
	tests := []struct {
		name   string
		merged bool
		state  string
		want   platform.PullRequestState
	}{
		{"open", false, "open", platform.PullRequestOpen},
		{"merged", true, "closed", platform.PullRequestMerged},
		{"closed_not_merged", false, "closed", platform.PullRequestClosed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newTestProvider(t, func(w http.ResponseWriter, _ *http.Request) {
				writeJSON(t, w, http.StatusOK, map[string]any{"merged": tt.merged, "state": tt.state})
			})

			got, err := p.PullRequestState(context.Background(), "org/repo", "7")
			if err != nil {
				t.Fatalf("PullRequestState: %v", err)
			}
			if got != tt.want {
				t.Errorf("PullRequestState() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPullRequestState_NotFound(t *testing.T) {
	p := newTestProvider(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, http.StatusNotFound, map[string]string{"message": "Not Found"})
	})

	_, err := p.PullRequestState(context.Background(), "org/repo", "999")
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("PullRequestState() error = %v, want wrapping *APIError with status 404", err)
	}
}

func TestAPIError_MessageIncludesContext(t *testing.T) {
	err := &APIError{Method: http.MethodGet, Path: "/repos/org/repo/pulls/7", StatusCode: 404, Body: `{"message":"Not Found"}`}
	msg := err.Error()
	if !strings.Contains(msg, "404") || !strings.Contains(msg, "/repos/org/repo/pulls/7") || !strings.Contains(msg, "Not Found") {
		t.Errorf("Error() = %q, want it to mention status, path and body", msg)
	}
}
