package service_order

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	pkgAuth "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/auth"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/json"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order/interfaces"
	"go.uber.org/zap"
)

const (
	soIDParam      = "id"
	workIDParam    = "workId"
	serviceIDParam = "serviceId"
	supplyIDParam  = "supplyId"
)

type soController struct {
	svc interfaces.ServiceOrderService
}

func NewController(svc interfaces.ServiceOrderService) *soController {
	return &soController{svc: svc}
}

func (ctrl *soController) Create(ctx context.Context, req *http.Request) (adapters.SOResponse, error) {
	var dto adapters.CreateSORequest
	if err := json.ParseJsonBodyToStruct(req.Body, &dto); err != nil {
		logger.Of(ctx).Debug("Failed to bind service order creation payload",
			zap.String("operation", "create_service_order"), zap.Error(err))
		return adapters.SOResponse{}, err
	}

	resp, err := ctrl.svc.Create(ctx, dto)
	if err != nil {
		logger.Of(ctx).Error(err)
		return adapters.SOResponse{}, err
	}
	return resp, nil
}

func (ctrl *soController) List(ctx context.Context, req *http.Request) (adapters.PaginatedSOResponse, error) {
	q := req.URL.Query()
	var page, pageSize int64 = 1, 10

	if p := q.Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = int64(v)
		}
	}
	if ps := q.Get("pageSize"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 {
			pageSize = int64(v)
		}
	}

	sortBy := q.Get("sort_by")
	sortOrder := q.Get("sort_order")

	params := adapters.SOFilterParams{
		Page:       page,
		PageSize:   pageSize,
		Status:     q.Get("status"),
		CustomerID: q.Get("customerId"),
		VehicleID:  q.Get("vehicleId"),
		SortBy:     sortBy,
		SortOrder:  sortOrder,
	}

	return ctrl.svc.List(ctx, params)
}

func (ctrl *soController) GetWorks(ctx context.Context, req *http.Request) (adapters.ListResponse[adapters.WorkItemResponse], error) {
	id, _ := ctx.Value(soIDParam).(string)
	if strings.TrimSpace(id) == "" {
		return adapters.ListResponse[adapters.WorkItemResponse]{}, domain.ErrInvalidServiceOrderId
	}

	resp, err := ctrl.svc.GetWorks(ctx, id)
	if err != nil {
		logger.Of(ctx).Debug("list service order works failed",
			zap.String("operation", "list_service_order_works"), zap.Error(err))
		return adapters.ListResponse[adapters.WorkItemResponse]{}, err
	}
	return resp, nil
}

func (ctrl *soController) AddWork(ctx context.Context, req *http.Request) error {
	id, _ := ctx.Value(soIDParam).(string)
	if strings.TrimSpace(id) == "" {
		return domain.ErrInvalidServiceOrderId
	}

	var dto adapters.AddWorksRequest
	if err := json.ParseJsonBodyToStruct(req.Body, &dto); err != nil {
		logger.Of(ctx).Debug("Failed to bind add works payload",
			zap.String("operation", "add_works_to_service_order"), zap.Error(err))
		return err
	}

	if err := ctrl.svc.AddWorks(ctx, id, dto.WorkIDs); err != nil {
		logger.Of(ctx).Debug("add works to service order failed",
			zap.String("operation", "add_works_to_service_order"), zap.Error(err))
		return err
	}
	return nil
}

func (ctrl *soController) DeleteWork(ctx context.Context, req *http.Request) error {
	soID, _ := ctx.Value(soIDParam).(string)
	workID, _ := ctx.Value(serviceIDParam).(string)

	if strings.TrimSpace(soID) == "" {
		return domain.ErrInvalidServiceOrderId
	}
	if strings.TrimSpace(workID) == "" {
		return domain.ErrInvalidWorkId
	}

	if err := ctrl.svc.RemoveWork(ctx, soID, workID); err != nil {
		logger.Of(ctx).Debug("remove work from service order failed",
			zap.String("operation", "remove_work_from_service_order"), zap.Error(err))
		return err
	}
	return nil
}

