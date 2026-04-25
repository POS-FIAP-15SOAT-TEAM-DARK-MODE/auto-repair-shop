package service_order_history

import (
	"context"
	"database/sql"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db/postgres"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/db"
	pgPkg "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/db/postgres"
)

type repository struct{}

func Repository() domain.ServiceOrderHistoryRepository {
	return &repository{}
}

func (r *repository) Search(ctx context.Context, params *domain.SearchServiceOrderHistoryParams) ([]domain.ServiceOrderHistory, error) {
	tx, err := postgres.GetOneTimeTransaction(ctx)
	if err != nil {
		return nil, err
	}

	qb := db.QueryBuilder(selectServiceOrderHistory).
		OrderBy("soh.created_at", db.ASC).
		AddPagination(params.PageSize, params.Page)

	if params.ID != "" {
		qb.Add("soh.service_order_id = ", params.ID)
	}

	query, args := qb.Build()

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pgPkg.Error(ctx, err)
	}

	defer func() { _ = rows.Close() }()

	histories := []domain.ServiceOrderHistory{}
	for rows.Next() {
		if err = rows.Err(); err != nil {
			return nil, pgPkg.Error(ctx, err)
		}

		var history domain.ServiceOrderHistory
		var previousStatus sql.NullString
		if err = rows.Scan(
			&history.ID,
			&history.ServiceOrderID,
			&previousStatus,
			&history.NewStatus,
			&history.CreatedAt,
		); err != nil {
			return nil, pgPkg.Error(ctx, err)
		}

		if previousStatus.Valid {
			history.PreviousStatus = domain.SERVICE_ORDER_STATUS(previousStatus.String)
		}

		histories = append(histories, history)
	}

	return histories, nil
}

func (r *repository) Count(ctx context.Context, params *domain.SearchServiceOrderHistoryParams) (int64, error) {
	tx, err := postgres.GetOneTimeTransaction(ctx)
	if err != nil {
		return 0, err
	}

	qb := db.QueryBuilder(countServiceOrderHistory)

	if params.ID != "" {
		qb.Add("soh.service_order_id = ", params.ID)
	}

	query, args := qb.Build()

	var total int64
	if err := tx.QueryRowContext(ctx, query, args...).Scan(&total); err != nil {
		return 0, pgPkg.Error(ctx, err)
	}

	return total, nil
}
