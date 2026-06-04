package factory

import (
	"context"
	"database/sql"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/app"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/auth"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/customer"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/infra/db/seed"
	pingHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/infra/handler/ping"
	container "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/infra/http"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/infra/http/middleware"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/infra/routing"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/infra/server"
	serviceOrder "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/service_order"
	serviceOrderHistory "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/service_order_history"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/supply"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/vehicle"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/work"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/pkg/env"
)

func HTTPServer() server.Server {
	middlewares := middlewaresContainer()
	db := app.DBConnect()
	seed.Run(context.Background(), db)
	routes := httpContainer(db)
	router := routing.SetupRouter(routes, middlewares)

	port := env.GetString("PORT", "8080")
	return server.NewHTTPServer("http-api", ":"+port, router)
}

func middlewaresContainer() *container.Middlewares {
	return &container.Middlewares{
		"Logger":       middleware.Logger(),
		"Recovery":     middleware.Recovery(),
		"ErrorHandler": middleware.ErrorHandler(),
	}
}

func httpContainer(db *sql.DB) *container.HandlersWrapper {
	return &container.HandlersWrapper{
		PingHandler:                newPingHandler(),
		UserHandler:                auth.NewHTTPController(db),
		CustomerHandler:            customer.NewHTTPController(db),
		WorkHandler:                work.NewHTTPController(db),
		VehicleHandler:             vehicle.NewHTTPController(db),
		SupplyHandler:              supply.NewHTTPController(db),
		ServiceOrderHandler:        serviceOrder.NewHTTPController(db),
		ServiceOrderHistoryHandler: serviceOrderHistory.NewHTTPController(db),
	}
}

func newPingHandler() container.PingHandler {
	return pingHandler.HttpHandler()
}