func (ctrl *soController) GetSupplies(ctx context.Context, req *http.Request) (adapters.ListResponse[adapters.SupplyItemResponse], error) {
	id, _ := ctx.Value(soIDParam).(string)
	if strings.TrimSpace(id) == "" {
		return adapters.ListResponse[adapters.SupplyItemResponse]{}, domain.ErrInvalidServiceOrderId
	}

	resp, err := ctrl.svc.GetSupplies(ctx, id)
	if err != nil {
		logger.Of(ctx).Debug("list service order supplies failed",
			zap.String("operation", "list_service_order_supplies"), zap.Error(err))
		return adapters.ListResponse[adapters.SupplyItemResponse]{}, err
	}
	return resp, nil
}

func (ctrl *soController) AddSupplies(ctx context.Context, req *http.Request) error {
	id, _ := ctx.Value(soIDParam).(string)
	if strings.TrimSpace(id) == "" {
		return domain.ErrInvalidServiceOrderId
	}

	var dto adapters.AddSuppliesRequest
	if err := json.ParseJsonBodyToStruct(req.Body, &dto); err != nil {
		logger.Of(ctx).Debug("Failed to bind add supplies payload",
			zap.String("operation", "add_supplies_to_service_order"), zap.Error(err))
		return err
	}

	if err := ctrl.svc.AddSupplies(ctx, id, dto.Supplies); err != nil {
		logger.Of(ctx).Debug("add supplies to service order failed",
			zap.String("operation", "add_supplies_to_service_order"), zap.Error(err))
		return err
	}
	return nil
}

func (ctrl *soController) DeleteSupply(ctx context.Context, req *http.Request) error {
	soID, _ := ctx.Value(soIDParam).(string)
	supplyID, _ := ctx.Value(supplyIDParam).(string)

	if strings.TrimSpace(soID) == "" {
		return domain.ErrInvalidServiceOrderId
	}
	if strings.TrimSpace(supplyID) == "" {
		return domain.ErrInvalidSupplyID
	}

	if err := ctrl.svc.RemoveSupply(ctx, soID, supplyID); err != nil {
		logger.Of(ctx).Debug("remove supply from service order failed",
			zap.String("operation", "remove_supply_from_service_order"), zap.Error(err))
		return err
	}
	return nil
}

func (ctrl *soController) Receive(ctx context.Context, req *http.Request) error {
	id, _ := ctx.Value(soIDParam).(string)
	if err := ctrl.svc.Receive(ctx, id); err != nil {
		logger.Of(ctx).Debug("receive service order failed",
			zap.String("operation", "receive_service_order"), zap.Error(err))
		return err
	}
	return nil
}

func (ctrl *soController) SendToCustomerApproval(ctx context.Context, req *http.Request) error {
	id, _ := ctx.Value(soIDParam).(string)
	if strings.TrimSpace(id) == "" {
		return domain.ErrInvalidServiceOrderId
	}
	if err := ctrl.svc.SendToCustomerApproval(ctx, id); err != nil {
		logger.Of(ctx).Debug("send service order to customer approval failed",
			zap.String("operation", "send_service_order_to_customer_approval"), zap.Error(err))
		return err
	}
	return nil
}

func (ctrl *soController) SendToDiagnosis(ctx context.Context, req *http.Request) error {
	id, _ := ctx.Value(soIDParam).(string)
	if strings.TrimSpace(id) == "" {
		return domain.ErrInvalidServiceOrderId
	}
	if err := ctrl.svc.SendToDiagnosis(ctx, id); err != nil {
		logger.Of(ctx).Debug("send service order to diagnosis failed",
			zap.String("operation", "send_service_order_to_diagnosis"), zap.Error(err))
		return err
	}
	return nil
}

func (ctrl *soController) Finish(ctx context.Context, req *http.Request) error {
	id, _ := ctx.Value(soIDParam).(string)
	if strings.TrimSpace(id) == "" {
		return domain.ErrInvalidServiceOrderId
	}
	if err := ctrl.svc.Finish(ctx, id); err != nil {
		logger.Of(ctx).Debug("finish service order failed",
			zap.String("operation", "finish_service_order"), zap.Error(err))
		return err
	}
	return nil
}

func (ctrl *soController) Accept(ctx context.Context, req *http.Request) error {
	id, _ := ctx.Value(soIDParam).(string)
	if strings.TrimSpace(id) == "" {
		return domain.ErrInvalidServiceOrderId
	}

	claims, err := pkgAuth.GetClaimsFromHeader(req)
	if err != nil {
		return err
	}

	if err = ctrl.svc.Accept(ctx, id, claims.UserId); err != nil {
		logger.Of(ctx).Debug("accept service order failed",
			zap.String("operation", "accept_service_order"), zap.Error(err))
		return err
	}
	return nil
}

