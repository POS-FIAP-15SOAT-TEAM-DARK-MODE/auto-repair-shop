package service_order

import (
	"context"
	"strings"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow"
)

type svc struct {
	uow             uow.Executor
	repo            domain.ServiceOrderRepository
	workRepo        domain.WorkRepository
	customerService domain.CustomerService
	vehicleService  domain.VehicleService
}

func Service(
	uow uow.Executor,
	repo domain.ServiceOrderRepository,
	workRepo domain.WorkRepository,
	customerService domain.CustomerService,
	vehicleService domain.VehicleService) *svc {
	return &svc{uow, repo, workRepo, customerService, vehicleService}
}

func (s *svc) Create(ctx context.Context, customerId string, vehicleId string) (domain.ServiceOrder, error) {
	customer, err := s.customerService.GetByID(ctx, customerId)
	if err != nil {
		return domain.ServiceOrder{}, err
	}

	vehicle, err := s.vehicleService.FindByID(ctx, vehicleId)
	if err != nil {
		return domain.ServiceOrder{}, err
	}

	so := domain.NewServiceOrder(&customer, vehicle)
	if err = s.uow.Execute(ctx, func(ctx context.Context) error {
		return s.repo.Save(ctx, so)
	}); err != nil {
		return domain.ServiceOrder{}, err
	}

	return *so, nil
}

func (s *svc) ListWorks(ctx context.Context, serviceOrderID string) ([]domain.Work, error) {
	serviceOrderID = strings.TrimSpace(serviceOrderID)
	if serviceOrderID == "" {
		return nil, domain.ErrInvalidServiceOrderId
	}

	var works []domain.Work
	err := s.uow.Execute(ctx, func(ctx context.Context) error {
		ok, err := s.repo.ExistsByID(ctx, serviceOrderID)
		if err != nil {
			return err
		}
		if !ok {
			return domain.ErrServiceOrderNotFound
		}
		works, err = s.repo.ListWorksByServiceOrderID(ctx, serviceOrderID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return works, nil
}

func (s *svc) AddWorks(ctx context.Context, serviceOrderID string, workIDs []string) error {
	serviceOrderID = strings.TrimSpace(serviceOrderID)
	if serviceOrderID == "" {
		return domain.ErrInvalidServiceOrderId
	}
	if len(workIDs) == 0 {
		return domain.ErrEmptyServicesList
	}

	return s.uow.Execute(ctx, func(ctx context.Context) error {
		ok, err := s.repo.ExistsByID(ctx, serviceOrderID)
		if err != nil {
			return err
		}
		if !ok {
			return domain.ErrServiceOrderNotFound
		}

		for _, rawID := range workIDs {
			workID := strings.TrimSpace(rawID)
			if workID == "" {
				return domain.ErrInvalidWorkId
			}
			w, err := s.workRepo.FindByID(ctx, workID)
			if err != nil {
				return err
			}
			if err = s.repo.AddWorkLink(ctx, serviceOrderID, w.ID, w.Price); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *svc) RemoveWork(ctx context.Context, serviceOrderID, workID string) error {
	serviceOrderID = strings.TrimSpace(serviceOrderID)
	workID = strings.TrimSpace(workID)
	if serviceOrderID == "" {
		return domain.ErrInvalidServiceOrderId
	}
	if workID == "" {
		return domain.ErrInvalidWorkId
	}

	return s.uow.Execute(ctx, func(ctx context.Context) error {
		ok, err := s.repo.ExistsByID(ctx, serviceOrderID)
		if err != nil {
			return err
		}
		if !ok {
			return domain.ErrServiceOrderNotFound
		}
		return s.repo.RemoveWorkLink(ctx, serviceOrderID, workID)
	})
}
