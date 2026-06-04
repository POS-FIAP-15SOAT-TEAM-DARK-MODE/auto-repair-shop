package service_order

import (
	"context"
	"fmt"
	"math"
	"strings"

	customerDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/customer/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/pkg/uow"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/service_order/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/service_order/domain"
	supplyInterfaces "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/supply/interfaces"
	vehicleInterfaces "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/vehicle/interfaces"
	workInterfaces "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/work/interfaces"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// CustomerFinder is the minimal customer interface the SO service needs.
type CustomerFinder interface {
	GetByID(ctx context.Context, id string) (customerDomain.Customer, error)
	GetByUserID(ctx context.Context, userID string) (customerDomain.Customer, error)
}

type soService struct {
	uow            uow.Executor
	repo           domain.ServiceOrderRepository
	workRepo       workInterfaces.WorkRepository
	supplyService  supplyInterfaces.SupplyService
	wsoHistoryRepo domain.WorkSOHistoryRepository
	customerFinder CustomerFinder
	vehicleService vehicleInterfaces.VehicleService
}

func NewService(
	u uow.Executor,
	repo domain.ServiceOrderRepository,
	workRepo workInterfaces.WorkRepository,
	supplyService supplyInterfaces.SupplyService,
	wsoHistoryRepo domain.WorkSOHistoryRepository,
	customerFinder CustomerFinder,
	vehicleService vehicleInterfaces.VehicleService,
) *soService {
	return &soService{u, repo, workRepo, supplyService, wsoHistoryRepo, customerFinder, vehicleService}
}

func (s *soService) logValidationError(ctx context.Context, operation string, err error) {
	logger.Of(ctx).Warn("service_order.validation_failed",
		zap.String("operation", operation),
		zap.String("entity", "service_order"),
		zap.Error(err),
	)
}

func (s *soService) logStatusTransition(ctx context.Context, operation, soID string, prev, next domain.SERVICE_ORDER_STATUS) {
	logger.Of(ctx).Info("service_order.status_transition",
		zap.String("operation", operation),
		zap.String("entity", "service_order"),
		zap.String("service_order_id", soID),
		zap.String("previous_status", prev.String()),
		zap.String("new_status", next.String()),
	)
}

func (s *soService) logWorkStatusTransition(ctx context.Context, operation, soID, workID string, prev, next domain.WORK_SERVICE_ORDER_STATUS) {
	logger.Of(ctx).Info("work_service_order.status_transition",
		zap.String("operation", operation),
		zap.String("entity", "work_service_order"),
		zap.String("service_order_id", soID),
		zap.String("work_id", workID),
		zap.String("previous_status", prev.String()),
		zap.String("new_status", next.String()),
	)
}

func (s *soService) Create(ctx context.Context, req adapters.CreateSORequest) (adapters.SOResponse, error) {
	customer, err := s.customerFinder.GetByID(ctx, req.CustomerID)
	if err != nil {
		return adapters.SOResponse{}, err
	}

	vehicle, err := s.vehicleService.FindById(ctx, req.VehicleID)
	if err != nil {
		return adapters.SOResponse{}, err
	}

	so := domain.NewServiceOrder(&customer, &vehicle)
	if err = s.uow.Execute(ctx, func(ctx context.Context) error {
		return s.repo.Save(ctx, so)
	}); err != nil {
		return adapters.SOResponse{}, err
	}

	logger.Of(ctx).Info("service_order.created",
		zap.String("operation", "create_service_order"),
		zap.String("entity", "service_order"),
		zap.String("service_order_id", so.ID),
		zap.String("status", so.Status.String()),
	)

	return adapters.SOResponse{ID: so.ID, Status: so.Status.String()}, nil
}

func (s *soService) List(ctx context.Context, params adapters.SOFilterParams) (adapters.PaginatedSOResponse, error) {
	domainParams := &domain.ServiceOrderFilterParams{
		Page:       params.Page,
		PageSize:   params.PageSize,
		Limit:      params.PageSize,
		Offset:     (params.Page - 1) * params.PageSize,
		Status:     params.Status,
		CustomerID: params.CustomerID,
		VehicleID:  params.VehicleID,
	}

	var total int64
	var items []domain.ServiceOrder

	eg := errgroup.Group{}
	eg.Go(func() error {
		t, err := s.repo.Count(ctx, domainParams)
		total = t
		return err
	})
	eg.Go(func() error {
		list, err := s.repo.Search(ctx, domainParams)
		items = list
		return err
	})

	if err := eg.Wait(); err != nil {
		return adapters.PaginatedSOResponse{}, fmt.Errorf("fail to list service orders: %w", err)
	}

	totalPages := int64(math.Ceil(float64(total) / float64(params.PageSize)))

	listItems := make([]adapters.SOListItemResponse, 0, len(items))
	for _, so := range items {
		listItems = append(listItems, mapSOToListItem(so))
	}

	return adapters.PaginatedSOResponse{
		Items:      listItems,
		TotalItems: total,
		TotalPages: totalPages,
		Page:       params.Page,
		PageSize:   params.PageSize,
	}, nil
}

func (s *soService) GetWorks(ctx context.Context, soID string) (adapters.ListResponse[adapters.WorkItemResponse], error) {
	soID = strings.TrimSpace(soID)
	if soID == "" {
		s.logValidationError(ctx, "list_service_order_works", domain.ErrInvalidServiceOrderId)
		return adapters.ListResponse[adapters.WorkItemResponse]{}, domain.ErrInvalidServiceOrderId
	}

	var works []adapters.WorkItemResponse
	err := s.uow.Execute(ctx, func(ctx context.Context) error {
		ok, _, err := s.repo.ExistsByID(ctx, soID)
		if err != nil {
			return err
		}
		if !ok {
			return domain.ErrServiceOrderNotFound
		}
		domainWorks, err := s.repo.ListWorksByServiceOrderID(ctx, soID)
		if err != nil {
			return err
		}
		works = mapWorksToResponse(domainWorks)
		return nil
	})
	if err != nil {
		return adapters.ListResponse[adapters.WorkItemResponse]{}, err
	}

	n := int64(len(works))
	return adapters.ListResponse[adapters.WorkItemResponse]{
		Items:      works,
		TotalItems: n,
		TotalPages: 1,
		PageSize:   n,
		Page:       1,
	}, nil
}

func (s *soService) AddWorks(ctx context.Context, soID string, workIDs []string) error {
	soID = strings.TrimSpace(soID)
	if soID == "" {
		s.logValidationError(ctx, "add_works_to_service_order", domain.ErrInvalidServiceOrderId)
		return domain.ErrInvalidServiceOrderId
	}

	if len(workIDs) == 0 {
		s.logValidationError(ctx, "add_works_to_service_order", domain.ErrEmptyServicesList)
		return domain.ErrEmptyServicesList
	}

	err := s.uow.Execute(ctx, func(ctx context.Context) error {
		ok, status, err := s.repo.ExistsByID(ctx, soID)
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

			if e = s.repo.AddWorkLink(ctx, soID, w.ID, w.Price); e != nil {
				return e
			}

			if e = s.wsoHistoryRepo.Insert(ctx, soID, w.ID, nil, domain.WORK_SERVICE_ORDER_STATUS_AWAITING_START); e != nil {
				return e
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	go func() {
		if err := s.reviewOSPricing(context.Background(), soID); err != nil {
			logger.Global().Error(err, zap.String("serviceOrderID", soID))
		}
	}()

	return nil
}

func (s *soService) RemoveWork(ctx context.Context, soID, workID string) error {
	soID = strings.TrimSpace(soID)
	if soID == "" {
		s.logValidationError(ctx, "remove_work_from_service_order", domain.ErrInvalidServiceOrderId)
		return domain.ErrInvalidServiceOrderId
	}

	workID = strings.TrimSpace(workID)
	if workID == "" {
		s.logValidationError(ctx, "remove_work_from_service_order", domain.ErrInvalidWorkId)
		return domain.ErrInvalidWorkId
	}

	err := s.uow.Execute(ctx, func(ctx context.Context) error {
		ok, status, err := s.repo.ExistsByID(ctx, soID)
		if err != nil {
			return err
		}
		if !ok {
			return domain.ErrServiceOrderNotFound
		}
		if status != domain.SERVICE_ORDER_STATUS_NEW {
			return domain.ErrServiceOrderNotNew
		}
		return s.repo.RemoveWorkLink(ctx, soID, workID)
	})
	if err != nil {
		return err
	}

	go func() {
		if err := s.reviewOSPricing(context.Background(), soID); err != nil {
			logger.Global().Error(err, zap.String("serviceOrderID", soID))
		}
	}()

	return nil
}

func (s *soService) GetSupplies(ctx context.Context, soID string) (adapters.ListResponse[adapters.SupplyItemResponse], error) {
	soID = strings.TrimSpace(soID)
	if soID == "" {
		s.logValidationError(ctx, "list_service_order_supplies", domain.ErrInvalidServiceOrderId)
		return adapters.ListResponse[adapters.SupplyItemResponse]{}, domain.ErrInvalidServiceOrderId
	}

	ok, _, err := s.repo.ExistsByID(ctx, soID)
	if err != nil {
		return adapters.ListResponse[adapters.SupplyItemResponse]{}, err
	}
	if !ok {
		return adapters.ListResponse[adapters.SupplyItemResponse]{}, domain.ErrServiceOrderNotFound
	}

	domainSupplies, err := s.repo.ListSuppliesByServiceOrderID(ctx, soID)
	if err != nil {
		return adapters.ListResponse[adapters.SupplyItemResponse]{}, err
	}

	supplies := mapSuppliesToResponse(domainSupplies)
	n := int64(len(supplies))
	return adapters.ListResponse[adapters.SupplyItemResponse]{
		Items:      supplies,
		TotalItems: n,
		TotalPages: 1,
		PageSize:   n,
		Page:       1,
	}, nil
}

func (s *soService) AddSupplies(ctx context.Context, soID string, supplies []adapters.AddSupplyItem) error {
	soID = strings.TrimSpace(soID)
	if soID == "" {
		s.logValidationError(ctx, "add_supplies_to_service_order", domain.ErrInvalidServiceOrderId)
		return domain.ErrInvalidServiceOrderId
	}

	if len(supplies) == 0 {
		s.logValidationError(ctx, "add_supplies_to_service_order", domain.ErrEmptyServicesList)
		return domain.ErrEmptyServicesList
	}

	err := s.uow.Execute(ctx, func(ctx context.Context) error {
		ok, status, err := s.repo.ExistsByID(ctx, soID)
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
			supplyID := strings.TrimSpace(sup.ID)
			if supplyID == "" {
				return domain.ErrInvalidSupplyID
			}
			if sup.Amount <= 0 {
				s.logValidationError(ctx, "add_supplies_to_service_order", domain.ErrInvalidSupplyAmount)
				return domain.ErrInvalidSupplyAmount
			}

			supply, e := s.supplyService.FindById(ctx, supplyID)
			if e != nil {
				return e
			}
			if supply.StockQuantity < sup.Amount {
				return domain.ErrSupplyOutOfStock
			}

			if err = s.repo.AddSupplyLink(ctx, soID, sup.ID, sup.Amount, supply.UnitPrice); err != nil {
				return err
			}
			if err = s.supplyService.DecrementStockQuantity(ctx, supply.ID, sup.Amount); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	go func() {
		if err := s.reviewOSPricing(context.Background(), soID); err != nil {
			logger.Global().Error(err, zap.String("serviceOrderID", soID))
		}
	}()

	return nil
}

func (s *soService) RemoveSupply(ctx context.Context, soID, supplyID string) error {
	soID = strings.TrimSpace(soID)
	if soID == "" {
		return domain.ErrInvalidServiceOrderId
	}
	supplyID = strings.TrimSpace(supplyID)
	if supplyID == "" {
		return domain.ErrInvalidSupplyID
	}

	err := s.uow.Execute(ctx, func(ctx context.Context) error {
		ok, status, err := s.repo.ExistsByID(ctx, soID)
		if err != nil {
			return err
		}
		if !ok {
			return domain.ErrServiceOrderNotFound
		}
		if status != domain.SERVICE_ORDER_STATUS_IN_DIAGNOSIS {
			return domain.ErrServiceOrderNotInDiagnosis
		}

		qty, err := s.repo.RemoveSupplyLink(ctx, soID, supplyID)
		if err != nil {
			return err
		}
		return s.supplyService.IncrementStockQuantity(ctx, supplyID, qty)
	})
	if err != nil {
		return err
	}

	go func() {
		if err := s.reviewOSPricing(context.Background(), soID); err != nil {
			logger.Global().Error(err, zap.String("serviceOrderID", soID))
		}
	}()

	return nil
}

func (s *soService) Receive(ctx context.Context, soID string) error {
	soID = strings.TrimSpace(soID)
	if soID == "" {
		return domain.ErrInvalidServiceOrderId
	}

	return s.uow.Execute(ctx, func(ctx context.Context) error {
		so, err := s.repo.FindByID(ctx, soID)
		if err != nil {
			return err
		}
		if so.Status != domain.SERVICE_ORDER_STATUS_NEW {
			return domain.ErrServiceOrderNotNew
		}
		prev := so.Status
		so.Status = domain.SERVICE_ORDER_STATUS_RECEIVED
		s.logStatusTransition(ctx, "receive_service_order", so.ID, prev, so.Status)
		return s.repo.Save(ctx, &so)
	})
}

func (s *soService) SendToDiagnosis(ctx context.Context, soID string) error {
	soID = strings.TrimSpace(soID)
	if soID == "" {
		return domain.ErrInvalidServiceOrderId
	}

	return s.uow.Execute(ctx, func(ctx context.Context) error {
		so, err := s.repo.FindByID(ctx, soID)
		if err != nil {
			return err
		}
		if so.Status != domain.SERVICE_ORDER_STATUS_RECEIVED {
			return domain.ErrServiceOrderNotInReceived
		}
		prev := so.Status
		so.Status = domain.SERVICE_ORDER_STATUS_IN_DIAGNOSIS
		s.logStatusTransition(ctx, "send_service_order_to_diagnosis", so.ID, prev, so.Status)
		return s.repo.Save(ctx, &so)
	})
}

func (s *soService) SendToCustomerApproval(ctx context.Context, soID string) error {
	soID = strings.TrimSpace(soID)
	if soID == "" {
		return domain.ErrInvalidServiceOrderId
	}

	err := s.uow.Execute(ctx, func(ctx context.Context) error {
		so, err := s.repo.FindByID(ctx, soID)
		if err != nil {
			return err
		}
		if so.Status != domain.SERVICE_ORDER_STATUS_IN_DIAGNOSIS {
			return domain.ErrServiceOrderNotInDiagnosis
		}
		prev := so.Status
		so.Status = domain.SERVICE_ORDER_STATUS_AWAITING_APPROVAL
		if err = s.repo.Save(ctx, &so); err != nil {
			return err
		}
		s.logStatusTransition(ctx, "send_service_order_to_customer_approval", so.ID, prev, so.Status)
		return nil
	})
	if err != nil {
		return err
	}

	return s.reviewOSPricing(ctx, soID)
}

func (s *soService) Accept(ctx context.Context, soID, userID string) error {
	soID = strings.TrimSpace(soID)
	if soID == "" {
		return domain.ErrInvalidServiceOrderId
	}

	return s.uow.Execute(ctx, func(ctx context.Context) error {
		eg := errgroup.Group{}
		var so domain.ServiceOrder
		eg.Go(func() (err error) {
			so, err = s.repo.FindByID(ctx, soID)
			if err != nil {
				return
			}
			if so.Customer == nil {
				return domain.ErrServiceOrderNotFound
			}
			return
		})

		var cust customerDomain.Customer
		eg.Go(func() (err error) {
			cust, err = s.customerFinder.GetByUserID(ctx, userID)
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

		prev := so.Status
		so.Status = domain.SERVICE_ORDER_STATUS_IN_PROGRESS

		if err := s.repo.Save(ctx, &so); err != nil {
			return err
		}
		s.logStatusTransition(ctx, "accept_service_order", so.ID, prev, so.Status)
		return nil
	})
}

func (s *soService) Reject(ctx context.Context, soID, userID string) error {
	soID = strings.TrimSpace(soID)
	if soID == "" {
		return domain.ErrInvalidServiceOrderId
	}

	return s.uow.Execute(ctx, func(ctx context.Context) error {
		eg := errgroup.Group{}
		var so domain.ServiceOrder
		eg.Go(func() (err error) {
			so, err = s.repo.FindByID(ctx, soID)
			if err != nil {
				return
			}
			if so.Customer == nil {
				return domain.ErrServiceOrderNotFound
			}
			return
		})

		var cust customerDomain.Customer
		eg.Go(func() (err error) {
			cust, err = s.customerFinder.GetByUserID(ctx, userID)
			return
		})

		var domainSupplies []adapters.SupplyItemResponse
		eg.Go(func() error {
			supplies, err := s.repo.ListSuppliesByServiceOrderID(ctx, soID)
			if err != nil {
				return err
			}
			domainSupplies = mapSuppliesToResponse(supplies)
			return nil
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

		for _, sup := range domainSupplies {
			if err := s.supplyService.IncrementStockQuantity(ctx, sup.ID, sup.Amount); err != nil {
				return err
			}
		}

		prev := so.Status
		so.Status = domain.SERVICE_ORDER_STATUS_REJECTED

		if err := s.repo.Save(ctx, &so); err != nil {
			return err
		}
		s.logStatusTransition(ctx, "reject_service_order", so.ID, prev, so.Status)
		return nil
	})
}

func (s *soService) Finish(ctx context.Context, soID string) error {
	soID = strings.TrimSpace(soID)
	if soID == "" {
		return domain.ErrInvalidServiceOrderId
	}

	return s.uow.Execute(ctx, func(ctx context.Context) error {
		so, err := s.repo.FindByID(ctx, soID)
		if err != nil {
			return err
		}
		if so.Status != domain.SERVICE_ORDER_STATUS_IN_PROGRESS {
			return domain.ErrServiceOrderNotInProgress
		}
		prev := so.Status
		so.Status = domain.SERVICE_ORDER_STATUS_COMPLETED
		s.logStatusTransition(ctx, "finish_service_order", so.ID, prev, so.Status)
		return s.repo.Save(ctx, &so)
	})
}

func (s *soService) Deliver(ctx context.Context, soID string) error {
	soID = strings.TrimSpace(soID)
	if soID == "" {
		return domain.ErrInvalidServiceOrderId
	}

	return s.uow.Execute(ctx, func(ctx context.Context) error {
		so, err := s.repo.FindByID(ctx, soID)
		if err != nil {
			return err
		}
		if so.Status != domain.SERVICE_ORDER_STATUS_COMPLETED {
			return domain.ErrServiceOrderNotCompleted
		}
		prev := so.Status
		so.Status = domain.SERVICE_ORDER_STATUS_DELIVERED
		if err := s.repo.Save(ctx, &so); err != nil {
			return err
		}
		s.logStatusTransition(ctx, "deliver_service_order", so.ID, prev, so.Status)
		return nil
	})
}

func (s *soService) Cancel(ctx context.Context, soID string) error {
	soID = strings.TrimSpace(soID)
	if soID == "" {
		return domain.ErrInvalidServiceOrderId
	}

	return s.uow.Execute(ctx, func(ctx context.Context) error {
		so, err := s.repo.FindByID(ctx, soID)
		if err != nil {
			return err
		}
		if so.Customer == nil {
			return domain.ErrServiceOrderNotFound
		}
		if !so.Status.IsCancelable() {
			return domain.ErrServiceOrderNotCancelable
		}
		prev := so.Status
		so.Status = domain.SERVICE_ORDER_STATUS_CANCELLED
		if err := s.repo.Save(ctx, &so); err != nil {
			return err
		}
		s.logStatusTransition(ctx, "cancel_service_order", so.ID, prev, so.Status)
		return nil
	})
}

func (s *soService) GetFullByID(ctx context.Context, soID string) (adapters.SODetailResponse, error) {
	soID = strings.TrimSpace(soID)
	if soID == "" {
		return adapters.SODetailResponse{}, domain.ErrInvalidServiceOrderId
	}

	eg := errgroup.Group{}
	var so domain.ServiceOrder
	var works []adapters.WorkItemResponse
	var supplies []adapters.SupplyItemResponse

	eg.Go(func() (err error) {
		so, err = s.repo.FindByID(ctx, soID)
		return
	})
	eg.Go(func() error {
		domainWorks, err := s.repo.ListWorksByServiceOrderID(ctx, soID)
		if err != nil {
			return err
		}
		works = mapWorksToResponse(domainWorks)
		return nil
	})
	eg.Go(func() error {
		domainSupplies, err := s.repo.ListSuppliesByServiceOrderID(ctx, soID)
		if err != nil {
			return err
		}
		supplies = mapSuppliesToResponse(domainSupplies)
		return nil
	})

	if err := eg.Wait(); err != nil {
		return adapters.SODetailResponse{}, err
	}

	return mapSOToDetailResponse(so, works, supplies), nil
}

func (s *soService) GetAverageExecutionTime(ctx context.Context, workIDs []string) ([]adapters.WorkExecutionTimeResponse, error) {
	result, err := s.repo.AverageExecutionTimeInHours(ctx, workIDs)
	if err != nil {
		return nil, err
	}

	items := make([]adapters.WorkExecutionTimeResponse, 0, len(result))
	for _, w := range result {
		items = append(items, adapters.WorkExecutionTimeResponse{
			WorkID:               w.WorkID,
			WorkName:             w.WorkName,
			AverageExecutionTime: w.AverageHours,
		})
	}
	return items, nil
}

func (s *soService) NextWork(ctx context.Context, soID, workID string) error {
	soID = strings.TrimSpace(soID)
	workID = strings.TrimSpace(workID)
	if soID == "" {
		return domain.ErrInvalidServiceOrderId
	}
	if workID == "" {
		return domain.ErrInvalidWorkId
	}

	return s.uow.Execute(ctx, func(ctx context.Context) error {
		results, err := s.wsoHistoryRepo.Search(ctx, domain.SearchWorkSOHistoryParams{
			ServiceOrderID: soID,
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
		s.logWorkStatusTransition(ctx, "next_work_service_order", soID, workID, prev, next)
		return s.wsoHistoryRepo.Insert(ctx, soID, workID, &prev, next)
	})
}

func (s *soService) CancelWork(ctx context.Context, soID, workID string) error {
	soID = strings.TrimSpace(soID)
	workID = strings.TrimSpace(workID)
	if soID == "" {
		return domain.ErrInvalidServiceOrderId
	}
	if workID == "" {
		return domain.ErrInvalidWorkId
	}

	return s.uow.Execute(ctx, func(ctx context.Context) error {
		results, err := s.wsoHistoryRepo.Search(ctx, domain.SearchWorkSOHistoryParams{
			ServiceOrderID: soID,
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
		s.logWorkStatusTransition(ctx, "cancel_work_service_order", soID, workID, prev, domain.WORK_SERVICE_ORDER_STATUS_CANCELLED)
		return s.wsoHistoryRepo.Insert(ctx, soID, workID, &prev, domain.WORK_SERVICE_ORDER_STATUS_CANCELLED)
	})
}

func (s *soService) GetStatus(ctx context.Context, soID string) (adapters.SOResponse, error) {
	soID = strings.TrimSpace(soID)
	if soID == "" {
		return adapters.SOResponse{}, domain.ErrInvalidServiceOrderId
	}

	var status domain.SERVICE_ORDER_STATUS
	err := s.uow.Execute(ctx, func(ctx context.Context) error {
		ok, st, err := s.repo.ExistsByID(ctx, soID)
		if err != nil {
			return err
		}
		if !ok {
			return domain.ErrServiceOrderNotFound
		}
		status = st
		return nil
	})
	if err != nil {
		logger.Of(ctx).Error(fmt.Errorf("error check service order status: %w", err))
		return adapters.SOResponse{}, err
	}

	return adapters.SOResponse{ID: soID, Status: status.String()}, nil
}

func (s *soService) reviewOSPricing(ctx context.Context, soID string) error {
	soID = strings.TrimSpace(soID)
	if soID == "" {
		return domain.ErrInvalidServiceOrderId
	}

	so, err := s.repo.FindByID(ctx, soID)
	if err != nil {
		return err
	}

	so.ResetPricing()

	eg := errgroup.Group{}
	eg.Go(func() error {
		works, e := s.repo.ListWorksByServiceOrderID(ctx, soID)
		if e != nil {
			return e
		}
		for _, work := range works {
			so.SumWorkValue(work)
		}
		return nil
	})
	eg.Go(func() error {
		supplies, e := s.repo.ListSuppliesByServiceOrderID(ctx, soID)
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
