package main

import (
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/docs/config"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/factory"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/env"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/routing"
	"go.uber.org/zap"
)

func main() {
	defer logger.Sync()

	cfg := config.LoadConfig()
	dbConn := db.Connect(cfg)

	container := factory.HttpContainer(dbConn)
	middlewares := factory.MiddlewaresContainer()
	router := routing.SetupRouter(container, middlewares)

	port := env.GetString("PORT", "8080")
	if err := router.Run(":" + port); err != nil {
		logger.Error("Fail to start application", zap.String("port", port), zap.Error(err))
	}

}
