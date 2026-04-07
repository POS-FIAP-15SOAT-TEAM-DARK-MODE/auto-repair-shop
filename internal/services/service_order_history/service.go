package service_order_history

import (
	"context"
	"fmt"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow"
)

type svc struct {
	uow  uow.Executor
	repo domain.ServiceOrderHistoryRepository
}

func Service(
	uow uow.Executor,
	repo domain.ServiceOrderHistoryRepository) *svc {
	return &svc{uow, repo}
}

func (s *svc) GetHistoryByID(ctx context.Context, id string) ([]domain.ServiceOrderHistory, error) {
	var soHistory []domain.ServiceOrderHistory
	err := s.uow.Execute(ctx, func(ctx context.Context) error {
		result, err := s.repo.Find(ctx, domain.FindServiceOrderHistoryParams{ID: id})
		if err != nil {
			return err
		}
		soHistory = result
		return nil
	})
	if err != nil {
		logger.Of(ctx).Error(fmt.Errorf("error get service order history: %w", err))
		return nil, err
	}

	return soHistory, nil
}
