package service_order_history

import (
	"context"
	"fmt"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow"
	"go.uber.org/zap"
)

type svc struct {
	uow  uow.Executor
	repo domain.ServiceOrderHistoryRepository
}

func Service(uow uow.Executor, repo domain.ServiceOrderHistoryRepository) *svc {
	return &svc{uow, repo}
}

func (s *svc) GetHistoryByID(ctx context.Context, params *domain.SearchServiceOrderHistoryParams) ([]domain.ServiceOrderHistoryItem, error) {
	items, err := s.repo.Search(ctx, params)
	if err != nil {
		logger.Of(ctx).Error(err)
		logger.Of(ctx).Debug("Failed to fetch service order history",
			zap.String("operation", "get_service_order_history"),
			zap.Error(err),
			zap.String("entity", "service_order"),
		)
		return nil, fmt.Errorf("fail to fetch service order history: %w", err)
	}

	if !hasInProgressTransition(items) {
		return items, nil
	}

	transitions, err := s.repo.SearchWorkTransitionsByServiceOrderID(ctx, params.ID)
	if err != nil {
		logger.Of(ctx).Error(err)
		logger.Of(ctx).Debug("Failed to fetch work transitions for service order history",
			zap.String("operation", "get_service_order_history"),
			zap.Error(err),
			zap.String("entity", "service_order"),
		)
		return nil, fmt.Errorf("fail to fetch work transitions: %w", err)
	}

	attachWorkTransitions(items, transitions)
	return items, nil
}

func hasInProgressTransition(items []domain.ServiceOrderHistoryItem) bool {
	for i := range items {
		if items[i].NewStatus == domain.SERVICE_ORDER_STATUS_IN_PROGRESS {
			return true
		}
	}
	return false
}

func attachWorkTransitions(items []domain.ServiceOrderHistoryItem, groups []domain.WorkTransitionGroup) {
	for i := range items {
		if items[i].NewStatus == domain.SERVICE_ORDER_STATUS_IN_PROGRESS {
			items[i].WorkTransitions = groups
			break
		}
	}
}
