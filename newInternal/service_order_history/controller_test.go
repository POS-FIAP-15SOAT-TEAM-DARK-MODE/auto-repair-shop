package service_order_history_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	soDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/service_order/domain"
	soHistory "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/service_order_history"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/service_order_history/domain"
	soHistoryMocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/service_order_history/interfaces/mocks"
)

func TestController_GetHistoryByID(t *testing.T) {
	baseTime := time.Date(2026, 4, 5, 12, 34, 56, 0, time.UTC)

	inProgressItem := domain.ServiceOrderHistoryItem{
		PreviousStatus: soDomain.SERVICE_ORDER_STATUS_AWAITING_APPROVAL,
		NewStatus:      soDomain.SERVICE_ORDER_STATUS_IN_PROGRESS,
		CreatedAt:      baseTime,
		WorkTransitions: []domain.WorkTransitionGroup{
			{
				WorkID: "work-1",
				Status: []domain.WorkStatusItem{
					{
						NewStatus: soDomain.SERVICE_ORDER_STATUS_NEW,
						CreatedAt: baseTime,
					},
					{
						PreviousStatus: soDomain.SERVICE_ORDER_STATUS_NEW,
						NewStatus:      soDomain.SERVICE_ORDER_STATUS_IN_PROGRESS,
						CreatedAt:      baseTime.Add(time.Hour),
					},
				},
			},
		},
	}

	tests := []struct {
		name          string
		ctxID         string
		setupMock     func(m *soHistoryMocks.ServiceOrderHistoryService)
		expectedLen   int
		expectedError error
	}{
		{
			name:          "missing id returns ErrServiceOrderIDRequired",
			ctxID:         "",
			setupMock:     nil,
			expectedLen:   0,
			expectedError: domain.ErrServiceOrderIDRequired,
		},
		{
			name:  "service error propagates",
			ctxID: "so-2",
			setupMock: func(m *soHistoryMocks.ServiceOrderHistoryService) {
				m.EXPECT().GetHistoryByID(mock.Anything, mock.MatchedBy(func(p *domain.SearchParams) bool {
					return p.ID == "so-2"
				})).Return(nil, errors.New("db error"))
			},
			expectedLen:   0,
			expectedError: errors.New("db error"),
		},
		{
			name:  "success returns mapped response list",
			ctxID: "so-1",
			setupMock: func(m *soHistoryMocks.ServiceOrderHistoryService) {
				m.EXPECT().GetHistoryByID(mock.Anything, mock.MatchedBy(func(p *domain.SearchParams) bool {
					return p.ID == "so-1"
				})).Return([]domain.ServiceOrderHistoryItem{inProgressItem}, nil)
			},
			expectedLen:   1,
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := soHistoryMocks.NewServiceOrderHistoryService(t)
			if tt.setupMock != nil {
				tt.setupMock(mockSvc)
			}

			ctrl := soHistory.NewController(mockSvc)

			ctx := context.WithValue(context.Background(), "id", tt.ctxID) //nolint:staticcheck
			req := httptest.NewRequest(http.MethodGet, "/service-order/"+tt.ctxID+"/history", nil)

			resp, err := ctrl.GetHistoryByID(ctx, req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.Len(t, resp, tt.expectedLen)

				if tt.expectedLen > 0 {
					first := resp[0]
					assert.Equal(t, "AWAITING_APPROVAL", first.PreviousStatus)
					assert.Equal(t, "IN_PROGRESS", first.NewStatus)
					assert.Len(t, first.WorkTransitions, 1)
					assert.Equal(t, "work-1", first.WorkTransitions[0].WorkID)
					assert.Len(t, first.WorkTransitions[0].Status, 2)
					assert.Equal(t, "IN_PROGRESS", first.WorkTransitions[0].Status[1].NewStatus)
				}
			}
		})
	}
}
