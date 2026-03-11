package db

import (
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // driver Postgres
)

func NewPostgresDB(dsn string) *sqlx.DB {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("falha ao conectar no Postgres: %v", err)
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)

	return db
}
