package routing

import (
	"github.com/gin-gonic/gin"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/container"
)

func SetupRouter(c *container.HTTP) *gin.Engine {
	router := gin.Default()

	router.GET("/ping", c.PingHandler.Ping)

	return router
}
