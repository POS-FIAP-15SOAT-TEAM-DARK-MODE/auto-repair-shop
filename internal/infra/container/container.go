package container

import (
	pingHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/ping"
	userHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/user"
)

type HTTP struct {
	PingHandler *pingHandler.Handler
	UserHandler *userHandler.Handler
}
