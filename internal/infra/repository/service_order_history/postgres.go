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

		var (
			history        domain.ServiceOrderHistory
			previousStatus sql.NullString
			newStatus      sql.NullString
			createdAt      sql.NullTime
		)

		if err = rows.Scan(
			&history.ID,
			&history.ServiceOrderID,
			&previousStatus,
			&newStatus,
			&createdAt,
		); err != nil {
			return nil, pgPkg.Error(ctx, err)
		}

		if previousStatus.Valid {
			history.PreviousStatus = domain.SERVICE_ORDER_STATUS(previousStatus.String)
		}

		if newStatus.Valid {
			history.NewStatus = domain.SERVICE_ORDER_STATUS(newStatus.String)
		}

		if createdAt.Valid {
			history.CreatedAt = createdAt.Time
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

func (r *repository) WorkTimelineByServiceOrderID(ctx context.Context, serviceOrderID string) ([]domain.WorkStatusTimeline, error) {
	tx, err := postgres.GetOneTimeTransaction(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, selectWorkTimelineByServiceOrderID, serviceOrderID)
	if err != nil {
		return nil, pgPkg.Error(ctx, err)
	}
	defer func() { _ = rows.Close() }()

	timelines := []domain.WorkStatusTimeline{}
	indexByWorkID := map[string]int{}

	for rows.Next() {
		if err = rows.Err(); err != nil {
			return nil, pgPkg.Error(ctx, err)
		}

		var (
			workID         string
			previousStatus sql.NullString
			newStatus      sql.NullString
			createdAt      sql.NullTime
		)

		if err = rows.Scan(&workID, &previousStatus, &newStatus, &createdAt); err != nil {
			return nil, pgPkg.Error(ctx, err)
		}

		idx, ok := indexByWorkID[workID]
		if !ok {
			timelines = append(timelines, domain.WorkStatusTimeline{
				WorkID:  workID,
				History: []domain.WorkStatusHistoryEntry{},
			})
			idx = len(timelines) - 1
			indexByWorkID[workID] = idx
		}

		if !newStatus.Valid {
			continue
		}

		entry := domain.WorkStatusHistoryEntry{
			NewStatus: domain.SERVICE_ORDER_STATUS(newStatus.String),
			CreatedAt: createdAt.Time,
		}
		if previousStatus.Valid {
			entry.PreviousStatus = domain.SERVICE_ORDER_STATUS(previousStatus.String)
		}

		timelines[idx].History = append(timelines[idx].History, entry)
	}

	return timelines, nil
}
