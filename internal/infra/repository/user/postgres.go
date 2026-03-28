package user

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db/postgres"
	pgPkg "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/db/postgres"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"go.uber.org/zap"
)

type repo struct{}

func Repository() domain.UserRepository {
	return &repo{}
}

func (u *repo) Create(ctx context.Context, c *domain.User) error {
	tx, err := postgres.GetTransaction(ctx)
	if err != nil {
		return err
	}

	logger.Of(ctx).Debug("Executing query", zap.String("query", createUserQuery), zap.Any("params", c))
	if _, err := tx.ExecContext(ctx, createUserQuery, c.ID, c.Name, c.Email, c.Password); err != nil {
		return pgPkg.Error(ctx, err)
	}
	return nil
}
