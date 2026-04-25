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
	if err := s.checkID(serviceOrderID); err != nil {
		return nil, err
	}

	var works []domain.Work
	if err := s.uow.Execute(ctx, func(ctx context.Context) error {
		ok, _, err := s.repo.ExistsByID(ctx, serviceOrderID)
		if err != nil {
			return err
		}
		if !ok {
			return domain.ErrServiceOrderNotFound
		}
		works, err = s.repo.ListWorksByServiceOrderID(ctx, serviceOrderID)
		return err
	}); err != nil {
		return nil, err
	}
	return works, nil
}

func (s *svc) AddWorks(ctx context.Context, serviceOrderID string, workIDs []string) error {
	if err := s.checkID(serviceOrderID); err != nil {
		return err
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
	if err := s.checkID(serviceOrderID); err != nil {
		return err
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
	if err := s.checkID(serviceOrderID); err != nil {
		return nil, err
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
	if len(supplies) == 0 {
		return domain.ErrEmptyServicesList
	}

	if err := s.checkID(serviceOrderID); err != nil {
		return err
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

			supply, e := s.supplyRepo.FindById(ctx, supplyId)
			if e != nil {
				return e
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
	if err := s.checkID(serviceOrderID); err != nil {
		return err
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

		if err = s.repo.RemoveSupplyLink(ctx, serviceOrderID, supplyID); err != nil {
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

func (s *svc) SendToCustomerApproval(ctx context.Context, serviceOrderID string) error {
	if err := s.checkID(serviceOrderID); err != nil {
		return err
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
	if err := s.checkID(serviceOrderID); err != nil {
		return err
	}

	so, err := s.repo.FindByID(ctx, serviceOrderID)
	if err != nil {
		return err
	}

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

func (s *svc) checkID(serviceOrderID string) error {
	serviceOrderID = strings.TrimSpace(serviceOrderID)
	if serviceOrderID == "" {
		return domain.ErrInvalidServiceOrderId
	}
	return nil
}

func (s *svc) Accept(ctx context.Context, serviceOrderID, userID string) error {
	eg := errgroup.Group{}

	var so domain.ServiceOrder
	eg.Go(func() (err error) {
		so, err = s.repo.FindByID(ctx, serviceOrderID)
		return
	})

	var cust domain.Customer
	eg.Go(func() (err error) {
		cust, err = s.customerService.GetByUserID(ctx, userID)
		return
	})

	if err := eg.Wait(); err != nil {
		return err
	}

	if so.Customer.ID != cust.ID {
		return domain.ErrInvalidCustomerId
	}

	if so.Status != domain.SERVICE_ORDER_STATUS_AWAITING_APPROVAL {
		return domain.ErrServiceOrderNotAwaitingApproval
	}

	so.Status = domain.SERVICE_ORDER_STATUS_IN_PROGRESS

	// TODO: NOTIFY MECHANICAL TO START SERVICE

	return s.uow.Execute(ctx, func(c context.Context) error {
		return s.repo.Save(c, &so)
	})
}
func (s *svc) Reject(ctx context.Context, serviceOrderID, userID string) error {
	eg := errgroup.Group{}

	var so domain.ServiceOrder
	eg.Go(func() (err error) {
		so, err = s.repo.FindByID(ctx, serviceOrderID)
		return
	})

	var cust domain.Customer
	eg.Go(func() (err error) {
		cust, err = s.customerService.GetByUserID(ctx, userID)
		return
	})

	if err := eg.Wait(); err != nil {
		return err
	}

	if so.Customer.ID != cust.ID {
		return domain.ErrInvalidCustomerId
	}

	if so.Status != domain.SERVICE_ORDER_STATUS_AWAITING_APPROVAL {
		return domain.ErrServiceOrderNotAwaitingApproval
	}

	so.Status = domain.SERVICE_ORDER_STATUS_REJECTED

	// TODO: RELEASE STOCK

	return s.uow.Execute(ctx, func(c context.Context) error {
		return s.repo.Save(c, &so)
	})
}
