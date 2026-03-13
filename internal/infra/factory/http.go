package factory

import (
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/container"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db"
	customerHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/customer"
	pingHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/ping"
	userHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/user"
	customerRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/customer"
	userRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/user"
	pkgdb "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/db"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/env"
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
		PostgresDB:       env.GetString("POSTGRES_DB", "autorepairshop"),
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
		PostgresDB:       env.GetString("POSTGRES_DB", "autorepairshop"),
	})

	transactor := db.NewTransactor(
		dbConn,
		func(ex pkgdb.Executor) domain.UserRepository {
			return userRepo.NewUserRepository(ex)
		},
		func(ex pkgdb.Executor) domain.CustomerRepository {
			return customerRepo.Repository(ex)
		},
	)

	customerService := customerSvc.Service(transactor)
	return customerHandler.NewHandler(customerService)
}
