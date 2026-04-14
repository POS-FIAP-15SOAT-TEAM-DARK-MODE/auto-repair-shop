package domain

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
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
	SupplyUpdate struct {
		ID            string
		Name          *string // ponteiro — nil = não enviado
		Description   *string
		UnitPrice     *decimal.Decimal
		StockQuantity *int
	}
)

type SupplyService interface {
	Create(ctx context.Context, req *Supply) error
	List(ctx context.Context) ([]*Supply, error)
	Update(ctx context.Context, req *SupplyUpdate) error
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=SupplyService --with-expecter
//go:generate go run github.com/vektra/mockery/v2@latest --name=SupplyRepository --with-expecter
type SupplyRepository interface {
	Create(ctx context.Context, c *Supply) error
	List(ctx context.Context) ([]*Supply, error)
	Update(ctx context.Context, c *SupplyUpdate) error
	GetByID(ctx context.Context, id string) (*Supply, error) // ← adiciona
}

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
func UpdateSupply(id string, name *string, description *string, unitPrice *decimal.Decimal, stockQuantity *int) *SupplyUpdate {
	return &SupplyUpdate{
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

func (s *SupplyUpdate) Validate() error {
	var errs []error
	if s.Name != nil && len(*s.Name) < 3 {
		errs = append(errs, ErrInvalidSupplyName)
	}
	if s.Description != nil && len(*s.Description) < 10 {
		errs = append(errs, ErrInvalidSupplyDescription)
	}
	if s.UnitPrice != nil && s.UnitPrice.LessThanOrEqual(decimal.Zero) {
		errs = append(errs, ErrInvalidSupplyUnitPrice)
	}
	if s.StockQuantity != nil && *s.StockQuantity < 0 {
		errs = append(errs, ErrInvalidSupplyStockQuantity)
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
