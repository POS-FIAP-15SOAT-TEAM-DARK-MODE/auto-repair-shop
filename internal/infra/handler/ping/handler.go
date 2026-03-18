package ping

import (
	"net/http"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/container"
	"github.com/gin-gonic/gin"
)

type httpHandler struct{}

func HttpHandler() container.PingHttpHandler {
	return &httpHandler{}
}

func (h *httpHandler) Ping() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	}
}
