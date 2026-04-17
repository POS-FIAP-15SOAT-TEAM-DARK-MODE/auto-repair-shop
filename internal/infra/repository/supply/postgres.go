package supply

import (
	"context"
	"fmt"
	"strings"

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
	logger.Of(ctx).Debug("Executing query", zap.String("query", createSupplyQuery), zap.Any("params", supply))
	if _, err = tx.ExecContext(ctx, createSupplyQuery, supply.ID, supply.Name, supply.Description, supply.UnitPrice, supply.StockQuantity, supply.Version); err != nil {
		return pgPkg.Error(ctx, err)
	}
	return nil
}

func (u *repo) Update(ctx context.Context, supply *domain.SupplyUpdate) error {
	tx, err := postgres.GetTransaction(ctx)
	if err != nil {
		return err
	}

	query, args := buildUpdateSupplyQuery(supply)
	logger.Of(ctx).Debug("Executing query", zap.String("query", query), zap.Any("params", supply))

	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return pgPkg.Error(ctx, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return pgPkg.Error(ctx, err)
	}

	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func buildUpdateSupplyQuery(supply *domain.SupplyUpdate) (string, []any) {
	setClauses := []string{}
	args := []any{}
	i := 1

	if supply.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", i))
		args = append(args, *supply.Name)
		i++
	}
	if supply.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", i))
		args = append(args, *supply.Description)
		i++
	}
	if supply.UnitPrice != nil {
		setClauses = append(setClauses, fmt.Sprintf("unit_price = $%d", i))
		args = append(args, *supply.UnitPrice)
		i++
	}
	if supply.StockQuantity != nil {
		setClauses = append(setClauses, fmt.Sprintf("stock_quantity = $%d", i))
		args = append(args, *supply.StockQuantity)
		i++
	}

	if supply.Name != nil || supply.UnitPrice != nil {
		setClauses = append(setClauses, "version = version + 1")
	}

	query := fmt.Sprintf(
		`UPDATE "supply" SET %s WHERE id = $%d`,
		strings.Join(setClauses, ", "), i,
	)
	args = append(args, supply.ID)

	return query, args
}

func (r *repo) Count(ctx context.Context, params *domain.SearchSupplyParams) (int64, error) {
	tx, err := postgres.GetOneTimeTransaction(ctx)
	if err != nil {
		return 0, err
	}

	queryBuilder := db.QueryBuilder(countSuppliesQuery)

	if params.Status != "" {
		queryBuilder.Add("status = ", domain.StringToWorkStatus(params.Status).Bool())
	}

	query, args := queryBuilder.Build()

	var total int64
	if err := tx.QueryRowContext(ctx, query, args...).Scan(&total); err != nil {
		return 0, pgPkg.Error(ctx, err)
	}

	return total, nil
}
func (r *repo) Search(ctx context.Context, params *domain.SearchSupplyParams) ([]domain.Supply, error) {
	tx, err := postgres.GetOneTimeTransaction(ctx)
	if err != nil {
		return nil, err
	}

	queryBuilder := db.QueryBuilder(searchQuery).
		OrderBy("s.created_at", db.ASC).
		AddPagination(params.Limit, params.Offset)

	if params.Status != "" {
		queryBuilder.Add("status = ", domain.StringToWorkStatus(params.Status).Bool())
	}

	query, args := queryBuilder.Build()

	logger.Of(ctx).Debug("Executing query", zap.String("query", query), zap.Any("params", args))
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pgPkg.Error(ctx, err)
	}

	supplies := make([]domain.Supply, 0, params.Limit)
	for rows.Next() {
		if err = rows.Err(); err != nil {
			return nil, pgPkg.Error(ctx, err)
		}

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

	return supplies, nil
}

func (r *repo) Delete(ctx context.Context, id string) error {
	tx, err := postgres.GetTransaction(ctx)
	if err != nil {
		return err
	}

	logger.Of(ctx).Debug("Executing query", zap.String("query", deleteQuery), zap.Any("params", id))
	if _, err = tx.ExecContext(ctx, deleteQuery, id); err != nil {
		return pgPkg.Error(ctx, err)
	}

	return nil
}
