package web

import (
	"errors"
	"net/http"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/json"
)

type errorResponse struct {
	StatusCode int      `json:"code"`
	Errors     []string `json:"errors,omitempty"`
	Error      string   `json:"error,omitempty"`
}

// Error builds an errorResponse for the given error and decides the status code.
func Error(err error) (int, errorResponse) {
	status := getHTTPStatus(err)
	stackedErrors, errorMsg := buildErrorMessages(err, status)

	return status, errorResponse{
		StatusCode: status,
		Errors:     stackedErrors,
		Error:      errorMsg,
	}
}

// getHTTPStatus determines the HTTP status code for a given error.
func getHTTPStatus(err error) int {
	switch {
	case isBadRequestError(err):
		return http.StatusBadRequest
	case isInternalServerError(err):
		return http.StatusInternalServerError
	case isConflictError(err):
		return http.StatusConflict
	case isUnprocessableEntityError(err):
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}

// buildErrorMessages formats the error messages based on status code.
func buildErrorMessages(err error, status int) ([]string, string) {
	switch status {
	case http.StatusBadRequest:
		// For 4XX BadRequest and 5XX Internal, provide stack of errors.
		return unwrapAll(err), ""
	case http.StatusConflict, http.StatusUnprocessableEntity, http.StatusInternalServerError:
		// For 409/422, return plain error message.
		return nil, err.Error()
	default:
		// For any other case, generic error message.
		return nil, "Internal Server Error"
	}
}

func isBadRequestError(err error) bool {
	return errors.Is(err, domain.ErrPasswordDontMatch) ||
		errors.Is(err, json.ErrJSONSyntax) ||
		errors.Is(err, json.ErrJSONType) ||
		errors.Is(err, json.ErrJSONUnexpectedEOF) ||
		errors.Is(err, json.ErrJSONEmptyBody) ||
		errors.Is(err, json.ErrWrongPayloadFormat) ||
		errors.Is(err, domain.ErrPasswordTooLong)
}

func isInternalServerError(err error) bool {
	return errors.Is(err, json.ErrJSONUnsupportedValue) ||
		errors.Is(err, json.ErrJSONUnsupportedType) ||
		errors.Is(err, json.ErrJSONInvalidUnmarshal)
}

func isConflictError(err error) bool {
	return errors.Is(err, domain.ErrDataConflict) ||
		errors.Is(err, domain.ErrInfraConflict)
}

func isUnprocessableEntityError(err error) bool {
	return errors.Is(err, domain.ErrDataViolation)
}

func unwrapAll(err error) []string {
	if err == nil {
		return nil
	}

	type multiUnwrap interface {
		Unwrap() []error
	}
	if m, ok := err.(multiUnwrap); ok {
		var all []string
		for _, e := range m.Unwrap() {
			all = append(all, unwrapAll(e)...)
		}
		return all
	}
	return []string{err.Error()}
}
