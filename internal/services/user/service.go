package user

import (
	"context"
	"fmt"
	"slices"

	"time"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/auth"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow"
	"golang.org/x/crypto/bcrypt"
)

type service struct {
	uow          uow.Executor
	repo         domain.UserRepository
	expiresIn    time.Duration
	hashPassword func(*domain.User, context.Context) error
}

func Service(uow uow.Executor, repo domain.UserRepository, expiresIn time.Duration) domain.UserService {
	return &service{
		uow:          uow,
		repo:         repo,
		expiresIn:    expiresIn,
		hashPassword: (*domain.User).HashPassword,
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

	roles, err := s.repo.GetRolesByUserId(ctx, loggedUser.ID)
	if err != nil {
		return err
	}

	expiresAt := time.Now().Add(s.expiresIn)
	token, err := auth.GenerateToken(loggedUser.ID, roles, expiresAt)
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

func (s *service) UpdateRole(ctx context.Context, id string, role domain.Role) error {
	if !isValidRole(role) {
		return domain.ValidationError{Message: "invalid role"}
	}

	if err := s.uow.Execute(ctx, func(txCtx context.Context) error {
		if err := s.repo.UpdateRole(txCtx, id, role); err != nil {
			return fmt.Errorf("repository.UpdateRole failed: %w", err)
		}
		return nil
	}); err != nil {
		logger.Of(ctx).Error(err)
		return err
	}

	return nil
}

func isValidRole(role domain.Role) bool {
	validRoles := []domain.Role{domain.ADMIN, domain.ATTENDANT, domain.MECHANIC, domain.CUSTOMER}
	return slices.Contains(validRoles, role)
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
	stored, err := s.repo.GetByEmail(ctx, loggedUser.Email)
	if err != nil {
		return err
	}

	if stored == nil {
		return domain.ErrInvalidUserCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(stored.Password), []byte(loggedUser.Password)); err != nil {
		return domain.ErrInvalidUserCredentials
	}

	loggedUser.User = *stored
	return nil
}
