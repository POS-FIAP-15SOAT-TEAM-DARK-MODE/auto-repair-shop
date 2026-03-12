package db

import (
	"fmt"
	"log"
	"sync"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Config struct {
	PostgresUser     string
	PostgresPassword string
	PostgresHost     string
	PostgresPort     string
	PostgresDB       string
}

var (
	db   *sqlx.DB
	once sync.Once
)

func Connect(cfg *Config) *sqlx.DB {
	once.Do(func() {
		db = connectDB(cfg)
	})

	return db
}

func connectDB(cfg *Config) *sqlx.DB {
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
