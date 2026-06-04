package web

import (
	"errors"
	"net/http"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/app"

	authDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/auth/domain"
	customerDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/customer/domain"
	jsonv2 "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/pkg/json"
	soDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/service_order/domain"
	soHistoryDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/service_order_history/domain"
	supplyDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/supply/domain"
	vehicleDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/vehicle/domain"
	workDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/work/domain"
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
	case isForbidden(err):
		return http.StatusForbidden
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
	case http.StatusUnauthorized, http.StatusForbidden, http.StatusBadRequest:
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
	return errors.Is(err, authDomain.ErrUserPasswordDontMatch) ||
		errors.Is(err, authDomain.ErrUserPasswordTooLong) ||
		errors.Is(err, authDomain.ErrEmptyUserName) ||
		errors.Is(err, authDomain.ErrInvalidUserName) ||
		errors.Is(err, authDomain.ErrEmptyUserEmail) ||
		errors.Is(err, authDomain.ErrInvalidUserEmail) ||
		errors.Is(err, authDomain.ErrEmptyUserPassword) ||
		errors.Is(err, authDomain.ErrInvalidUserPassword) ||
		errors.Is(err, authDomain.ErrUserPasswordTooShort) ||
		errors.Is(err, workDomain.ErrEmptyWorkName) ||
		errors.Is(err, workDomain.ErrWorkNameShorterThenRequired) ||
		errors.Is(err, workDomain.ErrEmptyWorkDescription) ||
		errors.Is(err, workDomain.ErrWorkDescriptionShorterThenRequired) ||
		errors.Is(err, workDomain.ErrInvalidWorkId) ||
		errors.Is(err, soDomain.ErrInvalidServiceOrderId) ||
		errors.Is(err, soDomain.ErrEmptyServicesList) ||
		errors.Is(err, soHistoryDomain.ErrServiceOrderIDRequired) ||
		errors.Is(err, supplyDomain.ErrInvalidSupplyID) ||
		errors.Is(err, supplyDomain.ErrInvalidSupplyAmount) ||
		errors.Is(err, jsonv2.ErrJSONSyntax) ||
		errors.Is(err, jsonv2.ErrJSONType) ||
		errors.Is(err, jsonv2.ErrJSONUnexpectedEOF) ||
		errors.Is(err, jsonv2.ErrJSONEmptyBody) ||
		errors.Is(err, jsonv2.ErrWrongPayloadFormat) ||
		errors.Is(err, vehicleDomain.ErrVehicleInvalidPlate) ||
		errors.Is(err, vehicleDomain.ErrRequiredVehicleBrand) ||
		errors.Is(err, vehicleDomain.ErrRequiredVehicleModel) ||
		errors.Is(err, vehicleDomain.ErrParamVehicleYear) ||
		errors.Is(err, vehicleDomain.ErrInvalidVehicleId) ||
		errors.Is(err, vehicleDomain.ErrInvalidSearchVehicleParams) ||
		errors.Is(err, customerDomain.ErrPhoneRequired) ||
		errors.Is(err, customerDomain.ErrCompanyNameRequired) ||
		errors.Is(err, customerDomain.ErrCompanyNameNotAllowed) ||
		errors.Is(err, customerDomain.ErrInvalidCustomerType) ||
		errors.Is(err, customerDomain.ErrInvalidDocumentFormat) ||
		errors.Is(err, customerDomain.ErrInvalidCustomerId)
}

func isNotFoundError(err error) bool {
	return errors.Is(err, workDomain.ErrWorkNotFound) ||
		errors.Is(err, supplyDomain.ErrSupplyNotFound) ||
		errors.Is(err, soDomain.ErrServiceOrderNotFound) ||
		errors.Is(err, soDomain.ErrServiceOrderWorkNotFound) ||
		errors.Is(err, soDomain.ErrServiceOrderSupplyNotFound) ||
		errors.Is(err, soDomain.ErrWorkServiceOrderNotFound) ||
		errors.Is(err, vehicleDomain.ErrVehicleNotFound) ||
		errors.Is(err, customerDomain.ErrCustomerNotFound)
}

func isUnauthorizedError(err error) bool {
	return errors.Is(err, authDomain.ErrInvalidUserCredentials)
}

func isInternalServerError(err error) bool {
	return errors.Is(err, jsonv2.ErrJSONUnsupportedValue) ||
		errors.Is(err, jsonv2.ErrJSONUnsupportedType) ||
		errors.Is(err, jsonv2.ErrJSONInvalidUnmarshal)
}

func isConflictError(err error) bool {
	return errors.Is(err, app.ErrDataConflict) ||
		errors.Is(err, app.ErrInfraConflict) ||
		errors.Is(err, customerDomain.ErrCustomerHasServiceOrders) ||
		errors.Is(err, soDomain.ErrServiceOrderNotNew) ||
		errors.Is(err, soDomain.ErrServiceOrderNotInReceived) ||
		errors.Is(err, soDomain.ErrServiceOrderNotInDiagnosis) ||
		errors.Is(err, soDomain.ErrServiceOrderNotInProgress) ||
		errors.Is(err, soDomain.ErrServiceOrderNotCompleted) ||
		errors.Is(err, soDomain.ErrServiceOrderNotCancelable) ||
		errors.Is(err, soDomain.ErrServiceOrderNotAwaitingApproval) ||
		errors.Is(err, soDomain.ErrWorkServiceOrderAlreadyCompleted) ||
		errors.Is(err, soDomain.ErrWorkServiceOrderAlreadyCancelled) ||
		errors.Is(err, soDomain.ErrWorkServiceOrderInvalidStatus)
}

func isUnprocessableEntityError(err error) bool {
	return errors.Is(err, app.ErrDataViolation) ||
		errors.Is(err, workDomain.ErrWorkPriceLessThenOrEqualZero) ||
		errors.Is(err, workDomain.ErrInvalidWorkPriceValue) ||
		errors.Is(err, workDomain.ErrInvalidWorkStatusValue) ||
		errors.Is(err, supplyDomain.ErrSupplyOutOfStock)
}

func isForbidden(err error) bool {
	return errors.Is(err, customerDomain.ErrInvalidCustomerProperty)
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