func (ctrl *soController) Reject(ctx context.Context, req *http.Request) error {
	id, _ := ctx.Value(soIDParam).(string)
	if strings.TrimSpace(id) == "" {
		return domain.ErrInvalidServiceOrderId
	}

	claims, err := pkgAuth.GetClaimsFromHeader(req)
	if err != nil {
		return err
	}

	if err = ctrl.svc.Reject(ctx, id, claims.UserId); err != nil {
		logger.Of(ctx).Debug("reject service order failed",
			zap.String("operation", "reject_service_order"), zap.Error(err))
		return err
	}
	return nil
}

func (ctrl *soController) Deliver(ctx context.Context, req *http.Request) error {
	id, _ := ctx.Value(soIDParam).(string)
	if strings.TrimSpace(id) == "" {
		return domain.ErrInvalidServiceOrderId
	}
	if err := ctrl.svc.Deliver(ctx, id); err != nil {
		logger.Of(ctx).Debug("deliver service order failed",
			zap.String("operation", "deliver_service_order"), zap.Error(err))
		return err
	}
	return nil
}

func (ctrl *soController) Cancel(ctx context.Context, req *http.Request) error {
	id, _ := ctx.Value(soIDParam).(string)
	if strings.TrimSpace(id) == "" {
		return domain.ErrInvalidServiceOrderId
	}
	if err := ctrl.svc.Cancel(ctx, id); err != nil {
		logger.Of(ctx).Debug("cancel service order failed",
			zap.String("operation", "cancel_service_order"), zap.Error(err))
		return err
	}
	return nil
}

func (ctrl *soController) GetFullByID(ctx context.Context, req *http.Request) (adapters.SODetailResponse, error) {
	id, _ := ctx.Value(soIDParam).(string)
	if strings.TrimSpace(id) == "" {
		return adapters.SODetailResponse{}, domain.ErrInvalidServiceOrderId
	}

	resp, err := ctrl.svc.GetFullByID(ctx, id)
	if err != nil {
		logger.Of(ctx).Debug("get full service order by id failed",
			zap.String("operation", "get_full_service_order_by_id"), zap.Error(err))
		return adapters.SODetailResponse{}, err
	}
	return resp, nil
}

func (ctrl *soController) GetAverageExecutionTime(ctx context.Context, req *http.Request) ([]adapters.WorkExecutionTimeResponse, error) {
	workIDs := req.URL.Query()["work_id"]
	if workIDs == nil {
		workIDs = []string{}
	}

	resp, err := ctrl.svc.GetAverageExecutionTime(ctx, workIDs)
	if err != nil {
		logger.Of(ctx).Debug("get average execution time failed",
			zap.String("operation", "get_average_execution_time"), zap.Error(err))
		return nil, err
	}
	return resp, nil
}

func (ctrl *soController) NextWork(ctx context.Context, req *http.Request) error {
	soID, _ := ctx.Value(soIDParam).(string)
	workID, _ := ctx.Value(workIDParam).(string)

	if err := ctrl.svc.NextWork(ctx, soID, workID); err != nil {
		logger.Of(ctx).Debug("advance work status failed",
			zap.String("operation", "next_work_status"), zap.Error(err))
		return err
	}
	return nil
}

func (ctrl *soController) CancelWork(ctx context.Context, req *http.Request) error {
	soID, _ := ctx.Value(soIDParam).(string)
	workID, _ := ctx.Value(workIDParam).(string)

	if err := ctrl.svc.CancelWork(ctx, soID, workID); err != nil {
		logger.Of(ctx).Debug("cancel work failed",
			zap.String("operation", "cancel_work_status"), zap.Error(err))
		return err
	}
	return nil
}

func (ctrl *soController) GetStatus(ctx context.Context, req *http.Request) (adapters.SOResponse, error) {
	id, _ := ctx.Value(soIDParam).(string)

	resp, err := ctrl.svc.GetStatus(ctx, id)
	if err != nil {
		logger.Of(ctx).Debug("get service order status failed",
			zap.String("operation", "get_service_order_status"), zap.Error(err))
		return adapters.SOResponse{}, err
	}
	return resp, nil
}
