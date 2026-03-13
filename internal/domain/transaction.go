package domain

import "context"

// TxFunc is the unit of work executed inside a transaction.
// The ctx carries the active *sqlx.Tx so repositories extract it transparently.
type TxFunc func(ctx context.Context) error

// Transactor manages database transactions without leaking DB concerns into the domain.
//
//go:generate mockery --name=Transactor --with-expecter
type Transactor interface {
	WithTransaction(ctx context.Context, fn TxFunc) error
}
