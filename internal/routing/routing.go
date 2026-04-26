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

	for name, mid := range *m {
		logger.Global().Info("Starting new middleware", zap.String("name", name))
		router.Use(mid)
	}

	mountSwaggerUI(router)

	router.GET("/ping", c.PingHandler.Ping)

	v1 := router.Group("/v1")

	v1.POST("/auth/register", middleware.Auth(role.ADMIN), c.UserHandler.Create)
	v1.POST("/auth/login", c.UserHandler.Login)

	v1.POST("/customers", middleware.Auth(role.AttendantRoles...), c.CustomerHandler.Create)
	v1.GET("/customers/:id", middleware.Auth(role.AttendantRoles...), c.CustomerHandler.GetByID)
	v1.GET("/customers", middleware.Auth(role.AttendantRoles...), c.CustomerHandler.GetByDocument)
	v1.PUT("/customers/:id", middleware.Auth(role.AttendantRoles...), c.CustomerHandler.Update)
	v1.DELETE("/customers/:id", middleware.Auth(role.AttendantRoles...), c.CustomerHandler.Delete)

	v1.POST("/works", middleware.Auth(role.AttendantRoles...), c.WorkHandler.Create)
	v1.GET("/works", middleware.Auth(role.AttendantAndMechanicRoles...), c.WorkHandler.List)
	v1.PUT("/works/:id", middleware.Auth(role.AttendantAndMechanicRoles...), c.WorkHandler.Update)
	v1.DELETE("/works/:id", middleware.Auth(role.AttendantAndMechanicRoles...), c.WorkHandler.Delete)

	v1.POST("/vehicles", middleware.Auth(role.AttendantAndMechanicRoles...), c.VehicleHandler.Create)
	v1.GET("/vehicles", middleware.Auth(role.AttendantAndMechanicRoles...), c.VehicleHandler.FindByLicensePlate)
	v1.PUT("/vehicles/:id", middleware.Auth(role.AttendantAndMechanicRoles...), c.VehicleHandler.Update)
	v1.DELETE("/vehicles/:id", middleware.Auth(role.AttendantAndMechanicRoles...), c.VehicleHandler.Delete)
	v1.GET("/vehicles/:customerId", middleware.Auth(role.AttendantRoles...), c.VehicleHandler.FindByCustomer)

	v1.POST("/supplies", middleware.Auth(role.AttendantAndMechanicRoles...), c.SupplyHandler.Create)
	v1.GET("/supplies", middleware.Auth(role.AttendantAndMechanicRoles...), c.SupplyHandler.List)
	v1.PUT("/supplies/:id", middleware.Auth(role.AttendantAndMechanicRoles...), c.SupplyHandler.Update)
	v1.DELETE("/supplies/:id", middleware.Auth(role.AttendantAndMechanicRoles...), c.SupplyHandler.Delete)

	v1.POST("/service-order", middleware.Auth(role.AttendantRoles...), c.ServiceOrderHandler.Create)
	v1.GET("/service-order/:id/history", middleware.Auth(role.AttendantRoles...), c.ServiceOrderHistoryHandler.GetHistoryByID)
	v1.PUT("/service-order/:id/send", middleware.Auth(role.AttendantAndMechanicRoles...), c.ServiceOrderHandler.SendToCustomerApproval)
	v1.PUT("/service-order/:id/accept", middleware.Auth(role.CustomerRoles...), c.ServiceOrderHandler.Accept)
	v1.PUT("/service-order/:id/reject", middleware.Auth(role.CustomerRoles...), c.ServiceOrderHandler.Reject)
	v1.PUT("/service-order/:id/deliver", middleware.Auth(role.AttendantRoles...), c.ServiceOrderHandler.Deliver)
	v1.PUT("/service-order/:id/cancel", middleware.Auth(role.AttendantRoles...), c.ServiceOrderHandler.Cancel)

	v1.GET("/service-order/:id/services", middleware.Auth(role.AttendantAndMechanicRoles...), c.ServiceOrderHandler.GetWorks)
	v1.POST("/service-order/:id/services", middleware.Auth(role.AttendantRoles...), c.ServiceOrderHandler.AddWork)
	v1.DELETE("/service-order/:id/services/:serviceId", middleware.Auth(role.AttendantRoles...), c.ServiceOrderHandler.DeleteWork)

	v1.GET("/service-order/:id/supplies", middleware.Auth(role.AttendantAndMechanicRoles...), c.ServiceOrderHandler.GetSupplies)
	v1.POST("/service-order/:id/supplies", middleware.Auth(role.MechanicRoles...), c.ServiceOrderHandler.AddSupplies)
	v1.DELETE("/service-order/:id/supplies/:supplyId", middleware.Auth(role.MechanicRoles...), c.ServiceOrderHandler.DeleteSupply)

	v1.GET("/reports/average-execution-time", middleware.Auth(role.AttendantRoles...), c.ServiceOrderHandler.GetAverageExecutionTime)

	return router
}
