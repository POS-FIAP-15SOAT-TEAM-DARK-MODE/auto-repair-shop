package customer

import (
	"context"
	"database/sql"
	"errors"

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
		mapped := pgPkg.Error(ctx, err)
		if !errors.Is(mapped, domain.ErrDataConflict) && !errors.Is(mapped, domain.ErrDataViolation) {
			logger.Of(ctx).Warn("customer repository: failed to create customer",
				zap.String("operation", "create_customer"),
				zap.String("entity_id", customer.ID),
				zap.Error(err),
			)
		}
		return mapped
	}

	return nil
}

func (r *repository) GetByID(ctx context.Context, id string) (domain.Customer, error) {
	db, err := postgres.GetOneTimeTransaction(ctx)
	if err != nil {
		return domain.Customer{}, err
	}

	var c domain.Customer
	c.User = &domain.User{}

	err = db.QueryRowContext(ctx, getCustomerByIDQuery, id).Scan(
		&c.ID, &c.UserID, &c.Type,
		&c.CPF, &c.CNPJ, &c.CompanyName, &c.Phone,
		&c.User.Name, &c.User.Email,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Customer{}, domain.ErrCustomerNotFound
		}
		logger.Of(ctx).Warn("customer repository: failed to get customer by id",
			zap.String("operation", "get_customer_by_id"),
			zap.String("entity_id", id),
			zap.Error(err),
		)
		return domain.Customer{}, pgPkg.Error(ctx, err)
	}

	c.User.ID = c.UserID
	return c, nil
}

func (r *repository) GetByDocument(ctx context.Context, document string) (domain.Customer, error) {
	db, err := postgres.GetOneTimeTransaction(ctx)
	if err != nil {
		return domain.Customer{}, err
	}

	var c domain.Customer
	c.User = &domain.User{}

	err = db.QueryRowContext(ctx, getCustomerByDocumentQuery, document).Scan(
		&c.ID, &c.UserID, &c.Type,
		&c.CPF, &c.CNPJ, &c.CompanyName, &c.Phone,
		&c.User.Name, &c.User.Email,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Customer{}, domain.ErrCustomerNotFound
		}
		logger.Of(ctx).Warn("customer repository: failed to get customer by document",
			zap.String("operation", "get_customer_by_document"),
			zap.Error(err),
		)
		return domain.Customer{}, pgPkg.Error(ctx, err)
	}

	c.User.ID = c.UserID
	return c, nil
}

func (r *repository) Update(ctx context.Context, id, phone string) error {
	tx, err := postgres.GetTransaction(ctx)
	if err != nil {
		return err
	}

	if _, err = tx.ExecContext(ctx, updateCustomerQuery, phone, id); err != nil {
		return pgPkg.Error(ctx, err)
	}

	return nil
}

func (r *repository) Delete(ctx context.Context, id string) error {
	tx, err := postgres.GetTransaction(ctx)
	if err != nil {
		return err
	}

	var count int
	if err = tx.QueryRowContext(ctx, countServiceOrdersByCustomerQuery, id).Scan(&count); err != nil {
		return pgPkg.Error(ctx, err)
	}
	if count > 0 {
		return domain.ErrCustomerHasServiceOrders
	}

	if _, err = tx.ExecContext(ctx, deleteCustomerQuery, id); err != nil {
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
