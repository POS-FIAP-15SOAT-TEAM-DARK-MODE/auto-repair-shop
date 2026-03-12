package factory

import (
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/container"
	pingHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/ping"
	userHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/user"
	userRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/user"
	pingSvc "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/ping"
	userSvc "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/user"
	"github.com/jmoiron/sqlx"
)

func HttpContainer(db *sqlx.DB) *container.HTTP {
	return &container.HTTP{
		PingHandler: newPingHandler(),
		UserHandler: newUserHandler(db),
	}
}

func newPingHandler() *pingHandler.Handler {
	pingService := pingSvc.Service()
	return pingHandler.NewHandler(pingService)
}

func newUserHandler(db *sqlx.DB) *userHandler.Handler {
	userRepository := userRepo.NewSqlxUserRepository(db)
	userService := userSvc.NewUserService(userRepository) // ponteiro
	return userHandler.NewHandler(*userService)           // passa ponteiro
}
