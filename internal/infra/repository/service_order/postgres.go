package service_order

import (
	"context"
	"database/sql"
	"errors"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db/postgres"
	dbPkg "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/db"
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

	_, err = tx.ExecContext(ctx, insertServiceOrderStatusQuery,
		domain.NewHistoryServiceOrderID(),
		so.ID,
		domain.GetPreviousStatus(so.Status),
		so.Status)
	if err != nil {
		return pgPkg.Error(ctx, err)
	}

	return nil
}

func (r *repository) ExistsByID(ctx context.Context, id string) (bool, domain.SERVICE_ORDER_STATUS, error) {
	tx, err := postgres.GetOneTimeTransaction(ctx)
	if err != nil {
		return false, "", err
	}

	var status domain.SERVICE_ORDER_STATUS
	if err = tx.QueryRowContext(ctx, serviceOrderExistsQuery, id).Scan(&status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, "", domain.ErrServiceOrderNotFound
		}
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

func (r *repository) RemoveSupplyLink(ctx context.Context, serviceOrderID string, supplyID string) (int, error) {
	tx, err := postgres.GetTransaction(ctx)
	if err != nil {
		return 0, err
	}

	var qty int
	err = tx.QueryRowContext(ctx, deleteServiceOrderSuppliesQuery, serviceOrderID, supplyID).Scan(&qty)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, domain.ErrServiceOrderSupplyNotFound
		}
		return 0, pgPkg.Error(ctx, err)
	}
	return qty, nil
}

func (r *repository) FindByID(ctx context.Context, id string) (domain.ServiceOrder, error) {
	db, err := postgres.GetOneTimeTransaction(ctx)
	if err != nil {
		return domain.ServiceOrder{}, err
	}

	var so domain.ServiceOrder
	so.Customer = new(domain.Customer)
	so.Customer.User = new(domain.User)
	so.Vehicle = new(domain.Vehicle)
	var status string
	if err = db.QueryRowContext(ctx, serviceOrderFindByIDQuery, id).Scan(
		&so.ID, &status, &so.TotalAmount,
		&so.Customer.ID, &so.Customer.UserID, &so.Customer.Type,
		&so.Customer.CPF, &so.Customer.CNPJ, &so.Customer.CompanyName, &so.Customer.Phone,
		&so.Customer.User.Name, &so.Customer.User.Email,
		&so.Vehicle.ID, &so.Vehicle.LicensePlate, &so.Vehicle.Model,
		&so.Vehicle.Brand, &so.Vehicle.Year,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ServiceOrder{}, domain.ErrServiceOrderNotFound
		}
		return domain.ServiceOrder{}, pgPkg.Error(ctx, err)
	}

	so.Status = domain.StringToServiceOrderStatus(status)

	return so, nil
}

func (r *repository) Count(ctx context.Context, params *domain.ServiceOrderFilterParams) (int64, error) {
	tx, err := postgres.GetOneTimeTransaction(ctx)
	if err != nil {
		return 0, err
	}

	queryBuilder := dbPkg.QueryBuilder(countServiceOrderQuery)
	queryBuilder.Add("so.status =", params.Status)
	queryBuilder.Add("so.customer_id =", params.CustomerID)
	queryBuilder.Add("so.vehicle_id =", params.VehicleID)
	query, args := queryBuilder.Build()

	var total int64
	if err = tx.QueryRowContext(ctx, query, args...).Scan(&total); err != nil {
		return 0, pgPkg.Error(ctx, err)
	}

	return total, nil
}

func (r *repository) Search(ctx context.Context, params *domain.ServiceOrderFilterParams) ([]domain.ServiceOrder, error) {
	tx, err := postgres.GetOneTimeTransaction(ctx)
	if err != nil {
		return nil, err
	}

	queryBuilder := dbPkg.QueryBuilder(searchServiceOrderQuery).
		OrderBy("so.created_at", dbPkg.ASC).
		AddPagination(params.Limit, params.Offset)
	queryBuilder.Add("so.status =", params.Status)
	queryBuilder.Add("so.customer_id =", params.CustomerID)
	queryBuilder.Add("so.vehicle_id =", params.VehicleID)
	query, args := queryBuilder.Build()

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pgPkg.Error(ctx, err)
	}
	defer func() { _ = rows.Close() }()

	items := make([]domain.ServiceOrder, 0, params.Limit)
	for rows.Next() {
		var so domain.ServiceOrder
		so.Customer = new(domain.Customer)
		so.Customer.User = new(domain.User)
		so.Vehicle = new(domain.Vehicle)
		var status string

		if err = rows.Scan(
			&so.ID,
			&status,
			&so.TotalAmount,
			&so.Customer.ID,
			&so.Customer.UserID,
			&so.Customer.Type,
			&so.Customer.CPF,
			&so.Customer.CNPJ,
			&so.Customer.CompanyName,
			&so.Customer.Phone,
			&so.Customer.User.Name,
			&so.Customer.User.Email,
			&so.Vehicle.ID,
			&so.Vehicle.LicensePlate,
			&so.Vehicle.Brand,
			&so.Vehicle.Model,
			&so.Vehicle.Year,
		); err != nil {
			return nil, pgPkg.Error(ctx, err)
		}

		so.Status = domain.StringToServiceOrderStatus(status)
		items = append(items, so)
	}

	if err = rows.Err(); err != nil {
		return nil, pgPkg.Error(ctx, err)
	}

	return items, nil
}
