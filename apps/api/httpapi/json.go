package httpapi

import (
	"encoding/json"
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
