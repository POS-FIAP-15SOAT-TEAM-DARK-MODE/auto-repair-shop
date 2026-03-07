package container

import pingHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/ping"

type HTTP struct {
	PingHandler *pingHandler.Handler
}
