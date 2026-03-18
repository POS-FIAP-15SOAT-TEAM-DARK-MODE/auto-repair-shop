package customer

import (
	"net/http"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	pkgjson "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/json"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/web"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type handler struct {
	service domain.CustomerService
}

func NewHandler(service domain.CustomerService) *handler {
	return &handler{
		service: service,
	}
}

func (h *handler) Create() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		var body createCustomerRequest

		if err := c.ShouldBindJSON(&body); err != nil {
			jsonErr := pkgjson.CheckJsonError(err)
			logger.Of(ctx).Warn("create customer: invalid JSON payload",
				zap.String("operation", "create_customer"),
				zap.String("error", jsonErr.Error()),
			)
			status, response := web.Error(jsonErr)
			c.JSON(status, response)
			return
		}

		if err := body.validate(); err != nil {
			logger.Of(ctx).Warn("create customer: validation failed",
				zap.String("operation", "create_customer"),
				zap.String("error", err.Error()),
			)
			status, response := web.Error(err)
			c.JSON(status, response)
			return
		}

		customer, err := body.Domain()
		if err != nil {
			logger.Of(ctx).Warn("create customer: domain mapping failed",
				zap.String("operation", "create_customer"),
				zap.String("error", err.Error()),
			)
			status, response := web.Error(err)
			c.JSON(status, response)
			return
		}

		if err = h.service.Create(ctx, customer); err != nil {
			status, response := web.Error(err)
			c.JSON(status, response)
			return
		}

		c.JSON(http.StatusCreated, toResponse(customer))
	}
}
