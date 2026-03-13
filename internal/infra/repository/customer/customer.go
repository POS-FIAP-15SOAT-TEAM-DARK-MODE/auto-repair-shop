package customer

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	pkgdb "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/db"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/db/postgres"
)

type repository struct {
	db pkgdb.Executor
}

func Repository(ex pkgdb.Executor) *repository {
	return &repository{db: ex}
}

func (r *repository) Create(ctx context.Context, customer *domain.Customer) error {
	exec := pkgdb.ExtractExecutor(ctx, r.db)

	stmt, err := exec.PrepareContext(ctx, createCustomerQuery)
	if err != nil {
		return err
	}
	defer func() { _ = stmt.Close() }()

	_, err = stmt.Exec(
		customer.ID,
		customer.UserID,
		customer.Type,
		nullableString(customer.CPF),
		nullableString(customer.CNPJ),
		nullableString(customer.CompanyName),
		customer.Phone,
	)
	if err != nil {
		return postgres.Error(err)
	}

	return nil
}

// nullableString converts an empty string to nil so the DB receives NULL
// instead of an empty string, preserving CHECK and IS NULL constraints.
func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
