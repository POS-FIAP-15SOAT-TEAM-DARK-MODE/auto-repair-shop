package interfaces

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order_history/domain"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=ServiceOrderHistoryService --with-expecter
type ServiceOrderHistoryService interface {
	GetHistoryByID(ctx context.Context, params *domain.SearchParams) ([]domain.ServiceOrderHistoryItem, error)
}
