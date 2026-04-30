package work_service_order_history

import (
	"context"
	"database/sql"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db/postgres"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/db"
	pgPkg "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/db/postgres"
)

type repository struct{}

func Repository() domain.WorkServiceOrderHistoryRepository {
	return &repository{}
}

func (r *repository) Insert(ctx context.Context, serviceOrderID, workID string, previousStatus *domain.WORK_SERVICE_ORDER_STATUS, newStatus domain.WORK_SERVICE_ORDER_STATUS) error {
	tx, err := postgres.GetTransaction(ctx)
	if err != nil {
		return err
	}

	var prev sql.NullString
	if previousStatus != nil {
		prev = sql.NullString{String: previousStatus.String(), Valid: true}
	}

	_, err = tx.ExecContext(ctx, insertWorkSOHistory,
		domain.NewHistoryServiceOrderID(),
		workID,
		serviceOrderID,
		prev,
		newStatus.String(),
	)
	return pgPkg.Error(ctx, err)
}

func (r *repository) Search(ctx context.Context, params domain.SearchWorkSOHistoryParams) ([]domain.WorkServiceOrderHistory, error) {
	tx, err := postgres.GetOneTimeTransaction(ctx)
	if err != nil {
		return nil, err
	}

	qb := db.QueryBuilder(searchWorkSOHistory).
		OrderBy("wsosh.created_at", db.DESC)

	if params.ServiceOrderID != "" {
		qb.Add("wsosh.service_order_id =", params.ServiceOrderID)
	}

	if params.WorkID != "" {
		qb.Add("wsosh.work_id =", params.WorkID)
	}

	query, args := qb.Build()

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pgPkg.Error(ctx, err)
	}
	defer func() { _ = rows.Close() }()

	var results []domain.WorkServiceOrderHistory
	for rows.Next() {
		if err = rows.Err(); err != nil {
			return nil, pgPkg.Error(ctx, err)
		}

		var (
			h          domain.WorkServiceOrderHistory
			prevStatus sql.NullString
		)

		if err = rows.Scan(&h.ID, &h.WorkID, &h.ServiceOrderID, &prevStatus, &h.Status, &h.CreatedAt); err != nil {
			return nil, pgPkg.Error(ctx, err)
		}

		if prevStatus.Valid {
			s := domain.WORK_SERVICE_ORDER_STATUS(prevStatus.String)
			h.PreviousStatus = &s
		}

		results = append(results, h)
	}

	return results, nil
}
