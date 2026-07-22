package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// writeJSON encodes v as the response body with the given status code.
// v may be nil for a bodyless response (e.g. 204 No Content).
func writeJSON(w http.ResponseWriter, status int, v any) {
	if v == nil {
		w.WriteHeader(status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// decodeJSON decodes the request body into v.
func decodeJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// decodeOptionalJSON decodes the request body into v, leaving v at its
// zero value if the body is empty — several operations in docs/api/*.md
// have an entirely optional body (e.g. only an optional "actor" field, or
// none at all for activate); domain validation still catches any field
// that turns out to be required (ErrMissingField etc.), so an empty body
// is not itself an error here.
func decodeOptionalJSON(r *http.Request, v any) error {
	err := decodeJSON(r, v)
	if errors.Is(err, io.EOF) {
		return nil
	}
	return err
}

// writeInvalidBody responds 400 for a request body that failed to
// decode — distinct from writeError's domain-error table, since a
// malformed body is not a domain error.
func writeInvalidBody(w http.ResponseWriter, err error) {
	writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body: " + err.Error()})
}
