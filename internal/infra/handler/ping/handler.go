package ping

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
)

type Handler struct {
	svc domain.PingService
}

func NewHandler(uc domain.PingService) *Handler {
	return &Handler{svc: uc}
}

func (h *Handler) Ping(c *gin.Context) {
	result := h.svc.Execute()
	logger.Info("ping called", zap.String("response", result.Message))
	c.JSON(http.StatusOK, gin.H{"message": result.Message})
}
