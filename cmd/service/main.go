package main

import (
	"fmt"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/factory"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/env"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/routing"
	"go.uber.org/zap"
)

func main() {
	defer logger.Global().Sync()

	container := factory.HttpContainer()
	middlewares := factory.MiddlewaresContainer()
	router := routing.SetupRouter(container, middlewares)

	port := env.GetString("PORT", "8080")
	if err := router.Run(":" + port); err != nil {
		err = fmt.Errorf("fail to start application: %w", err)
		logger.Global().Error(err, zap.String("port", port))
	}
}
