package services

import (
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
)

type UserService struct {
	repo domain.UserRepository
}

func NewUserService(repo domain.UserRepository) domain.UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Create(user *domain.User) (*domain.User, error) {
	if err := user.Validate(); err != nil {
		return nil, err
	}
	if err := user.HashPassword(); err != nil {
		return nil, err
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}
