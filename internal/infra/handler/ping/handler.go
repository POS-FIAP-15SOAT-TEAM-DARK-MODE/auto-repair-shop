package ping

import (
	goHttp "net/http"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/http"
	"github.com/gin-gonic/gin"
)

type httpHandler struct{}

func HttpHandler() http.PingHandler {
	return &httpHandler{}
}

func (h *httpHandler) Ping(c *gin.Context) {
	c.JSON(goHttp.StatusOK, gin.H{"message": "pong"})
}
