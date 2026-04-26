package service_order

import (
	"net/http"
	"strings"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/auth"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/json"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/web"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	serviceOrderIDParam = "id"
	workIDPathParam     = "serviceId"
	supplyIDPathParam   = "supplyId"
)

type handler struct {
	svc domain.ServiceOrderService
}

func HttpHandler(svc domain.ServiceOrderService) *handler {
	return &handler{svc: svc}
}

func (h *handler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	dto, err := mapBodyToRequestDTO(c)
	if err != nil {
		err = json.CheckJsonError(err)
		status, response := web.Error(err)
		logger.Of(ctx).Debug("Failed to bind work creation payload",
			zap.String("operation", "create_work"),
			zap.Error(err),
			zap.String("entity", "work"),
		)
		c.JSON(status, response)
		return
	}

	logger.Of(ctx).Debug("create request",
		zap.String("operation", "create_service_order"),
		zap.String("entity", "service_order"),
		zap.String("client_id", dto.ClientID),
		zap.String("vehicle_id", dto.VehicleID),
	)
	res, err := h.svc.Create(ctx, dto.ClientID, dto.VehicleID)
	if err != nil {
		status, response := web.Error(err)
		logger.Of(ctx).Error(err)
		logger.Of(ctx).Debug("Service Order creation failed in service layer",
			zap.String("operation", "create_service_order"),
			zap.Error(err),
			zap.String("entity", "service_order"),
		)
		c.JSON(status, response)
		return
	}

	logger.Of(ctx).Debug("create response",
		zap.String("operation", "create_service_order"),
		zap.String("entity", "service_order"),
		zap.String("service_order_id", res.ID),
		zap.String("status", res.Status.String()),
	)
	c.JSON(http.StatusCreated, mapResponseDTOFromDomain(res))
}

func (h *handler) List(c *gin.Context) {
	ctx := c.Request.Context()
	params := mapListServiceOrderParamsToDomain(c)
	response, err := h.svc.List(ctx, params)
	if err != nil {
		status, resp := web.Error(err)
		c.JSON(status, resp)
		return
	}

	c.JSON(http.StatusOK, mapServiceOrderListResponse(response))
}

func (h *handler) GetWorks(c *gin.Context) {
	ctx := c.Request.Context()
	id := strings.TrimSpace(c.Param(serviceOrderIDParam))
	if id == "" {
		status, response := web.Error(domain.ErrInvalidServiceOrderId)
		c.JSON(status, response)
		return
	}

	works, err := h.svc.ListWorks(ctx, id)
	if err != nil {
		status, response := web.Error(err)
		logger.Of(ctx).Debug("list service order works failed",
			zap.String("operation", "list_service_order_works"),
			zap.Error(err),
			zap.String("entity", "service_order"),
		)
		c.JSON(status, response)
		return
	}

	c.JSON(http.StatusOK, mapWorksListToResponse(works))
}

