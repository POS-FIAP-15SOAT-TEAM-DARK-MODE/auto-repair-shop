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
	domainmocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain/mocks"
	handler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/service_order_history"
)

func TestGetHistoryByID_Handler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	expHistory := domain.ServiceOrderHistory{
		ID:             "history-id-1",
		PreviousStatus: domain.SERVICE_ORDER_STATUS_RECEIVED,
		NewStatus:      domain.SERVICE_ORDER_STATUS_IN_DIAGNOSIS,
		CreatedAt:      time.Date(2026, 4, 5, 12, 34, 56, 0, time.UTC),
	}

	tests := []struct {
		name         string
		setupMock    func(m *domainmocks.ServiceOrderHistoryService)
		route        string
		expectedCode int
		expectedLen  int
		expectedErr  string
	}{
		{
			name:         "missing id -> bad request",
			setupMock:    nil,
			route:        "/v1/service-order//history",
			expectedCode: http.StatusBadRequest,
			expectedLen:  0,
			expectedErr:  domain.ErrServiceOrderIDRequired.Error(),
		},
		{
			name: "service error -> conflict",
			setupMock: func(m *domainmocks.ServiceOrderHistoryService) {
				m.EXPECT().GetHistoryByID(mock.Anything, mock.MatchedBy(func(p *domain.SearchServiceOrderHistoryParams) bool {
					return p != nil && p.ID == "so-2"
				})).Return(nil, domain.ErrDataConflict)
			},
			route:        "/v1/service-order/so-2/history",
			expectedCode: http.StatusConflict,
			expectedLen:  0,
			expectedErr:  domain.ErrDataConflict.Error(),
		},
		{
			name: "success",
			setupMock: func(m *domainmocks.ServiceOrderHistoryService) {
				m.EXPECT().GetHistoryByID(mock.Anything, mock.MatchedBy(func(p *domain.SearchServiceOrderHistoryParams) bool {
					return p != nil && p.ID == "so-1"
				})).Return(&domain.PaginatorResponse[domain.ServiceOrderHistory]{
					Items:      []domain.ServiceOrderHistory{expHistory},
					TotalItems: 1,
					TotalPages: 1,
					PageSize:   10,
					Page:       1,
				}, nil)
			},
			route:        "/v1/service-order/so-1/history",
			expectedCode: http.StatusOK,
			expectedLen:  1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mSvc := domainmocks.NewServiceOrderHistoryService(t)
			if tt.setupMock != nil {
				tt.setupMock(mSvc)
			}

			h := handler.HttpHandler(mSvc)
			router := gin.New()
			router.GET("/v1/service-order/:id/history", h.GetHistoryByID)

			req := httptest.NewRequest(http.MethodGet, tt.route, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code)

			var resp map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			if tt.expectedLen > 0 {
				items, ok := resp["items"].([]interface{})
				if !ok {
					t.Fatalf("expected items array in response")
				}
				assert.Len(t, items, tt.expectedLen)
			}

			if tt.expectedErr != "" {
				// errors may be in "errors" array or "error" string depending on status code
				if errs, ok := resp["errors"].([]interface{}); ok && len(errs) > 0 {
					assert.Contains(t, errs[0], tt.expectedErr)
				} else if errStr, ok := resp["error"].(string); ok {
					assert.Equal(t, tt.expectedErr, errStr)
				} else {
					t.Fatalf("expected error message in response")
				}
			}
		})
	}
}
