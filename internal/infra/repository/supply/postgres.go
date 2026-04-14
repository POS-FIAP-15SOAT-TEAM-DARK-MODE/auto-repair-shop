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

func (u *repo) List(ctx context.Context, params *domain.ListSupplyParams) (*domain.PaginatorResponse[domain.Supply], error) {
	db, err := postgres.GetOneTimeTransaction(ctx)
	if err != nil {
		return nil, err
	}

	offset := (params.Page - 1) * params.PageSize

	// total de registros
	var total int64
	if err := db.QueryRowContext(ctx, countSuppliesQuery).Scan(&total); err != nil {
		return nil, pgPkg.Error(ctx, err)
	}

	rows, err := db.QueryContext(ctx, listSuppliesQuery, params.PageSize, offset)
	if err != nil {
		return nil, pgPkg.Error(ctx, err)
	}
	defer rows.Close()

	var supplies []domain.Supply
	for rows.Next() {
		supply := domain.Supply{}
		if err = rows.Scan(&supply.ID, &supply.Name, &supply.Description, &supply.UnitPrice, &supply.StockQuantity, &supply.Version); err != nil {
			return nil, pgPkg.Error(ctx, err)
		}
		supplies = append(supplies, supply)
	}

	totalPages := (total + params.PageSize - 1) / params.PageSize

	return &domain.PaginatorResponse[domain.Supply]{
		Items:      supplies,
		TotalItems: total,
		TotalPages: totalPages,
		PageSize:   params.PageSize,
		Page:       params.Page,
	}, nil
}
