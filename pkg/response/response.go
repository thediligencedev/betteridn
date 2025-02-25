package response

import (
	"encoding/json"
	"net/http"
)

// RespondWithError sends a JSON error response: { "message": <error> }
func RespondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	resp := map[string]string{"message": message}
	_ = json.NewEncoder(w).Encode(resp)
}

// RespondWithJSON encodes any payload as JSON with a given status code.
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	_ = json.NewEncoder(w).Encode(payload)
}
