package factory

import (
	"context"
	"database/sql"
	"time"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db/postgres"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db/seed"
	customerHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/customer"
	pingHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/ping"
	userHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/user"
	workHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/work"
	container "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/http"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/http/middleware"
	customerRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/customer"
	soHistoryRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/service_order_history"
	userRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/user"
	workRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/work"
	workSOHistoryRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/work_service_order_history"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/server"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/env"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/routing"
	customerSvc "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/customer"
	userSvc "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/user"
	workSvc "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/work"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/pkg/id"
	uow2 "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/pkg/uow"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/supply"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/vehicle"
	vehicleRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/vehicle/repository"

	soHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/service_order"
	soRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/service_order"
	soSvc "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/service_order"
	supplyRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/supply/repository"

	soHistoryHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/service_order_history"
	soHistorySvc "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/service_order_history"
)

func HTTPServer() server.Server {
	middlewares := middlewaresContainer()
	db := postgres.Connect()
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
		UserHandler:                newUserHandler(db),
		CustomerHandler:            newCustomerHandler(db),
		WorkHandler:                newWorkHandler(db),
		VehicleHandler:             vehicle.NewHTTPController(db),
		SupplyHandler:              supply.NewHTTPController(db),
		ServiceOrderHandler:        newServiceOrderHandler(db),
		ServiceOrderHistoryHandler: newServiceOrderHistoryHandler(db),
	}
}

func newPingHandler() container.PingHandler {
	return pingHandler.HttpHandler()
}

func newUserHandler(db *sql.DB) container.UserHandler {
	uow := postgres.NewTransactionalUoW(db)

	expiresIn := env.GetTimeDuration("JWT_EXPIRES_IN", 24*time.Hour)

	userRepository := userRepo.Repository()
	userService := userSvc.Service(uow, userRepository, expiresIn)
	return userHandler.HttpHandler(userService)
}

func newCustomerHandler(db *sql.DB) container.CustomerHandler {
	uow := postgres.NewTransactionalUoW(db)
	userRepository := userRepo.Repository()
	customerRepository := customerRepo.Repository()

	return customerHandler.NewHandler(customerSvc.Service(uow, userRepository, customerRepository))
}

func newWorkHandler(db *sql.DB) container.WorkHandler {
	uow := postgres.NewTransactionalUoW(db)
	repo := workRepo.Repository()
	svc := workSvc.Service(uow, repo)
	return workHandler.HttpHandler(svc)
}

func newServiceOrderHandler(db *sql.DB) container.ServiceOrderHandler {
	uow := postgres.NewTransactionalUoW(db)

	idGen := id.NewIDGenerator()
	u := uow2.NewTransactionalUoW(db)

	repo := soRepo.Repository()
	workRepository := workRepo.Repository()
	workHistoryRepository := soHistoryRepo.Repository()
	workSOHistoryRepository := workSOHistoryRepo.Repository()
	userRepository := userRepo.Repository()
	customerRepository := customerRepo.Repository()

	supplyRepository := supplyRepo.NewPostgres()
	supplyService := supply.NewService(u, supplyRepository, idGen)

	vehicleRepository := vehicleRepo.NewPostgres()
	vehicleSvc := vehicle.NewService(u, vehicleRepository, idGen)

	customer := customerSvc.Service(uow, userRepository, customerRepository)
	svc := soSvc.Service(uow, repo, workRepository, supplyService, workHistoryRepository, workSOHistoryRepository, customer, vehicleSvc)

	return soHandler.HttpHandler(svc)
}

func newServiceOrderHistoryHandler(db *sql.DB) container.ServiceOrderHistoryHandler {
	uow := postgres.NewTransactionalUoW(db)
	repo := soHistoryRepo.Repository()
	svc := soHistorySvc.Service(uow, repo)
	return soHistoryHandler.HttpHandler(svc)
}
