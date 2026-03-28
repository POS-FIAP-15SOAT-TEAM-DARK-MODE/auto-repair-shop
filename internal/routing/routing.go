package routing

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/http"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
)

func SetupRouter(c *http.HandlersWrapper, m *http.Middlewares) *gin.Engine {
	router := gin.New()

	for name, middleware := range *m {
		logger.Global().Info("Starting new middleware", zap.String("name", name))
		router.Use(middleware)
	}

	mountSwaggerUI(router)

	router.GET("/ping", c.PingHandler.Ping())

	v1 := router.Group("/v1")

	v1.POST("/auth/register", c.UserHandler.Create())

	v1.POST("/customers", c.CustomerHandler.Create())

	v1.POST("/works", c.WorkHandler.Create())
	v1.GET("/works", c.WorkHandler.List())
	v1.PUT("/works/:id", c.WorkHandler.Update())
	v1.DELETE("/works/:id", c.WorkHandler.Delete())

	return router
}
