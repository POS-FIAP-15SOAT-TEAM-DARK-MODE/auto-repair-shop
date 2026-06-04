package domain

import (
	"context"
	"sync"

	customerDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/customer/domain"
	supplyDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/supply/domain"
	vehicleDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/vehicle/domain"
	workDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/work/domain"
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

type ServiceOrder struct {
	ID          string
	Status      SERVICE_ORDER_STATUS
	Customer    *customerDomain.Customer
	Vehicle     *vehicleDomain.Vehicle
	Services    []workDomain.Work
	Supplies    []supplyDomain.Supply
	TotalAmount decimal.Decimal
	sumLocker   sync.Locker
}

type AddSupply struct {
	ID     string
	Amount int
}

type FullServiceOrder struct {
	ServiceOrder
	Works    []workDomain.Work
	Supplies []supplyDomain.Supply
}

type ServiceOrderFilterParams struct {
	Page       int64
	PageSize   int64
	Limit      int64
	Offset     int64
	Status     string
	CustomerID string
	VehicleID  string
}

type WorkExecutionTime struct {
	WorkID       string
	WorkName     string
	AverageHours float64
}

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

func (so *ServiceOrder) SumWorkValue(work workDomain.Work) {
	so.PrepareForSum()

	so.sumLocker.Lock()
	so.TotalAmount = so.TotalAmount.Add(work.Price)
	so.sumLocker.Unlock()
}

func (so *ServiceOrder) SumSupplyValue(supply supplyDomain.Supply) {
	so.PrepareForSum()

	supplyAmount := decimal.NewFromInt(int64(supply.StockQuantity))
	supplyPrice := supply.UnitPrice.Mul(supplyAmount)
	so.sumLocker.Lock()
	so.TotalAmount = so.TotalAmount.Add(supplyPrice)
	so.sumLocker.Unlock()
}

func NewServiceOrder(customer *customerDomain.Customer, vehicle *vehicleDomain.Vehicle) *ServiceOrder {
	return &ServiceOrder{
		ID:          ulid.Make().String(),
		Status:      SERVICE_ORDER_STATUS_NEW,
		Customer:    customer,
		Vehicle:     vehicle,
		TotalAmount: decimal.Zero,
	}
}

func NewHistoryServiceOrderID() string {
	return ulid.Make().String()
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=ServiceOrderRepository --with-expecter
type ServiceOrderRepository interface {
	Save(context.Context, *ServiceOrder) error
	ExistsByID(ctx context.Context, id string) (bool, SERVICE_ORDER_STATUS, error)
	FindByID(ctx context.Context, id string) (ServiceOrder, error)
	Search(ctx context.Context, params *ServiceOrderFilterParams) ([]ServiceOrder, error)
	Count(ctx context.Context, params *ServiceOrderFilterParams) (int64, error)
	ListWorksByServiceOrderID(ctx context.Context, serviceOrderID string) ([]workDomain.Work, error)
	AddWorkLink(ctx context.Context, serviceOrderID, workID string, unitPrice decimal.Decimal) error
	RemoveWorkLink(ctx context.Context, serviceOrderID, workID string) error
	ListSuppliesByServiceOrderID(ctx context.Context, serviceOrderID string) ([]supplyDomain.Supply, error)
	AddSupplyLink(ctx context.Context, serviceOrderID, supplyID string, amount int, unitPrice decimal.Decimal) error
	RemoveSupplyLink(ctx context.Context, serviceOrderID, supplyID string) (int, error)
	AverageExecutionTimeInHours(ctx context.Context, workIDs []string) ([]WorkExecutionTime, error)
}
