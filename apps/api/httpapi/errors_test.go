package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/domain/artifact"
	"ai-studio-os/internal/domain/execution"
	"ai-studio-os/internal/domain/project"
	"ai-studio-os/internal/domain/task"
	"ai-studio-os/internal/domain/workflow"
)

func TestStatusFor(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"not found", application.ErrNotFound, http.StatusNotFound},
		{"wrapped not found", fmt.Errorf("get project: %w", application.ErrNotFound), http.StatusNotFound},
		{"project missing field", project.ErrMissingField, http.StatusBadRequest},
		{"task missing field", task.ErrMissingField, http.StatusBadRequest},
		{"artifact missing field", artifact.ErrMissingField, http.StatusBadRequest},
		{"execution missing field", execution.ErrMissingField, http.StatusBadRequest},
		{"artifact payload required", artifact.ErrPayloadRequired, http.StatusBadRequest},
		{"project archived", project.ErrArchived, http.StatusConflict},
		{"project already active", project.ErrAlreadyActive, http.StatusConflict},
		{"project no repository", project.ErrNoRepository, http.StatusConflict},
		{"project not active", application.ErrProjectNotActive, http.StatusConflict},
		{"executor not assignable", application.ErrExecutorNotAssignable, http.StatusConflict},
		{"execution not running (application)", application.ErrExecutionNotRunning, http.StatusConflict},
		{"transition not allowed", workflow.ErrTransitionNotAllowed, http.StatusConflict},
		{"artifact published", artifact.ErrPublished, http.StatusConflict},
		{"execution not queued", execution.ErrNotQueued, http.StatusConflict},
		{"execution not running (domain)", execution.ErrNotRunning, http.StatusConflict},
		{"execution terminal", execution.ErrTerminal, http.StatusConflict},
		{"unknown error", errors.New("boom"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := statusFor(tt.err); got != tt.want {
				t.Errorf("statusFor(%v) = %d, want %d", tt.err, got, tt.want)
			}
		})
	}
}

func TestWriteError_WritesStatusAndJSONBody(t *testing.T) {
	rec := httptest.NewRecorder()
	writeError(rec, application.ErrNotFound)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if body["error"] != application.ErrNotFound.Error() {
		t.Errorf("body[error] = %q, want %q", body["error"], application.ErrNotFound.Error())
	}
}
