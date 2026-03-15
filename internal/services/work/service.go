package work

import (
	"context"
	"fmt"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow"
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
