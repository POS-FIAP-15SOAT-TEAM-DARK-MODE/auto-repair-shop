package main

import (
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/factory"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/routing"
)

func main() {
	defer logger.Sync()

	container := factory.HttpContainer()
	middlewares := factory.MiddlewaresContainer()
	router := routing.SetupRouter(container, middlewares)
	router.Run(":8080")
}
