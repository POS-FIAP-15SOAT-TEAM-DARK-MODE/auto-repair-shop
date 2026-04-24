package supply

import (
	"net/http"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/web"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const idPathParamKey = "id"

type handler struct {
	svc domain.SupplyService
}

func HttpHandler(svc domain.SupplyService) *handler {
	return &handler{svc}
}

func (h *handler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	req, err := mapBodyToRequestDTO(c)
	if err != nil {
		status, response := web.Error(err)
		logger.Of(ctx).Debug("Failed to bind supply creation payload",
			zap.String("operation", "create_supply"),
			zap.Error(err),
			zap.String("entity", "supply"),
		)
		c.JSON(status, response)
		return
	}

	body := req.MapToDomain()
	logger.Of(ctx).Debug("create request", zap.Any("service", body))
	if err = h.svc.Create(ctx, body); err != nil {
		status, response := web.Error(err)
		logger.Of(ctx).Error(err)
		logger.Of(ctx).Debug("Supply creation failed in service layer",
			zap.String("operation", "create_supply"),
			zap.Error(err),
			zap.String("entity", "supply"),
		)
		c.JSON(status, response)
		return
	}

	logger.Of(ctx).Debug("create response", zap.Any("service", body))
	c.JSON(http.StatusCreated, mapSupplyToResponseDTO(body))
}

func (h *handler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param(idPathParamKey)
	if id == "" {
		status, response := web.Error(domain.ErrInvalidSupplyID)
		logger.Of(ctx).Debug("Supply invalid supply id",
			zap.String("operation", "update_supply"),
			zap.Error(domain.ErrInvalidSupplyID),
			zap.String("entity", "supply"),
			zap.String("entity.id", id),
		)
		c.JSON(status, response)
		return
	}

	dto, err := mapBodyToRequestDTO(c)
	if err != nil {
		status, response := web.Error(err)
		logger.Of(ctx).Debug("Failed to bind supply update payload",
			zap.String("operation", "update_supply"),
			zap.Error(err),
			zap.String("entity", "supply"),
		)
		c.JSON(status, response)
		return
	}

	body := dto.MapToDomain()
	body.ID = id
	logger.Of(ctx).Debug("update request", zap.Any("service", body))
	if err = h.svc.Update(ctx, body); err != nil {
		status, response := web.Error(err)
		logger.Of(ctx).Error(err)
		logger.Of(ctx).Debug("Supply update failed in service layer",
			zap.String("operation", "update_supply"),
			zap.Error(err),
			zap.String("entity", "supply"),
		)
		c.JSON(status, response)
		return
	}

	logger.Of(ctx).Debug("update response", zap.Any("service", body))
	c.JSON(http.StatusOK, mapSupplyToResponseDTO(body))

}

func (h *handler) List(c *gin.Context) {
	params := mapListParamsToDomain(c)
	ctx := c.Request.Context()

	logger.Of(ctx).Debug("list request", zap.Any("params", params))
	response, err := h.svc.List(ctx, params)
	if err != nil {
		status, errResp := web.Error(err)
		logger.Of(ctx).Error(err)
		logger.Of(ctx).Debug("Failed to list supplies in service layer",
			zap.String("operation", "list_supplies"),
			zap.Error(err),
			zap.String("entity", "supply"),
		)
		c.JSON(status, errResp)
		return
	}

	logger.Of(ctx).Debug("list response", zap.Any("service", response))
	c.JSON(http.StatusOK, mapListResponseDTOFromDomain(response))

}
