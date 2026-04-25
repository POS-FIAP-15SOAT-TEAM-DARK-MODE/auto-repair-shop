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
			name:   "success without ID filter returns all rows",
			params: &domain.SearchServiceOrderHistoryParams{},
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
			name:   "success with ID filter binds the id arg",
			params: &domain.SearchServiceOrderHistoryParams{ID: "so-99"},
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"id", "service_order_id", "previous_status", "new_status", "created_at"}).
					AddRow("hist-3", "so-99", domain.SERVICE_ORDER_STATUS_NEW, domain.SERVICE_ORDER_STATUS_IN_PROGRESS, time.Now())
				testMock.ExpectQuery(`SELECT soh.id, soh.service_order_id, soh.previous_status, soh.new_status, soh.created_at FROM service_order_status_history soh`).
					WithArgs("so-99").
					WillReturnRows(rows)
			},
			expectError:   false,
			expectLen:     1,
			expectOrderID: "so-99",
		},
		{
			name:   "query error",
			params: &domain.SearchServiceOrderHistoryParams{},
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

func TestPostgresRepository_SearchWorkTransitions(t *testing.T) {
	tests := []struct {
		name        string
		serviceID   string
		mockSetup   func()
		expectError bool
		expectLen   int
	}{
		{
			name:      "success returns all rows for the service order",
			serviceID: "so-1",
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"work_id", "previous_status", "new_status", "created_at"}).
					AddRow("w-1", domain.SERVICE_ORDER_STATUS_AWAITING_APPROVAL, domain.SERVICE_ORDER_STATUS_IN_PROGRESS, time.Now()).
					AddRow("w-2", domain.SERVICE_ORDER_STATUS_AWAITING_APPROVAL, domain.SERVICE_ORDER_STATUS_IN_PROGRESS, time.Now())
				testMock.ExpectQuery(`SELECT wsosh.work_id, wsosh.previous_status, wsosh.new_status, wsosh.created_at FROM work_service_order_status_history wsosh`).
					WithArgs("so-1").
					WillReturnRows(rows)
			},
			expectError: false,
			expectLen:   2,
		},
		{
			name:      "query error",
			serviceID: "so-1",
			mockSetup: func() {
				testMock.ExpectQuery(`SELECT wsosh.work_id, wsosh.previous_status, wsosh.new_status, wsosh.created_at FROM work_service_order_status_history wsosh`).
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

			items, err := repo.SearchWorkTransitions(context.Background(), tt.serviceID)
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Len(t, items, tt.expectLen)
		})
	}
}
