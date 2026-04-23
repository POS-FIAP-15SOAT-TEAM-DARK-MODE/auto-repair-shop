package repository

import (
	"context"
	"database/sql"
	"errors"

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

func (u *repo) FindById(ctx context.Context, id string) (domain.Supply, error) {
	tx, err := postgres.GetTransaction(ctx)
	if err != nil {
		return domain.Supply{}, err
	}

	var sup domain.Supply
	err = tx.QueryRowContext(ctx, findByIDQuery, id).Scan(
		&sup.ID,
		&sup.Name,
		&sup.Description,
		&sup.UnitPrice,
		&sup.StockQuantity,
		&sup.Version,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Supply{}, domain.ErrSupplyNotFound
	}
	if err != nil {
		return domain.Supply{}, pgPkg.Error(ctx, err)
	}

	return sup, nil
}
