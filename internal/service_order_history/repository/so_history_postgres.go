package repository

import (
	"context"
	"database/sql"

	dbPkg "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/db"
	pgPkg "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/db/postgres"
	uowPkg "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow"
	soDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order_history/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order_history/interfaces"
)

const (
	selectSOHistory = `SELECT soh.previous_status, soh.new_status, soh.created_at FROM service_order_status_history soh`

	selectWorkSOHistory = `SELECT wsosh.work_id, wsosh.previous_status, wsosh.new_status, wsosh.created_at FROM work_service_order_status_history wsosh`
)

type postgres struct{}

func NewPostgres() interfaces.ServiceOrderHistoryRepository {
	return &postgres{}
}

func (r *postgres) Search(ctx context.Context, params *domain.SearchParams) ([]domain.ServiceOrderHistoryItem, error) {
	db, err := uowPkg.GetOneTimeTransaction(ctx)
	if err != nil {
		return nil, err
	}

	qb := dbPkg.QueryBuilder(selectSOHistory).
		OrderBy("soh.created_at", dbPkg.ASC)

	if params.ID != "" {
		qb.Add("soh.service_order_id =", params.ID)
	}

	query, args := qb.Build()

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pgPkg.Error(ctx, err)
	}
	defer func() { _ = rows.Close() }()

	var items []domain.ServiceOrderHistoryItem
	for rows.Next() {
		if err = rows.Err(); err != nil {
			return nil, pgPkg.Error(ctx, err)
		}

		var (
			item           domain.ServiceOrderHistoryItem
			previousStatus sql.NullString
		)

		if err = rows.Scan(&previousStatus, &item.NewStatus, &item.CreatedAt); err != nil {
			return nil, pgPkg.Error(ctx, err)
		}

		if previousStatus.Valid {
			item.PreviousStatus = soDomain.SERVICE_ORDER_STATUS(previousStatus.String)
		}

		items = append(items, item)
	}

	return items, nil
}

func (r *postgres) SearchWorkTransitionsByServiceOrderID(ctx context.Context, serviceOrderID string) ([]domain.WorkTransitionGroup, error) {
	db, err := uowPkg.GetOneTimeTransaction(ctx)
	if err != nil {
		return nil, err
	}

	qb := dbPkg.QueryBuilder(selectWorkSOHistory).
		OrderBy("wsosh.created_at", dbPkg.ASC).
		Add("wsosh.service_order_id =", serviceOrderID)

	query, args := qb.Build()

	rows, err := db.QueryContext(ctx, query, args...)
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

		if err = rows.Scan(&workID, &previousStatus, &item.NewStatus, &item.CreatedAt); err != nil {
			return nil, pgPkg.Error(ctx, err)
		}

		if previousStatus.Valid {
			item.PreviousStatus = soDomain.SERVICE_ORDER_STATUS(previousStatus.String)
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
