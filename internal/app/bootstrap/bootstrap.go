package bootstrap

import (
	"context"
	"database/sql"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/app"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/app/container"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/app/middleware"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/app/routing"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/auth"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/customer"
	serviceOrder "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order"
	serviceOrderHistory "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order_history"
	soHistoryRepository "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order_history/repository"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/supply"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/vehicle"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/work"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/env"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/metrics"
)

func HTTPServer() app.Server {
	prometheus.MustRegister(metrics.NewStatusDurationCollector(soHistoryRepository.NewMetricsRepository()))

	middlewares := middlewaresContainer()
	db := app.DBConnect()
	app.RunSeed(context.Background(), db)
	routes := httpContainer(db)
	router := routing.SetupRouter(routes, middlewares)

	port := env.GetString("PORT", "8080")
	return app.NewHTTPServer("http-api", ":"+port, router)
}

func middlewaresContainer() *container.Middlewares {
	return &container.Middlewares{
		"Logger":       middleware.Logger(),
		"Recovery":     middleware.Recovery(),
		"Metrics":      middleware.Metrics(),
		"ErrorHandler": middleware.ErrorHandler(),
	}
}

func httpContainer(db *sql.DB) *container.HandlersWrapper {
	return &container.HandlersWrapper{
		UserHandler:                auth.NewHTTPController(db),
		CustomerHandler:            customer.NewHTTPController(db),
		WorkHandler:                work.NewHTTPController(db),
		VehicleHandler:             vehicle.NewHTTPController(db),
		SupplyHandler:              supply.NewHTTPController(db),
		ServiceOrderHandler:        serviceOrder.NewHTTPController(db),
		ServiceOrderHistoryHandler: serviceOrderHistory.NewHTTPController(db),
	}
}
