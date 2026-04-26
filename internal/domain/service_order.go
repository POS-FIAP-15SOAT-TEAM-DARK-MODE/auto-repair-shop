package domain

import (
	"context"
	"sync"

	"github.com/oklog/ulid/v2"
	"github.com/shopspring/decimal"
)

type SERVICE_ORDER_STATUS string

const (
	SERVICE_ORDER_STATUS_NEW               SERVICE_ORDER_STATUS = "NEW"
	SERVICE_ORDER_STATUS_RECEIVED          SERVICE_ORDER_STATUS = "RECEIVED"
	SERVICE_ORDER_STATUS_IN_DIAGNOSIS      SERVICE_ORDER_STATUS = "IN_DIAGNOSIS"
	SERVICE_ORDER_STATUS_AWAITING_APPROVAL SERVICE_ORDER_STATUS = "AWAITING_APPROVAL"
	SERVICE_ORDER_STATUS_REJECTED          SERVICE_ORDER_STATUS = "REJECTED"
	SERVICE_ORDER_STATUS_IN_PROGRESS       SERVICE_ORDER_STATUS = "IN_PROGRESS"
	SERVICE_ORDER_STATUS_COMPLETED         SERVICE_ORDER_STATUS = "COMPLETED"
	SERVICE_ORDER_STATUS_DELIVERED         SERVICE_ORDER_STATUS = "DELIVERED"
	SERVICE_ORDER_STATUS_CANCELLED         SERVICE_ORDER_STATUS = "CANCELLED"
)

func (s SERVICE_ORDER_STATUS) String() string {
	return string(s)
}

func (s SERVICE_ORDER_STATUS) IsCancelable() bool {
	return s != SERVICE_ORDER_STATUS_REJECTED &&
		s != SERVICE_ORDER_STATUS_COMPLETED &&
		s != SERVICE_ORDER_STATUS_DELIVERED &&
		s != SERVICE_ORDER_STATUS_CANCELLED
}

func StringToServiceOrderStatus(val string) SERVICE_ORDER_STATUS {
	return SERVICE_ORDER_STATUS(val)
}

type (
	ServiceOrder struct {
		ID          string
		Status      SERVICE_ORDER_STATUS
		Customer    *Customer
		Vehicle     *Vehicle
		Services    []Work
		Supplies    []Supply
		TotalAmount decimal.Decimal
		sumLocker   sync.Locker
	}

	AddSupply struct {
		ID     string
		Amount int
	}
)

func (so *ServiceOrder) ResetPricing() {
	so.PrepareForSum()

	so.sumLocker.Lock()
	so.TotalAmount = decimal.Zero
	so.sumLocker.Unlock()
}

func (so *ServiceOrder) PrepareForSum() {
	sync.OnceFunc(func() {
		if so.sumLocker == nil {
			so.sumLocker = &sync.Mutex{}
		}
	})()
}

func (so *ServiceOrder) SumWorkValue(work Work) {
	so.PrepareForSum()

	so.sumLocker.Lock()
	so.TotalAmount = so.TotalAmount.Add(work.Price)
	so.sumLocker.Unlock()
}

func (so *ServiceOrder) SumSupplyValue(supply Supply) {
	so.PrepareForSum()

	supplyAmount := decimal.NewFromInt(int64(supply.StockQuantity))
	supplyPrice := supply.UnitPrice.Mul(supplyAmount)
	so.sumLocker.Lock()
	so.TotalAmount = so.TotalAmount.Add(supplyPrice)
	so.sumLocker.Unlock()
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=ServiceOrderService --with-expecter
//go:generate go run github.com/vektra/mockery/v2@latest --name=ServiceOrderRepository --with-expecter
type (
	ServiceOrderService interface {
		Create(ctx context.Context, customerId, vehicleId string) (ServiceOrder, error)
		ListWorks(ctx context.Context, serviceOrderID string) ([]Work, error)
		AddWorks(ctx context.Context, serviceOrderID string, workIDs []string) error
		RemoveWork(ctx context.Context, serviceOrderID, workID string) error
		ListSupplies(ctx context.Context, serviceOrderID string) ([]Supply, error)
		AddSupplies(ctx context.Context, serviceOrderID string, supplies []AddSupply) error
		RemoveSupply(ctx context.Context, serviceOrderID, supplyID string) error
		SendToCustomerApproval(ctx context.Context, serviceOrderID string) error
		Accept(ctx context.Context, serviceOrderID, userID string) error
		Reject(ctx context.Context, serviceOrderID, userID string) error
		Deliver(ctx context.Context, serviceOrderID string) error
		Cancel(ctx context.Context, serviceOrderID string) error
		AverageExecutionTime(ctx context.Context) (float64, error)
	}

	ServiceOrderRepository interface {
		Save(context.Context, *ServiceOrder) error
		ExistsByID(ctx context.Context, id string) (bool, SERVICE_ORDER_STATUS, error)
		FindByID(ctx context.Context, id string) (ServiceOrder, error)
		ListWorksByServiceOrderID(ctx context.Context, serviceOrderID string) ([]Work, error)
		AddWorkLink(ctx context.Context, serviceOrderID, workID string, unitPrice decimal.Decimal) error
		RemoveWorkLink(ctx context.Context, serviceOrderID, workID string) error
		ListSuppliesByServiceOrderID(ctx context.Context, serviceOrderID string) ([]Supply, error)
		AddSupplyLink(ctx context.Context, serviceOrderID, supplyID string, amount int, unitPrice decimal.Decimal) error
		RemoveSupplyLink(ctx context.Context, serviceOrderID, supplyID string) (int, error)
		AverageExecutionTimeInHours(ctx context.Context) (float64, error)
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

func GetPreviousStatus(currentStatus SERVICE_ORDER_STATUS) *SERVICE_ORDER_STATUS {
	var prev SERVICE_ORDER_STATUS
	switch currentStatus {
	case SERVICE_ORDER_STATUS_RECEIVED:
		prev = SERVICE_ORDER_STATUS_NEW
	case SERVICE_ORDER_STATUS_IN_DIAGNOSIS:
		prev = SERVICE_ORDER_STATUS_RECEIVED
	case SERVICE_ORDER_STATUS_AWAITING_APPROVAL:
		prev = SERVICE_ORDER_STATUS_IN_DIAGNOSIS
	case SERVICE_ORDER_STATUS_IN_PROGRESS:
		prev = SERVICE_ORDER_STATUS_AWAITING_APPROVAL
	case SERVICE_ORDER_STATUS_COMPLETED:
		prev = SERVICE_ORDER_STATUS_IN_PROGRESS
	case SERVICE_ORDER_STATUS_DELIVERED:
		prev = SERVICE_ORDER_STATUS_COMPLETED
	case SERVICE_ORDER_STATUS_REJECTED:
		prev = SERVICE_ORDER_STATUS_AWAITING_APPROVAL
	default:
		return nil
	}
	return &prev
}

func NewHistoryServiceOrderID() string {
	return ulid.Make().String()
}
