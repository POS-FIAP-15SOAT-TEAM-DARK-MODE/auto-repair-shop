package db

import (
	"fmt"
	"log"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/docs/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // driver Postgres
)

func Connect(cfg *config.Config) *sqlx.DB {
	dataDb := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.PostgresUser,
		cfg.PostgresPassword,
		cfg.PostgresHost,
		cfg.PostgresPort,
		cfg.PostgresDB,
	)

	db, err := sqlx.Connect("postgres", dataDb)
	if err != nil {
		log.Fatal("Fail to connect to database", err)
	}

	return db
}
