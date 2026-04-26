package service_order

import (
	"context"
	"strings"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type svc struct {
	uow             uow.Executor
	repo            domain.ServiceOrderRepository
	workRepo        domain.WorkRepository
	supplyRepo      domain.SupplyRepository
	workHistoryRepo domain.ServiceOrderHistoryRepository
	customerService domain.CustomerService
	vehicleService  domain.VehicleService
}

func Service(
	uow uow.Executor,
	repo domain.ServiceOrderRepository,
	workRepo domain.WorkRepository,
	supplyRepo domain.SupplyRepository,
	workHistoryRepo domain.ServiceOrderHistoryRepository,
	customerService domain.CustomerService,
	vehicleService domain.VehicleService) *svc {
	return &svc{uow, repo, workRepo, supplyRepo, workHistoryRepo, customerService, vehicleService}
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

	err := s.uow.Execute(ctx, func(ctx context.Context) error {
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

			w, e := s.workRepo.FindByID(ctx, workID)
			if e != nil {
				return e
			}

			if e = s.repo.AddWorkLink(ctx, serviceOrderID, w.ID, w.Price); e != nil {
				return e
			}

			if e = s.workHistoryRepo.InsertWorkHistory(ctx, serviceOrderID, w.ID, domain.SERVICE_ORDER_STATUS_NEW); e != nil {
				return e
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	go func(id string) {
		if err = s.reviewOSPricing(context.Background(), id); err != nil {
			logger.Global().Error(err, zap.String("serviceOrderID", serviceOrderID))
		}
	}(serviceOrderID)

	return nil
}

func (s *svc) RemoveWork(ctx context.Context, serviceOrderID, workID string) error {
	serviceOrderID = strings.TrimSpace(serviceOrderID)
	if serviceOrderID == "" {
		return domain.ErrInvalidServiceOrderId
	}

	workID = strings.TrimSpace(workID)
	if workID == "" {
		return domain.ErrInvalidWorkId
	}

	err := s.uow.Execute(ctx, func(ctx context.Context) error {
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

		if err = s.repo.RemoveWorkLink(ctx, serviceOrderID, workID); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	go func(id string) {
		if err = s.reviewOSPricing(context.Background(), id); err != nil {
			logger.Global().Error(err, zap.String("serviceOrderID", serviceOrderID))
		}
	}(serviceOrderID)

	return nil
}

func (s *svc) ListSupplies(ctx context.Context, serviceOrderID string) ([]domain.Supply, error) {
	serviceOrderID = strings.TrimSpace(serviceOrderID)
	if serviceOrderID == "" {
		return nil, domain.ErrInvalidServiceOrderId
	}

	ok, _, err := s.repo.ExistsByID(ctx, serviceOrderID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrServiceOrderNotFound
	}

	return s.repo.ListSuppliesByServiceOrderID(ctx, serviceOrderID)
}

func (s *svc) AddSupplies(ctx context.Context, serviceOrderID string, supplies []domain.AddSupply) error {
	serviceOrderID = strings.TrimSpace(serviceOrderID)
	if serviceOrderID == "" {
		return domain.ErrInvalidServiceOrderId
	}

	if len(supplies) == 0 {
		return domain.ErrEmptyServicesList
	}

	err := s.uow.Execute(ctx, func(ctx context.Context) error {
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

			if sup.Amount <= 0 {
				return domain.ErrInvalidSupplyAmount
			}

			supply, e := s.supplyRepo.FindById(ctx, supplyId)
			if e != nil {
				return e
			}

			if supply.StockQuantity < sup.Amount {
				return domain.ErrSupplyOutOfStock
			}

			if err = s.repo.AddSupplyLink(ctx, serviceOrderID, sup.ID, sup.Amount, supply.UnitPrice); err != nil {
				return err
			}

			if err = s.supplyRepo.DecrementStock(ctx, supply.ID, sup.Amount); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	go func(id string) {
		if err = s.reviewOSPricing(context.Background(), id); err != nil {
			logger.Global().Error(err, zap.String("serviceOrderID", serviceOrderID))
		}
	}(serviceOrderID)

	return nil
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

	err := s.uow.Execute(ctx, func(ctx context.Context) error {
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

		qty, err := s.repo.RemoveSupplyLink(ctx, serviceOrderID, supplyID)
		if err != nil {
			return err
		}

		return s.supplyRepo.RestoreStock(ctx, supplyID, qty)
	})

	if err != nil {
		return err
	}

	go func(id string) {
		if err = s.reviewOSPricing(context.Background(), id); err != nil {
			logger.Global().Error(err, zap.String("serviceOrderID", serviceOrderID))
		}
	}(serviceOrderID)

	return nil
}

func (s *svc) SendToCustomerApproval(ctx context.Context, serviceOrderID string) error {
	serviceOrderID = strings.TrimSpace(serviceOrderID)
	if serviceOrderID == "" {
		return domain.ErrInvalidServiceOrderId
	}

	err := s.uow.Execute(ctx, func(ctx context.Context) error {
		so, err := s.repo.FindByID(ctx, serviceOrderID)
		if err != nil {
			return err
		}

		if so.Status != domain.SERVICE_ORDER_STATUS_IN_DIAGNOSIS {
			return domain.ErrServiceOrderNotInDiagnosis
		}

		so.Status = domain.SERVICE_ORDER_STATUS_AWAITING_APPROVAL

		// TODO: SEND CUSTOMER NOTIFICATION (EMAIL / SMS / ETC)

		if err = s.repo.Save(ctx, &so); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	return s.reviewOSPricing(ctx, serviceOrderID)
}

func (s *svc) reviewOSPricing(ctx context.Context, serviceOrderID string) error {
	serviceOrderID = strings.TrimSpace(serviceOrderID)
	if serviceOrderID == "" {
		return domain.ErrInvalidServiceOrderId
	}

	so, err := s.repo.FindByID(ctx, serviceOrderID)
	if err != nil {
		return err
	}

	so.ResetPricing()

	var eg errgroup.Group
	eg.Go(func() error {
		works, e := s.repo.ListWorksByServiceOrderID(ctx, serviceOrderID)
		if e != nil {
			return e
		}

		for _, work := range works {
			so.SumWorkValue(work)
		}
		return nil
	})

	eg.Go(func() error {
		supplies, e := s.repo.ListSuppliesByServiceOrderID(ctx, serviceOrderID)
		if e != nil {
			return e
		}

		for _, supply := range supplies {
			so.SumSupplyValue(supply)
		}
		return nil
	})

	if err = eg.Wait(); err != nil {
		return err
	}

	return s.uow.Execute(ctx, func(c context.Context) error {
		return s.repo.Save(c, &so)
	})
}

func (s *svc) Accept(c context.Context, serviceOrderID, userID string) error {
	serviceOrderID = strings.TrimSpace(serviceOrderID)
	if serviceOrderID == "" {
		return domain.ErrInvalidServiceOrderId
	}

	return s.uow.Execute(c, func(ctx context.Context) error {
		eg := errgroup.Group{}
		var so domain.ServiceOrder
		eg.Go(func() (err error) {
			so, err = s.repo.FindByID(ctx, serviceOrderID)
			if err != nil {
				return
			}

			if so.Customer == nil {
				return domain.ErrServiceOrderNotFound
			}
			return
		})

		var cust domain.Customer
		eg.Go(func() (err error) {
			cust, err = s.customerService.GetByUserID(ctx, userID)
			if err != nil {
				return
			}

			if cust.ID == "" {
				return domain.ErrCustomerNotFound
			}
			return
		})

		if err := eg.Wait(); err != nil {
			return err
		}

		if so.Customer.ID != cust.ID {
			return domain.ErrInvalidCustomerProperty
		}

		if so.Status != domain.SERVICE_ORDER_STATUS_AWAITING_APPROVAL {
			return domain.ErrServiceOrderNotAwaitingApproval
		}

		so.Status = domain.SERVICE_ORDER_STATUS_IN_PROGRESS

		// TODO: NOTIFY MECHANICAL TO START SERVICE (SMS / EMAIL / WPP)
		return s.repo.Save(ctx, &so)
	})
}

func (s *svc) Reject(c context.Context, serviceOrderID, userID string) error {
	serviceOrderID = strings.TrimSpace(serviceOrderID)
	if serviceOrderID == "" {
		return domain.ErrInvalidServiceOrderId
	}

	return s.uow.Execute(c, func(ctx context.Context) error {
		eg := errgroup.Group{}
		var so domain.ServiceOrder
		eg.Go(func() (err error) {
			so, err = s.repo.FindByID(ctx, serviceOrderID)
			if err != nil {
				return
			}

			if so.Customer == nil {
				return domain.ErrServiceOrderNotFound
			}
			return
		})

		var cust domain.Customer
		eg.Go(func() (err error) {
			cust, err = s.customerService.GetByUserID(ctx, userID)
			if err != nil {
				return
			}

			if cust.ID == "" {
				return domain.ErrCustomerNotFound
			}
			return
		})

		var supplies []domain.Supply
		eg.Go(func() (err error) {
			supplies, err = s.ListSupplies(ctx, serviceOrderID)
			return
		})

		if err := eg.Wait(); err != nil {
			return err
		}

		if so.Customer.ID != cust.ID {
			return domain.ErrInvalidCustomerProperty
		}

		if so.Status != domain.SERVICE_ORDER_STATUS_AWAITING_APPROVAL {
			return domain.ErrServiceOrderNotAwaitingApproval
		}

		so.Status = domain.SERVICE_ORDER_STATUS_REJECTED

		for _, sup := range supplies {
			if err := s.supplyRepo.RestoreStock(ctx, sup.ID, sup.StockQuantity); err != nil {
				return err
			}
		}

		return s.repo.Save(ctx, &so)
	})
}

func (s *svc) Deliver(c context.Context, serviceOrderID string) error {
	serviceOrderID = strings.TrimSpace(serviceOrderID)
	if serviceOrderID == "" {
		return domain.ErrInvalidServiceOrderId
	}

	return s.uow.Execute(c, func(ctx context.Context) error {
		so, err := s.repo.FindByID(ctx, serviceOrderID)
		if err != nil {
			return err
		}

		if so.Status != domain.SERVICE_ORDER_STATUS_COMPLETED {
			return domain.ErrServiceOrderNotCompleted
		}

		so.Status = domain.SERVICE_ORDER_STATUS_DELIVERED

		return s.repo.Save(ctx, &so)
	})
}

func (s *svc) Cancel(c context.Context, serviceOrderID string) error {
	serviceOrderID = strings.TrimSpace(serviceOrderID)
	if serviceOrderID == "" {
		return domain.ErrInvalidServiceOrderId
	}

	return s.uow.Execute(c, func(ctx context.Context) error {
		so, err := s.repo.FindByID(ctx, serviceOrderID)
		if err != nil {
			return err
		}

		if so.Customer == nil {
			return domain.ErrServiceOrderNotFound
		}

		if so.Status.IsCancelable() {
			return domain.ErrServiceOrderNotCancelable
		}

		so.Status = domain.SERVICE_ORDER_STATUS_CANCELLED

		return s.repo.Save(ctx, &so)
	})
}
