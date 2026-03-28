package user

import (
	"context"
	"fmt"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow"
)

type service struct {
	uow          uow.Executor
	repo         domain.UserRepository
	hashPassword func(*domain.User, context.Context) error
}

func Service(uow uow.Executor, repo domain.UserRepository) domain.UserService {
	return &service{
		uow:          uow,
		repo:         repo,
		hashPassword: (*domain.User).HashPassword,
	}
}

func (s *service) Create(ctx context.Context, user *domain.User) error {
	if err := user.Validate(); err != nil {
		err = fmt.Errorf("user validation failed: %w", err)
		logger.Of(ctx).Error(err)
		return err
	}
	if err := s.hashPassword(user, ctx); err != nil {
		err = fmt.Errorf("user password hashing failed: %w", err)
		logger.Of(ctx).Error(err)
		return err
	}

	if err := s.uow.Execute(ctx, s.createRepositoryStep(user)); err != nil {
		logger.Of(ctx).Error(err)
		return err
	}

	return nil
}

func (s *service) createRepositoryStep(user *domain.User) func(context.Context) error {
	return func(ctx context.Context) error {
		if err := s.repo.Create(ctx, user); err != nil {
			return fmt.Errorf("repository.Create failed: %w", err)
		}
		return nil
	}
}
