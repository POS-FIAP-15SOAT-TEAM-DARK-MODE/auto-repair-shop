package main

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/app/bootstrap"
	pkgApp "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/app"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
)

func main() {
	defer logger.Global().Sync()

	httpServer, err := bootstrap.HTTPServer()
	if err != nil {
		logger.Global().Fatal("invalid gateway configuration", zap.Error(err))
	}

	manager := pkgApp.NewLifecycleManager(5 * time.Second)
	manager.Add(httpServer)

	if err := manager.Run(context.Background()); err != nil {
		logger.Global().Fatal("application failed")
	}
}
