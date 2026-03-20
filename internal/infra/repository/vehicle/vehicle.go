package vehicle

import (
	"context"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db/postgres"
	pgPkg "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/db/postgres"
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
		vehicle.Id,
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
