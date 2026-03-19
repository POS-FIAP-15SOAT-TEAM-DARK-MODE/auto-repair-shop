package user

import (
	"context"
	"fmt"

	"time"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/auth"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow"
	"golang.org/x/crypto/bcrypt"
)

type service struct {
	repo      domain.UserRepository
	expiresIn time.Duration
	uow       uow.Executor
}

func Service(uow uow.Executor, repo domain.UserRepository, expiresIn time.Duration) domain.UserService {
	return &service{
		uow:       uow,
		repo:      repo,
		expiresIn: expiresIn,
	}
}

func (s *service) Login(ctx context.Context, user *domain.User) (*domain.LoginResponse, error) {
	stored, err := s.validateUserCredentials(ctx, user)
	if err != nil {
		return nil, err
	}

	roles, err := s.repo.GetRolesById(ctx, stored.ID)
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(s.expiresIn)
	token, err := auth.GenerateToken(stored.ID, roles, expiresAt)
	if err != nil {
		return nil, err
	}

	return &domain.LoginResponse{
		Token:     token,
		ExpiresIn: int(s.expiresIn.Seconds()),
	}, nil
}

func (s *service) Create(ctx context.Context, user *domain.User) error {
	if err := user.Validate(); err != nil {
		err = fmt.Errorf("user validation failed: %w", err)
		logger.Of(ctx).Error(err)
		return err
	}
	if err := user.HashPassword(ctx); err != nil {
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

func (s *service) validateUserCredentials(ctx context.Context, user *domain.User) (*domain.User, error) {
	stored, err := s.repo.GetByEmail(ctx, user.Email)
	if err != nil {
		return nil, err
	}

	if stored == nil {
		return nil, domain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(stored.Password), []byte(user.Password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	return stored, nil
}
