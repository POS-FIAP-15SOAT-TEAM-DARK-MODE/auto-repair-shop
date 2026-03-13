package db

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	pkgdb "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/db"
	"github.com/jmoiron/sqlx"
)

type transactor struct {
	db              *sqlx.DB
	newUserRepo     func(pkgdb.Executor) domain.UserRepository
	newCustomerRepo func(pkgdb.Executor) domain.CustomerRepository
}

func NewTransactor(
	db *sqlx.DB,
	newUserRepo func(pkgdb.Executor) domain.UserRepository,
	newCustomerRepo func(pkgdb.Executor) domain.CustomerRepository,
) domain.Transactor {
	return &transactor{
		db:              db,
		newUserRepo:     newUserRepo,
		newCustomerRepo: newCustomerRepo,
	}
}

func (t *transactor) WithTransaction(ctx context.Context, fn domain.TxFunc) error {
	tx, err := t.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(t.newUserRepo(tx), t.newCustomerRepo(tx)); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
