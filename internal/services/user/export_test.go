package user

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow"
)

func NewServiceForTest(uow uow.Executor, repo domain.UserRepository, hashFunc func(*domain.User, context.Context) error) domain.UserService {
	return &service{
		uow:          uow,
		repo:         repo,
		hashPassword: hashFunc,
	}
}
