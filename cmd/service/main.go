package main

import (
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/factory"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/env"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/routing"
	"go.uber.org/zap"
)

func main() {
	defer logger.Sync()

	container := factory.HttpContainer()
	router := routing.SetupRouter(container)

	port := env.GetString("PORT", "8080")
	if err := router.Run(":" + port); err != nil {
		logger.Error("Fail to start application", zap.String("port", port), zap.Error(err))
	}
}
