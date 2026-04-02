package vehicle

import (
	"context"
	"database/sql"
	"errors"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db/postgres"
	pgPkg "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/db/postgres"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"go.uber.org/zap"
)

type vehicleRepository struct{}

func NewVehicleRepository() *vehicleRepository {
	return &vehicleRepository{}
}

// TODO: Add query building in this repository

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

func (r *vehicleRepository) Find(ctx context.Context, licensePlate string) (*domain.Vehicle, error) {
	tx, err := postgres.GetOneTimeTransaction(ctx)
	if err != nil {
		return nil, err
	}

	v := &domain.Vehicle{}
	if err = tx.QueryRowContext(ctx, selectVehicle, licensePlate).Scan(&v.ID, &v.LicensePlate, &v.Brand, &v.Model, &v.Year, &v.CustomerId); err != nil {
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
