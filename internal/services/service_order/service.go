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
	supplyRepo      domain.SupplyRepository
	customerService domain.CustomerService
	vehicleService  domain.VehicleService
}

func Service(
	uow uow.Executor,
	repo domain.ServiceOrderRepository,
	workRepo domain.WorkRepository,
	supplyRepo domain.SupplyRepository,
	customerService domain.CustomerService,
	vehicleService domain.VehicleService) *svc {
	return &svc{uow, repo, workRepo, supplyRepo, customerService, vehicleService}
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
		ok, _, err := s.repo.ExistsByID(ctx, serviceOrderID)
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
		ok, status, err := s.repo.ExistsByID(ctx, serviceOrderID)
		if err != nil {
			return err
		}
		if !ok {
			return domain.ErrServiceOrderNotFound
		}

		if status != domain.SERVICE_ORDER_STATUS_NEW {
			return domain.ErrServiceOrderNotNew
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
		ok, status, err := s.repo.ExistsByID(ctx, serviceOrderID)
		if err != nil {
			return err
		}
		if !ok {
			return domain.ErrServiceOrderNotFound
		}
		if status != domain.SERVICE_ORDER_STATUS_NEW {
			return domain.ErrServiceOrderNotNew
		}
		return s.repo.RemoveWorkLink(ctx, serviceOrderID, workID)
	})
}

func (s *svc) ListSupplies(ctx context.Context, serviceOrderID string) ([]domain.Supply, error) {
	serviceOrderID = strings.TrimSpace(serviceOrderID)
	if serviceOrderID == "" {
		return nil, domain.ErrInvalidServiceOrderId
	}

	var supplies []domain.Supply
	err := s.uow.Execute(ctx, func(ctx context.Context) error {
		ok, _, err := s.repo.ExistsByID(ctx, serviceOrderID)
		if err != nil {
			return err
		}
		if !ok {
			return domain.ErrServiceOrderNotFound
		}
		supplies, err = s.repo.ListSuppliesByServiceOrderID(ctx, serviceOrderID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return supplies, nil
}

func (s *svc) AddSupplies(ctx context.Context, serviceOrderID string, supplies []domain.AddSupply) error {
	serviceOrderID = strings.TrimSpace(serviceOrderID)
	if serviceOrderID == "" {
		return domain.ErrInvalidServiceOrderId
	}
	if len(supplies) == 0 {
		return domain.ErrEmptyServicesList
	}

	return s.uow.Execute(ctx, func(ctx context.Context) error {
		ok, status, err := s.repo.ExistsByID(ctx, serviceOrderID)
		if err != nil {
			return err
		}

		if !ok {
			return domain.ErrServiceOrderNotFound
		}

		if status != domain.SERVICE_ORDER_STATUS_IN_DIAGNOSIS {
			return domain.ErrServiceOrderNotInDiagnosis
		}

		for _, sup := range supplies {
			supplyId := strings.TrimSpace(sup.ID)
			if supplyId == "" {
				return domain.ErrInvalidSupplyID
			}

			supply, err := s.supplyRepo.FindById(ctx, supplyId)
			if err != nil {
				return err
			}

			amount := sup.Amount
			if amount > supply.StockQuantity {
				amount = supply.StockQuantity
			}

			if err = s.repo.AddSupplyLink(ctx, serviceOrderID, sup.ID, amount, supply.UnitPrice); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *svc) RemoveSupply(ctx context.Context, serviceOrderID, supplyID string) error {
	serviceOrderID = strings.TrimSpace(serviceOrderID)
	if serviceOrderID == "" {
		return domain.ErrInvalidServiceOrderId
	}

	supplyID = strings.TrimSpace(supplyID)
	if supplyID == "" {
		return domain.ErrInvalidSupplyID
	}

	return s.uow.Execute(ctx, func(ctx context.Context) error {
		ok, status, err := s.repo.ExistsByID(ctx, serviceOrderID)
		if err != nil {
			return err
		}
		if !ok {
			return domain.ErrServiceOrderNotFound
		}
		if status != domain.SERVICE_ORDER_STATUS_IN_DIAGNOSIS {
			return domain.ErrServiceOrderNotInDiagnosis
		}
		return s.repo.RemoveSupplyLink(ctx, serviceOrderID, supplyID)
	})
}

func (s *svc) SendToCustomerApproval(ctx context.Context, serviceOrderID string) error {
	serviceOrderID = strings.TrimSpace(serviceOrderID)
	if serviceOrderID == "" {
		return domain.ErrInvalidServiceOrderId
	}

	return s.uow.Execute(ctx, func(ctx context.Context) error {
		so, err := s.repo.FindByID(ctx, serviceOrderID)
		if err != nil {
			return err
		}

		if so.Status != domain.SERVICE_ORDER_STATUS_IN_DIAGNOSIS {
			return domain.ErrServiceOrderNotInDiagnosis
		}

		so.Status = domain.SERVICE_ORDER_STATUS_AWAITING_APPROVAL

		// TODO: SEND CUSTOMER NOTIFICATION (EMAIL / SMS / ETC)

		return s.repo.Save(ctx, &so)
	})
}
