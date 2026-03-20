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

func (s *service) Login(ctx context.Context, loggedUser *domain.LoggedUser) error {
	if err := loggedUser.Validate(); err != nil {
		return err
	}

	err := s.authenticateUser(ctx, loggedUser)
	if err != nil {
		return err
	}

	roles, err := s.repo.GetRolesById(ctx, loggedUser.User.ID)
	if err != nil {
		return err
	}

	expiresAt := time.Now().Add(s.expiresIn)
	token, err := auth.GenerateToken(loggedUser.User.ID, roles, expiresAt)
	if err != nil {
		return err
	}

	loggedUser.SessionToken = token
	loggedUser.SessionExpiresIn = int(s.expiresIn.Seconds())

	return nil
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

func (s *service) authenticateUser(ctx context.Context, loggedUser *domain.LoggedUser) error {
	stored, err := s.repo.GetByEmail(ctx, loggedUser.User.Email)
	if err != nil {
		return err
	}

	if stored == nil {
		return domain.ErrInvalidUserCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(stored.Password), []byte(loggedUser.User.Password)); err != nil {
		return domain.ErrInvalidUserCredentials
	}

	loggedUser.User = *stored
	return nil
}
