package repository

import (
	"context"
	"fmt"
	"strings"

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

func (u *repo) List(ctx context.Context) ([]*domain.Supply, error) {
	tx, err := postgres.GetTransaction(ctx)
	if err != nil {
		return nil, err
	}
	logger.Of(ctx).Debug("Executing query", zap.String("query", listSuppliesQuery))
	rows, err := tx.QueryContext(ctx, listSuppliesQuery)
	if err != nil {
		return nil, pgPkg.Error(ctx, err)
	}
	defer rows.Close()

	var supplies []*domain.Supply
	for rows.Next() {
		supply := &domain.Supply{}
		if err = rows.Scan(&supply.ID, &supply.Name, &supply.Description, &supply.UnitPrice, &supply.StockQuantity, &supply.Version); err != nil {
			return nil, pgPkg.Error(ctx, err)
		}
		supplies = append(supplies, supply)
	}

	return supplies, nil
}

func (u *repo) GetByID(ctx context.Context, id string) (*domain.Supply, error) {
	tx, err := postgres.GetTransaction(ctx)
	if err != nil {
		return nil, err
	}
	logger.Of(ctx).Debug("Executing query", zap.String("query", getSupplyByIDQuery), zap.String("id", id))
	row := tx.QueryRowContext(ctx, getSupplyByIDQuery, id)
	supply := &domain.Supply{}
	if err = row.Scan(&supply.ID, &supply.Name, &supply.Description, &supply.UnitPrice, &supply.StockQuantity, &supply.Version); err != nil {
		return nil, pgPkg.Error(ctx, err)
	}
	return supply, nil
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
