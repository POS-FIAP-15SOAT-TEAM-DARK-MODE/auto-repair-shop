package factory

import (
	_ "github.com/lib/pq"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/container"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db/postgres"
	pingHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/ping"
	pingSvc "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/ping"

	serviceHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/service"
	serviceRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/service"
	serviceSvc "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/service"
)

func HttpContainer() *container.HTTP {
	return &container.HTTP{
		PingHandler:    newPingHandler(),
		ServiceHandler: newServiceHandler(),
	}
}

func newPingHandler() *pingHandler.Handler {
	pingService := pingSvc.Service()
	return pingHandler.NewHandler(pingService)
}

func newServiceHandler() *serviceHandler.HttpHandler {
	db := postgres.Connect()
	uow := postgres.NewTransactionalUoW(db)
	repo := serviceRepo.Repository()
	svc := serviceSvc.Service(uow, repo)
	return serviceHandler.Handler(svc)
}
