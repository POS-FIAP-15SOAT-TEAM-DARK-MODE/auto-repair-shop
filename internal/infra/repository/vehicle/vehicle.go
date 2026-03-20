package vehicle

import (
	"context"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
)

type vehicleRepository struct{}

func NewVehicleRepository() *vehicleRepository {
	return &vehicleRepository{}
}

func (r *vehicleRepository) Save(ctx context.Context, vehicle *domain.Vehicle) error {
	return nil
}
