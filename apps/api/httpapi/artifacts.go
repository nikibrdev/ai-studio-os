package httpapi

import (
	"net/http"
	"time"

	"ai-studio-os/internal/application"
	"ai-studio-os/internal/domain/artifact"
)

// registerArtifactRoutes wires the three Artifact operations from
// docs/api/artifacts.md. Payload fields are []byte — encoding/json
// already (de)serializes []byte as a base64 string, matching the spec's
// "string (base64)" without any manual encoding here.
func registerArtifactRoutes(mux *http.ServeMux, deps Deps) {
	mux.HandleFunc("GET /artifacts/{id}", handleGetArtifact(deps))
	mux.HandleFunc("POST /artifacts", handleRecordDraftArtifact(deps))
	mux.HandleFunc("PATCH /artifacts/{id}", handleUpdateArtifactDraft(deps))
	mux.HandleFunc("POST /artifacts/{id}/publish", handlePublishArtifact(deps))
}

type recordDraftArtifactRequest struct {
	ID          string          `json:"id"`
	ProjectID   string          `json:"projectId"`
	ExecutionID string          `json:"executionId"`
	Type        artifact.Type   `json:"type"`
	Origin      artifact.Origin `json:"origin"`
	Author      artifact.Author `json:"author"`
	Payload     []byte          `json:"payload"`
	Actor       string          `json:"actor"`
}

type artifactResponse struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`
	Type      string `json:"type"`
	Origin    string `json:"origin"`
	Author    string `json:"author"`
	State     string `json:"state"`

	// Payload and CreatedAt are filled on read (TASK-100) — a human needs the
	// text of a QA report, not just its metadata. As a string, not base64: the
	// artifacts this serves are text, and encoding would only make them harder
	// to read. Omitted from creation responses, where the caller already has
	// the payload it just sent.
	Payload   string `json:"payload,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
}

func handleRecordDraftArtifact(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req recordDraftArtifactRequest
		if err := decodeOptionalJSON(r, &req); err != nil {
			writeInvalidBody(w, err)
			return
		}

		a, err := deps.Results.RecordDraftArtifact(r.Context(), application.RecordDraftArtifactParams{
			ID: req.ID, ProjectID: req.ProjectID, ExecutionID: req.ExecutionID,
			Type: req.Type, Origin: req.Origin, Author: req.Author, Payload: req.Payload, Actor: req.Actor,
		})
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, artifactResponse{
			ID: a.ID(), ProjectID: a.ProjectID(), Type: string(a.ArtifactType()),
			Origin: string(a.Origin()), Author: string(a.Author()), State: string(a.State()),
		})
	}
}

type updateArtifactDraftRequest struct {
	Payload []byte          `json:"payload"`
	Author  artifact.Author `json:"author"`
}

func handleUpdateArtifactDraft(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req updateArtifactDraftRequest
		if err := decodeOptionalJSON(r, &req); err != nil {
			writeInvalidBody(w, err)
			return
		}

		if err := deps.Results.UpdateArtifactDraft(r.Context(), r.PathValue("id"), req.Payload, req.Author); err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusNoContent, nil)
	}
}

func handlePublishArtifact(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req actorRequest
		if err := decodeOptionalJSON(r, &req); err != nil {
			writeInvalidBody(w, err)
			return
		}

		if err := deps.Results.PublishArtifact(r.Context(), r.PathValue("id"), req.Actor); err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusNoContent, nil)
	}
}

func handleGetArtifact(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a, err := deps.Results.Artifact(r.Context(), r.PathValue("id"))
		if err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, artifactResponse{
			ID:        a.ID(),
			ProjectID: a.ProjectID(),
			Type:      string(a.ArtifactType()),
			Origin:    string(a.Origin()),
			Author:    string(a.Author()),
			State:     string(a.State()),
			Payload:   string(a.Payload()),
			CreatedAt: a.CreatedAt().Format(time.RFC3339),
		})
	}
}
