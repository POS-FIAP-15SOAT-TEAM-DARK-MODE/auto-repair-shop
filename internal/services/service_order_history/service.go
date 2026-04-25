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

func (s *svc) GetHistoryByID(ctx context.Context, params *domain.SearchServiceOrderHistoryParams) (*domain.ServiceOrderHistoryResponse, error) {
	response, err := s.getHistoryResponse(ctx, params)
	if err != nil {
		logger.Of(ctx).Error(err,
			zap.String("operation", "get_service_order_history"),
			zap.String("entity", "service_order"),
			zap.String("entity_id", params.ID),
		)
		return nil, err
	}

	return response, nil
}

func (s *svc) getHistoryResponse(ctx context.Context, params *domain.SearchServiceOrderHistoryParams) (*domain.ServiceOrderHistoryResponse, error) {
	eg, egCtx := errgroup.WithContext(ctx)

	var (
		total int64
		items []domain.ServiceOrderHistory
		works []domain.WorkStatusTimeline
	)

	eg.Go(func() error {
		t, err := s.repo.Count(egCtx, params)
		if err != nil {
			return err
		}
		total = t
		return nil
	})

	eg.Go(func() error {
		list, err := s.repo.Search(egCtx, params)
		if err != nil {
			return err
		}
		items = list
		return nil
	})

	eg.Go(func() error {
		timeline, err := s.repo.WorkTimelineByServiceOrderID(egCtx, params.ID)
		if err != nil {
			return err
		}
		works = timeline
		return nil
	})

	if err := eg.Wait(); err != nil {
		return nil, fmt.Errorf("fail to get service order history: %w", err)
	}

	pageSize := params.PageSize
	totalPages := int64(0)
	if pageSize > 0 {
		totalPages = int64(math.Ceil(float64(total) / float64(pageSize)))
	}

	return &domain.ServiceOrderHistoryResponse{
		Page: domain.PaginatorResponse[domain.ServiceOrderHistory]{
			Items:      items,
			TotalItems: total,
			TotalPages: totalPages,
			Page:       params.Page,
			PageSize:   pageSize,
		},
		Works: works,
	}, nil
}
