package repository

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/service_order/domain"
)

type wsoHistoryMemory struct{}

func NewWSOHistoryMemory() domain.WorkSOHistoryRepository {
	return &wsoHistoryMemory{}
}

func (r *wsoHistoryMemory) Insert(_ context.Context, _, _ string, _ *domain.WORK_SERVICE_ORDER_STATUS, _ domain.WORK_SERVICE_ORDER_STATUS) error {
	return nil
}

func (r *wsoHistoryMemory) Search(_ context.Context, _ domain.SearchWorkSOHistoryParams) ([]domain.WorkServiceOrderHistory, error) {
	return nil, nil
}
