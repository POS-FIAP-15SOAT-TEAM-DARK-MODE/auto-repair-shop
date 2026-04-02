package supply

import (
	"net/http"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/web"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type handler struct {
	service domain.SupplyService
}

func HttpHandler(service domain.SupplyService) *handler {
	return &handler{
		service: service,
	}
}

func (h *handler) Create(c *gin.Context) {
	supplyRequest := c.Request.Context()
	req, err := mapBodyToCreateSupplyRequest(c)
	if err != nil {
		status, response := web.Error(err)
		logger.Of(supplyRequest).Debug("Failed to bind supply creation payload",
			zap.String("operation", "create_supply"),
			zap.Error(err),
			zap.String("entity", "supply"),
		)
		c.JSON(status, response)
		return
	}

	supply := mapSupplyRequestDTOToDomain(req)
	if err := h.service.Create(supplyRequest, supply); err != nil {
		status, response := web.Error(err)
		logger.Of(supplyRequest).Error(err)
		logger.Of(supplyRequest).Debug("Supply creation failed in service layer",
			zap.String("operation", "create_supply"),
			zap.Error(err),
			zap.String("entity", "supply"),
		)
		c.JSON(status, response)
		return
	}

	logger.Of(supplyRequest).Debug("create response", zap.Any("service", supply))
	c.JSON(http.StatusCreated, mapSupplyToResponseDTO(supply))
}
