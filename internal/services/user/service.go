package user

import (
	"time"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/auth"
	"golang.org/x/crypto/bcrypt"
)

type service struct {
	repo      domain.UserRepository
	expiresIn time.Duration
	secretKey string
}

func Service(repo domain.UserRepository, expiresIn time.Duration, secretKey string) domain.UserService {
	return &service{repo: repo, expiresIn: expiresIn, secretKey: secretKey}
}

func (s *service) Login(user *domain.User) (*domain.LoginResponse, error) {
	stored, err := s.validateUserCredentials(user)
	if err != nil {
		return nil, err
	}

	roles, err := s.repo.GetRolesById(stored.Id)
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(s.expiresIn)
	token, err := auth.GenerateToken(s.secretKey, stored.Id, roles, expiresAt)
	if err != nil {
		return nil, err
	}

	return &domain.LoginResponse{
		Token:     token,
		ExpiresIn: int(s.expiresIn.Seconds()),
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

func (s *service) validateUserCredentials(user *domain.User) (*domain.User, error) {
	stored, err := s.repo.GetByEmail(user.Email)
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
