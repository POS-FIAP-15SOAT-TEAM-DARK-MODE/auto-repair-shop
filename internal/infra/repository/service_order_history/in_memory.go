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

func (r *memory_repo) Search(_ context.Context, params *domain.SearchServiceOrderHistoryParams) ([]domain.ServiceOrderHistory, error) {
	return []domain.ServiceOrderHistory{}, nil
}

func (r *memory_repo) SearchWorkTransitions(_ context.Context, _ string) ([]domain.WorkServiceOrderHistory, error) {
	return []domain.WorkServiceOrderHistory{}, nil
}
