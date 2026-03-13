package domain

import (
	"errors"
	"strings"

	"github.com/shopspring/decimal"
)

type ServiceService interface {
	Create(*Service) (*Service, error)
}

type ServiceRepository interface {
	Save(*Service) error
}

const (
	nameMinSize        = 3
	descriptionMinSize = 10
)

var (
	ErrEmptyName                      = errors.New("service name can't be empty")
	ErrNameShorterThenRequired        = errors.New("service name should have at least 3 letters")
	ErrEmptyDescription               = errors.New("service description can't be empty")
	ErrDescriptionShorterThenRequired = errors.New("service description should have at least 10 characters")
	ErrPriceLessThenOrEqualZero       = errors.New("service price should be bigger then 0")
	ErrInvalidStatusValue             = errors.New("service status should be ACTIVE or INACTIVE")
)

type Service struct {
	ID          string
	Name        string
	Description string
	Price       decimal.Decimal
	Status      ServiceStatus
}

type ServiceStatus bool

const (
	ACTIVE   ServiceStatus = true
	INACTIVE ServiceStatus = false

	ActiveString   = "ACTIVE"
	InactiveString = "INACTIVE"
)

func (s ServiceStatus) String() string {
	if s {
		return ActiveString
	}
	return InactiveString
}

func (s ServiceStatus) Bool() bool {
	if s {
		return bool(ACTIVE)
	}
	return bool(INACTIVE)
}

func BoolToServiceStatus(v bool) ServiceStatus {
	if v {
		return ACTIVE
	}
	return INACTIVE
}

func StringToServiceStatus(v string) ServiceStatus {
	if strings.ToUpper(v) == ActiveString {
		return ACTIVE
	}
	return INACTIVE
}

func ValidServiceStatusStringValue(v string) error {
	if v != ActiveString && v != InactiveString {
		return ErrInvalidStatusValue
	}
	return nil
}

func (s *Service) Validate() error {
	s.Name = strings.TrimSpace(s.Name)
	if s.Name == "" {
		return ErrEmptyName
	}

	if len(s.Name) < nameMinSize {
		return ErrNameShorterThenRequired
	}

	s.Description = strings.TrimSpace(s.Description)
	if s.Description == "" {
		return ErrEmptyDescription
	}

	if len(s.Description) < descriptionMinSize {
		return ErrDescriptionShorterThenRequired
	}

	if decimal.Zero.LessThanOrEqual(s.Price) {
		return ErrPriceLessThenOrEqualZero
	}

	return nil
}
