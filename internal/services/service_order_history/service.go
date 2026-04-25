package service_order_history

import (
	"context"
	"fmt"
	"math"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type svc struct {
	uow  uow.Executor
	repo domain.ServiceOrderHistoryRepository
}

func Service(uow uow.Executor, repo domain.ServiceOrderHistoryRepository) *svc {
	return &svc{uow, repo}
}

func (s *svc) GetHistoryByID(ctx context.Context, params *domain.SearchServiceOrderHistoryParams) (*domain.PaginatorResponse[domain.ServiceOrderHistory], error) {
	response, err := s.getPaginatedList(ctx, params)
	if err != nil {
		logger.Of(ctx).Error(err)
		logger.Of(ctx).Debug("Failed to get paginated list of service order histories",
			zap.String("operation", "get_service_order_history"),
			zap.Error(err),
			zap.String("entity", "service_order"),
		)
		return nil, err
	}

	return response, nil
}

func (s *svc) getPaginatedList(ctx context.Context, params *domain.SearchServiceOrderHistoryParams) (*domain.PaginatorResponse[domain.ServiceOrderHistory], error) {
	var eg errgroup.Group
	var total int64
	var items []domain.ServiceOrderHistory

	eg.Go(func() error {
		t, err := s.repo.Count(ctx, params)
		if err != nil {
			return err
		}
		total = t
		return nil
	})

	eg.Go(func() error {
		list, err := s.repo.Search(ctx, params)
		if err != nil {
			return err
		}
		items = list
		return nil
	})

	if err := eg.Wait(); err != nil {
		return nil, fmt.Errorf("fail to get paginated list of service order histories: %w", err)
	}

	return &domain.PaginatorResponse[domain.ServiceOrderHistory]{
		Items:      items,
		TotalItems: total,
		TotalPages: int64(math.Ceil(float64(total) / float64(params.PageSize))),
		Page:       params.Page,
		PageSize:   params.PageSize,
	}, nil
}
