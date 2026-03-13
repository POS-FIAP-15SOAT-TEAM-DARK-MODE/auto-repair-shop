package db

import "context"

type txKey struct{}

// WithTx stores an Executor (a *sqlx.Tx) in the context so repositories can
// retrieve it transparently without the caller knowing about sqlx.
func WithTx(ctx context.Context, exec Executor) context.Context {
	return context.WithValue(ctx, txKey{}, exec)
}

// ExtractExecutor returns the Executor stored in ctx by WithTx.
// Falls back to the provided default when no transaction is active.
func ExtractExecutor(ctx context.Context, fallback Executor) Executor {
	if exec, ok := ctx.Value(txKey{}).(Executor); ok {
		return exec
	}
	return fallback
}
