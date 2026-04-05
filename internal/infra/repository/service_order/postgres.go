package service_order

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db/postgres"
	pgPkg "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/db/postgres"
)

type repository struct{}

func Repository() domain.ServiceOrderRepository {
	return &repository{}
}

func (r *repository) Save(ctx context.Context, so *domain.ServiceOrder) error {
	tx, err := postgres.GetTransaction(ctx)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, insertServiceOrderQuery,
		so.ID,
		so.Customer.ID,
		so.Vehicle.ID,
		so.Status.String(),
		so.TotalAmount,
	)
	if err != nil {
		return pgPkg.Error(ctx, err)
	}

	return nil
}

func (r *repository) GetHistoryByID(ctx context.Context, id string) ([]domain.ServiceOrderHistory, error) {
	tx, err := postgres.GetOneTimeTransaction(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, getServiceOrderHistoryByIDQuery, id)
	if err != nil {
		return nil, pgPkg.Error(ctx, err)
	}

	defer func() { _ = rows.Close() }()

	histories := []domain.ServiceOrderHistory{}
	for rows.Next() {
		var history domain.ServiceOrderHistory
		if err := rows.Scan(&history.ID, &history.PreviousStatus, &history.NewStatus, &history.CreatedAt); err != nil {
			return nil, pgPkg.Error(ctx, err)
		}
		histories = append(histories, history)
	}

	return histories, nil
}
