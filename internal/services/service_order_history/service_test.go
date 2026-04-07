package service_order_history_test

import (
	"context"
	"testing"
	"time"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain/mocks"
	uow "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow"
	service "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/service_order_history"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_GetHistoryByID_Success(t *testing.T) {
	// table-driven tests for GetHistoryByID
	expSuccess := []domain.ServiceOrderHistory{
		{
			ID:             "h-1",
			PreviousStatus: domain.SERVICE_ORDER_STATUS_RECEIVED,
			NewStatus:      domain.SERVICE_ORDER_STATUS_IN_DIAGNOSIS,
			CreatedAt:      time.Date(2026, 4, 5, 12, 34, 56, 0, time.UTC),
		},
	}

	tests := []struct {
		name        string
		serviceID   string
		mockSetup   func(m *mocks.ServiceOrderHistoryRepository)
		expected    []domain.ServiceOrderHistory
		expectedErr error
	}{
		{
			name:      "success",
			serviceID: "so-1",
			mockSetup: func(m *mocks.ServiceOrderHistoryRepository) {
				m.EXPECT().Find(mock.Anything, domain.FindServiceOrderHistoryParams{ID: "so-1"}).Return(expSuccess, nil)
			},
			expected:    expSuccess,
			expectedErr: nil,
		},
		{
			name:      "repository error",
			serviceID: "so-2",
			mockSetup: func(m *mocks.ServiceOrderHistoryRepository) {
				m.EXPECT().Find(mock.Anything, domain.FindServiceOrderHistoryParams{ID: "so-2"}).Return(nil, domain.ErrInfraConflict)
			},
			expected:    nil,
			expectedErr: domain.ErrInfraConflict,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewServiceOrderHistoryRepository(t)
			tt.mockSetup(repo)

			s := service.Service(repo)
			history, err := s.GetHistoryByID(context.Background(), tt.serviceID)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, history)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, history)
			}
		})
	}
}

func runAllSteps() func(context.Context, ...uow.Step) error {
	return func(ctx context.Context, steps ...uow.Step) error {
		for _, step := range steps {
			if err := step(ctx); err != nil {
				return err
			}
		}
		return nil
	}
}
