package service_order

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type svc struct {
	uow               uow.Executor
	repo              domain.ServiceOrderRepository
	workRepo          domain.WorkRepository
	supplyRepo        domain.SupplyRepository
	workHistoryRepo   domain.ServiceOrderHistoryRepository
	workSOHistoryRepo domain.WorkServiceOrderHistoryRepository
	customerService   domain.CustomerService
	vehicleService    domain.VehicleService
}

func Service(
	uow uow.Executor,
	repo domain.ServiceOrderRepository,
	workRepo domain.WorkRepository,
	supplyRepo domain.SupplyRepository,
	workHistoryRepo domain.ServiceOrderHistoryRepository,
	workSOHistoryRepo domain.WorkServiceOrderHistoryRepository,
	customerService domain.CustomerService,
	vehicleService domain.VehicleService) *svc {
	return &svc{uow, repo, workRepo, supplyRepo, workHistoryRepo, workSOHistoryRepo, customerService, vehicleService}
}

func (s *svc) logValidationError(ctx context.Context, operation string, err error) {
	logger.Of(ctx).Warn("service_order.validation_failed",
		zap.String("operation", operation),
		zap.String("entity", "service_order"),
		zap.Error(err),
	)
}

func (s *svc) logCreateSuccess(ctx context.Context, serviceOrderID string, status domain.SERVICE_ORDER_STATUS) {
	logger.Of(ctx).Info("service_order.created",
		zap.String("operation", "create_service_order"),
		zap.String("entity", "service_order"),
		zap.String("service_order_id", serviceOrderID),
		zap.String("status", status.String()),
	)
}

func (s *svc) logStatusTransition(ctx context.Context, operation string, serviceOrderID string, previousStatus, newStatus domain.SERVICE_ORDER_STATUS) {
	logger.Of(ctx).Info("service_order.status_transition",
		zap.String("operation", operation),
		zap.String("entity", "service_order"),
		zap.String("service_order_id", serviceOrderID),
		zap.String("previous_status", previousStatus.String()),
		zap.String("new_status", newStatus.String()),
	)
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

	s.logCreateSuccess(ctx, so.ID, so.Status)
	return *so, nil
}

func (s *svc) List(ctx context.Context, params *domain.ServiceOrderFilterParams) (*domain.PaginatorResponse[domain.ServiceOrder], error) {
	var total int64
	var items []domain.ServiceOrder

	var eg errgroup.Group
	eg.Go(func() error {
		t, err := s.repo.Count(ctx, params)
		total = t
		return err
	})
	eg.Go(func() error {
		list, err := s.repo.Search(ctx, params)
		items = list
		return err
	})

	if err := eg.Wait(); err != nil {
		return nil, fmt.Errorf("fail to list service orders: %w", err)
	}

	totalPages := int64(math.Ceil(float64(total) / float64(params.PageSize)))

	return &domain.PaginatorResponse[domain.ServiceOrder]{
		Items:      items,
		TotalItems: total,
		TotalPages: totalPages,
		Page:       params.Page,
		PageSize:   params.PageSize,
	}, nil
}

func (s *svc) ListWorks(ctx context.Context, serviceOrderID string) ([]domain.Work, error) {
	serviceOrderID = strings.TrimSpace(serviceOrderID)
	if serviceOrderID == "" {
		s.logValidationError(ctx, "list_service_order_works", domain.ErrInvalidServiceOrderId)
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
		s.logValidationError(ctx, "add_works_to_service_order", domain.ErrInvalidServiceOrderId)
		return domain.ErrInvalidServiceOrderId
	}

	if len(workIDs) == 0 {
		s.logValidationError(ctx, "add_works_to_service_order", domain.ErrEmptyServicesList)
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

			if e = s.workSOHistoryRepo.Insert(ctx, serviceOrderID, w.ID, nil, domain.WORK_SERVICE_ORDER_STATUS_AWAITING_START); e != nil {
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
		s.logValidationError(ctx, "remove_work_from_service_order", domain.ErrInvalidServiceOrderId)
		return domain.ErrInvalidServiceOrderId
	}

	workID = strings.TrimSpace(workID)
	if workID == "" {
		s.logValidationError(ctx, "remove_work_from_service_order", domain.ErrInvalidWorkId)
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
		s.logValidationError(ctx, "list_service_order_supplies", domain.ErrInvalidServiceOrderId)
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
		s.logValidationError(ctx, "add_supplies_to_service_order", domain.ErrInvalidServiceOrderId)
		return domain.ErrInvalidServiceOrderId
	}

	if len(supplies) == 0 {
		s.logValidationError(ctx, "add_supplies_to_service_order", domain.ErrEmptyServicesList)
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
				s.logValidationError(ctx, "add_supplies_to_service_order", domain.ErrInvalidSupplyAmount)
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

		previousStatus := so.Status
		so.Status = domain.SERVICE_ORDER_STATUS_AWAITING_APPROVAL

		// TODO: SEND CUSTOMER NOTIFICATION (EMAIL / SMS / ETC)

		if err = s.repo.Save(ctx, &so); err != nil {
			return err
		}

		s.logStatusTransition(ctx, "send_service_order_to_customer_approval", so.ID, previousStatus, so.Status)
		return nil
	})

	if err != nil {
		return err
	}

	return s.reviewOSPricing(ctx, serviceOrderID)
}

