package customer

import (
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/web"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	service domain.CustomerService
}

func NewHandler(service domain.CustomerService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Create(c *gin.Context) {
	var body CreateCustomerRequest

	if err := c.ShouldBindJSON(&body); err != nil {
		logger.Warn("create customer: invalid JSON payload",
			zap.String("operation", "create_customer"),
			zap.String("error", err.Error()),
		)
		web.Error(c, domain.ValidationError{Message: "invalid JSON payload"})
		return
	}

	if err := body.Validate(); err != nil {
		logger.Warn("create customer: validation failed",
			zap.String("operation", "create_customer"),
			zap.String("error", err.Error()),
		)
		web.Error(c, err)
		return
	}

	customer, err := body.toDomain()
	if err != nil {
		logger.Warn("create customer: validation failed",
			zap.String("operation", "create_customer"),
			zap.String("error", err.Error()),
		)
		web.Error(c, err)
		return
	}

	if err = h.service.Create(c.Request.Context(), customer); err != nil {
		web.Error(c, err)
		return
	}

	c.JSON(201, toResponse(customer))
}
