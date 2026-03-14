package service

import (
	"context"
	"fmt"
	"math"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"golang.org/x/sync/errgroup"
)

type service struct {
	uow  domain.Executor
	repo domain.ServiceRepository
}

func Service(uow domain.Executor, repo domain.ServiceRepository) domain.ServiceService {
	return &service{uow, repo}
}

func (s *service) Create(ctx context.Context, svc *domain.Service) (*domain.Service, error) {
	if err := svc.Validate(); err != nil {
		err = fmt.Errorf("service validation failed: %w", err)
		logger.Of(ctx).Error(err)
		return nil, err
	}

	if err := s.uow.Execute(ctx, s.saveRepositoryStep(svc)); err != nil {
		logger.Of(ctx).Error(err)
		return nil, err
	}

	return svc, nil
}

func (s *service) saveRepositoryStep(svc *domain.Service) func(context.Context) error {
	return func(ctx context.Context) error {
		if err := s.repo.Save(ctx, svc); err != nil {
			return fmt.Errorf("repository.Save failed: %w", err)
		}
		return nil
	}
}

func (s *service) List(c context.Context, params *domain.ListServiceParams) (*domain.PaginatorResponse[domain.Service], error) {
	var response *domain.PaginatorResponse[domain.Service]
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

func (s *service) getPaginatedList(ctx context.Context, params *domain.ListServiceParams) (*domain.PaginatorResponse[domain.Service], error) {
	p := params.SearchServiceParams()

	var total int64
	var items []domain.Service
	var eg errgroup.Group

	eg.Go(func() error {
		t, err := s.repo.Count(ctx, p)
		if err != nil {
			return err
		}
		total = t
		return nil
	})

	eg.Go(func() error {
		list, err := s.repo.Search(ctx, p)
		if err != nil {
			return err
		}
		items = list
		return err
	})

	if err := eg.Wait(); err != nil {
		return nil, fmt.Errorf("fail to list services: %w", err)
	}

	return &domain.PaginatorResponse[domain.Service]{
		Items:      items,
		TotalItems: total,
		TotalPages: int64(math.Ceil(float64(total) / float64(params.PageSize))),
		Page:       params.Page,
		PageSize:   params.PageSize,
	}, nil
}
