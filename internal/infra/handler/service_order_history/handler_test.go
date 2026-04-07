package service_order_history_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain/mocks"
	handler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/service_order_history"
)

func TestGetHistoryByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		serviceID      string
		mockSetup      func(m *mocks.ServiceOrderHistoryService)
		expectedCode   int
		expectedItems  int
		expectedErrMsg string
	}{
		{
			name:           "Should fail gracefully when the service order id is not provided",
			serviceID:      "",
			expectedCode:   http.StatusBadRequest,
			expectedItems:  0,
			expectedErrMsg: domain.ErrServiceOrderIDRequired.Error(),
		}, {
			name:      "Should fail gracefully when service returns an error",
			serviceID: "so-2",
			mockSetup: func(m *mocks.ServiceOrderHistoryService) {
				m.EXPECT().GetHistoryByID(mock.Anything, "so-2").Return(nil, domain.ErrDataConflict)
			},
			expectedCode:   http.StatusConflict,
			expectedItems:  0,
			expectedErrMsg: domain.ErrDataConflict.Error(),
		},
		{
			name:      "Should return the history when service returns it",
			serviceID: "so-1",
			mockSetup: func(m *mocks.ServiceOrderHistoryService) {
				m.EXPECT().GetHistoryByID(mock.Anything, "so-1").Return([]domain.ServiceOrderHistory{
					{
						ID:             "history-id-1",
						PreviousStatus: domain.SERVICE_ORDER_STATUS_RECEIVED,
						NewStatus:      domain.SERVICE_ORDER_STATUS_IN_DIAGNOSIS,
						CreatedAt:      time.Date(2026, 4, 5, 12, 34, 56, 0, time.UTC),
					},
				}, nil)
			},
			expectedCode:  http.StatusOK,
			expectedItems: 1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mSvc := mocks.NewServiceOrderHistoryService(t)
			if tt.mockSetup != nil {
				tt.mockSetup(mSvc)
			}

			h := handler.HttpHandler(mSvc)
			router := gin.New()
			router.GET("/v1/service-order/:id/history", h.GetHistoryByID)

			req := httptest.NewRequest(http.MethodGet, "/v1/service-order/"+tt.serviceID+"/history", nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != tt.expectedCode {
				t.Fatalf("expected status %d, got %d", tt.expectedCode, rec.Code)
			}

			var resp struct {
				Items  []domain.ServiceOrderHistory `json:"items,omitempty"`
				Code   int                          `json:"code,omitempty"`
				Errors []string                     `json:"errors,omitempty"`
				Error  string                       `json:"error,omitempty"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			assert.Equal(t, tt.expectedItems, len(resp.Items), "unexpected number of history items")
			if tt.expectedErrMsg != "" {
				// Accept error either in Errors array (bad request style) or in Error field (conflict/internal style)
				if len(resp.Errors) > 0 {
					assert.Contains(t, resp.Errors, tt.expectedErrMsg)
				} else {
					assert.Equal(t, tt.expectedErrMsg, resp.Error)
				}
				assert.Equal(t, tt.expectedCode, resp.Code)
			}
		})
	}
}
