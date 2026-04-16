package supply

import (
	"context"
	"fmt"
	"math"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow"
)

type service struct {
	uow  uow.Executor
	repo domain.SupplyRepository
}

func Service(uow uow.Executor, repo domain.SupplyRepository) domain.SupplyService {
	return &service{uow, repo}
}

func (s *service) Create(ctx context.Context, req *domain.Supply) error {
	if err := req.Validate(); err != nil {
		err := fmt.Errorf("supply validation failed: %w", err)
		logger.Of(ctx).Error(err)
		return err
	}

	if err := s.uow.Execute(ctx, s.saveRepositoryStep(req)); err != nil {
		logger.Of(ctx).Error(err)
		return err
	}
	return nil
}

func (s *service) saveRepositoryStep(supply *domain.Supply) func(context.Context) error {
	return func(ctx context.Context) error {
		if err := s.repo.Save(ctx, supply); err != nil {
			return fmt.Errorf("repository.Save failed: %w", err)
		}
		return nil
	}
}

func (s *service) List(c context.Context, params *domain.ListSupplyParams) (*domain.PaginatorResponse[domain.Supply], error) {
	var response *domain.PaginatorResponse[domain.Supply]
	if err := s.uow.Execute(c, func(ctx context.Context) error {
		r, err := s.getPaginatedList(ctx, params)
		if err != nil {
			return err
		}
		response = r
		return nil
	}); err != nil {
		logger.Of(c).Error(err)
		return nil, err
	}

	return response, nil
}

func (s *service) getPaginatedList(ctx context.Context, params *domain.ListSupplyParams) (*domain.PaginatorResponse[domain.Supply], error) {
	p := params.SearchSupplyParams()

	total, err := s.repo.Count(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("fail to count supplies: %w", err)
	}

	items, err := s.repo.Search(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("fail to search supplies: %w", err)
	}

	return &domain.PaginatorResponse[domain.Supply]{
		Items:      items,
		TotalItems: total,
		TotalPages: int64(math.Ceil(float64(total) / float64(params.PageSize))),
		Page:       params.Page,
		PageSize:   params.PageSize,
	}, nil
}
func (s *service) Update(ctx context.Context, supply *domain.Supply) error {
	if err := supply.Validate(); err != nil {
		err = fmt.Errorf("supply validation failed: %w", err)
		logger.Of(ctx).Error(err)
		return err
	}

	if err := s.uow.Execute(ctx, s.saveRepositoryStep(supply)); err != nil {
		logger.Of(ctx).Error(err)
		return err
	}

	return nil
}
