package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	soDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order_history/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubOneTimeTransaction swaps getOneTimeTransaction for the duration of the
// test so the repository queries a sqlmock DB instead of a real Postgres.
func stubOneTimeTransaction(t *testing.T, db *sql.DB) {
	t.Helper()
	orig := getOneTimeTransaction
	getOneTimeTransaction = func(_ context.Context) (*sql.DB, error) { return db, nil }
	t.Cleanup(func() { getOneTimeTransaction = orig })
}

func TestPostgres_AverageDurationByStatusInHours(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	stubOneTimeTransaction(t, db)

	rows := sqlmock.NewRows([]string{"new_status", "avg_hours"}).
		AddRow("IN_DIAGNOSIS", 2.5).
		AddRow("COMPLETED", 10.0)
	mock.ExpectQuery("service_order_status_history").WillReturnRows(rows)

	repo := NewMetricsRepository()
	result, err := repo.AverageDurationByStatusInHours(context.Background())
	require.NoError(t, err)

	assert.Equal(t, []domain.StatusDuration{
		{Status: soDomain.SERVICE_ORDER_STATUS_IN_DIAGNOSIS, AverageHours: 2.5},
		{Status: soDomain.SERVICE_ORDER_STATUS_COMPLETED, AverageHours: 10.0},
	}, result)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgres_AverageDurationByStatusInHours_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	stubOneTimeTransaction(t, db)

	mock.ExpectQuery("service_order_status_history").WillReturnError(errors.New("query boom"))

	repo := NewMetricsRepository()
	_, err = repo.AverageDurationByStatusInHours(context.Background())
	assert.Error(t, err)
}

func TestPostgres_AverageDurationByStatusInHours_NoTransaction(t *testing.T) {
	orig := getOneTimeTransaction
	getOneTimeTransaction = func(_ context.Context) (*sql.DB, error) { return nil, errors.New("no db") }
	t.Cleanup(func() { getOneTimeTransaction = orig })

	repo := NewMetricsRepository()
	_, err := repo.AverageDurationByStatusInHours(context.Background())
	assert.Error(t, err)
}
