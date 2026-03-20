package vehicle

import (
	"context"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow"
)

type vehicleService struct {
	uow                uow.Executor
	vehicleRepository  domain.VehicleRepository
	customerRepository domain.CustomerRepository
}

func NewService(uow uow.Executor, vehicleRepository domain.VehicleRepository) *vehicleService {
	return &vehicleService{
		uow:               uow,
		vehicleRepository: vehicleRepository,
	}
}

func (s *vehicleService) Create(ctx context.Context, vehicle *domain.Vehicle) error {
	// TODO: check customer_id is in database

	return nil
}
