package work

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
	svc domain.WorkService
}

func HttpHandler(svc domain.WorkService) *handler {
	return &handler{svc}
}

func (h *handler) Create() gin.HandlerFunc {
	return func(c *gin.Context) {
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

		if err := dto.Validate(); err != nil {
			status, response := web.Error(err)
			logger.Of(ctx).Debug("Work validation failed",
				zap.String("operation", "create_work"),
				zap.Error(err),
				zap.String("entity", "work"),
			)
			c.JSON(status, response)
			return
		}

		body, err := dto.MapToDomain()
		if err != nil {
			status, response := web.Error(err)
			logger.Of(ctx).Debug("Work parse to domain failed",
				zap.String("operation", "create_work"),
				zap.Error(err),
				zap.String("entity", "work"),
			)
			c.JSON(status, response)
			return
		}

		logger.Of(ctx).Debug("create request", zap.Any("service", body))
		if err := h.svc.Create(ctx, body); err != nil {
			status, response := web.Error(err)
			logger.Of(ctx).Error(err)
			logger.Of(ctx).Debug("Work creation failed in service layer",
				zap.String("operation", "create_work"),
				zap.Error(err),
				zap.String("entity", "work"),
			)
			c.JSON(status, response)
			return
		}

		logger.Of(ctx).Debug("create response", zap.Any("service", body))
		c.JSON(http.StatusCreated, mapResponseDTOFromDomain(body))
	}
}
