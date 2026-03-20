package customer

import (
	"context"
	"go.uber.org/zap"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db/postgres"
	pgPkg "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/db/postgres"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
)

type repository struct{}

func Repository() *repository {
	return &repository{}
}

func (r *repository) Create(ctx context.Context, customer *domain.Customer) error {
	tx, err := postgres.GetTransaction(ctx)
	if err != nil {
		return err
	}

	logger.Of(ctx).Debug("Executing query", zap.String("query", createCustomerQuery), zap.Any("params", customer))
	_, err = tx.ExecContext(ctx, createCustomerQuery,
		customer.ID,
		customer.UserID,
		customer.Type,
		nullableString(customer.CPF),
		nullableString(customer.CNPJ),
		nullableString(customer.CompanyName),
		customer.Phone,
	)
	if err != nil {
		logger.Of(ctx).Warn("customer repository: failed to create customer",
			zap.String("operation", "create_customer"),
			zap.String("entity_id", customer.ID),
			zap.Error(err),
		)
		return pgPkg.Error(ctx, err)
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
