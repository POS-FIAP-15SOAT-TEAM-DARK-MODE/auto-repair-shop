package repository

import (
	"context"
	"database/sql"

	dbPkg "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/pkg/db"
	pgPkg "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/pkg/db/postgres"
	uowPkg "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/pkg/uow"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/service_order/domain"
)

const (
	insertWorkSOHistory = `
		INSERT INTO work_service_order_status_history (id, work_id, service_order_id, previous_status, new_status)
		VALUES ($1, $2, $3, $4, $5)`

	searchWorkSOHistory = `
		SELECT wsosh.id, wsosh.work_id, wsosh.service_order_id, wsosh.previous_status, wsosh.new_status, wsosh.created_at
		FROM work_service_order_status_history wsosh`
)

type wsoHistoryPostgres struct{}

func NewWSOHistoryPostgres() domain.WorkSOHistoryRepository {
	return &wsoHistoryPostgres{}
}

func (r *wsoHistoryPostgres) Insert(ctx context.Context, serviceOrderID, workID string, previousStatus *domain.WORK_SERVICE_ORDER_STATUS, newStatus domain.WORK_SERVICE_ORDER_STATUS) error {
	tx, err := uowPkg.GetTransaction(ctx)
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

func (r *wsoHistoryPostgres) Search(ctx context.Context, params domain.SearchWorkSOHistoryParams) ([]domain.WorkServiceOrderHistory, error) {
	db, err := uowPkg.GetOneTimeTransaction(ctx)
	if err != nil {
		return nil, err
	}

	qb := dbPkg.QueryBuilder(searchWorkSOHistory).
		OrderBy("wsosh.created_at", dbPkg.DESC)

	if params.ServiceOrderID != "" {
		qb.Add("wsosh.service_order_id =", params.ServiceOrderID)
	}
	if params.WorkID != "" {
		qb.Add("wsosh.work_id =", params.WorkID)
	}

	query, args := qb.Build()

	rows, err := db.QueryContext(ctx, query, args...)
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
