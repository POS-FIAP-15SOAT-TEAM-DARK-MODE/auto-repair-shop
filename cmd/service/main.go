package main

import (
	"context"
	"time"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/infra/factory"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/pkg/app"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/pkg/logger"
)

func main() {
	defer logger.Global().Sync()

	httpServer := factory.HTTPServer()
	manager := app.NewLifecycleManager(5 * time.Second)
	manager.Add(httpServer)

	if err := manager.Run(context.Background()); err != nil {
		logger.Global().Fatal("application failed")
	}
}