func (h *handler) AddWork(c *gin.Context) {
	ctx := c.Request.Context()
	id := strings.TrimSpace(c.Param(serviceOrderIDParam))
	if id == "" {
		status, response := web.Error(domain.ErrInvalidServiceOrderId)
		c.JSON(status, response)
		return
	}

	dto, err := mapAddWorksBody(c)
	if err != nil {
		err = json.CheckJsonError(err)
		status, response := web.Error(err)
		logger.Of(ctx).Debug("Failed to bind add works payload",
			zap.String("operation", "add_works_to_service_order"),
			zap.Error(err),
			zap.String("entity", "service_order"),
		)
		c.JSON(status, response)
		return
	}

	if err := h.svc.AddWorks(ctx, id, dto.Services); err != nil {
		status, response := web.Error(err)
		logger.Of(ctx).Debug("add works to service order failed",
			zap.String("operation", "add_works_to_service_order"),
			zap.Error(err),
			zap.String("entity", "service_order"),
		)
		c.JSON(status, response)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *handler) DeleteWork(c *gin.Context) {
	ctx := c.Request.Context()
	soID := strings.TrimSpace(c.Param(serviceOrderIDParam))
	workID := strings.TrimSpace(c.Param(workIDPathParam))
	if soID == "" {
		status, response := web.Error(domain.ErrInvalidServiceOrderId)
		c.JSON(status, response)
		return
	}
	if workID == "" {
		status, response := web.Error(domain.ErrInvalidWorkId)
		c.JSON(status, response)
		return
	}

	if err := h.svc.RemoveWork(ctx, soID, workID); err != nil {
		status, response := web.Error(err)
		logger.Of(ctx).Debug("remove work from service order failed",
			zap.String("operation", "remove_work_from_service_order"),
			zap.Error(err),
			zap.String("entity", "service_order"),
		)
		c.JSON(status, response)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *handler) GetSupplies(c *gin.Context) {
	ctx := c.Request.Context()
	id := strings.TrimSpace(c.Param(serviceOrderIDParam))
	if id == "" {
		status, response := web.Error(domain.ErrInvalidServiceOrderId)
		c.JSON(status, response)
		return
	}

	supplies, err := h.svc.ListSupplies(ctx, id)
	if err != nil {
		status, response := web.Error(err)
		logger.Of(ctx).Debug("list service order supplies failed",
			zap.String("operation", "list_service_order_supplies"),
			zap.Error(err),
			zap.String("entity", "service_order"),
		)
		c.JSON(status, response)
		return
	}

	c.JSON(http.StatusOK, mapSuppliesListToResponse(supplies))
}

func (h *handler) AddSupplies(c *gin.Context) {
	ctx := c.Request.Context()
	id := strings.TrimSpace(c.Param(serviceOrderIDParam))
	if id == "" {
		status, response := web.Error(domain.ErrInvalidServiceOrderId)
		c.JSON(status, response)
		return
	}

	dto, err := mapAddSuppliesBody(c)
	if err != nil {
		err = json.CheckJsonError(err)
		status, response := web.Error(err)
		logger.Of(ctx).Debug("Failed to bind add supplies payload",
			zap.String("operation", "add_supplies_to_service_order"),
			zap.Error(err),
			zap.String("entity", "service_order"),
		)
		c.JSON(status, response)
		return
	}

	if err := h.svc.AddSupplies(ctx, id, dto); err != nil {
		status, response := web.Error(err)
		logger.Of(ctx).Debug("add supplies to service order failed",
			zap.String("operation", "add_supplies_to_service_order"),
			zap.Error(err),
			zap.String("entity", "service_order"),
		)
		c.JSON(status, response)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *handler) DeleteSupply(c *gin.Context) {
	soID := strings.TrimSpace(c.Param(serviceOrderIDParam))
	if soID == "" {
		status, response := web.Error(domain.ErrInvalidServiceOrderId)
		c.JSON(status, response)
		return
	}

	ctx := c.Request.Context()
	supplyID := strings.TrimSpace(c.Param(supplyIDPathParam))
	if supplyID == "" {
		status, response := web.Error(domain.ErrInvalidSupplyID)
		c.JSON(status, response)
		return
	}

	if err := h.svc.RemoveSupply(ctx, soID, supplyID); err != nil {
		status, response := web.Error(err)
		logger.Of(ctx).Debug("remove supply from service order failed",
			zap.String("operation", "remove_supply_from_service_order"),
			zap.Error(err),
			zap.String("entity", "service_order"),
		)
		c.JSON(status, response)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *handler) SendToCustomerApproval(c *gin.Context) {
	ctx := c.Request.Context()
	id := strings.TrimSpace(c.Param(serviceOrderIDParam))
	if id == "" {
		status, response := web.Error(domain.ErrInvalidServiceOrderId)
		c.JSON(status, response)
		return
	}

	if err := h.svc.SendToCustomerApproval(ctx, id); err != nil {
		status, response := web.Error(err)
		logger.Of(ctx).Debug("send service order to customer approval failed",
			zap.String("operation", "send_service_order_to_customer_approval"),
			zap.Error(err),
			zap.String("entity", "service_order"),
		)
		c.JSON(status, response)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *handler) Finish(c *gin.Context) {
	ctx := c.Request.Context()
	id := strings.TrimSpace(c.Param(serviceOrderIDParam))
	if id == "" {
		status, response := web.Error(domain.ErrInvalidServiceOrderId)
		c.JSON(status, response)
		return
	}

	if err := h.svc.Finish(ctx, id); err != nil {
		status, response := web.Error(err)
		logger.Of(ctx).Debug("finish service order failed",
			zap.String("operation", "finish_service_order"),
			zap.Error(err),
			zap.String("entity", "service_order"),
		)
		c.JSON(status, response)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *handler) SendToDiagnosis(c *gin.Context) {
	ctx := c.Request.Context()
	id := strings.TrimSpace(c.Param(serviceOrderIDParam))
	if id == "" {
		status, response := web.Error(domain.ErrInvalidServiceOrderId)
		c.JSON(status, response)
		return
	}

	if err := h.svc.SendToDiagnosis(ctx, id); err != nil {
		status, response := web.Error(err)
		logger.Of(ctx).Debug("send service order to diagnosis failed",
			zap.String("operation", "send_service_order_to_diagnosis"),
			zap.Error(err),
			zap.String("entity", "service_order"),
		)
		c.JSON(status, response)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *handler) Accept(c *gin.Context) {
	ctx := c.Request.Context()
	id := strings.TrimSpace(c.Param(serviceOrderIDParam))
	if id == "" {
		status, response := web.Error(domain.ErrInvalidServiceOrderId)
		c.JSON(status, response)
		return
	}

	userClaims, err := auth.GetClaimsFromHeader(c.Request)
	if err != nil {
		status, response := web.Error(err)
		c.JSON(status, response)
		return
	}

	if err = h.svc.Accept(ctx, id, userClaims.UserId); err != nil {
		status, response := web.Error(err)
		logger.Of(ctx).Debug("send service order to customer approval failed",
			zap.String("operation", "send_service_order_to_customer_approval"),
			zap.Error(err),
			zap.String("entity", "service_order"),
		)
		c.JSON(status, response)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *handler) Reject(c *gin.Context) {
	ctx := c.Request.Context()
	id := strings.TrimSpace(c.Param(serviceOrderIDParam))
	if id == "" {
		status, response := web.Error(domain.ErrInvalidServiceOrderId)
		c.JSON(status, response)
		return
	}

	userClaims, err := auth.GetClaimsFromHeader(c.Request)
	if err != nil {
		status, response := web.Error(err)
		c.JSON(status, response)
		return
	}

	if err = h.svc.Reject(ctx, id, userClaims.UserId); err != nil {
		status, response := web.Error(err)
		logger.Of(ctx).Debug("send service order to customer approval failed",
			zap.String("operation", "send_service_order_to_customer_approval"),
			zap.Error(err),
			zap.String("entity", "service_order"),
		)
		c.JSON(status, response)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *handler) Deliver(c *gin.Context) {
	ctx := c.Request.Context()
	id := strings.TrimSpace(c.Param(serviceOrderIDParam))
	if id == "" {
		status, response := web.Error(domain.ErrInvalidServiceOrderId)
		c.JSON(status, response)
		return
	}

	if err := h.svc.Deliver(ctx, id); err != nil {
		status, response := web.Error(err)
		logger.Of(ctx).Debug("finish an already completed service order",
			zap.String("operation", "deliver_service_order"),
			zap.Error(err),
			zap.String("entity", "service_order"),
		)
		c.JSON(status, response)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *handler) GetAverageExecutionTime(c *gin.Context) {
	ctx := c.Request.Context()
	workIDs := c.QueryArray("work_id")
	if workIDs == nil {
		workIDs = []string{}
	}
	avg, err := h.svc.AverageExecutionTime(ctx, workIDs)
	if err != nil {
		status, response := web.Error(err)
		logger.Of(ctx).Debug("get average execution time failed",
			zap.String("operation", "get_average_execution_time"),
			zap.Error(err),
			zap.String("entity", "service_order"),
		)
		c.JSON(status, response)
		return
	}
	items := make([]workExecutionTimeResponse, 0, len(avg))
	for _, w := range avg {
		items = append(items, workExecutionTimeResponse{
			WorkID:               w.WorkID,
			WorkName:             w.WorkName,
			AverageExecutionTime: w.AverageHours,
		})
	}
	c.JSON(http.StatusOK, items)
}

func (h *handler) Cancel(c *gin.Context) {
	ctx := c.Request.Context()
	id := strings.TrimSpace(c.Param(serviceOrderIDParam))
	if id == "" {
		status, response := web.Error(domain.ErrInvalidServiceOrderId)
		c.JSON(status, response)
		return
	}
	if err := h.svc.Cancel(ctx, id); err != nil {
		status, response := web.Error(err)
		logger.Of(ctx).Debug("cancel service order",
			zap.String("operation", "cancel_service_order"),
			zap.Error(err),
			zap.String("entity", "service_order"),
		)
		c.JSON(status, response)
		return
	}

	c.Status(http.StatusOK)
}

func (h *handler) GetFullByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := strings.TrimSpace(c.Param(serviceOrderIDParam))
	if id == "" {
		status, response := web.Error(domain.ErrInvalidServiceOrderId)
		c.JSON(status, response)
		return
	}
	os, err := h.svc.GetFullOSByID(ctx, id)
	if err != nil {
		status, response := web.Error(err)
		logger.Of(ctx).Debug("get full service order by id",
			zap.String("operation", "get_full_service_order_by_id"),
			zap.Error(err),
			zap.String("entity", "service_order"),
		)
		c.JSON(status, response)
		return
	}

	c.JSON(http.StatusOK, mapServiceOrderDetailResponse(os))
}

func (h *handler) GetStatus(c *gin.Context) {
	id := strings.TrimSpace(c.Param(serviceOrderIDParam))
	if id == "" {
		status, response := web.Error(domain.ErrInvalidServiceOrderId)
		c.JSON(status, response)
		return
	}
	ctx := c.Request.Context()

	statusOS, err := h.svc.GetStatus(ctx, id)
	if err != nil {
		status, response := web.Error(err)
		logger.Of(ctx).Debug("get full service order by id",
			zap.String("operation", "get_service_order_status"),
			zap.Error(err),
			zap.String("entity", "service_order"),
		)
		c.JSON(status, response)
		return
	}
	domainServiceOrder := domain.ServiceOrder{ID: id, Status: statusOS}

	c.JSON(http.StatusOK, mapResponseDTOFromDomain(domainServiceOrder))
}
