package factory

import (
	"time"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/container"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db/postgres"
	customerHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/customer"
	pingHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/ping"
	userHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/user"
	customerRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/customer"
	userRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/user"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/env"
	customerSvc "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/customer"
	userSvc "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/user"

	workHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/work"
	workRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/work"
	workSvc "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/work"

	vehicleHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/vehicle"
	vehicleRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/vehicle"
	vehicleSvc "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/vehicle"

	supplyHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/supply"
	supplyRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/supply"
	supplySvc "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/supply"
)

func HttpContainer() *container.HTTP {
	return &container.HTTP{
		PingHandler:     newPingHandler(),
		UserHandler:     newUserHandler(),
		CustomerHandler: newCustomerHandler(),
		WorkHandler:     newWorkHandler(),
		VehicleHandler:  newVehicleHandler(),
		SupplyHandler:   newSupplyHandler(),
	}
}

func newPingHandler() container.PingHttpHandler {
	return pingHandler.HttpHandler()
}

func newUserHandler() container.UserHttpHandler {
	db := postgres.Connect()
	uow := postgres.NewTransactionalUoW(db)

	expiresIn := env.GetTimeDuration("JWT_EXPIRES_IN", 24*time.Hour)

	userRepository := userRepo.Repository()
	userService := userSvc.Service(uow, userRepository, expiresIn)
	return userHandler.HttpHandler(userService)
}

func newCustomerHandler() container.CustomerHttpHandler {
	db := postgres.Connect()
	uow := postgres.NewTransactionalUoW(db)
	userRepository := userRepo.Repository()
	customerRepository := customerRepo.Repository()

	return customerHandler.NewHandler(customerSvc.Service(uow, userRepository, customerRepository))
}

func newWorkHandler() container.WorkHttpHandler {
	db := postgres.Connect()
	uow := postgres.NewTransactionalUoW(db)
	repo := workRepo.Repository()
	svc := workSvc.Service(uow, repo)
	return workHandler.HttpHandler(svc)
}

func newVehicleHandler() container.VehicleHttpHandler {
	db := postgres.Connect()
	uow := postgres.NewTransactionalUoW(db)
	vehicleRepository := vehicleRepo.NewVehicleRepository()
	service := vehicleSvc.NewService(uow, vehicleRepository)
	return vehicleHandler.HttpHandler(service)
}

func newSupplyHandler() container.SupplyHttpHandler {
	db := postgres.Connect()
	uow := postgres.NewTransactionalUoW(db)

	supplyRepository := supplyRepo.Repository()
	service := supplySvc.Service(uow, supplyRepository)
	return supplyHandler.HttpHandler(service)
}
