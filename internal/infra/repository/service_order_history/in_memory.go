package service_order

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
)

type memory_repo struct {
	data map[string]domain.ServiceOrder
}

func MemoryRepository() domain.ServiceOrderHistoryRepository {
	return &memory_repo{
		data: make(map[string]domain.ServiceOrder),
	}
}

func (r *memory_repo) Find(_ context.Context, params domain.FindServiceOrderHistoryParams) ([]domain.ServiceOrderHistory, error) {
	return []domain.ServiceOrderHistory{}, nil
}
