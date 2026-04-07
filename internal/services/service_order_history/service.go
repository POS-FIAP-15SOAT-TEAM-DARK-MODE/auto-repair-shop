package service_order_history

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
)

type svc struct {
	repo domain.ServiceOrderHistoryRepository
}

func Service(
	repo domain.ServiceOrderHistoryRepository) *svc {
	return &svc{repo}
}

func (s *svc) GetHistoryByID(ctx context.Context, id string) ([]domain.ServiceOrderHistory, error) {
	return s.repo.Find(ctx, domain.FindServiceOrderHistoryParams{ID: id})
}
