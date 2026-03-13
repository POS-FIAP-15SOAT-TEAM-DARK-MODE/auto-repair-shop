package service

import (
	"context"
	"fmt"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
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
