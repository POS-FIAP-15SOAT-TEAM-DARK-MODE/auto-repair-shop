package routing

import (
	"context"
	http2 "net/http"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/web"
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

	v1.POST("/auth/register", middleware.Auth(role.ADMIN), GinHandler(c.UserHandler.Register, http2.StatusCreated))
	v1.POST("/auth/login", GinHandler(c.UserHandler.Login))
	v1.PATCH("/users/:id/role", middleware.Auth(role.ADMIN), GinOnlyErrorHandler(c.UserHandler.ChangeRole))

	v1.GET("/customers", middleware.Auth(role.AttendantRoles...), c.CustomerHandler.GetByDocument)
	v1.GET("/customers/:id", middleware.Auth(role.AttendantRoles...), c.CustomerHandler.GetByID)
	v1.POST("/customers", middleware.Auth(role.AttendantRoles...), c.CustomerHandler.Create)
	v1.PUT("/customers/:id", middleware.Auth(role.AttendantRoles...), c.CustomerHandler.Update)
	v1.DELETE("/customers/:id", middleware.Auth(role.AttendantRoles...), c.CustomerHandler.Delete)

	v1.GET("/works", middleware.Auth(role.AttendantAndMechanicRoles...), c.WorkHandler.List)
	v1.POST("/works", middleware.Auth(role.AttendantRoles...), c.WorkHandler.Create)
	v1.PUT("/works/:id", middleware.Auth(role.AttendantRoles...), c.WorkHandler.Update)
	v1.DELETE("/works/:id", middleware.Auth(role.AttendantRoles...), c.WorkHandler.Delete)

	v1.GET("/vehicles", middleware.Auth(role.AttendantAndMechanicRoles...), GinHandler(c.VehicleHandler.List))
	v1.POST("/vehicles", middleware.Auth(role.AttendantRoles...), GinHandler(c.VehicleHandler.Create, http2.StatusCreated))
	v1.PUT("/vehicles/:id", middleware.Auth(role.AttendantRoles...), GinHandler(c.VehicleHandler.Edit))
	v1.DELETE("/vehicles/:id", middleware.Auth(role.AttendantRoles...), GinOnlyErrorHandler(c.VehicleHandler.Delete))

	v1.GET("/supplies", middleware.Auth(role.AttendantAndMechanicRoles...), GinHandler(c.SupplyHandler.List))
	v1.POST("/supplies", middleware.Auth(role.AttendantAndMechanicRoles...), GinHandler(c.SupplyHandler.Create, http2.StatusCreated))
	v1.PUT("/supplies/:id", middleware.Auth(role.AttendantAndMechanicRoles...), GinHandler(c.SupplyHandler.Update))
	v1.DELETE("/supplies/:id", middleware.Auth(role.AttendantAndMechanicRoles...), GinOnlyErrorHandler(c.SupplyHandler.Delete))

	v1.POST("/service-order", middleware.Auth(role.AttendantRoles...), c.ServiceOrderHandler.Create)
	v1.POST("/service-order/:id/works", middleware.Auth(role.AttendantRoles...), c.ServiceOrderHandler.AddWork)
	v1.DELETE("/service-order/:id/works/:serviceId", middleware.Auth(role.AttendantRoles...), c.ServiceOrderHandler.DeleteWork)
	v1.PUT("/service-order/:id/received", middleware.Auth(role.AttendantRoles...), c.ServiceOrderHandler.Receive)
	v1.PUT("/service-order/:id/start-diagnosis", middleware.Auth(role.MechanicRoles...), c.ServiceOrderHandler.SendToDiagnosis)
	v1.POST("/service-order/:id/supplies", middleware.Auth(role.MechanicRoles...), c.ServiceOrderHandler.AddSupplies)
	v1.DELETE("/service-order/:id/supplies/:supplyId", middleware.Auth(role.MechanicRoles...), c.ServiceOrderHandler.DeleteSupply)
	v1.PUT("/service-order/:id/send", middleware.Auth(role.MechanicRoles...), c.ServiceOrderHandler.SendToCustomerApproval)
	v1.PUT("/service-order/:id/accept", middleware.Auth(role.CustomerRoles...), c.ServiceOrderHandler.Accept)
	v1.PUT("/service-order/:id/reject", middleware.Auth(role.CustomerRoles...), c.ServiceOrderHandler.Reject)
	v1.PUT("/service-order/:id/cancel", middleware.Auth(role.AttendantAndMechanicRoles...), c.ServiceOrderHandler.Cancel)
	v1.PUT("/service-order/:id/work/:workId/next", middleware.Auth(role.MechanicRoles...), c.ServiceOrderHandler.NextWork)
	v1.PUT("/service-order/:id/work/:workId/cancel", middleware.Auth(role.MechanicRoles...), c.ServiceOrderHandler.CancelWork)
	v1.PUT("/service-order/:id/finish", middleware.Auth(role.MechanicRoles...), c.ServiceOrderHandler.Finish)
	v1.PUT("/service-order/:id/deliver", middleware.Auth(role.AttendantRoles...), c.ServiceOrderHandler.Deliver)

	v1.GET("/service-order", middleware.Auth(role.AttendantAndMechanicRoles...), c.ServiceOrderHandler.List)
	v1.GET("/service-order/:id", middleware.Auth(role.AttendantAndMechanicRoles...), c.ServiceOrderHandler.GetFullByID)
	v1.GET("/service-order/:id/history", middleware.Auth(role.AttendantAndMechanicRoles...), c.ServiceOrderHistoryHandler.GetHistoryByID)
	v1.GET("/service-order/:id/works", middleware.Auth(role.AttendantAndMechanicRoles...), c.ServiceOrderHandler.GetWorks)
	v1.GET("/service-order/:id/supplies", middleware.Auth(role.AttendantAndMechanicRoles...), c.ServiceOrderHandler.GetSupplies)
	v1.GET("/reports/average-execution-time", middleware.Auth(role.AttendantRoles...), c.ServiceOrderHandler.GetAverageExecutionTime)

	v1.GET("/service-order/:id/status", c.ServiceOrderHandler.GetStatus)

	return router
}

func GinOnlyErrorHandler(cb func(context.Context, *http2.Request) error, customStatus ...int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		// context keys are gin param names (strings) intentionally;
		// collision risk is acceptable here since params are route-scoped.
		for _, v := range c.Params {
			ctx = context.WithValue(ctx, v.Key, v.Value) //nolint:staticcheck
		}

		if err := cb(ctx, c.Request); err != nil {
			status, errResp := web.Error(err)
			c.JSON(status, errResp)
			return
		}

		if len(customStatus) == 0 {
			c.Status(http2.StatusNoContent)
			c.Writer.WriteHeaderNow()
			return
		}

		c.Status(customStatus[0])
		c.Writer.WriteHeaderNow()
	}
}

func GinHandler[T any](cb func(context.Context, *http2.Request) (T, error), customStatus ...int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		// context keys are gin param names (strings) intentionally;
		// collision risk is acceptable here since params are route-scoped.
		for _, v := range c.Params {
			ctx = context.WithValue(ctx, v.Key, v.Value) //nolint:staticcheck
		}

		resp, err := cb(ctx, c.Request)
		if err != nil {
			status, errResp := web.Error(err)
			c.JSON(status, errResp)
			return
		}

		if len(customStatus) == 0 {
			c.JSON(http2.StatusOK, resp)
			return
		}

		c.JSON(customStatus[0], resp)
	}
}
