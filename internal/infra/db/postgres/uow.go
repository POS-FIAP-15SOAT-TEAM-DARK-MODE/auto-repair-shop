package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
)

type txKey struct{}

var ErrMissingPostgresTransaction = errors.New("transaction not found")

func NewTransactionalUoW(db *sql.DB) domain.Executor {
	return &domain.UnitOfWork{
		OnStart:   onStart(db),
		OnSuccess: onSuccess,
		OnFailure: onFailure,
	}
}

func onStart(db *sql.DB) func(context.Context) (context.Context, error) {
	return func(ctx context.Context) (context.Context, error) {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return ctx, fmt.Errorf("begin tx: %w", err)
		}
		return withTx(ctx, tx), nil
	}
}

func onSuccess(ctx context.Context) error {
	tx, found := txFrom(ctx)
	if !found {
		return ErrMissingPostgresTransaction
	}
	return tx.Commit()
}

func onFailure(ctx context.Context, cause error) error {
	tx, found := txFrom(ctx)
	if !found {
		return ErrMissingPostgresTransaction
	}
	return tx.Rollback()
}

func withTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func txFrom(ctx context.Context) (*sql.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(*sql.Tx)
	return tx, tx != nil && ok
}

func GetTransaction(ctx context.Context) (*sql.Tx, error) {
	tx, found := txFrom(ctx)
	if !found {
		return nil, ErrMissingPostgresTransaction
	}
	return tx, nil
}
