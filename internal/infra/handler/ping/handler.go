package ping

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
)

type Handler struct {
	svc domain.PingService
}

func NewHandler(uc domain.PingService) *Handler {
	return &Handler{svc: uc}
}

func (h *Handler) Ping(c *gin.Context) {
	result := h.svc.Execute()
	c.JSON(http.StatusOK, gin.H{"message": result.Message})
}
