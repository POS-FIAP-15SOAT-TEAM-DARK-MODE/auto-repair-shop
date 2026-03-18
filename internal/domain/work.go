package domain

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=WorkService --with-expecter
type WorkService interface {
	Create(context.Context, *Work) error
	List(context.Context, *ListWorkParams) (*PaginatorResponse[Work], error)
}

type ListWorkParams struct {
	PageSize int64
	Page     int64
	Status   string
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=WorkRepository --with-expecter
type WorkRepository interface {
	Save(context.Context, *Work) error
	Search(context.Context, *SearchWorkParams) ([]Work, error)
	Count(context.Context, *SearchWorkParams) (int64, error)
}

type SearchWorkParams struct {
	Limit  int64
	Offset int64
	Status string
}

func (l ListWorkParams) SearchWorkParams() *SearchWorkParams {
	return &SearchWorkParams{
		Limit:  l.PageSize,
		Offset: (l.Page - 1) * l.PageSize,
		Status: l.Status,
	}
}

const (
	nameMinSize        = 3
	descriptionMinSize = 10
)

func NewWork(name, description, rawPrice string, status WorkStatus) (*Work, error) {
	price, err := decimal.NewFromString(rawPrice)
	if err != nil {
		return nil, err
	}

	return &Work{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		Price:       price,
		Status:      status,
	}, nil
}

type Work struct {
	ID          string
	Name        string
	Description string
	Price       decimal.Decimal
	Status      WorkStatus
}

type WorkStatus bool

const (
	ACTIVE   WorkStatus = true
	INACTIVE WorkStatus = false

	ActiveString   = "ACTIVE"
	InactiveString = "INACTIVE"
)

func (s WorkStatus) String() string {
	if s {
		return ActiveString
	}
	return InactiveString
}

func (s WorkStatus) Bool() bool {
	if s {
		return bool(ACTIVE)
	}
	return bool(INACTIVE)
}

func BoolToWorkStatus(v bool) WorkStatus {
	if v {
		return ACTIVE
	}
	return INACTIVE
}

func StringToWorkStatus(v string) WorkStatus {
	if strings.ToUpper(v) == ActiveString {
		return ACTIVE
	}
	return INACTIVE
}

func ValidWorkStatusStringValue(v string) error {
	if v != ActiveString && v != InactiveString {
		return ErrInvalidWorkStatusValue
	}
	return nil
}

func (s *Work) Validate() error {
	var errs []error

	s.Name = strings.TrimSpace(s.Name)
	if s.Name == "" {
		errs = append(errs, ErrEmptyWorkName)
	} else if len(s.Name) < nameMinSize {
		errs = append(errs, ErrWorkNameShorterThenRequired)
	}

	s.Description = strings.TrimSpace(s.Description)
	if s.Description == "" {
		errs = append(errs, ErrEmptyWorkDescription)
	} else if len(s.Description) < descriptionMinSize {
		errs = append(errs, ErrWorkDescriptionShorterThenRequired)
	}

	if s.Price.LessThanOrEqual(decimal.Zero) {
		errs = append(errs, ErrWorkPriceLessThenOrEqualZero)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
