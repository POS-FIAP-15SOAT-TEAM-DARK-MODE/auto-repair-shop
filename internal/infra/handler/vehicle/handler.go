package vehicle

import (
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/web"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

type handler struct {
}

func HttpHandler() *handler {
	return &handler{}
}

func (h *handler) Create() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		reqDTO, err := validateAndCreateDTO(c)
		if err != nil {
			status, resp := web.Error(err)
			logger.Of(ctx).Debug("Vehicle validation error",
				zap.String("operation", "create_vehicle"),
				zap.Error(err),
				zap.String("entity", "vehicle"),
			)
			c.JSON(status, resp)
			return
		}

		// TODO: check costumer_id is in database

		_ = reqDTO.Domain()

		c.Status(http.StatusCreated)
	}
}
