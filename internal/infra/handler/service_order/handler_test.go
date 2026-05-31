package service_order

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	domainmocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain/mocks"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/auth"
)

func validBearerToken(t *testing.T) string {
	t.Helper()
	tok, err := auth.GenerateToken("user-1", []domain.Role{domain.CUSTOMER}, time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	return "Bearer " + tok
}

func setupRouter(h *handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/service-orders", h.Create)
	r.GET("/service-orders", h.List)
	r.GET("/service-orders/:id/services", h.GetWorks)
	r.POST("/service-orders/:id/services", h.AddWork)
	r.DELETE("/service-orders/:id/services/:serviceId", h.DeleteWork)
	r.GET("/service-orders/:id/supplies", h.GetSupplies)
	r.POST("/service-orders/:id/supplies", h.AddSupplies)
	r.DELETE("/service-orders/:id/supplies/:supplyId", h.DeleteSupply)
	r.PATCH("/service-orders/:id/send-to-customer-approval", h.SendToCustomerApproval)
	r.PATCH("/service-orders/:id/finish", h.Finish)
	r.PATCH("/service-orders/:id/receive", h.Receive)
	r.PATCH("/service-orders/:id/send-to-diagnosis", h.SendToDiagnosis)
	r.PATCH("/service-orders/:id/accept", h.Accept)
	r.PATCH("/service-orders/:id/reject", h.Reject)
	r.PATCH("/service-orders/:id/deliver", h.Deliver)
	r.PATCH("/service-orders/:id/cancel", h.Cancel)
	r.POST("/service-orders/:id/works/:workId/next", h.NextWork)
	r.POST("/service-orders/:id/works/:workId/cancel", h.CancelWork)
	r.GET("/service-orders/:id/full", h.GetFullByID)
	return r
}

func jsonBody(v any) *bytes.Buffer {
	b, _ := json.Marshal(v)
	return bytes.NewBuffer(b)
}

func TestGetAverageExecutionTime_NoFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	expected := []domain.WorkExecutionTime{
		{WorkID: "w-1", WorkName: "Troca de óleo", AverageHours: 1.5},
		{WorkID: "w-2", WorkName: "Revisão de freios", AverageHours: 3.0},
	}

	mockSvc := domainmocks.NewServiceOrderService(t)
	mockSvc.EXPECT().AverageExecutionTime(mock.Anything, []string{}).Return(expected, nil)

	h := HttpHandler(mockSvc)
	router := gin.New()
	router.GET("/reports/average-execution-time", h.GetAverageExecutionTime)

	req := httptest.NewRequest(http.MethodGet, "/reports/average-execution-time", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp []workExecutionTimeResponse
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp, 2)
	assert.Equal(t, "w-1", resp[0].WorkID)
	assert.Equal(t, 1.5, resp[0].AverageExecutionTime)
}

func TestGetAverageExecutionTime_FilterByWorkID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	expected := []domain.WorkExecutionTime{
		{WorkID: "w-1", WorkName: "Troca de óleo", AverageHours: 1.5},
	}

	mockSvc := domainmocks.NewServiceOrderService(t)
	mockSvc.EXPECT().AverageExecutionTime(mock.Anything, []string{"w-1"}).Return(expected, nil)

	h := HttpHandler(mockSvc)
	router := gin.New()
	router.GET("/reports/average-execution-time", h.GetAverageExecutionTime)

	req := httptest.NewRequest(http.MethodGet, "/reports/average-execution-time?work_id=w-1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp []workExecutionTimeResponse
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp, 1)
	assert.Equal(t, "w-1", resp[0].WorkID)
}

func TestGetAverageExecutionTime_FilterByMultipleWorkIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	expected := []domain.WorkExecutionTime{
		{WorkID: "w-1", WorkName: "Troca de óleo", AverageHours: 1.5},
		{WorkID: "w-2", WorkName: "Revisão de freios", AverageHours: 3.0},
	}

	mockSvc := domainmocks.NewServiceOrderService(t)
	mockSvc.EXPECT().AverageExecutionTime(mock.Anything, []string{"w-1", "w-2"}).Return(expected, nil)

	h := HttpHandler(mockSvc)
	router := gin.New()
	router.GET("/reports/average-execution-time", h.GetAverageExecutionTime)

	req := httptest.NewRequest(http.MethodGet, "/reports/average-execution-time?work_id=w-1&work_id=w-2", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp []workExecutionTimeResponse
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp, 2)
}

