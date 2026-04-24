package service_order

import (
	"context"
	"database/sql"
	"errors"

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

func (r *repository) ExistsByID(ctx context.Context, id string) (bool, domain.SERVICE_ORDER_STATUS, error) {
	tx, err := postgres.GetTransaction(ctx)
	if err != nil {
		return false, "", err
	}

	var status domain.SERVICE_ORDER_STATUS
	if err = tx.QueryRowContext(ctx, serviceOrderExistsQuery, id).Scan(&status); err != nil {
		return false, "", pgPkg.Error(ctx, err)
	}
	return status != "", status, nil
}

func (r *repository) ListWorksByServiceOrderID(ctx context.Context, serviceOrderID string) ([]domain.Work, error) {
	tx, err := postgres.GetOneTimeTransaction(ctx)
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

func (r *repository) ListSuppliesByServiceOrderID(ctx context.Context, serviceOrderID string) ([]domain.Supply, error) {
	tx, err := postgres.GetOneTimeTransaction(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, listSuppliesByServiceOrderQuery, serviceOrderID)
	if err != nil {
		return nil, pgPkg.Error(ctx, err)
	}
	defer func() { _ = rows.Close() }()

	var supplies []domain.Supply
	for rows.Next() {
		var sup domain.Supply
		if err = rows.Scan(&sup.ID, &sup.Name, &sup.Description, &sup.UnitPrice, &sup.StockQuantity, &sup.Version); err != nil {
			return nil, pgPkg.Error(ctx, err)
		}
		supplies = append(supplies, sup)
	}
	if err = rows.Err(); err != nil {
		return nil, pgPkg.Error(ctx, err)
	}
	return supplies, nil
}

func (r *repository) AddSupplyLink(ctx context.Context, serviceOrderID string, supplyID string, amount int, unitPrice decimal.Decimal) error {
	tx, err := postgres.GetTransaction(ctx)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, insertServiceOrderSuppliesQuery,
		uuid.NewString(),
		serviceOrderID,
		supplyID,
		amount,
		unitPrice,
	)
	if err != nil {
		return pgPkg.Error(ctx, err)
	}
	return nil
}

func (r *repository) RemoveSupplyLink(ctx context.Context, serviceOrderID string, supplyID string) error {
	tx, err := postgres.GetTransaction(ctx)
	if err != nil {
		return err
	}

	res, err := tx.ExecContext(ctx, deleteServiceOrderSuppliesQuery, serviceOrderID, supplyID)
	if err != nil {
		return pgPkg.Error(ctx, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return pgPkg.Error(ctx, err)
	}
	if n == 0 {
		return domain.ErrServiceOrderSupplyNotFound
	}
	return nil
}

func (r *repository) FindByID(ctx context.Context, id string) (domain.ServiceOrder, error) {
	db, err := postgres.GetOneTimeTransaction(ctx)
	if err != nil {
		return domain.ServiceOrder{}, err
	}

	var so domain.ServiceOrder
	var ct domain.Customer
	var vh domain.Vehicle
	var status string
	if err = db.QueryRowContext(ctx, serviceOrderFindByIDQuery, id).Scan(
		&so.ID,
		&status,
		&so.TotalAmount,
		&ct.ID,
		&vh.ID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ServiceOrder{}, domain.ErrServiceOrderNotFound
		}
		return domain.ServiceOrder{}, pgPkg.Error(ctx, err)
	}

	so.Status = domain.StringToServiceOrderStatus(status)
	so.Customer = &ct
	so.Vehicle = &vh

	return so, nil
}
