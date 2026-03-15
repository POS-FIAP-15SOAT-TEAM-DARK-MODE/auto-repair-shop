package user

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
)

type service struct {
	repo domain.UserRepository
}

func NewUserService(repo domain.UserRepository) domain.UserService {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	if err := user.Validate(); err != nil {
		return nil, err
	}
	if err := user.HashPassword(); err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}
