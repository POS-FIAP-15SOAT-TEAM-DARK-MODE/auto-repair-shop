package supply

import (
	"context"
	"database/sql"
	"errors"

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
		w.Version,
	); err != nil {
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
	if params.Version != "" {
		queryBuilder.Add("s.version = ", params.Version)
	}
	query, args := queryBuilder.Build()

	var total int64
	if err = tx.QueryRowContext(ctx, query, args...).Scan(&total); err != nil {
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
	if params.Version != "" {
		queryBuilder.Add("s.version = ", params.Version)
	}
	query, args := queryBuilder.Build()

	logger.Of(ctx).Debug("Executing query", zap.String("query", query), zap.Any("params", args))
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pgPkg.Error(ctx, err)
	}
	defer func() { _ = rows.Close() }()

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

func (r *repo) FindById(ctx context.Context, id string) (domain.Supply, error) {
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
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Supply{}, domain.ErrSupplyNotFound
		}

		return domain.Supply{}, pgPkg.Error(ctx, err)
	}

	return sup, nil
}

func (r *repo) DecrementStock(ctx context.Context, id string, amount int) error {
	tx, err := postgres.GetTransaction(ctx)
	if err != nil {
		return err
	}

	res, err := tx.ExecContext(ctx, decrementStockQuery, id, amount)
	if err != nil {
		return pgPkg.Error(ctx, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return pgPkg.Error(ctx, err)
	}
	if n == 0 {
		return domain.ErrSupplyOutOfStock
	}
	return nil
}

func (r *repo) RestoreStock(ctx context.Context, id string, amount int) error {
	tx, err := postgres.GetTransaction(ctx)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, restoreStockQuery, id, amount)
	if err != nil {
		return pgPkg.Error(ctx, err)
	}
	return nil
}
