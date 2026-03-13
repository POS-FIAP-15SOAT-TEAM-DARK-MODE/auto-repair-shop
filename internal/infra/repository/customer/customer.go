package customer

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	pkgdb "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/db"
)

type repository struct {
	db pkgdb.Executor
}

func Repository(ex pkgdb.Executor) *repository {
	return &repository{db: ex}
}

func (r *repository) Create(customer *domain.Customer) error {
	stmt, err := r.db.PrepareContext(context.Background(), createCustomerQuery)
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
		return pkgdb.Error(err)
	}

	return nil
}

func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
