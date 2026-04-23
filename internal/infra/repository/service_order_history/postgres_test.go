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
		name          string
		params        *domain.SearchServiceOrderHistoryParams
		mockSetup     func()
		expectError   bool
		expectLen     int
		expectOrderID string
	}{
		{
			name:   "success without ID filter",
			params: &domain.SearchServiceOrderHistoryParams{Page: 1, PageSize: 10},
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"id", "service_order_id", "previous_status", "new_status", "created_at"}).
					AddRow("hist-1", "so-1", domain.SERVICE_ORDER_STATUS_NEW, domain.SERVICE_ORDER_STATUS_RECEIVED, time.Now()).
					AddRow("hist-2", "so-1", domain.SERVICE_ORDER_STATUS_RECEIVED, domain.SERVICE_ORDER_STATUS_IN_DIAGNOSIS, time.Now())
				testMock.ExpectQuery(`SELECT soh.id, soh.service_order_id, soh.previous_status, soh.new_status, soh.created_at FROM service_order_status_history soh`).
					WillReturnRows(rows)
			},
			expectError: false,
			expectLen:   2,
		},
		{
			name:   "success with ID filter",
			params: &domain.SearchServiceOrderHistoryParams{ID: "so-99", Page: 1, PageSize: 10},
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"id", "service_order_id", "previous_status", "new_status", "created_at"}).
					AddRow("hist-3", "so-99", domain.SERVICE_ORDER_STATUS_NEW, domain.SERVICE_ORDER_STATUS_IN_PROGRESS, time.Now())
				testMock.ExpectQuery(`SELECT soh.id, soh.service_order_id, soh.previous_status, soh.new_status, soh.created_at FROM service_order_status_history soh`).
					WithArgs("so-99", int64(10), int64(1)).
					WillReturnRows(rows)
			},
			expectError:   false,
			expectLen:     1,
			expectOrderID: "so-99",
		},
		{
			name:   "query error",
			params: &domain.SearchServiceOrderHistoryParams{Page: 1, PageSize: 10},
			mockSetup: func() {
				testMock.ExpectQuery(`SELECT soh.id, soh.service_order_id, soh.previous_status, soh.new_status, soh.created_at FROM service_order_status_history soh`).
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
			if tt.expectOrderID != "" {
				assert.Equal(t, tt.expectOrderID, histories[0].ServiceOrderID)
			}
		})
	}
}

func TestPostgresRepository_Count(t *testing.T) {
	tests := []struct {
		name        string
		params      *domain.SearchServiceOrderHistoryParams
		mockSetup   func()
		expectError bool
		expectCount int64
	}{
		{
			name:   "success without ID filter",
			params: &domain.SearchServiceOrderHistoryParams{},
			mockSetup: func() {
				testMock.ExpectQuery(`SELECT COUNT\(soh.id\) FROM service_order_status_history soh`).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
			},
			expectError: false,
			expectCount: 5,
		},
		{
			name:   "success with ID filter",
			params: &domain.SearchServiceOrderHistoryParams{ID: "so-42"},
			mockSetup: func() {
				testMock.ExpectQuery(`SELECT COUNT\(soh.id\) FROM service_order_status_history soh`).
					WithArgs("so-42").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
			},
			expectError: false,
			expectCount: 3,
		},
		{
			name:   "query error",
			params: &domain.SearchServiceOrderHistoryParams{},
			mockSetup: func() {
				testMock.ExpectQuery(`SELECT COUNT\(soh.id\) FROM service_order_status_history soh`).
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

			count, err := repo.Count(context.Background(), tt.params)
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expectCount, count)
		})
	}
}
