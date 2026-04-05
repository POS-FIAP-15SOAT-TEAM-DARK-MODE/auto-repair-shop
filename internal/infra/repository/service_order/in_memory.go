package service_order

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
)

type memory_repo struct {
	data map[string]domain.ServiceOrder
}

func MemoryRepository() domain.ServiceOrderRepository {
	return &memory_repo{
		data: make(map[string]domain.ServiceOrder),
	}
}

func (r *memory_repo) Save(_ context.Context, so *domain.ServiceOrder) error {
	r.data[so.ID] = *so
	return nil
}

func (r *memory_repo) GetHistoryByID(_ context.Context, id string) ([]domain.ServiceOrderHistory, error) {
	return nil, nil
}
