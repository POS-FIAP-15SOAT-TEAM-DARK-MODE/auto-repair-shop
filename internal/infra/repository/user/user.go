package repository

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	pkgdb "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/db"
)

type SqlxUserRepository struct {
	db pkgdb.Executor
}

func NewUserRepository(ex pkgdb.Executor) *SqlxUserRepository {
	return &SqlxUserRepository{db: ex}
}

func (u *SqlxUserRepository) Create(ctx context.Context, c *domain.User) error {
	exec := pkgdb.ExtractExecutor(ctx, u.db)
	_, err := exec.Exec(CreateUser, c.ID, c.Name, c.Email, c.Password)
	if err != nil {
		return pkgdb.Error(err)
	}
	return nil
}
