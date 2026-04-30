package work_service_order_history

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
)

type memoryRepo struct{}

func MemoryRepository() domain.WorkServiceOrderHistoryRepository {
	return &memoryRepo{}
}

func (r *memoryRepo) Insert(_ context.Context, _, _ string, _ *domain.WORK_SERVICE_ORDER_STATUS, _ domain.WORK_SERVICE_ORDER_STATUS) error {
	return nil
}

func (r *memoryRepo) Search(_ context.Context, _ domain.SearchWorkSOHistoryParams) ([]domain.WorkServiceOrderHistory, error) {
	return nil, nil
}
