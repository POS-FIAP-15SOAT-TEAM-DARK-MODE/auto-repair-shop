package service_order

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	domainmocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain/mocks"
)

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
