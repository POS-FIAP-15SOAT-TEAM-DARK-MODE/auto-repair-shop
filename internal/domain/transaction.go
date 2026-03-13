package domain

import "context"

// TxFunc is the unit of work executed inside a transaction.
// Repositories provided are scoped to the same transaction.
type TxFunc func(userRepo UserRepository, customerRepo CustomerRepository) error

// Transactor manages database transactions without leaking DB concerns into the domain.
//
//go:generate mockery --name=Transactor --with-expecter
type Transactor interface {
	WithTransaction(ctx context.Context, fn TxFunc) error
}
