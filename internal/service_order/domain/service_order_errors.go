package domain

import (
	"errors"

	customerDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/customer/domain"
	supplyDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/supply/domain"
	workDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/work/domain"
)

// Errors aliased from other domains so callers can use SO domain as single import.
var (
	ErrCustomerNotFound        = customerDomain.ErrCustomerNotFound
	ErrInvalidCustomerProperty = customerDomain.ErrInvalidCustomerProperty
	ErrWorkNotFound            = workDomain.ErrWorkNotFound
	ErrInvalidWorkId           = workDomain.ErrInvalidWorkId
	ErrSupplyOutOfStock        = supplyDomain.ErrSupplyOutOfStock
	ErrSupplyNotFound          = supplyDomain.ErrSupplyNotFound
	ErrInvalidSupplyAmount     = supplyDomain.ErrInvalidSupplyAmount
	ErrInvalidSupplyID         = supplyDomain.ErrInvalidSupplyID
)

// Service order errors.
var (
	ErrInvalidServiceOrderId           = errors.New("invalid service order id")
	ErrServiceOrderNotFound            = errors.New("service order not found")
	ErrServiceOrderNotNew              = errors.New("service order isn't at NEW state")
	ErrServiceOrderNotInReceived       = errors.New("service order isn't in RECEIVED state")
	ErrServiceOrderNotInDiagnosis      = errors.New("service order isn't at IN_DIAGNOSIS state")
	ErrServiceOrderNotAwaitingApproval = errors.New("service order isn't at AWAITING_APPROVAL state")
	ErrServiceOrderNotInProgress       = errors.New("service order isn't at IN_PROGRESS state")
	ErrServiceOrderNotCompleted        = errors.New("service order isn't at COMPLETED state")
	ErrServiceOrderNotCancelable       = errors.New("service order state isn't cancelable")
	ErrServiceOrderWorkNotFound        = errors.New("work is not linked to this service order")
	ErrServiceOrderSupplyNotFound      = errors.New("supply is not linked to this service order")
	ErrEmptyServicesList               = errors.New("services list cannot be empty")
)

// Work-service-order history errors.
var (
	ErrWorkServiceOrderNotFound         = errors.New("work has no history in this service order")
	ErrWorkServiceOrderAlreadyCompleted = errors.New("work has already been completed")
	ErrWorkServiceOrderAlreadyCancelled = errors.New("work has already been cancelled")
	ErrWorkServiceOrderInvalidStatus    = errors.New("invalid work status transition")
)
