package factory

import (
	"context"
	"database/sql"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/app"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/auth"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/customer"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db/seed"
	pingHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/ping"
	container "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/http"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/http/middleware"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/routing"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/server"
	serviceOrder "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order"
	serviceOrderHistory "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order_history"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/supply"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/vehicle"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/work"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/env"
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
