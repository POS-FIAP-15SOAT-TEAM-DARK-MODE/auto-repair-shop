package service_order_history_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	postgresdb "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db/postgres"
	sohrepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/service_order_history"
	"github.com/stretchr/testify/assert"
)

var testMock sqlmock.Sqlmock

func TestMain(m *testing.M) {
	db, mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}
	testMock = mock
	defer func() { _ = db.Close() }()

	postgresdb.ConnectWithDB(db)
	os.Exit(m.Run())
}

func TestPostgresRepository_Search(t *testing.T) {
	tests := []struct {
		name        string
		params      *domain.SearchServiceOrderHistoryParams
		mockSetup   func()
		expectError bool
		expectLen   int
	}{
		{
			name:   "success without ID filter returns all rows",
			params: &domain.SearchServiceOrderHistoryParams{},
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"previous_status", "new_status", "created_at"}).
					AddRow(domain.SERVICE_ORDER_STATUS_NEW, domain.SERVICE_ORDER_STATUS_RECEIVED, time.Now()).
					AddRow(domain.SERVICE_ORDER_STATUS_RECEIVED, domain.SERVICE_ORDER_STATUS_IN_DIAGNOSIS, time.Now())
				testMock.ExpectQuery(`SELECT soh.previous_status, soh.new_status, soh.created_at FROM service_order_status_history soh`).
					WillReturnRows(rows)
			},
			expectError: false,
			expectLen:   2,
		},
		{
			name:   "success with ID filter binds the id arg",
			params: &domain.SearchServiceOrderHistoryParams{ID: "so-99"},
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"previous_status", "new_status", "created_at"}).
					AddRow(domain.SERVICE_ORDER_STATUS_NEW, domain.SERVICE_ORDER_STATUS_IN_PROGRESS, time.Now())
				testMock.ExpectQuery(`SELECT soh.previous_status, soh.new_status, soh.created_at FROM service_order_status_history soh`).
					WithArgs("so-99").
					WillReturnRows(rows)
			},
			expectError: false,
			expectLen:   1,
		},
		{
			name:   "query error",
			params: &domain.SearchServiceOrderHistoryParams{},
			mockSetup: func() {
				testMock.ExpectQuery(`SELECT soh.previous_status, soh.new_status, soh.created_at FROM service_order_status_history soh`).
					WillReturnError(errors.New("query failed"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := sohrepo.Repository()

			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			histories, err := repo.Search(context.Background(), tt.params)
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Len(t, histories, tt.expectLen)
		})
	}
}

