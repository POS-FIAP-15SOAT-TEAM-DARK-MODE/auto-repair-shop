package interfaces

import (
	"context"
	"net/http"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/service_order_history/adapters"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=ServiceOrderHistoryHTTPController --with-expecter
type ServiceOrderHistoryHTTPController interface {
	GetHistoryByID(ctx context.Context, req *http.Request) ([]adapters.SOHistoryResponse, error)
}
