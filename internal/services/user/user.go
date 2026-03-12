package services

import (
	"fmt"
	"strings"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/user/dto"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo domain.UserRepository
}

func NewUserService(repo domain.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(req dto.UserRequest) (*domain.User, error) {
	if strings.TrimSpace(req.Password) == "" || strings.TrimSpace(req.ConfirmPassword) == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}

	if strings.TrimSpace(req.Password) != strings.TrimSpace(req.ConfirmPassword) {
		return nil, fmt.Errorf("passwords do not match")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	customer, err := domain.NewUser(req.Name, req.Email, string(hash))

	if err != nil {
		return nil, err
	}
	return s.repo.CreateUser(customer)
}
