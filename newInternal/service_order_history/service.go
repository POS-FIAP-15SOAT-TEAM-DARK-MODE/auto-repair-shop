package service_order_history

import (
	"context"
	"fmt"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/pkg/logger"
	soDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/service_order/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/service_order_history/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/service_order_history/interfaces"
	"go.uber.org/zap"
)

type svc struct {
	repo interfaces.ServiceOrderHistoryRepository
}

func NewService(repo interfaces.ServiceOrderHistoryRepository) interfaces.ServiceOrderHistoryService {
	return &svc{repo: repo}
}

func (s *svc) GetHistoryByID(ctx context.Context, params *domain.SearchParams) ([]domain.ServiceOrderHistoryItem, error) {
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
		if items[i].NewStatus == soDomain.SERVICE_ORDER_STATUS_IN_PROGRESS {
			return true
		}
	}
	return false
}

func attachWorkTransitions(items []domain.ServiceOrderHistoryItem, groups []domain.WorkTransitionGroup) {
	for i := range items {
		if items[i].NewStatus == soDomain.SERVICE_ORDER_STATUS_IN_PROGRESS {
			items[i].WorkTransitions = groups
			break
		}
	}
}
