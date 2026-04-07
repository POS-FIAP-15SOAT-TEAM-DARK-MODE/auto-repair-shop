package service_order_history

import (
	"net/http"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/web"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type handler struct {
	svc domain.ServiceOrderHistoryService
}

func HttpHandler(svc domain.ServiceOrderHistoryService) *handler {
	return &handler{svc: svc}
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

	response := mapResponseDTOFromDomainList(history)

	c.JSON(http.StatusOK, gin.H{"items": response})
}
