package domain

import (
	"context"
	"errors"

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
		Change(context.Context, *Supply) error
	}

	SupplyRepository interface {
		Save(context.Context, *Supply) error
		Search(context.Context, *ListSupplyParams) ([]Supply, error)
		Count(context.Context, *ListSupplyParams) (int64, error)
		Change(context.Context, *Supply) error
		FindById(ctx context.Context, id string) (Supply, error)
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
		Status   string
	}
	UpdateSupply struct {
		ID            string
		Name          string
		Description   string
		UnitPrice     decimal.Decimal
		StockQuantity int
	}
)

func (l ListSupplyParams) Offset() int64 {
	return (l.Page - 1) * l.PageSize
}

func UpdateSupplyConstructor(id string, name string, description string, unitPrice decimal.Decimal, stockQuantity int) *UpdateSupply {
	return &UpdateSupply{
		ID:            id,
		Name:          name,
		Description:   description,
		UnitPrice:     unitPrice,
		StockQuantity: stockQuantity,
	}
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

func (s *UpdateSupply) Validate() error {
	var errs []error
	if s.Name != "" && len(s.Name) < 3 {
		errs = append(errs, ErrInvalidSupplyName)
	}
	if s.Description != "" && len(s.Description) < 10 {
		errs = append(errs, ErrInvalidSupplyDescription)
	}
	if s.UnitPrice != (decimal.Decimal{}) && s.UnitPrice.LessThanOrEqual(decimal.Zero) {
		errs = append(errs, ErrInvalidSupplyUnitPrice)
	}
	if s.StockQuantity != 0 && s.StockQuantity < 0 {
		errs = append(errs, ErrInvalidSupplyStockQuantity)
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
