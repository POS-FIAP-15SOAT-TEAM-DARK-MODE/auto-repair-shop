package web

import (
	"encoding/json"
	"net/http"
)

// RequestIDHeader carries the request correlation id through the gateway and
// into every upstream service.
const RequestIDHeader = "X-Request-Id"

// ErrorBody is the error envelope returned by the gateway, for example
// {"code":502,"errors":["upstream unavailable"]}.
type ErrorBody struct {
	Code   int      `json:"code"`
	Errors []string `json:"errors"`
}

// WriteError writes an ErrorBody with the given status and messages.
func WriteError(w http.ResponseWriter, status int, messages ...string) {
	WriteJSON(w, status, ErrorBody{Code: status, Errors: messages})
}

// WriteJSON writes body as JSON with the given status.
func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
