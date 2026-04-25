package supply

import (
	"context"
	"fmt"
	"math"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow"
	"golang.org/x/sync/errgroup"
)

type service struct {
	uow  uow.Executor
	repo domain.SupplyRepository
}

func Service(uow uow.Executor, repo domain.SupplyRepository) domain.SupplyService {
	return &service{uow, repo}
}

func (s *service) Create(ctx context.Context, supply *domain.Supply) error {
	if err := s.uow.Execute(ctx, s.saveRepositoryStep(supply)); err != nil {
		logger.Of(ctx).Error(err)
		return err
	}
	return nil
}

func (s *service) Update(ctx context.Context, supply *domain.Supply) error {
	if err := s.uow.Execute(ctx, s.saveRepositoryStep(supply)); err != nil {
		logger.Of(ctx).Error(err)
		return err
	}
	return nil
}

func (s *service) saveRepositoryStep(supply *domain.Supply) func(context.Context) error {
	return func(ctx context.Context) error {
		if err := supply.Validate(); err != nil {
			err = fmt.Errorf("supply validation failed: %w", err)
			logger.Of(ctx).Error(err)
			return err
		}

		if err := s.repo.Save(ctx, supply); err != nil {
			return fmt.Errorf("repository.Save failed: %w", err)
		}
		return nil
	}
}

func (s *service) List(ctx context.Context, params *domain.ListSupplyParams) (*domain.PaginatorResponse[domain.Supply], error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}

	var total int64
	var items []domain.Supply
	var eg errgroup.Group

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
		return nil, fmt.Errorf("fail to list supplies: %w", err)
	}

	return &domain.PaginatorResponse[domain.Supply]{
		Items:      items,
		TotalItems: total,
		TotalPages: int64(math.Ceil(float64(total) / float64(params.PageSize))),
		Page:       params.Page,
		PageSize:   params.PageSize,
	}, nil
}
func (s *service) Delete(c context.Context, id string) error {
	if err := s.uow.Execute(c, func(ctx context.Context) error { return s.repo.Delete(ctx, id) }); err != nil {
		logger.Of(c).Error(err)
		return err
	}

	return nil
}