func TestGetAverageExecutionTime_EmptyWhenNoCompletedWorks(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := domainmocks.NewServiceOrderService(t)
	mockSvc.EXPECT().AverageExecutionTime(mock.Anything, []string{}).Return([]domain.WorkExecutionTime{}, nil)

	h := HttpHandler(mockSvc)
	router := gin.New()
	router.GET("/reports/average-execution-time", h.GetAverageExecutionTime)

	req := httptest.NewRequest(http.MethodGet, "/reports/average-execution-time", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp []workExecutionTimeResponse
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Empty(t, resp)
}

func TestGetAverageExecutionTime_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := domainmocks.NewServiceOrderService(t)
	mockSvc.EXPECT().AverageExecutionTime(mock.Anything, []string{}).Return(nil, errors.New("db error"))

	h := HttpHandler(mockSvc)
	router := gin.New()
	router.GET("/reports/average-execution-time", h.GetAverageExecutionTime)

	req := httptest.NewRequest(http.MethodGet, "/reports/average-execution-time", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestHandler_GetStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		serviceOrderID string
		mockSetup      func(m *domainmocks.ServiceOrderService)
		expectedStatus int
	}{
		{
			name:           "success - should return status",
			serviceOrderID: "123",
			mockSetup: func(m *domainmocks.ServiceOrderService) {
				m.EXPECT().
					GetStatus(
						mock.Anything,
						"123",
					).
					Return(domain.StringToServiceOrderStatus("NEW"), nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "error - service returns error",
			serviceOrderID: "123",
			mockSetup: func(m *domainmocks.ServiceOrderService) {
				m.EXPECT().
					GetStatus(
						mock.Anything,
						"123",
					).
					Return("", domain.ErrServiceOrderNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req, _ := http.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Content-Type", "application/json")

			c.Request = req

			c.Params = gin.Params{
				{Key: serviceOrderIDParam, Value: tt.serviceOrderID},
			}

			mockService := domainmocks.NewServiceOrderService(t)

			if tt.mockSetup != nil {
				tt.mockSetup(mockService)
			}

			h := HttpHandler(mockService)

			handlerFunc := h.GetStatus
			handlerFunc(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name       string
		body       any
		mockSetup  func(m *domainmocks.ServiceOrderService)
		wantStatus int
	}{
		{
			name: "success",
			body: map[string]string{"client": "c-1", "vehicle": "v-1"},
			mockSetup: func(m *domainmocks.ServiceOrderService) {
				m.EXPECT().Create(mock.Anything, "c-1", "v-1").
					Return(domain.ServiceOrder{ID: "so-1", Status: domain.StringToServiceOrderStatus("NEW")}, nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "invalid json",
			body:       "not-json",
			mockSetup:  nil,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			body: map[string]string{"client": "c-1", "vehicle": "v-1"},
			mockSetup: func(m *domainmocks.ServiceOrderService) {
				m.EXPECT().Create(mock.Anything, "c-1", "v-1").
					Return(domain.ServiceOrder{}, errors.New("internal"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := domainmocks.NewServiceOrderService(t)
			if tt.mockSetup != nil {
				tt.mockSetup(mockSvc)
			}
			router := setupRouter(HttpHandler(mockSvc))

			var body *bytes.Buffer
			if s, ok := tt.body.(string); ok {
				body = bytes.NewBufferString(s)
			} else {
				body = jsonBody(tt.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/service-orders", body)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestList(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		mockSetup  func(m *domainmocks.ServiceOrderService)
		wantStatus int
	}{
		{
			name:  "success",
			query: "",
			mockSetup: func(m *domainmocks.ServiceOrderService) {
				m.EXPECT().List(mock.Anything, mock.Anything).
					Return(&domain.PaginatorResponse[domain.ServiceOrder]{Items: []domain.ServiceOrder{}}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "success with items",
			query: "?page=1&pageSize=10",
			mockSetup: func(m *domainmocks.ServiceOrderService) {
				so := domain.ServiceOrder{
					ID:     "so-1",
					Status: domain.StringToServiceOrderStatus("NEW"),
					Customer: &domain.Customer{
						ID:   "c-1",
						Type: domain.IndividualCustomerType,
						User: &domain.User{Name: "Alice", Email: "alice@test.com"},
					},
					Vehicle: &domain.Vehicle{
						ID:           "v-1",
						LicensePlate: "ABC1D23",
						Brand:        "Honda",
						Model:        "Civic",
						Year:         2021,
					},
				}
				m.EXPECT().List(mock.Anything, mock.Anything).
					Return(&domain.PaginatorResponse[domain.ServiceOrder]{Items: []domain.ServiceOrder{so}, TotalItems: 1, TotalPages: 1, PageSize: 10, Page: 1}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "service error",
			query: "",
			mockSetup: func(m *domainmocks.ServiceOrderService) {
				m.EXPECT().List(mock.Anything, mock.Anything).
					Return(nil, errors.New("internal"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := domainmocks.NewServiceOrderService(t)
			tt.mockSetup(mockSvc)
			router := setupRouter(HttpHandler(mockSvc))

			req := httptest.NewRequest(http.MethodGet, "/service-orders"+tt.query, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestGetWorks(t *testing.T) {
	tests := []struct {
		name       string
		soID       string
		mockSetup  func(m *domainmocks.ServiceOrderService)
		wantStatus int
	}{
		{
			name: "success",
			soID: "so-1",
			mockSetup: func(m *domainmocks.ServiceOrderService) {
				m.EXPECT().ListWorks(mock.Anything, "so-1").Return([]domain.Work{}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "service error",
			soID: "so-1",
			mockSetup: func(m *domainmocks.ServiceOrderService) {
				m.EXPECT().ListWorks(mock.Anything, "so-1").Return(nil, domain.ErrServiceOrderNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := domainmocks.NewServiceOrderService(t)
			tt.mockSetup(mockSvc)
			router := setupRouter(HttpHandler(mockSvc))

			req := httptest.NewRequest(http.MethodGet, "/service-orders/"+tt.soID+"/services", nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestAddWork(t *testing.T) {
	tests := []struct {
		name       string
		soID       string
		body       any
		mockSetup  func(m *domainmocks.ServiceOrderService)
		wantStatus int
	}{
		{
			name: "success",
			soID: "so-1",
			body: map[string]any{"services": []string{"w-1"}},
			mockSetup: func(m *domainmocks.ServiceOrderService) {
				m.EXPECT().AddWorks(mock.Anything, "so-1", []string{"w-1"}).Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "invalid json",
			soID:       "so-1",
			body:       "bad",
			mockSetup:  nil,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			soID: "so-1",
			body: map[string]any{"services": []string{"w-1"}},
			mockSetup: func(m *domainmocks.ServiceOrderService) {
				m.EXPECT().AddWorks(mock.Anything, "so-1", []string{"w-1"}).Return(domain.ErrServiceOrderNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := domainmocks.NewServiceOrderService(t)
			if tt.mockSetup != nil {
				tt.mockSetup(mockSvc)
			}
			router := setupRouter(HttpHandler(mockSvc))

			var body *bytes.Buffer
			if s, ok := tt.body.(string); ok {
				body = bytes.NewBufferString(s)
			} else {
				body = jsonBody(tt.body)
			}
			req := httptest.NewRequest(http.MethodPost, "/service-orders/"+tt.soID+"/services", body)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestDeleteWork(t *testing.T) {
	tests := []struct {
		name       string
		soID       string
		workID     string
		mockSetup  func(m *domainmocks.ServiceOrderService)
		wantStatus int
	}{
		{
			name:   "success",
			soID:   "so-1",
			workID: "w-1",
			mockSetup: func(m *domainmocks.ServiceOrderService) {
				m.EXPECT().RemoveWork(mock.Anything, "so-1", "w-1").Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:   "service error",
			soID:   "so-1",
			workID: "w-1",
			mockSetup: func(m *domainmocks.ServiceOrderService) {
				m.EXPECT().RemoveWork(mock.Anything, "so-1", "w-1").Return(domain.ErrServiceOrderNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := domainmocks.NewServiceOrderService(t)
			tt.mockSetup(mockSvc)
			router := setupRouter(HttpHandler(mockSvc))

			req := httptest.NewRequest(http.MethodDelete, "/service-orders/"+tt.soID+"/services/"+tt.workID, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestGetSupplies(t *testing.T) {
	tests := []struct {
		name       string
		soID       string
		mockSetup  func(m *domainmocks.ServiceOrderService)
		wantStatus int
	}{
		{
			name: "success",
			soID: "so-1",
			mockSetup: func(m *domainmocks.ServiceOrderService) {
				m.EXPECT().ListSupplies(mock.Anything, "so-1").Return([]domain.Supply{}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "service error",
			soID: "so-1",
			mockSetup: func(m *domainmocks.ServiceOrderService) {
				m.EXPECT().ListSupplies(mock.Anything, "so-1").Return(nil, domain.ErrServiceOrderNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := domainmocks.NewServiceOrderService(t)
			tt.mockSetup(mockSvc)
			router := setupRouter(HttpHandler(mockSvc))

			req := httptest.NewRequest(http.MethodGet, "/service-orders/"+tt.soID+"/supplies", nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestAddSupplies(t *testing.T) {
	tests := []struct {
		name       string
		soID       string
		body       any
		mockSetup  func(m *domainmocks.ServiceOrderService)
		wantStatus int
	}{
		{
			name: "success",
			soID: "so-1",
			body: map[string]any{"supplies": []map[string]any{{"id": "s-1", "amount": 2}}},
			mockSetup: func(m *domainmocks.ServiceOrderService) {
				m.EXPECT().AddSupplies(mock.Anything, "so-1", []domain.AddSupply{{ID: "s-1", Amount: 2}}).Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "invalid json",
			soID:       "so-1",
			body:       "bad",
			mockSetup:  nil,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			soID: "so-1",
			body: map[string]any{"supplies": []map[string]any{{"id": "s-1", "amount": 2}}},
			mockSetup: func(m *domainmocks.ServiceOrderService) {
				m.EXPECT().AddSupplies(mock.Anything, "so-1", mock.Anything).Return(domain.ErrServiceOrderNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := domainmocks.NewServiceOrderService(t)
			if tt.mockSetup != nil {
				tt.mockSetup(mockSvc)
			}
			router := setupRouter(HttpHandler(mockSvc))

			var body *bytes.Buffer
			if s, ok := tt.body.(string); ok {
				body = bytes.NewBufferString(s)
			} else {
				body = jsonBody(tt.body)
			}
			req := httptest.NewRequest(http.MethodPost, "/service-orders/"+tt.soID+"/supplies", body)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestDeleteSupply(t *testing.T) {
	tests := []struct {
		name       string
		soID       string
		supplyID   string
		mockSetup  func(m *domainmocks.ServiceOrderService)
		wantStatus int
	}{
		{
			name:     "success",
			soID:     "so-1",
			supplyID: "s-1",
			mockSetup: func(m *domainmocks.ServiceOrderService) {
				m.EXPECT().RemoveSupply(mock.Anything, "so-1", "s-1").Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:     "service error",
			soID:     "so-1",
			supplyID: "s-1",
			mockSetup: func(m *domainmocks.ServiceOrderService) {
				m.EXPECT().RemoveSupply(mock.Anything, "so-1", "s-1").Return(domain.ErrServiceOrderNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := domainmocks.NewServiceOrderService(t)
			tt.mockSetup(mockSvc)
			router := setupRouter(HttpHandler(mockSvc))

			req := httptest.NewRequest(http.MethodDelete, "/service-orders/"+tt.soID+"/supplies/"+tt.supplyID, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestSendToCustomerApproval(t *testing.T) {
	tests := []struct {
		name       string
		soID       string
		mockSetup  func(m *domainmocks.ServiceOrderService)
		wantStatus int
	}{
		{
			name: "success",
			soID: "so-1",
			mockSetup: func(m *domainmocks.ServiceOrderService) {
				m.EXPECT().SendToCustomerApproval(mock.Anything, "so-1").Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "service error",
			soID: "so-1",
			mockSetup: func(m *domainmocks.ServiceOrderService) {
				m.EXPECT().SendToCustomerApproval(mock.Anything, "so-1").Return(errors.New("internal"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := domainmocks.NewServiceOrderService(t)
			tt.mockSetup(mockSvc)
			router := setupRouter(HttpHandler(mockSvc))

			req := httptest.NewRequest(http.MethodPatch, "/service-orders/"+tt.soID+"/send-to-customer-approval", nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestFinish(t *testing.T) {
	tests := []struct {
		name       string
		soID       string
		mockSetup  func(m *domainmocks.ServiceOrderService)
		wantStatus int
	}{
		{
			name: "success",
			soID: "so-1",
			mockSetup: func(m *domainmocks.ServiceOrderService) {
				m.EXPECT().Finish(mock.Anything, "so-1").Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "service error",
			soID: "so-1",
			mockSetup: func(m *domainmocks.ServiceOrderService) {
				m.EXPECT().Finish(mock.Anything, "so-1").Return(errors.New("internal"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := domainmocks.NewServiceOrderService(t)
			tt.mockSetup(mockSvc)
			router := setupRouter(HttpHandler(mockSvc))

			req := httptest.NewRequest(http.MethodPatch, "/service-orders/"+tt.soID+"/finish", nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestReceive(t *testing.T) {
	tests := []struct {
		name       string
		soID       string
		mockSetup  func(m *domainmocks.ServiceOrderService)
		wantStatus int
	}{
		{
			name: "success",
			soID: "so-1",
			mockSetup: func(m *domainmocks.ServiceOrderService) {
				m.EXPECT().Receive(mock.Anything, "so-1").Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "service error",
			soID: "so-1",
			mockSetup: func(m *domainmocks.ServiceOrderService) {
				m.EXPECT().Receive(mock.Anything, "so-1").Return(errors.New("internal"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := domainmocks.NewServiceOrderService(t)
			tt.mockSetup(mockSvc)
			router := setupRouter(HttpHandler(mockSvc))

			req := httptest.NewRequest(http.MethodPatch, "/service-orders/"+tt.soID+"/receive", nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestSendToDiagnosis(t *testing.T) {
	tests := []struct {
		name       string
		soID       string
		mockSetup  func(m *domainmocks.ServiceOrderService)
		wantStatus int
	}{
		{
			name: "success",
			soID: "so-1",
			mockSetup: func(m *domainmocks.ServiceOrderService) {
				m.EXPECT().SendToDiagnosis(mock.Anything, "so-1").Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "service error",
			soID: "so-1",
			mockSetup: func(m *domainmocks.ServiceOrderService) {
				m.EXPECT().SendToDiagnosis(mock.Anything, "so-1").Return(errors.New("internal"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := domainmocks.NewServiceOrderService(t)
			tt.mockSetup(mockSvc)
			router := setupRouter(HttpHandler(mockSvc))

			req := httptest.NewRequest(http.MethodPatch, "/service-orders/"+tt.soID+"/send-to-diagnosis", nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestAccept(t *testing.T) {
	tests := []struct {
		name       string
		soID       string
		addAuth    bool
		mockSetup  func(m *domainmocks.ServiceOrderService)
		wantStatus int
	}{
		{
			name:    "success",
			soID:    "so-1",
			addAuth: true,
			mockSetup: func(m *domainmocks.ServiceOrderService) {
				m.EXPECT().Accept(mock.Anything, "so-1", "user-1").Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "missing auth header",
			soID:       "so-1",
			addAuth:    false,
			mockSetup:  nil,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:    "service error",
			soID:    "so-1",
			addAuth: true,
			mockSetup: func(m *domainmocks.ServiceOrderService) {
				m.EXPECT().Accept(mock.Anything, "so-1", "user-1").Return(errors.New("internal"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := domainmocks.NewServiceOrderService(t)
			if tt.mockSetup != nil {
				tt.mockSetup(mockSvc)
			}
			router := setupRouter(HttpHandler(mockSvc))

			req := httptest.NewRequest(http.MethodPatch, "/service-orders/"+tt.soID+"/accept", nil)
			if tt.addAuth {
				req.Header.Set("Authorization", validBearerToken(t))
			}
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestReject(t *testing.T) {
	tests := []struct {
		name       string
		soID       string
		addAuth    bool
		mockSetup  func(m *domainmocks.ServiceOrderService)
		wantStatus int
	}{
		{
			name:    "success",
			soID:    "so-1",
			addAuth: true,
			mockSetup: func(m *domainmocks.ServiceOrderService) {
				m.EXPECT().Reject(mock.Anything, "so-1", "user-1").Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "missing auth header",
			soID:       "so-1",
			addAuth:    false,
			mockSetup:  nil,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:    "service error",
			soID:    "so-1",
			addAuth: true,
			mockSetup: func(m *domainmocks.ServiceOrderService) {
				m.EXPECT().Reject(mock.Anything, "so-1", "user-1").Return(errors.New("internal"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := domainmocks.NewServiceOrderService(t)
			if tt.mockSetup != nil {
				tt.mockSetup(mockSvc)
			}
			router := setupRouter(HttpHandler(mockSvc))

			req := httptest.NewRequest(http.MethodPatch, "/service-orders/"+tt.soID+"/reject", nil)
			if tt.addAuth {
				req.Header.Set("Authorization", validBearerToken(t))
			}
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestDeliver(t *testing.T) {
	tests := []struct {
		name       string
		soID       string
		mockSetup  func(m *domainmocks.ServiceOrderService)
		wantStatus int
	}{
		{
			name: "success",
			soID: "so-1",
			mockSetup: func(m *domainmocks.ServiceOrderService) {
				m.EXPECT().Deliver(mock.Anything, "so-1").Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "service error",
			soID: "so-1",
			mockSetup: func(m *domainmocks.ServiceOrderService) {
				m.EXPECT().Deliver(mock.Anything, "so-1").Return(errors.New("internal"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := domainmocks.NewServiceOrderService(t)
			tt.mockSetup(mockSvc)
			router := setupRouter(HttpHandler(mockSvc))

			req := httptest.NewRequest(http.MethodPatch, "/service-orders/"+tt.soID+"/deliver", nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestCancel(t *testing.T) {
	tests := []struct {
		name       string
		soID       string
		mockSetup  func(m *domainmocks.ServiceOrderService)
		wantStatus int
	}{
		{
			name: "success",
			soID: "so-1",
			mockSetup: func(m *domainmocks.ServiceOrderService) {
				m.EXPECT().Cancel(mock.Anything, "so-1").Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "service error",
			soID: "so-1",
			mockSetup: func(m *domainmocks.ServiceOrderService) {
				m.EXPECT().Cancel(mock.Anything, "so-1").Return(errors.New("internal"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := domainmocks.NewServiceOrderService(t)
			tt.mockSetup(mockSvc)
			router := setupRouter(HttpHandler(mockSvc))

			req := httptest.NewRequest(http.MethodPatch, "/service-orders/"+tt.soID+"/cancel", nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestNextWork(t *testing.T) {
	tests := []struct {
		name       string
		soID       string
		workID     string
		mockSetup  func(m *domainmocks.ServiceOrderService)
		wantStatus int
	}{
		{
			name:   "success",
			soID:   "so-1",
			workID: "w-1",
			mockSetup: func(m *domainmocks.ServiceOrderService) {
				m.EXPECT().NextWork(mock.Anything, "so-1", "w-1").Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:   "service error",
			soID:   "so-1",
			workID: "w-1",
			mockSetup: func(m *domainmocks.ServiceOrderService) {
				m.EXPECT().NextWork(mock.Anything, "so-1", "w-1").Return(errors.New("internal"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := domainmocks.NewServiceOrderService(t)
			tt.mockSetup(mockSvc)
			router := setupRouter(HttpHandler(mockSvc))

			req := httptest.NewRequest(http.MethodPost, "/service-orders/"+tt.soID+"/works/"+tt.workID+"/next", nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestCancelWork(t *testing.T) {
	tests := []struct {
		name       string
		soID       string
		workID     string
		mockSetup  func(m *domainmocks.ServiceOrderService)
		wantStatus int
	}{
		{
			name:   "success",
			soID:   "so-1",
			workID: "w-1",
			mockSetup: func(m *domainmocks.ServiceOrderService) {
				m.EXPECT().CancelWork(mock.Anything, "so-1", "w-1").Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:   "service error",
			soID:   "so-1",
			workID: "w-1",
			mockSetup: func(m *domainmocks.ServiceOrderService) {
				m.EXPECT().CancelWork(mock.Anything, "so-1", "w-1").Return(errors.New("internal"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := domainmocks.NewServiceOrderService(t)
			tt.mockSetup(mockSvc)
			router := setupRouter(HttpHandler(mockSvc))

			req := httptest.NewRequest(http.MethodPost, "/service-orders/"+tt.soID+"/works/"+tt.workID+"/cancel", nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestGetFullByID_Detailed(t *testing.T) {
	fullSO := domain.FullServiceOrder{
		ServiceOrder: domain.ServiceOrder{
			ID:     "so-1",
			Status: domain.StringToServiceOrderStatus("NEW"),
			Customer: &domain.Customer{
				ID:          "c-1",
				Type:        domain.IndividualCustomerType,
				CPF:         "12345678901",
				CompanyName: "Company",
				Phone:       "119",
				User:        &domain.User{Name: "Test User", Email: "test@test.com"},
			},
			Vehicle: &domain.Vehicle{
				ID:           "v-1",
				LicensePlate: "ABC1D23",
				Brand:        "Toyota",
				Model:        "Corolla",
				Year:         2020,
			},
		},
		Works: []domain.Work{
			{ID: "w-1", Name: "Work 1", Price: decimal.NewFromInt(100), Status: domain.ACTIVE},
		},
		Supplies: []domain.Supply{
			{ID: "s-1", Name: "Supply 1", UnitPrice: decimal.NewFromInt(10), StockQuantity: 5, Version: 1},
		},
	}

	t.Run("success", func(t *testing.T) {
		mockSvc := domainmocks.NewServiceOrderService(t)
		mockSvc.EXPECT().GetFullOSByID(mock.Anything, "so-1").Return(fullSO, nil)
		router := setupRouter(HttpHandler(mockSvc))

		req := httptest.NewRequest(http.MethodGet, "/service-orders/so-1/full", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp serviceOrderDetailResponse
		assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, "so-1", resp.ID)
		assert.Len(t, resp.Works, 1)
		assert.Len(t, resp.Supplies, 1)
		assert.Equal(t, "12345678901", resp.Customer.Document)
	})

	t.Run("empty id", func(t *testing.T) {
		mockSvc := domainmocks.NewServiceOrderService(t)
		router := setupRouter(HttpHandler(mockSvc))

		req := httptest.NewRequest(http.MethodGet, "/service-orders/%20/full", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("service error", func(t *testing.T) {
		mockSvc := domainmocks.NewServiceOrderService(t)
		mockSvc.EXPECT().GetFullOSByID(mock.Anything, "so-1").Return(domain.FullServiceOrder{}, assert.AnError)
		router := setupRouter(HttpHandler(mockSvc))

		req := httptest.NewRequest(http.MethodGet, "/service-orders/so-1/full", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestGetFullByID_Corporate(t *testing.T) {
	fullSO := domain.FullServiceOrder{
		ServiceOrder: domain.ServiceOrder{
			ID:     "so-1",
			Status: domain.StringToServiceOrderStatus("NEW"),
			Customer: &domain.Customer{
				ID:          "c-1",
				Type:        domain.CompanyCustomerType,
				CNPJ:        "12345678000199",
				CompanyName: "Company",
				User:        &domain.User{Name: "Test User", Email: "test@test.com"},
			},
			Vehicle: &domain.Vehicle{ID: "v-1"},
		},
	}

	mockSvc := domainmocks.NewServiceOrderService(t)
	mockSvc.EXPECT().GetFullOSByID(mock.Anything, "so-1").Return(fullSO, nil)
	router := setupRouter(HttpHandler(mockSvc))

	req := httptest.NewRequest(http.MethodGet, "/service-orders/so-1/full", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp serviceOrderDetailResponse
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "12345678000199", resp.Customer.Document)
}
