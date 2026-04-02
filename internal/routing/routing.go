package routing

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	role "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/http"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/http/middleware"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
)

func SetupRouter(c *http.HandlersWrapper, m *http.Middlewares) *gin.Engine {
	router := gin.New()

	for name, mw := range *m {
		logger.Global().Info("Starting new middleware", zap.String("name", name))
		router.Use(mw)
	}

	mountSwaggerUI(router)

	router.GET("/ping", c.PingHandler.Ping)

	v1 := router.Group("/v1")

	v1.POST("/auth/register", middleware.Auth(role.ADMIN), c.UserHandler.Create)
	v1.POST("/auth/login", c.UserHandler.Login)

	v1.POST("/customers", middleware.Auth(role.AttendantRoles...), c.CustomerHandler.Create)
	v1.GET("/customers/:id", middleware.Auth(role.AttendantRoles...), c.CustomerHandler.GetByID)
	v1.GET("/customers", middleware.Auth(role.AttendantRoles...), c.CustomerHandler.GetByDocument)

	v1.POST("/works", middleware.Auth(role.AttendantRoles...), c.WorkHandler.Create)
	v1.GET("/works", middleware.Auth(role.AttendantAndMechanicRoles...), c.WorkHandler.List)
	v1.PUT("/works/:id", middleware.Auth(role.AttendantAndMechanicRoles...), c.WorkHandler.Update)
	v1.DELETE("/works/:id", middleware.Auth(role.AttendantAndMechanicRoles...), c.WorkHandler.Delete)

	v1.POST("/vehicle", middleware.Auth(role.AttendantAndMechanicRoles...), c.VehicleHandler.Create)
	v1.GET("/vehicle", middleware.Auth(role.AttendantAndMechanicRoles...), c.VehicleHandler.FindByLicensePlate)
	v1.PUT("/vehicle/:id", c.VehicleHandler.Update)

	return router
}
