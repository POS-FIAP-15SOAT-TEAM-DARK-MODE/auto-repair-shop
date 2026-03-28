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

// Generical errors
var (
	ErrDataConflict  = errors.New("resource already exists")
	ErrDataViolation = errors.New("data violates business constraints")
	ErrInfraConflict = errors.New("something went wrong with our resources, try again")
)

// User errors
var (
	ErrEmptyUserName          = errors.New("user name cannot be empty")
	ErrInvalidUserCredentials = errors.New("invalid user credentials")
	ErrInvalidUserName        = errors.New("the username must be longer than 3 characters and can only contain letters and spaces")
	ErrEmptyUserEmail         = errors.New("user email cannot be empty")
	ErrInvalidUserEmail       = errors.New("invalid user email format")
	ErrEmptyUserPassword      = errors.New("user password cannot be empty")
	ErrInvalidUserPassword    = errors.New("password must have at least 8 characters and contain at least one uppercase letter, one lowercase letter, one digit, and one special character")
	ErrUserPasswordDontMatch  = errors.New("user passwords do not match")
	ErrUserPasswordTooLong    = errors.New("user password too long")
	ErrUserPasswordTooShort   = errors.New("user password too short")
)

// Work errors
var (
	ErrEmptyWorkName                      = errors.New("work name can't be empty")
	ErrWorkNameShorterThenRequired        = errors.New("work name should have at least 3 letters")
	ErrEmptyWorkDescription               = errors.New("work description can't be empty")
	ErrWorkDescriptionShorterThenRequired = errors.New("work description should have at least 10 characters")
	ErrWorkPriceLessThenOrEqualZero       = errors.New("work price should be bigger then 0")
	ErrInvalidWorkPriceValue              = errors.New("invalid work price value")
	ErrInvalidWorkStatusValue             = errors.New("work status should be ACTIVE or INACTIVE")
	ErrInvalidWorkId                      = errors.New("invalid work id")
)

// Customer errors
var (
	ErrPhoneRequired            = errors.New("phone is required")
	ErrCompanyNameRequired      = errors.New("company_name is required for COMPANY type")
	ErrInvalidCustomerType      = errors.New("invalid customer type")
	ErrInvalidCPF               = errors.New("invalid CPF")
	ErrInvalidCNPJ              = errors.New("invalid CNPJ")
	ErrCPFLength                = errors.New("CPF must have 11 digits")
	ErrCNPJLength               = errors.New("CNPJ must have 14 characters")
	ErrCustomerNotFound         = errors.New("customer not found")
	ErrCustomerHasServiceOrders = errors.New("customer has associated service orders")
	ErrInvalidDocumentFormat    = ValidationError{Message: "document must be a valid CPF (11 digits) or CNPJ (14 characters)"}
)

// Vehicle errors
var (
	ErrVehicleInvalidPlate         = errors.New("invalid vehicle plate")
	ErrVehicleNoCustomerAssociated = errors.New("no customer associated to the vehicle")
	ErrRequiredVehicleBrand        = errors.New("brand of the vehicle is required")
	ErrRequiredVehicleModel        = errors.New("model of the vehicle is required")
	ErrParamVehicleYear            = errors.New("year of the vehicle is not valid")
	ErrVehicleNotFound             = errors.New("vehicle not found")
)
