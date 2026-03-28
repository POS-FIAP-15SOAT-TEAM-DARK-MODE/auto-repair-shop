package customer_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	postgresdb "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db/postgres"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/customer"
	"github.com/stretchr/testify/assert"
)

func TestPostgresRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := customer.Repository()

	c, _ := domain.NewCustomer(
		"João Silva",
		"joao@example.com",
		"Secret@123",
		domain.IndividualCustomerType,
		"111.444.777-35",
		"",
		"11999999999",
	)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO customer").
		WithArgs(c.ID, c.UserID, c.Type, c.CPF, nil, nil, c.Phone).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		return repo.Create(ctx, &c)
	})

	assert.NoError(t, err)
}

func TestPostgresRepository_Create_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := customer.Repository()

	c, _ := domain.NewCustomer(
		"João Silva", "j@e.com", "Secret@123",
		domain.IndividualCustomerType, "11144477735", "", "11999999999",
	)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO customer").
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		return repo.Create(ctx, &c)
	})

	assert.Error(t, err)
}

func TestPostgresRepository_Create_MissingTx(t *testing.T) {
	repo := customer.Repository()
	c := &domain.Customer{}

	err := repo.Create(context.Background(), c)
	assert.Error(t, err)
}
