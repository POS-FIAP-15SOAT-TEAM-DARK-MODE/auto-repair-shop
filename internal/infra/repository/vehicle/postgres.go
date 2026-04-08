package vehicle

import (
	"context"
	"database/sql"
	"errors"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db/postgres"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/db"
	pgPkg "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/db/postgres"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"go.uber.org/zap"
)

type dbRunner interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type vehicleRepository struct{}

func NewVehicleRepository() *vehicleRepository {
	return &vehicleRepository{}
}

func (r *vehicleRepository) Save(ctx context.Context, vehicle *domain.Vehicle) error {
	tx, err := postgres.GetTransaction(ctx)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		insertNewVehicle,
		vehicle.ID,
		vehicle.LicensePlate,
		vehicle.Brand,
		vehicle.Model,
		vehicle.Year,
		vehicle.CustomerId,
	)
	if err != nil {
		return pgPkg.Error(ctx, err)
	}

	return nil
}

func (r *vehicleRepository) Find(ctx context.Context, params domain.FindVehicleParams) (*domain.Vehicle, error) {
	tx, err := postgres.GetOneTimeTransaction(ctx)
	if err != nil {
		return nil, err
	}

	qb := db.QueryBuilder(selectVehicle)
	if params.LicensePlate != "" {
		qb.Add("license_plate = ", params.LicensePlate)
	}

	if params.ID != "" {
		qb.Add("id = ", params.ID)
	}

	query, args := qb.Build()
	v := &domain.Vehicle{}
	if err = tx.QueryRowContext(ctx, query, args...).Scan(&v.ID, &v.LicensePlate, &v.Brand, &v.Model, &v.Year, &v.CustomerId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrVehicleNotFound
		}
		logger.Of(ctx).Warn("vehicle repository", zap.String("operation", "get_vehicle"), zap.Error(err))
		return nil, pgPkg.Error(ctx, err)
	}

	return v, nil
}

func (r *vehicleRepository) Update(ctx context.Context, vehicle *domain.Vehicle) error {
	tx, err := postgres.GetTransaction(ctx)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		updateVehicle,
		vehicle.ID,
		vehicle.LicensePlate,
		vehicle.Brand,
		vehicle.Model,
		vehicle.Year,
		vehicle.CustomerId,
	)
	if err != nil {
		return pgPkg.Error(ctx, err)
	}

	return nil
}

func (r *vehicleRepository) Delete(ctx context.Context, id string) error {
	tx, err := postgres.GetTransaction(ctx)
	if err != nil {
		return err
	}

	if _, err = tx.ExecContext(ctx, deleteVehicle, id); err != nil {
		logger.Of(ctx).Warn("vehicle repository", zap.String("operation", "delete"), zap.Error(err))
		return pgPkg.Error(ctx, err)
	}

	return nil
}

func (r *vehicleRepository) Search(ctx context.Context, params *domain.SearchVehicleParams) ([]domain.Vehicle, error) {
	runner := instanceDbRunner(ctx)

	qb := db.QueryBuilder(selectVehicle).AddPagination(params.Limit, params.Offset)
	if params.CustomerId != "" {
		qb.Add("customer_id = ", params.CustomerId)
	}
	query, args := qb.Build()

	rows, err := runner.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pgPkg.Error(ctx, err)
	}

	vehicles := make([]domain.Vehicle, 0, params.Limit)
	for rows.Next() {
		if err = rows.Err(); err != nil {
			return nil, pgPkg.Error(ctx, err)
		}

		var v domain.Vehicle
		if err = rows.Scan(
			&v.ID,
			&v.LicensePlate,
			&v.Model,
			&v.Brand,
			&v.Year,
			&v.CustomerId,
		); err != nil {
			return nil, pgPkg.Error(ctx, err)
		}
		vehicles = append(vehicles, v)
	}

	return vehicles, nil
}

func (r *vehicleRepository) Count(ctx context.Context, params *domain.SearchVehicleParams) (int64, error) {
	runner := instanceDbRunner(ctx)

	qb := db.QueryBuilder(countVehicles)
	if params.CustomerId != "" {
		qb.Add("customer_id = ", params.CustomerId)
	}
	query, args := qb.Build()

	var total int64
	if err := runner.QueryRowContext(ctx, query, args...).Scan(&total); err != nil {
		return 0, pgPkg.Error(ctx, err)
	}

	return total, nil
}

func instanceDbRunner(ctx context.Context) dbRunner {
	var runner dbRunner
	if tx, err := postgres.GetTransaction(ctx); err == nil {
		runner = tx
	} else {
		runner = postgres.Connect()
	}
	return runner
}
