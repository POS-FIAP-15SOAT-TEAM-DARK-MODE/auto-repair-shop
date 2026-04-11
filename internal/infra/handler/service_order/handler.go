package service_order

import (
	"net/http"
	"strings"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/json"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/web"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	serviceOrderIDParam = "id"
	workIDPathParam     = "serviceId"
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

	logger.Of(ctx).Debug("create request", zap.Any("service_order", dto))
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

	logger.Of(ctx).Debug("create response", zap.Any("service_order", res))
	c.JSON(http.StatusCreated, mapResponseDTOFromDomain(res))
}

func (h *handler) Get(c *gin.Context) {
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
