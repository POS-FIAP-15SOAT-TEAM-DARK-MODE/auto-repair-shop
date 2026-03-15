package factory

import (
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/container"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db"
	pingHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/ping"
	userHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/user"
	userRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/user"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/env"
	pingSvc "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/ping"
	userSvc "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/user"
)

func HttpContainer() *container.HTTP {
	return &container.HTTP{
		PingHandler: newPingHandler(),
		UserHandler: newUserHandler(),
	}
}

func newPingHandler() *pingHandler.Handler {
	pingService := pingSvc.Service()
	return pingHandler.NewHandler(pingService)
}

func newUserHandler() *userHandler.Handler {
	dbConn := db.Connect(&db.Config{
		PostgresUser:     env.GetString("POSTGRES_USER", "postgres"),
		PostgresPassword: env.GetString("POSTGRES_PASSWORD", "postgres"),
		PostgresHost:     env.GetString("POSTGRES_HOST", "localhost"),
		PostgresPort:     env.GetString("POSTGRES_PORT", "5432"),
		PostgresDB:       env.GetString("POSTGRES_DB", "auto_repair_shop"),
	})
	userRepository := userRepo.NewSqlxUserRepository(dbConn)
	userService := userSvc.Service(userRepository)
	return userHandler.NewHandler(userService)
}
