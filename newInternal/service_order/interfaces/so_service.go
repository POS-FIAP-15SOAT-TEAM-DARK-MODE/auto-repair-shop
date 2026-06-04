package interfaces

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/service_order/adapters"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=ServiceOrderService --with-expecter
type ServiceOrderService interface {
	Create(ctx context.Context, req adapters.CreateSORequest) (adapters.SOResponse, error)
	List(ctx context.Context, params adapters.SOFilterParams) (adapters.PaginatedSOResponse, error)
	GetWorks(ctx context.Context, soID string) (adapters.ListResponse[adapters.WorkItemResponse], error)
	AddWorks(ctx context.Context, soID string, workIDs []string) error
	RemoveWork(ctx context.Context, soID, workID string) error
	GetSupplies(ctx context.Context, soID string) (adapters.ListResponse[adapters.SupplyItemResponse], error)
	AddSupplies(ctx context.Context, soID string, supplies []adapters.AddSupplyItem) error
	RemoveSupply(ctx context.Context, soID, supplyID string) error
	Receive(ctx context.Context, soID string) error
	SendToCustomerApproval(ctx context.Context, soID string) error
	SendToDiagnosis(ctx context.Context, soID string) error
	Finish(ctx context.Context, soID string) error
	Accept(ctx context.Context, soID, userID string) error
	Reject(ctx context.Context, soID, userID string) error
	Deliver(ctx context.Context, soID string) error
	Cancel(ctx context.Context, soID string) error
	GetFullByID(ctx context.Context, soID string) (adapters.SODetailResponse, error)
	GetAverageExecutionTime(ctx context.Context, workIDs []string) ([]adapters.WorkExecutionTimeResponse, error)
	NextWork(ctx context.Context, soID, workID string) error
	CancelWork(ctx context.Context, soID, workID string) error
	GetStatus(ctx context.Context, soID string) (adapters.SOResponse, error)
}
