package supply

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db/postgres"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/db"
	pgPkg "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/db/postgres"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"

	"go.uber.org/zap"
)

type repo struct{}

func Repository() domain.SupplyRepository {
	return &repo{}
}

func (r *repo) Save(ctx context.Context, w *domain.Supply) error {
	tx, err := postgres.GetTransaction(ctx)
	if err != nil {
		return err
	}

	logger.Of(ctx).Debug("Executing query", zap.String("query", upsertQuery), zap.Any("params", w))
	if _, err = tx.ExecContext(
		ctx,
		upsertQuery,
		w.ID,
		w.Name,
		w.Description,
		w.UnitPrice,
		w.StockQuantity,
	); err != nil {
		return pgPkg.Error(ctx, err)
	}

	return nil
}

func (u *repo) Create(ctx context.Context, supply *domain.Supply) error {
	tx, err := postgres.GetTransaction(ctx)
	if err != nil {
		return err
	}
	logger.Of(ctx).Debug("Executing query", zap.String("query", upsertQuery), zap.Any("params", supply))
	if _, err = tx.ExecContext(ctx, upsertQuery, supply.ID, supply.Name, supply.Description, supply.UnitPrice, supply.StockQuantity); err != nil {
		return pgPkg.Error(ctx, err)
	}
	return nil
}

func (r *repo) Count(ctx context.Context, params *domain.ListSupplyParams) (int64, error) {
	tx, err := postgres.GetOneTimeTransaction(ctx)
	if err != nil {
		return 0, err
	}

	queryBuilder := db.QueryBuilder(countSuppliesQuery)

	query, args := queryBuilder.Build()

	var total int64
	if err := tx.QueryRowContext(ctx, query, args...).Scan(&total); err != nil {
		return 0, pgPkg.Error(ctx, err)
	}

	return total, nil
}
func (r *repo) Search(ctx context.Context, params *domain.ListSupplyParams) ([]domain.Supply, error) {
	tx, err := postgres.GetOneTimeTransaction(ctx)
	if err != nil {
		return nil, err
	}

	queryBuilder := db.QueryBuilder(searchQuery).
		OrderBy("s.created_at", db.ASC).
		AddPagination(params.PageSize, params.Offset())

	query, args := queryBuilder.Build()

	logger.Of(ctx).Debug("Executing query", zap.String("query", query), zap.Any("params", args))
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pgPkg.Error(ctx, err)
	}

	supplies := make([]domain.Supply, 0, params.PageSize)
	for rows.Next() {
		var supply domain.Supply
		if err = rows.Scan(
			&supply.ID,
			&supply.Name,
			&supply.Description,
			&supply.UnitPrice,
			&supply.StockQuantity,
			&supply.Version,
		); err != nil {
			return nil, pgPkg.Error(ctx, err)
		}

		supplies = append(supplies, supply)
	}

	if err = rows.Err(); err != nil {
		return nil, pgPkg.Error(ctx, err)
	}

	return supplies, nil
}

func (r *repo) Change(ctx context.Context, w *domain.Supply) error {
	tx, err := postgres.GetTransaction(ctx)
	if err != nil {
		return err
	}

	logger.Of(ctx).Debug("Executing query", zap.String("query", updateQuery), zap.Any("params", w))
	if _, err = tx.ExecContext(
		ctx,
		updateQuery,
		w.ID,
		w.Name,
		w.Description,
		w.UnitPrice,
		w.StockQuantity,
	); err != nil {
		return pgPkg.Error(ctx, err)
	}

	return nil
}
