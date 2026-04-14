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

func (h *handler) List(c *gin.Context) {
	supplyRequest := c.Request.Context()
	supplies, err := h.service.List(supplyRequest)
	if err != nil {
		status, response := web.Error(err)
		logger.Of(supplyRequest).Error(err)
		logger.Of(supplyRequest).Debug("Failed to list supplies in service layer",
			zap.String("operation", "list_supplies"),
			zap.Error(err),
			zap.String("entity", "supply"),
		)
		c.JSON(status, response)
		return
	}
	responseDTOs := make([]interface{}, len(supplies))
	for i, supply := range supplies {
		responseDTOs[i] = mapSupplyToResponseDTO(supply)
	}
	c.JSON(http.StatusOK, responseDTOs)
}

func (h *handler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id") // ← vem da rota, não do body

	req, err := mapBodyToUpdateSupplyRequest(c)
	if err != nil {
		status, response := web.Error(err)
		c.JSON(status, response)
		return
	}

	supply := mapUpdateSupplyRequestDTOToDomain(id, req)
	if err := h.service.Update(ctx, supply); err != nil {
		status, response := web.Error(err)
		c.JSON(status, response)
		return
	}

	c.JSON(http.StatusOK, supply)
}
