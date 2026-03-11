package factory

import (
	"fmt"
	"log"
	"os"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/container"
	pingHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/ping"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/user"
	userHandler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/user"
	userRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository"
	pingSvc "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/ping"
	userSvc "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/user"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func HttpContainer() *container.HTTP {

	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found: %v", err)
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"),
	)
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatal("Erro ao conectar no banco:", err)
	}
	return &container.HTTP{
		PingHandler: newPingHandler(),
		UserHandler: newUserHandler(db),
	}
}

func newPingHandler() *pingHandler.Handler {
	pingService := pingSvc.Service()
	return pingHandler.NewHandler(pingService)
}

func newUserHandler(db *sqlx.DB) *user.Handler {
	userRepository := userRepo.NewSqlxUserRepository(db)
	userService := userSvc.NewUserService(userRepository)
	return userHandler.NewHandler(*userService)
}
