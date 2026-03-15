package customer

import (
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	pkgjson "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/json"
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
		jsonErr := pkgjson.CheckJsonError(err)
		logger.Warn("create customer: invalid JSON payload",
			zap.String("operation", "create_customer"),
			zap.String("error", jsonErr.Error()),
		)
		status, response := web.Error(jsonErr)
		c.JSON(status, response)
		return
	}

	if err := body.Validate(); err != nil {
		logger.Warn("create customer: validation failed",
			zap.String("operation", "create_customer"),
			zap.String("error", err.Error()),
		)
		status, response := web.Error(err)
		c.JSON(status, response)
		return
	}

	customer, err := body.toDomain()
	if err != nil {
		logger.Warn("create customer: domain mapping failed",
			zap.String("operation", "create_customer"),
			zap.String("error", err.Error()),
		)
		status, response := web.Error(err)
		c.JSON(status, response)
		return
	}

	if err = h.service.Create(c.Request.Context(), customer); err != nil {
		status, response := web.Error(err)
		c.JSON(status, response)
		return
	}

	c.JSON(201, toResponse(customer))
}
