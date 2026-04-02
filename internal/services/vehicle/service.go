package vehicle

import (
	"context"
	"fmt"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow"
)

type vehicleService struct {
	uow               uow.Executor
	vehicleRepository domain.VehicleRepository
}

func NewService(uow uow.Executor, vehicleRepository domain.VehicleRepository) *vehicleService {
	return &vehicleService{
		uow:               uow,
		vehicleRepository: vehicleRepository,
	}
}

func (s *vehicleService) Create(ctx context.Context, vehicle *domain.Vehicle) error {
	if err := vehicle.Validate(); err != nil {
		errMessage := fmt.Errorf("vehicle validation failed: %w", err)
		logger.Of(ctx).Error(errMessage)
		return errMessage
	}

	return s.uow.Execute(ctx,
		func(txCtx context.Context) error {
			return s.vehicleRepository.Save(txCtx, vehicle)
		},
	)
}

func (s *vehicleService) FindByLicensePlate(ctx context.Context, licensePlate string) (*domain.Vehicle, error) {
	if err := domain.ValidateLicensePlate(licensePlate); err != nil {
		return nil, err
	}

	var result *domain.Vehicle

	err := s.uow.Execute(ctx, func(ctx context.Context) error {
		vehicle, err := s.vehicleRepository.Find(ctx, licensePlate)
		if err != nil {
			return err
		}
		result = vehicle
		return nil
	})
	if err != nil {
		logger.Of(ctx).Error(fmt.Errorf("error find vehicle: %w", err))
		return nil, err
	}

	return result, nil
}

func (s *vehicleService) Update(ctx context.Context, v *domain.Vehicle) error {
	if err := v.Validate(); err != nil {
		errMessage := fmt.Errorf("vehicle validation failed: %w", err)
		logger.Of(ctx).Error(errMessage)
		return errMessage
	}

	return s.uow.Execute(ctx,
		func(txCtx context.Context) error {
			return s.vehicleRepository.Update(txCtx, v)
		},
	)
}

func (s *vehicleService) Delete(c context.Context, id string) error {
	if err := s.uow.Execute(c, func(ctx context.Context) error { return s.vehicleRepository.Delete(ctx, id) }); err != nil {
		logger.Of(c).Error(err)
		return err
	}
	return nil
}
