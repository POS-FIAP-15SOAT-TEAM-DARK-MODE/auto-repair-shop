package routing

import (
	"context"
	http2 "net/http"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/web"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	authDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/auth/domain"
	newHttp "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/http"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/http/middleware"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
)

func SetupRouter(c *newHttp.HandlersWrapper, m *newHttp.Middlewares) *gin.Engine {
	router := gin.New()

	for name, mid := range *m {
		logger.Global().Info("Starting new middleware", zap.String("name", name))
		router.Use(mid)
	}

	mountSwaggerUI(router)

	router.GET("/ping", c.PingHandler.Ping)

	v1 := router.Group("/v1")

	v1.POST("/auth/register", middleware.Auth(authDomain.ADMIN), GinHandler(c.UserHandler.Register, http2.StatusCreated))
	v1.POST("/auth/login", GinHandler(c.UserHandler.Login))
	v1.PATCH("/users/:id/role", middleware.Auth(authDomain.ADMIN), GinOnlyErrorHandler(c.UserHandler.ChangeRole))

	v1.GET("/customers", middleware.Auth(authDomain.AttendantRoles...), GinHandler(c.CustomerHandler.GetByDocument))
	v1.GET("/customers/:id", middleware.Auth(authDomain.AttendantRoles...), GinHandler(c.CustomerHandler.GetByID))
	v1.POST("/customers", middleware.Auth(authDomain.AttendantRoles...), GinHandler(c.CustomerHandler.Create, http2.StatusCreated))
	v1.PUT("/customers/:id", middleware.Auth(authDomain.AttendantRoles...), GinHandler(c.CustomerHandler.Update))
	v1.DELETE("/customers/:id", middleware.Auth(authDomain.AttendantRoles...), GinOnlyErrorHandler(c.CustomerHandler.Delete))

	v1.GET("/works", middleware.Auth(authDomain.AttendantAndMechanicRoles...), GinHandler(c.WorkHandler.List))
	v1.POST("/works", middleware.Auth(authDomain.AttendantRoles...), GinHandler(c.WorkHandler.Create, http2.StatusCreated))
	v1.PUT("/works/:id", middleware.Auth(authDomain.AttendantRoles...), GinHandler(c.WorkHandler.Update))
	v1.DELETE("/works/:id", middleware.Auth(authDomain.AttendantRoles...), GinOnlyErrorHandler(c.WorkHandler.Delete))

	v1.GET("/vehicles", middleware.Auth(authDomain.AttendantAndMechanicRoles...), GinHandler(c.VehicleHandler.List))
	v1.POST("/vehicles", middleware.Auth(authDomain.AttendantRoles...), GinHandler(c.VehicleHandler.Create, http2.StatusCreated))
	v1.PUT("/vehicles/:id", middleware.Auth(authDomain.AttendantRoles...), GinHandler(c.VehicleHandler.Edit))
	v1.DELETE("/vehicles/:id", middleware.Auth(authDomain.AttendantRoles...), GinOnlyErrorHandler(c.VehicleHandler.Delete))

	v1.GET("/supplies", middleware.Auth(authDomain.AttendantAndMechanicRoles...), GinHandler(c.SupplyHandler.List))
	v1.POST("/supplies", middleware.Auth(authDomain.AttendantAndMechanicRoles...), GinHandler(c.SupplyHandler.Create, http2.StatusCreated))
	v1.PUT("/supplies/:id", middleware.Auth(authDomain.AttendantAndMechanicRoles...), GinHandler(c.SupplyHandler.Update))
	v1.DELETE("/supplies/:id", middleware.Auth(authDomain.AttendantAndMechanicRoles...), GinOnlyErrorHandler(c.SupplyHandler.Delete))

	v1.POST("/service-order", middleware.Auth(authDomain.AttendantRoles...), GinHandler(c.ServiceOrderHandler.Create, http2.StatusCreated))
	v1.POST("/service-order/:id/works", middleware.Auth(authDomain.AttendantRoles...), GinOnlyErrorHandler(c.ServiceOrderHandler.AddWork))
	v1.DELETE("/service-order/:id/works/:serviceId", middleware.Auth(authDomain.AttendantRoles...), GinOnlyErrorHandler(c.ServiceOrderHandler.DeleteWork))
	v1.PUT("/service-order/:id/received", middleware.Auth(authDomain.AttendantRoles...), GinOnlyErrorHandler(c.ServiceOrderHandler.Receive))
	v1.PUT("/service-order/:id/start-diagnosis", middleware.Auth(authDomain.MechanicRoles...), GinOnlyErrorHandler(c.ServiceOrderHandler.SendToDiagnosis))
	v1.POST("/service-order/:id/supplies", middleware.Auth(authDomain.MechanicRoles...), GinOnlyErrorHandler(c.ServiceOrderHandler.AddSupplies))
	v1.DELETE("/service-order/:id/supplies/:supplyId", middleware.Auth(authDomain.MechanicRoles...), GinOnlyErrorHandler(c.ServiceOrderHandler.DeleteSupply))
	v1.PUT("/service-order/:id/send", middleware.Auth(authDomain.MechanicRoles...), GinOnlyErrorHandler(c.ServiceOrderHandler.SendToCustomerApproval))
	v1.PUT("/service-order/:id/accept", middleware.Auth(authDomain.CustomerRoles...), GinOnlyErrorHandler(c.ServiceOrderHandler.Accept))
	v1.PUT("/service-order/:id/reject", middleware.Auth(authDomain.CustomerRoles...), GinOnlyErrorHandler(c.ServiceOrderHandler.Reject))
	v1.PUT("/service-order/:id/cancel", middleware.Auth(authDomain.AttendantAndMechanicRoles...), GinOnlyErrorHandler(c.ServiceOrderHandler.Cancel, http2.StatusOK))
	v1.PUT("/service-order/:id/work/:workId/next", middleware.Auth(authDomain.MechanicRoles...), GinOnlyErrorHandler(c.ServiceOrderHandler.NextWork))
	v1.PUT("/service-order/:id/work/:workId/cancel", middleware.Auth(authDomain.MechanicRoles...), GinOnlyErrorHandler(c.ServiceOrderHandler.CancelWork))
	v1.PUT("/service-order/:id/finish", middleware.Auth(authDomain.MechanicRoles...), GinOnlyErrorHandler(c.ServiceOrderHandler.Finish))
	v1.PUT("/service-order/:id/deliver", middleware.Auth(authDomain.AttendantRoles...), GinOnlyErrorHandler(c.ServiceOrderHandler.Deliver))

	v1.GET("/service-order", middleware.Auth(authDomain.AttendantAndMechanicRoles...), GinHandler(c.ServiceOrderHandler.List))
	v1.GET("/service-order/:id", middleware.Auth(authDomain.AttendantAndMechanicRoles...), GinHandler(c.ServiceOrderHandler.GetFullByID))
	v1.GET("/service-order/:id/history", middleware.Auth(authDomain.AttendantAndMechanicRoles...), GinHandler(c.ServiceOrderHistoryHandler.GetHistoryByID))
	v1.GET("/service-order/:id/works", middleware.Auth(authDomain.AttendantAndMechanicRoles...), GinHandler(c.ServiceOrderHandler.GetWorks))
	v1.GET("/service-order/:id/supplies", middleware.Auth(authDomain.AttendantAndMechanicRoles...), GinHandler(c.ServiceOrderHandler.GetSupplies))
	v1.GET("/reports/average-execution-time", middleware.Auth(authDomain.AttendantRoles...), GinHandler(c.ServiceOrderHandler.GetAverageExecutionTime))

	v1.GET("/service-order/:id/status", GinHandler(c.ServiceOrderHandler.GetStatus))

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
