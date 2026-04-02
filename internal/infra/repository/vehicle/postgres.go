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
