package service

import (
	"net/http"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type HttpHandler struct {
	svc domain.ServiceService
}

func Handler(svc domain.ServiceService) *HttpHandler {
	return &HttpHandler{svc}
}

func (h *HttpHandler) Create(c *gin.Context) {
	dto, err := mapBodyToRequestDTO(c)
	if err != nil {
		// TODO: replace by custom way to deal with error handling
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := dto.Validate(); err != nil {
		// TODO: replace by custom way to deal with error handling
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	body, err := dto.MapToDomain()
	if err != nil {
		// TODO: replace by custom way to deal with error handling
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	logger.Of(ctx).Debug("create request", zap.Any("service", body))
	svc, err := h.svc.Create(ctx, body)
	if err != nil {
		logger.Of(ctx).Error(err)
		// TODO: replace by custom way to deal with error handling
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Of(ctx).Debug("create response", zap.Any("service", svc))
	resp := mapResponseDTOFromDomain(svc)
	// TODO: replace by custom way to deal with response handling
	c.JSON(http.StatusCreated, resp)
}

func (h *HttpHandler) List(c *gin.Context) {
	params := mapListParamsToDomain(c)
	ctx := c.Request.Context()

	logger.Of(ctx).Debug("list request", zap.Any("params", params))
	response, err := h.svc.List(ctx, params)
	if err != nil {
		logger.Of(ctx).Error(err)
		// TODO: replace by custom way to deal with error handling
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Of(ctx).Debug("list response", zap.Any("service", response))
	// TODO: replace by custom way to deal with response handling
	c.JSON(http.StatusCreated, response)
}
