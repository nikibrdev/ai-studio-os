package httpapi

import (
	"errors"
	"net/http"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/domain/artifact"
	"ai-studio-os/internal/domain/execution"
	"ai-studio-os/internal/domain/project"
	"ai-studio-os/internal/domain/task"
	"ai-studio-os/internal/domain/workflow"
)

// badRequestErrors and conflictErrors implement the single error->HTTP
// status convention docs/api/README.md declares once for every operation
// — a handler never picks a status code itself, it calls writeError.
//
// Referencing these domain sentinel values here is a narrow, deliberate
// exception to apps/api depending only on internal/application
// (module-boundaries.md): internal/application's use-case methods return
// them unwrapped by design (EPIC-004), so any caller distinguishing error
// kinds already needs to know them. No domain business logic is invoked.
var badRequestErrors = []error{
	project.ErrMissingField,
	task.ErrMissingField,
	artifact.ErrMissingField,
	execution.ErrMissingField,
	artifact.ErrPayloadRequired,
}

var conflictErrors = []error{
	project.ErrArchived,
	project.ErrAlreadyActive,
	project.ErrNoRepository,
	application.ErrProjectNotActive,
	application.ErrExecutorNotAssignable,
	application.ErrExecutionNotRunning,
	workflow.ErrTransitionNotAllowed,
	artifact.ErrArchived,
	artifact.ErrPublished,
	execution.ErrNotQueued,
	execution.ErrNotRunning,
	execution.ErrTerminal,
}

// writeError maps err to its HTTP status (docs/api/README.md) and writes
// a JSON error body: {"error": "<message>"}.
func writeError(w http.ResponseWriter, err error) {
	writeJSON(w, statusFor(err), map[string]string{"error": err.Error()})
}

func statusFor(err error) int {
	if errors.Is(err, application.ErrNotFound) {
		return http.StatusNotFound
	}
	for _, e := range badRequestErrors {
		if errors.Is(err, e) {
			return http.StatusBadRequest
		}
	}
	for _, e := range conflictErrors {
		if errors.Is(err, e) {
			return http.StatusConflict
		}
	}
	return http.StatusInternalServerError
}
