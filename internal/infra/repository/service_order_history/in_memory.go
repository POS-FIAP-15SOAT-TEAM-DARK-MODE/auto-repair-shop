package service_order_history

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
)

type memory_repo struct {
	data map[string]domain.ServiceOrder
}

func MemoryRepository() domain.ServiceOrderHistoryRepository {
	return &memory_repo{
		data: nil,
	}
}

func (r *memory_repo) Search(_ context.Context, params *domain.SearchServiceOrderHistoryParams) ([]domain.ServiceOrderHistoryItem, error) {
	return []domain.ServiceOrderHistoryItem{}, nil
}

func (r *memory_repo) SearchWorkTransitionsByServiceOrderID(_ context.Context, _ string) ([]domain.WorkTransitionGroup, error) {
	return []domain.WorkTransitionGroup{}, nil
}

func (r *memory_repo) InsertWorkHistory(_ context.Context, _, _ string, _ domain.SERVICE_ORDER_STATUS) error {
	return nil
}
