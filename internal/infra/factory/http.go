package factory

import (
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/container"
	pingHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/ping"
	vehicleHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/vehicle"
	pingSvc "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/ping"
)

func HttpContainer() *container.HTTP {
	return &container.HTTP{
		PingHandler:    newPingHandler(),
		VehicleHandler: newVehicleHandler(),
	}
}

func newPingHandler() *pingHandler.Handler {
	pingService := pingSvc.Service()
	return pingHandler.NewHandler(pingService)
}

func newVehicleHandler() *vehicleHandler.Handler {
	return vehicleHandler.NewHandler()
}
