package factory

import (
	"database/sql"
	"time"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db/postgres"
	customerHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/customer"
	pingHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/ping"
	userHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/user"
	container "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/http"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/http/middleware"
	customerRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/customer"
	userRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/user"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/server"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/env"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/routing"
	customerSvc "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/customer"
	userSvc "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/user"

	workHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/work"
	workRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/work"
	workSvc "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/work"

	vehicleHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/vehicle"
	vehicleRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/vehicle"
	vehicleSvc "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/vehicle"
)

func HTTPServer() server.Server {
	middlewares := middlewaresContainer()
	db := postgres.Connect()
	httpContainer := httpContainer(db)
	router := routing.SetupRouter(httpContainer, middlewares)

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
		PingHandler:     newPingHandler(),
		UserHandler:     newUserHandler(db),
		CustomerHandler: newCustomerHandler(db),
		WorkHandler:     newWorkHandler(db),
		VehicleHandler:  newVehicleHandler(db),
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

func newVehicleHandler(db *sql.DB) container.VehicleHandler {
	uow := postgres.NewTransactionalUoW(db)
	vehicleRepository := vehicleRepo.NewVehicleRepository()
	service := vehicleSvc.NewService(uow, vehicleRepository)
	return vehicleHandler.HttpHandler(service)
}
