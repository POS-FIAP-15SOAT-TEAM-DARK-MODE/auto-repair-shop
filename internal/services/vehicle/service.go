package vehicle

import (
	"context"
	"fmt"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow"
)

type vehicleService struct {
	uow                uow.Executor
	vehicleRepository  domain.VehicleRepository
	customerRepository domain.CustomerRepository
}

func NewService(uow uow.Executor, vehicleRepository domain.VehicleRepository, customerRepository domain.CustomerRepository) *vehicleService {
	return &vehicleService{
		uow:                uow,
		vehicleRepository:  vehicleRepository,
		customerRepository: customerRepository,
	}
}

func (s *vehicleService) Create(ctx context.Context, vehicle *domain.Vehicle) error {
	if err := vehicle.Validate(); err != nil {
		errMessage := fmt.Errorf("vehicle validation failed: %w", err)
		logger.Of(ctx).Error(errMessage)
		return errMessage
	}

	// TODO: validate if customer_id exists in database
	if err := s.uow.Execute(
		ctx,
		s.createInsertStep(vehicle),
	); err != nil {
		logger.Of(ctx).Error(err)
		return err
	}
	return nil
}

func (s *vehicleService) createInsertStep(v *domain.Vehicle) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		if err := s.vehicleRepository.Save(ctx, v); err != nil {
			return fmt.Errorf("vehicle save failed: %w", err)
		}
		return nil
	}
}