func TestPostgresRepository_SearchWorkTransitions(t *testing.T) {
	now := time.Now()

	t.Run("groups rows by work_id preserving insertion order", func(t *testing.T) {
		repo := sohrepo.Repository()
		rows := sqlmock.NewRows([]string{"work_id", "previous_status", "new_status", "created_at"}).
			AddRow("w-1", nil, domain.SERVICE_ORDER_STATUS_NEW, now).
			AddRow("w-2", nil, domain.SERVICE_ORDER_STATUS_NEW, now).
			AddRow("w-1", domain.SERVICE_ORDER_STATUS_NEW, domain.SERVICE_ORDER_STATUS_IN_PROGRESS, now)
		testMock.ExpectQuery(`SELECT wsosh.work_id, wsosh.previous_status, wsosh.new_status, wsosh.created_at FROM work_service_order_status_history wsosh`).
			WithArgs("so-1").
			WillReturnRows(rows)

		groups, err := repo.SearchWorkTransitionsByServiceOrderID(context.Background(), "so-1")
		assert.NoError(t, err)
		assert.Len(t, groups, 2)

		assert.Equal(t, "w-1", groups[0].WorkID)
		assert.Len(t, groups[0].Status, 2)
		assert.Equal(t, domain.SERVICE_ORDER_STATUS_NEW, groups[0].Status[0].NewStatus)
		assert.Equal(t, domain.SERVICE_ORDER_STATUS_IN_PROGRESS, groups[0].Status[1].NewStatus)

		assert.Equal(t, "w-2", groups[1].WorkID)
		assert.Len(t, groups[1].Status, 1)
	})

	t.Run("empty result returns nil", func(t *testing.T) {
		repo := sohrepo.Repository()
		rows := sqlmock.NewRows([]string{"work_id", "previous_status", "new_status", "created_at"})
		testMock.ExpectQuery(`SELECT wsosh.work_id, wsosh.previous_status, wsosh.new_status, wsosh.created_at FROM work_service_order_status_history wsosh`).
			WithArgs("so-1").
			WillReturnRows(rows)

		groups, err := repo.SearchWorkTransitionsByServiceOrderID(context.Background(), "so-1")
		assert.NoError(t, err)
		assert.Nil(t, groups)
	})

	t.Run("query error", func(t *testing.T) {
		repo := sohrepo.Repository()
		testMock.ExpectQuery(`SELECT wsosh.work_id, wsosh.previous_status, wsosh.new_status, wsosh.created_at FROM work_service_order_status_history wsosh`).
			WillReturnError(errors.New("query failed"))

		_, err := repo.SearchWorkTransitionsByServiceOrderID(context.Background(), "so-1")
		assert.Error(t, err)
	})

	t.Run("scan error returns error", func(t *testing.T) {
		repo := sohrepo.Repository()
		rows := sqlmock.NewRows([]string{"work_id", "previous_status", "new_status", "created_at"}).
			AddRow("w-1", nil, "NEW", "not-a-time")
		testMock.ExpectQuery(`SELECT wsosh.work_id, wsosh.previous_status, wsosh.new_status, wsosh.created_at FROM work_service_order_status_history wsosh`).
			WithArgs("so-1").
			WillReturnRows(rows)

		_, err := repo.SearchWorkTransitionsByServiceOrderID(context.Background(), "so-1")
		assert.Error(t, err)
	})
}

func TestPostgresRepository_Search_ScanError(t *testing.T) {
	repo := sohrepo.Repository()
	rows := sqlmock.NewRows([]string{"previous_status", "new_status", "created_at"}).
		AddRow("NEW", "RECEIVED", "not-a-time")
	testMock.ExpectQuery(`SELECT soh.previous_status, soh.new_status, soh.created_at FROM service_order_status_history soh`).
		WillReturnRows(rows)

	_, err := repo.Search(context.Background(), &domain.SearchServiceOrderHistoryParams{})
	assert.Error(t, err)
}

func TestPostgresRepository_InsertWorkHistory(t *testing.T) {
	uow := postgresdb.NewTransactionalUoW(postgresdb.Connect())

	t.Run("success", func(t *testing.T) {
		testMock.ExpectBegin()
		testMock.ExpectExec("INSERT INTO work_service_order_status_history").
			WillReturnResult(sqlmock.NewResult(1, 1))
		testMock.ExpectCommit()

		repo := sohrepo.Repository()
		err := uow.Execute(context.Background(), func(ctx context.Context) error {
			return repo.InsertWorkHistory(ctx, "so-1", "w-1", domain.SERVICE_ORDER_STATUS_NEW)
		})

		assert.NoError(t, err)
		assert.NoError(t, testMock.ExpectationsWereMet())
	})

	t.Run("exec error", func(t *testing.T) {
		testMock.ExpectBegin()
		testMock.ExpectExec("INSERT INTO work_service_order_status_history").
			WillReturnError(errors.New("db error"))
		testMock.ExpectRollback()

		repo := sohrepo.Repository()
		err := uow.Execute(context.Background(), func(ctx context.Context) error {
			return repo.InsertWorkHistory(ctx, "so-1", "w-1", domain.SERVICE_ORDER_STATUS_NEW)
		})

		assert.Error(t, err)
	})
}

func TestPostgresRepository_InsertWorkHistory_NoTransaction(t *testing.T) {
	repo := sohrepo.Repository()
	err := repo.InsertWorkHistory(context.Background(), "so-1", "w-1", domain.SERVICE_ORDER_STATUS_NEW)
	assert.Error(t, err)
}
