package service_order

import (
	"net/http"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/json"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/web"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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

func (h *handler) GetHistoryByID(c *gin.Context) {
	ctx := c.Request.Context()

	req := getServiceOrderHistoryDTO{ID: c.Param("id")}
	if err := req.validate(); err != nil {
		status, response := web.Error(err)
		logger.Of(ctx).Error(err)
		logger.Of(ctx).Debug("Failed to bind service order history request",
			zap.String("operation", "get_service_order_history"),
			zap.Error(err),
			zap.String("entity", "work"),
		)
		c.JSON(status, response)
		return
	}

	history, err := h.svc.GetHistoryByID(ctx, req.ID)
	if err != nil {
		status, response := web.Error(err)
		logger.Of(ctx).Error(err)
		logger.Of(ctx).Debug("Service order history retrieval failed in service layer",
			zap.String("operation", "get_service_order_history"),
			zap.Error(err),
			zap.String("entity", "service_order"),
		)
		c.JSON(status, response)
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": history})
}
