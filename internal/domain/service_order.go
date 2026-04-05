package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type SERVICE_ORDER_STATUS string

const (
	SERVICE_ORDER_STATUS_NEW SERVICE_ORDER_STATUS = "NEW"
)

func (s SERVICE_ORDER_STATUS) String() string {
	return string(s)
}

type (
	ServiceOrder struct {
type ServiceOrder struct {
	ID          string
	Status      SERVICE_ORDER_STATUS
	Customer    *Customer
	Vehicle     *Vehicle
	Services    []Work
	Supplies    []Supply
	TotalAmount decimal.Decimal
}

	ServiceOrderHistory struct {
		ID             string
		PreviousStatus SERVICE_ORDER_STATUS
		NewStatus      SERVICE_ORDER_STATUS
		CreatedAt      time.Time
	}
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=ServiceOrderService --with-expecter
//go:generate go run github.com/vektra/mockery/v2@latest --name=ServiceOrderRepository --with-expecter
type (
	ServiceOrderService interface {
		Create(ctx context.Context, customerId, vehicleId string) (ServiceOrder, error)
		GetHistoryByID(ctx context.Context, id string) ([]ServiceOrderHistory, error)
	}
	ServiceOrderRepository interface {
		Save(context.Context, *ServiceOrder) error
		GetHistoryByID(ctx context.Context, id string) ([]ServiceOrderHistory, error)
	}
)

func NewServiceOrder(Customer *Customer, Vehicle *Vehicle) *ServiceOrder {
	return &ServiceOrder{
		ID:          uuid.NewString(),
		Status:      SERVICE_ORDER_STATUS_NEW,
		Customer:    Customer,
		Vehicle:     Vehicle,
		TotalAmount: decimal.Zero,
	}
}
