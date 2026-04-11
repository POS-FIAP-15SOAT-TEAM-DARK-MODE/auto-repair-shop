package domain

import (
	"context"

	"github.com/oklog/ulid/v2"
	"github.com/shopspring/decimal"
)

type SERVICE_ORDER_STATUS string

const (
	SERVICE_ORDER_STATUS_NEW SERVICE_ORDER_STATUS = "NEW"
)

func (s SERVICE_ORDER_STATUS) String() string {
	return string(s)
}

type ServiceOrder struct {
	ID          string
	Status      SERVICE_ORDER_STATUS
	Customer    *Customer
	Vehicle     *Vehicle
	Services    []Work
	Supplies    []Supply
	TotalAmount decimal.Decimal
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=ServiceOrderService --with-expecter
//go:generate go run github.com/vektra/mockery/v2@latest --name=ServiceOrderRepository --with-expecter
type (
	ServiceOrderService interface {
		Create(ctx context.Context, customerId, vehicleId string) (ServiceOrder, error)
		ListWorks(ctx context.Context, serviceOrderID string) ([]Work, error)
		AddWorks(ctx context.Context, serviceOrderID string, workIDs []string) error
		RemoveWork(ctx context.Context, serviceOrderID, workID string) error
	}
	ServiceOrderRepository interface {
		Save(context.Context, *ServiceOrder) error
		ExistsByID(ctx context.Context, id string) (bool, error)
		ListWorksByServiceOrderID(ctx context.Context, serviceOrderID string) ([]Work, error)
		AddWorkLink(ctx context.Context, serviceOrderID, workID string, unitPrice decimal.Decimal) error
		RemoveWorkLink(ctx context.Context, serviceOrderID, workID string) error
	}
)

func NewServiceOrder(Customer *Customer, Vehicle *Vehicle) *ServiceOrder {
	return &ServiceOrder{
		ID:          ulid.Make().String(),
		Status:      SERVICE_ORDER_STATUS_NEW,
		Customer:    Customer,
		Vehicle:     Vehicle,
		TotalAmount: decimal.Zero,
	}
}
