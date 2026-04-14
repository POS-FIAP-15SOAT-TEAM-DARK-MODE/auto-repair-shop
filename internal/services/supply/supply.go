package supply

import (
	"context"
	"fmt"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow"
)

type service struct {
	repo domain.SupplyRepository
	uow  uow.Executor
}

func Service(uow uow.Executor, repo domain.SupplyRepository) domain.SupplyService {
	return &service{
		uow:  uow,
		repo: repo,
	}
}

func (s *service) Create(ctx context.Context, req *domain.Supply) error {
	if err := req.Validate(); err != nil {
		err := fmt.Errorf("supply validation failed: %w", err)
		logger.Of(ctx).Error(err)
		return err
	}

	if err := s.uow.Execute(ctx, s.createRepositoryStep(req)); err != nil {
		logger.Of(ctx).Error(err)
		return err
	}
	return nil
}

func (s *service) List(ctx context.Context) ([]*domain.Supply, error) {
	var supplies []*domain.Supply

	if err := s.uow.Execute(ctx, func(ctx context.Context) error {
		var err error
		supplies, err = s.repo.List(ctx)
		if err != nil {
			return fmt.Errorf("failed to list supplies: %w", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return supplies, nil
}

func (s *service) createRepositoryStep(user *domain.Supply) func(context.Context) error {
	return func(ctx context.Context) error {
		if err := s.repo.Create(ctx, user); err != nil {
			return fmt.Errorf("failed to create supply in repository: %w", err)
		}
		return nil
	}
}

func (s *service) Update(ctx context.Context, req *domain.SupplyUpdate) error {
	if err := req.Validate(); err != nil {
		err := fmt.Errorf("supply validation failed: %w", err)
		logger.Of(ctx).Error(err)
		return err
	}

	if err := s.uow.Execute(ctx, func(ctx context.Context) error {
		if err := s.repo.Update(ctx, req); err != nil {
			return fmt.Errorf("failed to update supply in repository: %w", err)
		}
		return nil
	}); err != nil {
		logger.Of(ctx).Error(err)
		return err
	}
	return nil
}
