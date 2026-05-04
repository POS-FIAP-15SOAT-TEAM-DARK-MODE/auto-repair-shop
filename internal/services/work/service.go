package work

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
	repo domain.WorkRepository
}

func Service(uow uow.Executor, repo domain.WorkRepository) domain.WorkService {
	return &service{uow, repo}
}

func (s *service) Create(ctx context.Context, work *domain.Work) error {
	if err := work.Validate(); err != nil {
		err = fmt.Errorf("work validation failed: %w", err)
		logger.Of(ctx).Error(err)
		return err
	}

	if err := s.uow.Execute(ctx, s.saveRepositoryStep(work)); err != nil {
		logger.Of(ctx).Error(err)
		return err
	}

	return nil
}

func (s *service) saveRepositoryStep(work *domain.Work) func(context.Context) error {
	return func(ctx context.Context) error {
		if err := s.repo.Save(ctx, work); err != nil {
			return fmt.Errorf("repository.Save failed: %w", err)
		}
		return nil
	}
}

func (s *service) List(ctx context.Context, params *domain.ListWorkParams) (*domain.PaginatorResponse[domain.Work], error) {
	p := params.SearchWorkParams()
	var eg errgroup.Group
	response := &domain.PaginatorResponse[domain.Work]{
		Page:     params.Page,
		PageSize: params.PageSize,
	}

	eg.Go(func() error {
		t, err := s.repo.Count(ctx, p)
		if err != nil {
			return err
		}
		response.TotalItems = t
		response.TotalPages = int64(math.Ceil(float64(t) / float64(params.PageSize)))
		return nil
	})

	eg.Go(func() error {
		list, err := s.repo.Search(ctx, p)
		if err != nil {
			return err
		}
		response.Items = list
		return err
	})

	if err := eg.Wait(); err != nil {
		return nil, fmt.Errorf("fail to list works: %w", err)
	}

	return response, nil
}

func (s *service) Delete(c context.Context, id string) error {
	if err := s.uow.Execute(c, func(ctx context.Context) error { return s.repo.Delete(ctx, id) }); err != nil {
		logger.Of(c).Error(err)
		return err
	}

	return nil
}

func (s *service) Update(ctx context.Context, work *domain.Work) error {
	if err := work.Validate(); err != nil {
		err = fmt.Errorf("work validation failed: %w", err)
		logger.Of(ctx).Error(err)
		return err
	}

	if err := s.uow.Execute(ctx, s.saveRepositoryStep(work)); err != nil {
		logger.Of(ctx).Error(err)
		return err
	}

	return nil
}
