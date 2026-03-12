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

	v1 := router.Group("/v1")

	v1.POST("/v1/auth/register", c.UserHandler.CreateUser)

	v1.POST("/customers", c.CustomerHandler.Create)

	return router
}