func (s *svc) SendToDiagnosis(ctx context.Context, serviceOrderID string) error {
	serviceOrderID = strings.TrimSpace(serviceOrderID)
	if serviceOrderID == "" {
		return domain.ErrInvalidServiceOrderId
	}

	return s.uow.Execute(ctx, func(ctx context.Context) error {
		so, err := s.repo.FindByID(ctx, serviceOrderID)
		if err != nil {
			return err
		}

		if so.Status != domain.SERVICE_ORDER_STATUS_RECEIVED {
			return domain.ErrServiceOrderNotInReceived
		}

		so.Status = domain.SERVICE_ORDER_STATUS_IN_DIAGNOSIS

		return s.repo.Save(ctx, &so)
	})
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

		previousStatus := so.Status
		so.Status = domain.SERVICE_ORDER_STATUS_IN_PROGRESS

		if err := s.repo.Save(ctx, &so); err != nil {
			return err
		}

		s.logStatusTransition(ctx, "accept_service_order", so.ID, previousStatus, so.Status)
		// TODO: NOTIFY MECHANICAL TO START SERVICE (SMS / EMAIL / WPP)
		return nil
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

		previousStatus := so.Status
		so.Status = domain.SERVICE_ORDER_STATUS_REJECTED

		for _, sup := range supplies {
			if err := s.supplyRepo.RestoreStock(ctx, sup.ID, sup.StockQuantity); err != nil {
				return err
			}
		}

		if err := s.repo.Save(ctx, &so); err != nil {
			return err
		}

		s.logStatusTransition(ctx, "reject_service_order", so.ID, previousStatus, so.Status)
		return nil
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

		previousStatus := so.Status
		so.Status = domain.SERVICE_ORDER_STATUS_DELIVERED

		if err := s.repo.Save(ctx, &so); err != nil {
			return err
		}

		s.logStatusTransition(ctx, "deliver_service_order", so.ID, previousStatus, so.Status)
		return nil
	})
}

func (s *svc) AverageExecutionTime(ctx context.Context, workIDs []string) ([]domain.WorkExecutionTime, error) {
	return s.repo.AverageExecutionTimeInHours(ctx, workIDs)
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

		if !so.Status.IsCancelable() {
			return domain.ErrServiceOrderNotCancelable
		}

		previousStatus := so.Status
		so.Status = domain.SERVICE_ORDER_STATUS_CANCELLED

		if err := s.repo.Save(ctx, &so); err != nil {
			return err
		}

		s.logStatusTransition(ctx, "cancel_service_order", so.ID, previousStatus, so.Status)
		return nil
	})
}

func (s *svc) GetFullOSByID(ctx context.Context, serviceOrderID string) (domain.FullServiceOrder, error) {
	serviceOrderID = strings.TrimSpace(serviceOrderID)
	if serviceOrderID == "" {
		return domain.FullServiceOrder{}, domain.ErrInvalidServiceOrderId
	}

	var res domain.FullServiceOrder
	eg := errgroup.Group{}
	eg.Go(func() (err error) {
		res.ServiceOrder, err = s.repo.FindByID(ctx, serviceOrderID)
		return
	})
	eg.Go(func() (err error) {
		res.Works, err = s.repo.ListWorksByServiceOrderID(ctx, serviceOrderID)
		return
	})
	eg.Go(func() (err error) {
		res.Supplies, err = s.repo.ListSuppliesByServiceOrderID(ctx, serviceOrderID)
		return
	})

	if err := eg.Wait(); err != nil {
		return domain.FullServiceOrder{}, err
	}

	return res, nil
}

func (s *svc) NextWork(ctx context.Context, serviceOrderID, workID string) error {
	serviceOrderID = strings.TrimSpace(serviceOrderID)
	workID = strings.TrimSpace(workID)
	if serviceOrderID == "" {
		return domain.ErrInvalidServiceOrderId
	}
	if workID == "" {
		return domain.ErrInvalidWorkId
	}

	return s.uow.Execute(ctx, func(ctx context.Context) error {
		results, err := s.workSOHistoryRepo.Search(ctx, domain.SearchWorkSOHistoryParams{
			ServiceOrderID: serviceOrderID,
			WorkID:         workID,
		})
		if err != nil {
			return err
		}
		if len(results) == 0 {
			return domain.ErrWorkServiceOrderNotFound
		}

		latest := results[0]
		next, err := latest.NextStatus()
		if err != nil {
			return err
		}

		prev := latest.Status
		return s.workSOHistoryRepo.Insert(ctx, serviceOrderID, workID, &prev, next)
	})
}

func (s *svc) CancelWork(ctx context.Context, serviceOrderID, workID string) error {
	serviceOrderID = strings.TrimSpace(serviceOrderID)
	workID = strings.TrimSpace(workID)
	if serviceOrderID == "" {
		return domain.ErrInvalidServiceOrderId
	}
	if workID == "" {
		return domain.ErrInvalidWorkId
	}

	return s.uow.Execute(ctx, func(ctx context.Context) error {
		results, err := s.workSOHistoryRepo.Search(ctx, domain.SearchWorkSOHistoryParams{
			ServiceOrderID: serviceOrderID,
			WorkID:         workID,
		})
		if err != nil {
			return err
		}
		if len(results) == 0 {
			return domain.ErrWorkServiceOrderNotFound
		}

		latest := results[0]
		if latest.Status == domain.WORK_SERVICE_ORDER_STATUS_COMPLETED {
			return domain.ErrWorkServiceOrderAlreadyCompleted
		}
		if latest.Status == domain.WORK_SERVICE_ORDER_STATUS_CANCELLED {
			return domain.ErrWorkServiceOrderAlreadyCancelled
		}

		prev := latest.Status
		return s.workSOHistoryRepo.Insert(ctx, serviceOrderID, workID, &prev, domain.WORK_SERVICE_ORDER_STATUS_CANCELLED)
	})
}
