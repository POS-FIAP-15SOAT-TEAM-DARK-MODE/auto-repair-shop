package factory

import (
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/container"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db"
	customerHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/customer"
	pingHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/ping"
	userHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/user"
	userRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/user"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/env"
	customerRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/repository/customer"
	customerSvc "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/customer"
	pingSvc "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/ping"
	userSvc "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/user"
)

func HttpContainer() *container.HTTP {
	return &container.HTTP{
		PingHandler:     newPingHandler(),
		UserHandler:     newUserHandler(),
		CustomerHandler: newCustomerHandler(),
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
	userRepository := userRepo.NewUserRepository(dbConn)
	userService := userSvc.NewUserService(userRepository)
	return userHandler.NewHandler(userService)
}

func newCustomerHandler() *customerHandler.Handler {
	dbConn := db.Connect(&db.Config{
		PostgresUser:     env.GetString("POSTGRES_USER", "postgres"),
		PostgresPassword: env.GetString("POSTGRES_PASSWORD", "postgres"),
		PostgresHost:     env.GetString("POSTGRES_HOST", "localhost"),
		PostgresPort:     env.GetString("POSTGRES_PORT", "5432"),
		PostgresDB:       env.GetString("POSTGRES_DB", "auto_repair_shop"),
	})

	userRepository := userRepo.NewUserRepository(dbConn)
	customerRepository := customerRepo.Repository(dbConn)
	customerService := customerSvc.Service(customerRepository, userRepository)
	return customerHandler.NewHandler(customerService)
}
