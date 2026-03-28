package user

import (
	"context"
	"time"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow"
)

func NewServiceForTest(uow uow.Executor, repo domain.UserRepository, expiresIn time.Duration, hashFunc func(*domain.User, context.Context) error) domain.UserService {
	return &service{
		uow:          uow,
		repo:         repo,
		expiresIn:    expiresIn,
		hashPassword: hashFunc,
	}
}
