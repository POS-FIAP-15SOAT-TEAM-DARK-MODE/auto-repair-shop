package domain

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=ServiceService --with-expecter
type ServiceService interface {
	Create(context.Context, *Service) (*Service, error)
	List(context.Context, *ListServiceParams) (*PaginatorResponse[Service], error)
}

type ListServiceParams struct {
	PageSize int64
	Page     int64
	Status   string
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=ServiceRepository --with-expecter
type ServiceRepository interface {
	Save(context.Context, *Service) error
	Search(context.Context, *SearchServiceParams) ([]Service, error)
	Count(context.Context, *SearchServiceParams) (int64, error)
}

type SearchServiceParams struct {
	Limit  int64
	Offset int64
	Status string
}

func (l ListServiceParams) SearchServiceParams() *SearchServiceParams {
	return &SearchServiceParams{
		Limit:  l.PageSize,
		Offset: (l.Page - 1) * l.PageSize,
		Status: l.Status,
	}
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
	ErrInvalidPriceValue              = errors.New("invalid price value")
	ErrInvalidStatusValue             = errors.New("service status should be ACTIVE or INACTIVE")
)

func NewService(name, description, rawPrice string, status ServiceStatus) (*Service, error) {
	price, err := decimal.NewFromString(rawPrice)
	if err != nil {
		return nil, err
	}

	return &Service{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		Price:       price,
		Status:      status,
	}, nil
}

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

	if s.Price.LessThanOrEqual(decimal.Zero) {
		return ErrPriceLessThenOrEqualZero
	}

	return nil
}
