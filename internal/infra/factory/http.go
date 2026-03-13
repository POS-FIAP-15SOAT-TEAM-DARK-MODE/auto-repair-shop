package factory

import (
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/container"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db"
	customerHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/customer"
	pingHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/ping"
	userHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/user"
	customerRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/customer"
	userRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/user"
	customerSvc "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/customer"
	pingSvc "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/ping"
	userSvc "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/user"
	"github.com/jmoiron/sqlx"
)

func HttpContainer() *container.HTTP {
	dbConn := db.Connect()

	return &container.HTTP{
		PingHandler:     newPingHandler(),
		UserHandler:     newUserHandler(dbConn),
		CustomerHandler: newCustomerHandler(dbConn),
	}
}

func newPingHandler() *pingHandler.Handler {
	return pingHandler.NewHandler(pingSvc.Service())
}

func newUserHandler(dbConn *sqlx.DB) *userHandler.Handler {
	userRepository := userRepo.NewUserRepository(dbConn)
	return userHandler.NewHandler(userSvc.NewUserService(userRepository))
}

func newCustomerHandler(dbConn *sqlx.DB) *customerHandler.Handler {
	transactor := db.NewTransactor(dbConn)
	userRepository := userRepo.NewUserRepository(dbConn)
	customerRepository := customerRepo.Repository(dbConn)

	return customerHandler.NewHandler(customerSvc.Service(transactor, userRepository, customerRepository))
}
