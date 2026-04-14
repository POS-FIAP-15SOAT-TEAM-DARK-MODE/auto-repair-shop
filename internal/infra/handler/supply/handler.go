package supply

import (
	"net/http"
	"strconv"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/web"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type handler struct {
	service domain.SupplyService
}

const (
	defaultPage     = 1
	defaultPageSize = 10
	pageParam       = "page"
	pageSizeParam   = "pageSize"
)

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

func createSupplyListParams(c *gin.Context) *domain.ListSupplyParams {
	pg := int64(defaultPage)
	pgSize := int64(defaultPageSize)

	if page := c.Query(pageParam); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > defaultPage {
			pg = int64(p)
		}
	}
	if pageSize := c.Query(pageSizeParam); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 {
			pgSize = int64(ps)
		}
	}

	return &domain.ListSupplyParams{
		Page:     pg,
		PageSize: pgSize,
	}
}

func (h *handler) List(c *gin.Context) {
	ctx := c.Request.Context()
	params := createSupplyListParams(c)

	result, err := h.service.List(ctx, params)
	if err != nil {
		status, response := web.Error(err)
		logger.Of(ctx).Error(err)
		logger.Of(ctx).Debug("Failed to list supplies in service layer",
			zap.String("operation", "list_supplies"),
			zap.Error(err),
			zap.String("entity", "supply"),
		)
		c.JSON(status, response)
		return
	}

	c.JSON(http.StatusOK, domainListToResponseDTO(result))
}
