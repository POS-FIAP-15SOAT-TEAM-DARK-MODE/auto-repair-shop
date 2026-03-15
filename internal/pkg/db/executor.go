package db

import (
	"context"
	"database/sql"
)

// Executor is satisfied by both *sqlx.DB and *sqlx.Tx, allowing repositories
// to work transparently inside or outside a transaction.
type Executor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}
