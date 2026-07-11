package interfaces

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order_history/domain"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=ServiceOrderHistoryRepository --with-expecter
type ServiceOrderHistoryRepository interface {
	Search(ctx context.Context, params *domain.SearchParams) ([]domain.ServiceOrderHistoryItem, error)
	SearchWorkTransitionsByServiceOrderID(ctx context.Context, serviceOrderID string) ([]domain.WorkTransitionGroup, error)
}
