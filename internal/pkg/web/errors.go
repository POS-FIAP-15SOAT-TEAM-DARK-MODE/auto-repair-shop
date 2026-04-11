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
	case isUnauthorizedError(err):
		return http.StatusUnauthorized
	case isBadRequestError(err), IsJoined(err):
		return http.StatusBadRequest
	case isNotFoundError(err):
		return http.StatusNotFound
	case isInternalServerError(err):
		return http.StatusInternalServerError
	case isConflictError(err):
		return http.StatusConflict
	case isUnprocessableEntityError(err):
		return http.StatusUnprocessableEntity
	case isNotFoundError(err):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

func IsJoined(err error) bool {
	type unwrapper interface {
		Unwrap() []error
	}

	var uw unwrapper
	return errors.As(err, &uw)
}

// buildErrorMessages formats the error messages based on status code.
func buildErrorMessages(err error, status int) ([]string, string) {
	switch status {
	case http.StatusUnauthorized, http.StatusBadRequest:
		// For 4XX BadRequest provide stack of errors.
		return unwrapAll(err), ""
	case http.StatusNotFound:
		return nil, err.Error()
	case http.StatusConflict, http.StatusUnprocessableEntity, http.StatusInternalServerError:
		// For 409/422 and 5XX unwrap recursively and return the innermost (root cause) error message.
		root := err
		for {
			unwrapped := errors.Unwrap(root)
			if unwrapped == nil {
				break
			}
			root = unwrapped
		}
		return nil, root.Error()
	default:
		// For any other case, generic error message.
		return nil, "Internal Server Error"
	}
}

func isBadRequestError(err error) bool {
	var validationErr domain.ValidationError
	return errors.As(err, &validationErr) ||
		errors.Is(err, domain.ErrUserPasswordDontMatch) ||
		errors.Is(err, domain.ErrUserPasswordTooLong) ||
		errors.Is(err, domain.ErrEmptyUserName) ||
		errors.Is(err, domain.ErrInvalidUserName) ||
		errors.Is(err, domain.ErrEmptyUserEmail) ||
		errors.Is(err, domain.ErrInvalidUserEmail) ||
		errors.Is(err, domain.ErrEmptyUserPassword) ||
		errors.Is(err, domain.ErrInvalidUserPassword) ||
		errors.Is(err, domain.ErrUserPasswordTooShort) ||
		errors.Is(err, domain.ErrEmptyWorkName) ||
		errors.Is(err, domain.ErrWorkNameShorterThenRequired) ||
		errors.Is(err, domain.ErrEmptyWorkDescription) ||
		errors.Is(err, domain.ErrWorkDescriptionShorterThenRequired) ||
		errors.Is(err, domain.ErrInvalidWorkId) ||
		errors.Is(err, domain.ErrInvalidServiceOrderId) ||
		errors.Is(err, domain.ErrEmptyServicesList) ||
		errors.Is(err, json.ErrJSONSyntax) ||
		errors.Is(err, json.ErrJSONType) ||
		errors.Is(err, json.ErrJSONUnexpectedEOF) ||
		errors.Is(err, json.ErrJSONEmptyBody) ||
		errors.Is(err, json.ErrWrongPayloadFormat) ||
		errors.Is(err, domain.ErrVehicleInvalidPlate) ||
		errors.Is(err, domain.ErrInvalidVehicleId) ||
		errors.Is(err, domain.ErrPhoneRequired) ||
		errors.Is(err, domain.ErrInvalidCustomerId)
}

func isNotFoundError(err error) bool {
	return errors.Is(err, domain.ErrCustomerNotFound) ||
		errors.Is(err, domain.ErrVehicleNotFound) ||
		errors.Is(err, domain.ErrWorkNotFound) ||
		errors.Is(err, domain.ErrServiceOrderNotFound) ||
		errors.Is(err, domain.ErrServiceOrderWorkNotFound)
}

func isUnauthorizedError(err error) bool {
	return errors.Is(err, domain.ErrInvalidUserCredentials)
}

func isInternalServerError(err error) bool {
	return errors.Is(err, json.ErrJSONUnsupportedValue) ||
		errors.Is(err, json.ErrJSONUnsupportedType) ||
		errors.Is(err, json.ErrJSONInvalidUnmarshal)
}

func isConflictError(err error) bool {
	return errors.Is(err, domain.ErrDataConflict) ||
		errors.Is(err, domain.ErrInfraConflict) ||
		errors.Is(err, domain.ErrCustomerHasServiceOrders)
}

func isUnprocessableEntityError(err error) bool {
	return errors.Is(err, domain.ErrDataViolation) ||
		errors.Is(err, domain.ErrWorkPriceLessThenOrEqualZero) ||
		errors.Is(err, domain.ErrInvalidWorkPriceValue) ||
		errors.Is(err, domain.ErrInvalidWorkStatusValue)
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
