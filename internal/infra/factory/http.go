package factory

import (
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/container"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db/postgres"
	pingHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/ping"
	userHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/user"
	userRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/user"
	userSvc "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/user"

	workHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/work"
	workRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/work"
	workSvc "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/work"
)

func HttpContainer() *container.HTTP {
	return &container.HTTP{
		PingHandler: newPingHandler(),
		UserHandler: newUserHandler(),
		WorkHandler: newWorkHandler(),
	}
}

func newPingHandler() container.PingHttpHandler {
	return pingHandler.HttpHandler()
}

func newUserHandler() container.UserHttpHandler {
	db := postgres.Connect()
	uow := postgres.NewTransactionalUoW(db)
	userRepository := userRepo.Repository()
	userService := userSvc.Service(uow, userRepository)
	return userHandler.HttpHandler(userService)
}

func newWorkHandler() container.WorkHttpHandler {
	db := postgres.Connect()
	uow := postgres.NewTransactionalUoW(db)
	repo := workRepo.Repository()
	svc := workSvc.Service(uow, repo)
	return workHandler.HttpHandler(svc)
}
