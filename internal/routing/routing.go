package routing

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/container"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
)

func SetupRouter(c *container.HTTP, m *container.Middlewares) *gin.Engine {
	router := gin.New()

	for name, middleware := range *m {
		logger.Info("Starting new middleware", zap.String("name", name))
		router.Use(middleware)
	}

	mountSwaggerUI(router)

	router.GET("/ping", c.PingHandler.Ping)

	router.POST("/vehicle/create", c.VehicleHandler.Create)

	return router
}
