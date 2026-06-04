package interfaces

import (
	"context"
	"net/http"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/service_order/adapters"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=ServiceOrderHTTPController --with-expecter
type ServiceOrderHTTPController interface {
	Create(ctx context.Context, req *http.Request) (adapters.SOResponse, error)
	List(ctx context.Context, req *http.Request) (adapters.PaginatedSOResponse, error)
	GetWorks(ctx context.Context, req *http.Request) (adapters.ListResponse[adapters.WorkItemResponse], error)
	AddWork(ctx context.Context, req *http.Request) error
	DeleteWork(ctx context.Context, req *http.Request) error
	GetSupplies(ctx context.Context, req *http.Request) (adapters.ListResponse[adapters.SupplyItemResponse], error)
	AddSupplies(ctx context.Context, req *http.Request) error
	DeleteSupply(ctx context.Context, req *http.Request) error
	Receive(ctx context.Context, req *http.Request) error
	SendToCustomerApproval(ctx context.Context, req *http.Request) error
	SendToDiagnosis(ctx context.Context, req *http.Request) error
	Finish(ctx context.Context, req *http.Request) error
	Accept(ctx context.Context, req *http.Request) error
	Reject(ctx context.Context, req *http.Request) error
	Deliver(ctx context.Context, req *http.Request) error
	Cancel(ctx context.Context, req *http.Request) error
	GetFullByID(ctx context.Context, req *http.Request) (adapters.SODetailResponse, error)
	GetAverageExecutionTime(ctx context.Context, req *http.Request) ([]adapters.WorkExecutionTimeResponse, error)
	NextWork(ctx context.Context, req *http.Request) error
	CancelWork(ctx context.Context, req *http.Request) error
	GetStatus(ctx context.Context, req *http.Request) (adapters.SOResponse, error)
}
