package domain

import "errors"

// ValidationError carries a human-readable message for domain validation failures (e.g. invalid CPF).
// It maps to HTTP 400 Bad Request.
type ValidationError struct {
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return e.Message
}

// Sentinel errors used by the repository and service layers.
// These map to HTTP status codes in pkg/web/errors.go.
var (
	ErrDataConflict  = errors.New("resource already exists")
	ErrDataViolation = errors.New("data violates business constraints")
	ErrInfraConflict = errors.New("something went wrong with our resources, try again")
)
