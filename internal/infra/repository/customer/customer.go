package customer

import (
	"context"

	"go.uber.org/zap"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	pkgdb "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/db"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/db/postgres"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
)

type repository struct {
	db pkgdb.Executor
}

func Repository(ex pkgdb.Executor) *repository {
	return &repository{db: ex}
}

func (r *repository) Create(ctx context.Context, customer *domain.Customer) error {
	exec := pkgdb.ExtractExecutor(ctx, r.db)

	_, err := exec.ExecContext(ctx, createCustomerQuery,
		customer.ID,
		customer.UserID,
		customer.Type,
		nullableString(customer.CPF),
		nullableString(customer.CNPJ),
		nullableString(customer.CompanyName),
		customer.Phone,
	)
	if err != nil {
		logger.Error("customer repository: failed to create customer",
			zap.String("operation", "create_customer"),
			zap.String("entity_id", customer.ID),
			zap.Error(err),
		)
		return postgres.Error(err)
	}

	return nil
}

// nullableString converts an empty string to nil so the DB receives NULL
// instead of an empty string, preserving CHECK and IS NULL constraints.
func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}
