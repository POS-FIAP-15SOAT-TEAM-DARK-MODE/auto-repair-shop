package repository

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db/postgres"
	pgPkg "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/db/postgres"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"

	"go.uber.org/zap"
)

type repo struct{}

func Repository() domain.SupplyRepository {
	return &repo{}
}

func (u *repo) Create(ctx context.Context, supply *domain.Supply) error {
	tx, err := postgres.GetTransaction(ctx)
	if err != nil {
		return err
	}
	logger.Of(ctx).Debug("Executing query", zap.String("query", createSupplyQuery), zap.Any("params", supply))
	if _, err = tx.ExecContext(ctx, createSupplyQuery, supply.ID, supply.Name, supply.Description, supply.UnitPrice, supply.StockQuantity, supply.Version); err != nil {
		return pgPkg.Error(ctx, err)
	}
	return nil
}
