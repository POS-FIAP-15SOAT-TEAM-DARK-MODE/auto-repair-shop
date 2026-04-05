package service_order_test

import (
	"context"
	"testing"
	"time"

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

func TestPostgresRepository_GetHistoryByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	// wire the global connector to our mock
	postgresdb.ConnectWithDB(db)

	repo := service_order.Repository()

	rows := sqlmock.NewRows([]string{"id", "previous_status", "new_status", "created_at"}).
		AddRow("h-1", "RECEIVED", "IN_DIAGNOSIS", time.Date(2026, 4, 5, 12, 34, 56, 0, time.UTC))

	mock.ExpectQuery("SELECT id, previous_status, new_status, created_at FROM service_order_status_history").
		WithArgs("so-1").
		WillReturnRows(rows)

	hist, err := repo.GetHistoryByID(context.Background(), "so-1")
	assert.NoError(t, err)
	assert.Len(t, hist, 1)
	assert.Equal(t, "h-1", hist[0].ID)
}

func TestPostgresRepository_GetHistoryByID_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	postgresdb.ConnectWithDB(db)

	repo := service_order.Repository()

	mock.ExpectQuery("SELECT id, previous_status, new_status, created_at FROM service_order_status_history").
		WithArgs("so-2").
		WillReturnError(assert.AnError)

	hist, err := repo.GetHistoryByID(context.Background(), "so-2")
	assert.Error(t, err)
	assert.Nil(t, hist)
}
