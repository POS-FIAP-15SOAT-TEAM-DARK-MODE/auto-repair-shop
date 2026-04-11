package service_order

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db/postgres"
	pgPkg "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/db/postgres"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
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

func (r *repository) ExistsByID(ctx context.Context, id string) (bool, error) {
	tx, err := postgres.GetTransaction(ctx)
	if err != nil {
		return false, err
	}

	var exists bool
	if err = tx.QueryRowContext(ctx, serviceOrderExistsQuery, id).Scan(&exists); err != nil {
		return false, pgPkg.Error(ctx, err)
	}
	return exists, nil
}

func (r *repository) ListWorksByServiceOrderID(ctx context.Context, serviceOrderID string) ([]domain.Work, error) {
	tx, err := postgres.GetTransaction(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, listWorksByServiceOrderQuery, serviceOrderID)
	if err != nil {
		return nil, pgPkg.Error(ctx, err)
	}
	defer func() { _ = rows.Close() }()

	var works []domain.Work
	for rows.Next() {
		var w domain.Work
		if err = rows.Scan(&w.ID, &w.Name, &w.Description, &w.Price, &w.Status); err != nil {
			return nil, pgPkg.Error(ctx, err)
		}
		works = append(works, w)
	}
	if err = rows.Err(); err != nil {
		return nil, pgPkg.Error(ctx, err)
	}
	return works, nil
}

func (r *repository) AddWorkLink(ctx context.Context, serviceOrderID, workID string, unitPrice decimal.Decimal) error {
	tx, err := postgres.GetTransaction(ctx)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, insertServiceOrderWorkQuery,
		uuid.NewString(),
		serviceOrderID,
		workID,
		unitPrice,
	)
	if err != nil {
		return pgPkg.Error(ctx, err)
	}
	return nil
}

func (r *repository) RemoveWorkLink(ctx context.Context, serviceOrderID, workID string) error {
	tx, err := postgres.GetTransaction(ctx)
	if err != nil {
		return err
	}

	res, err := tx.ExecContext(ctx, deleteServiceOrderWorkQuery, serviceOrderID, workID)
	if err != nil {
		return pgPkg.Error(ctx, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return pgPkg.Error(ctx, err)
	}
	if n == 0 {
		return domain.ErrServiceOrderWorkNotFound
	}
	return nil
}
