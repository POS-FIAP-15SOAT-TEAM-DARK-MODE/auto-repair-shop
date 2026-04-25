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

	inProgressHistory := domain.ServiceOrderHistoryItem{
		PreviousStatus: domain.SERVICE_ORDER_STATUS_AWAITING_APPROVAL,
		NewStatus:      domain.SERVICE_ORDER_STATUS_IN_PROGRESS,
		CreatedAt:      time.Date(2026, 4, 5, 12, 34, 56, 0, time.UTC),
		WorkTransitions: []domain.WorkTransitionGroup{
			{
				WorkID: "work-1",
				Status: []domain.WorkStatusItem{
					{
						PreviousStatus: "",
						NewStatus:      domain.SERVICE_ORDER_STATUS_NEW,
						CreatedAt:      time.Date(2026, 4, 5, 12, 34, 56, 0, time.UTC),
					},
					{
						PreviousStatus: domain.SERVICE_ORDER_STATUS_NEW,
						NewStatus:      domain.SERVICE_ORDER_STATUS_IN_PROGRESS,
						CreatedAt:      time.Date(2026, 4, 5, 13, 0, 0, 0, time.UTC),
					},
				},
			},
		},
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
			name: "success with work transitions grouped under IN_PROGRESS",
			setupMock: func(m *domainmocks.ServiceOrderHistoryService) {
				m.EXPECT().GetHistoryByID(mock.Anything, mock.MatchedBy(func(p *domain.SearchServiceOrderHistoryParams) bool {
					return p != nil && p.ID == "so-1"
				})).Return([]domain.ServiceOrderHistoryItem{inProgressHistory}, nil)
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

			if tt.expectedLen > 0 {
				var items []map[string]any
				if err := json.Unmarshal(rec.Body.Bytes(), &items); err != nil {
					t.Fatalf("expected JSON array response: %v", err)
				}
				assert.Len(t, items, tt.expectedLen)

				first := items[0]
				workTransitions, ok := first["work_transitions"].([]interface{})
				if !ok {
					t.Fatalf("expected work_transitions array, got: %T", first["work_transitions"])
				}
				assert.Len(t, workTransitions, 1)

				group, ok := workTransitions[0].(map[string]interface{})
				if !ok {
					t.Fatalf("expected work_transitions[0] to be an object")
				}
				assert.Equal(t, "work-1", group["work_id"])

				statuses, ok := group["status"].([]interface{})
				if !ok {
					t.Fatalf("expected status array, got: %T", group["status"])
				}
				assert.Len(t, statuses, 2)

				s0, _ := statuses[0].(map[string]interface{})
				assert.Equal(t, "NEW", s0["new_status"])
				assert.Nil(t, s0["id"])

				s1, _ := statuses[1].(map[string]interface{})
				assert.Equal(t, "IN_PROGRESS", s1["new_status"])
				assert.Nil(t, s1["id"])
			}

			if tt.expectedErr != "" {
				var resp map[string]any
				if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to unmarshal error response: %v", err)
				}
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
