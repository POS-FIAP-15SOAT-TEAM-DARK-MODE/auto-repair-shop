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

func (r *repository) Search(ctx context.Context, params *domain.SearchServiceOrderHistoryParams) ([]domain.ServiceOrderHistoryItem, error) {
	tx, err := postgres.GetOneTimeTransaction(ctx)
	if err != nil {
		return nil, err
	}

	qb := db.QueryBuilder(selectServiceOrderHistory).
		OrderBy("soh.created_at", db.ASC)

	if params.ID != "" {
		qb.Add("soh.service_order_id = ", params.ID)
	}

	query, args := qb.Build()

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pgPkg.Error(ctx, err)
	}

	defer func() { _ = rows.Close() }()

	histories := []domain.ServiceOrderHistoryItem{}
	for rows.Next() {
		if err = rows.Err(); err != nil {
			return nil, pgPkg.Error(ctx, err)
		}

		var (
			history        domain.ServiceOrderHistoryItem
			previousStatus sql.NullString
		)

		if err = rows.Scan(
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

func (r *repository) SearchWorkTransitionsByServiceOrderID(ctx context.Context, serviceOrderID string) ([]domain.WorkTransitionGroup, error) {
	tx, err := postgres.GetOneTimeTransaction(ctx)
	if err != nil {
		return nil, err
	}

	qb := db.QueryBuilder(selectWorkServiceOrderHistory).
		OrderBy("wsosh.created_at", db.ASC).
		Add("wsosh.service_order_id = ", serviceOrderID)

	query, args := qb.Build()

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pgPkg.Error(ctx, err)
	}

	defer func() { _ = rows.Close() }()

	orderMap := make(map[string]int)
	var groups []domain.WorkTransitionGroup

	for rows.Next() {
		if err = rows.Err(); err != nil {
			return nil, pgPkg.Error(ctx, err)
		}

		var (
			workID         string
			item           domain.WorkStatusItem
			previousStatus sql.NullString
		)

		if err = rows.Scan(
			&workID,
			&previousStatus,
			&item.NewStatus,
			&item.CreatedAt,
		); err != nil {
			return nil, pgPkg.Error(ctx, err)
		}

		if previousStatus.Valid {
			item.PreviousStatus = domain.SERVICE_ORDER_STATUS(previousStatus.String)
		}

		idx, exists := orderMap[workID]
		if !exists {
			idx = len(groups)
			orderMap[workID] = idx
			groups = append(groups, domain.WorkTransitionGroup{WorkID: workID})
		}
		groups[idx].Status = append(groups[idx].Status, item)
	}

	return groups, nil
}
