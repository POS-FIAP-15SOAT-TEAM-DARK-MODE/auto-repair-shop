package user

import (
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
)

type service struct {
	repo domain.UserRepository
}

func Service(repo domain.UserRepository) domain.UserService {
	return &service{repo: repo}
}

func (s *service) Login(user *domain.User) (*domain.LoginResponse, error) {
	return &domain.LoginResponse{
		Token:     "token",
		ExpiresIn: 1000,
	}, nil
}

func (s *service) Create(user *domain.User) (*domain.User, error) {
	if err := user.Validate(); err != nil {
		// TODO: add logging
		return nil, err
	}
	if err := user.HashPassword(); err != nil {
		// TODO: add logging
		return nil, err
	}

	return s.repo.Create(user)
}
