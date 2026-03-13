package container

import (
	pingHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/ping"
	serviceHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/service"
)

type HTTP struct {
	PingHandler    *pingHandler.Handler
	ServiceHandler *serviceHandler.HttpHandler
}
