package domain

import (
	"context"
	"errors"
	"strconv"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func NewSupply(name, description string, unitPrice decimal.Decimal, stockQuantity, version int) *Supply {
	return &Supply{
		ID:            uuid.New().String(),
		Name:          name,
		Description:   description,
		UnitPrice:     unitPrice,
		StockQuantity: stockQuantity,
		Version:       version,
	}
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=SupplyService --with-expecter
//go:generate go run github.com/vektra/mockery/v2@latest --name=SupplyRepository --with-expecter
type (
	SupplyService interface {
		Create(ctx context.Context, c *Supply) error
		List(context.Context, *ListSupplyParams) (*PaginatorResponse[Supply], error)
		Update(context.Context, *Supply) error
		Delete(context.Context, string) error
	}

	SupplyRepository interface {
		Save(context.Context, *Supply) error
		Search(context.Context, *ListSupplyParams) ([]Supply, error)
		Count(context.Context, *ListSupplyParams) (int64, error)
		Delete(context.Context, string) error
		FindById(context.Context, string) (Supply, error)
		DecrementStock(ctx context.Context, id string, amount int) error
		RestoreStock(ctx context.Context, id string, amount int) error
	}
)

type (
	Supply struct {
		ID            string
		Name          string
		Description   string
		UnitPrice     decimal.Decimal
		StockQuantity int
		Version       int
	}
	ListSupplyParams struct {
		PageSize int64
		Page     int64
		Version  string
	}
)

func (l ListSupplyParams) Validate() error {
	if l.Version != "" {
		if _, err := strconv.Atoi(l.Version); err != nil {
			return ErrInvalidSupplyVersion
		}
	}
	return nil
}

func (l ListSupplyParams) Offset() int64 {
	return (l.Page - 1) * l.PageSize
}

func (s *Supply) IsValidName() error {
	if len(s.Name) < 3 {
		return ErrInvalidSupplyName
	}
	return nil
}

func (s *Supply) IsValidDescription() error {
	if len(s.Description) < 10 {
		return ErrInvalidSupplyDescription
	}
	return nil
}

func (s *Supply) IsValidUnitPrice() error {
	if s.UnitPrice.LessThanOrEqual(decimal.Zero) {
		return ErrInvalidSupplyUnitPrice
	}
	return nil
}

func (s *Supply) IsValidStockQuantity() error {
	if s.StockQuantity < 0 {
		return ErrInvalidSupplyStockQuantity
	}
	return nil
}

func (s *Supply) IsValidVersion() error {
	if s.Version < 0 {
		return ErrInvalidSupplyVersion
	}
	return nil
}

func (s *Supply) Validate() error {
	var errs []error
	if err := s.IsValidName(); err != nil {
		errs = append(errs, err)
	}
	if err := s.IsValidDescription(); err != nil {
		errs = append(errs, err)
	}
	if err := s.IsValidUnitPrice(); err != nil {
		errs = append(errs, err)
	}
	if err := s.IsValidStockQuantity(); err != nil {
		errs = append(errs, err)
	}
	if err := s.IsValidVersion(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
