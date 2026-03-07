package routing

import (
	"github.com/gin-gonic/gin"

	pingHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/ping"
	ping "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/ping"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	pingSrvc := ping.Service()
	ph := pingHandler.NewHandler(pingSrvc)

	router.GET("/ping", ph.Ping)

	return router
}
