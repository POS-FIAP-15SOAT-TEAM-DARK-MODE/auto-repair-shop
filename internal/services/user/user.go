package services

import (
	"fmt"
	"strings"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo domain.UserRepository
}

type CreateUserRequest struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

func NewUserService(repo domain.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(req CreateUserRequest) (*domain.User, error) {
	id := uuid.New().String()

	if strings.TrimSpace(req.Password) == "" || strings.TrimSpace(req.ConfirmPassword) == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}

	if strings.TrimSpace(req.Password) != strings.TrimSpace(req.ConfirmPassword) {
		return nil, fmt.Errorf("passwords do not match")
	}

	//Gera o hash da senha
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Tentar inserir o usuário no banco de dados, e se der erro de email duplicado, retornar um erro

	// Criar o objeto User usando a função construtora, que já garante as regras de negócio
	customer, err := domain.NewUser(id, req.Name, req.Email, string(hash))

	if err != nil {
		return nil, err
	}
	return s.repo.CreateUser(customer)
}
