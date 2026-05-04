package work_service_order_history_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	infraPostgres "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db/postgres"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/work_service_order_history"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupWSOMockDB(t *testing.T) (sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	infraPostgres.ConnectWithDB(db)
	return mock, func() { _ = db.Close() }
}

func wsoHistoryColumns() []string {
	return []string{"id", "work_id", "service_order_id", "previous_status", "new_status", "created_at"}
}

func TestPostgresWSO_Repository(t *testing.T) {
	r := work_service_order_history.Repository()
	assert.NotNil(t, r)
}

func TestPostgresWSO_Insert_Success(t *testing.T) {
	mock, close := setupWSOMockDB(t)
	defer close()

	from := domain.WORK_SERVICE_ORDER_STATUS_AWAITING_START
	uow := infraPostgres.NewTransactionalUoW(infraPostgres.Connect())

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO work_service_order_status_history").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	r := work_service_order_history.Repository()
	err := uow.Execute(context.Background(), func(ctx context.Context) error {
		return r.Insert(ctx, "so-1", "w-1", &from, domain.WORK_SERVICE_ORDER_STATUS_IN_PROGRESS)
	})

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresWSO_Insert_NilPreviousStatus(t *testing.T) {
	mock, close := setupWSOMockDB(t)
	defer close()

	uow := infraPostgres.NewTransactionalUoW(infraPostgres.Connect())

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO work_service_order_status_history").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	r := work_service_order_history.Repository()
	err := uow.Execute(context.Background(), func(ctx context.Context) error {
		return r.Insert(ctx, "so-1", "w-1", nil, domain.WORK_SERVICE_ORDER_STATUS_AWAITING_START)
	})

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresWSO_Insert_Error(t *testing.T) {
	mock, close := setupWSOMockDB(t)
	defer close()

	uow := infraPostgres.NewTransactionalUoW(infraPostgres.Connect())

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO work_service_order_status_history").
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	r := work_service_order_history.Repository()
	err := uow.Execute(context.Background(), func(ctx context.Context) error {
		return r.Insert(ctx, "so-1", "w-1", nil, domain.WORK_SERVICE_ORDER_STATUS_AWAITING_START)
	})

	assert.Error(t, err)
}

func TestPostgresWSO_Insert_NoTransaction(t *testing.T) {
	r := work_service_order_history.Repository()
	err := r.Insert(context.Background(), "so-1", "w-1", nil, domain.WORK_SERVICE_ORDER_STATUS_AWAITING_START)
	assert.Error(t, err)
}

func TestPostgresWSO_Search_Success(t *testing.T) {
	mock, close := setupWSOMockDB(t)
	defer close()

	now := time.Now()
	mock.ExpectQuery("SELECT wsosh.id").
		WillReturnRows(
			sqlmock.NewRows(wsoHistoryColumns()).
				AddRow("h-1", "w-1", "so-1", "AWAITING_START", "IN_PROGRESS", now),
		)

	r := work_service_order_history.Repository()
	results, err := r.Search(context.Background(), domain.SearchWorkSOHistoryParams{ServiceOrderID: "so-1"})

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "h-1", results[0].ID)
	assert.NotNil(t, results[0].PreviousStatus)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresWSO_Search_NilPreviousStatus(t *testing.T) {
	mock, close := setupWSOMockDB(t)
	defer close()

	now := time.Now()
	mock.ExpectQuery("SELECT wsosh.id").
		WillReturnRows(
			sqlmock.NewRows(wsoHistoryColumns()).
				AddRow("h-1", "w-1", "so-1", nil, "AWAITING_START", now),
		)

	r := work_service_order_history.Repository()
	results, err := r.Search(context.Background(), domain.SearchWorkSOHistoryParams{})

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Nil(t, results[0].PreviousStatus)
}

func TestPostgresWSO_Search_WithWorkID(t *testing.T) {
	mock, close := setupWSOMockDB(t)
	defer close()

	mock.ExpectQuery("SELECT wsosh.id").
		WillReturnRows(sqlmock.NewRows(wsoHistoryColumns()))

	r := work_service_order_history.Repository()
	results, err := r.Search(context.Background(), domain.SearchWorkSOHistoryParams{WorkID: "w-1"})

	assert.NoError(t, err)
	assert.Empty(t, results)
}

func TestPostgresWSO_Search_QueryError(t *testing.T) {
	mock, close := setupWSOMockDB(t)
	defer close()

	mock.ExpectQuery("SELECT wsosh.id").
		WillReturnError(errors.New("db error"))

	r := work_service_order_history.Repository()
	_, err := r.Search(context.Background(), domain.SearchWorkSOHistoryParams{ServiceOrderID: "so-1"})

	assert.Error(t, err)
}

func TestPostgresWSO_Search_ScanError(t *testing.T) {
	mock, close := setupWSOMockDB(t)
	defer close()

	mock.ExpectQuery("SELECT wsosh.id").
		WillReturnRows(
			sqlmock.NewRows(wsoHistoryColumns()).
				AddRow("h-1", "w-1", "so-1", "PREV", "NEW", "not-a-time"),
		)

	r := work_service_order_history.Repository()
	_, err := r.Search(context.Background(), domain.SearchWorkSOHistoryParams{})

	assert.Error(t, err)
}
