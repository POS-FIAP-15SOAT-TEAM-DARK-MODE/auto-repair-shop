package service_order_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	postgresdb "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db/postgres"
	service_order "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/service_order"
	"github.com/stretchr/testify/assert"
)

func TestPostgresRepository_Save(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := service_order.Repository()

	so := buildServiceOrder()

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO service_order").
		WithArgs(so.ID, so.Customer.ID, so.Vehicle.ID, so.Status.String(), so.TotalAmount).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		return repo.Save(ctx, so)
	})

	assert.NoError(t, err)
}

func TestPostgresRepository_Save_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := service_order.Repository()

	so := buildServiceOrder()

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO service_order").WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		return repo.Save(ctx, so)
	})

	assert.Error(t, err)
}

func TestPostgresRepository_Save_NoTransaction(t *testing.T) {
	repo := service_order.Repository()

	so := buildServiceOrder()
	err := repo.Save(context.Background(), so)

	assert.Error(t, err)
}
