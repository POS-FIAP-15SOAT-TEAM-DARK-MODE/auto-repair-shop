package postgres

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/env"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	_ "github.com/lib/pq"
)

var (
	instance *sql.DB
	once     sync.Once
)

func Connect() *sql.DB {
	once.Do(func() {
		connStr := fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=disable",
			env.GetString("POSTGRES_USER", "postgres"),
			env.GetString("POSTGRES_PASSWORD", "postgres"),
			env.GetString("POSTGRES_HOST", "localhost"),
			env.GetString("POSTGRES_PORT", "5432"),
			env.GetString("POSTGRES_DB", "autorepairshop"),
		)

		conn, err := sql.Open("postgres", connStr)
		if err != nil {
			logger.Global().Fatal(fmt.Errorf("failed to connect to database: %w", err).Error())
		}

		//TODO: Define pool configuration

		instance = conn
	})

	return instance
}
