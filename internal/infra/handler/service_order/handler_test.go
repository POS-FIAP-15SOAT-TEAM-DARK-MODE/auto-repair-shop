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

	domainmocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain/mocks"
)

func TestGetAverageExecutionTime_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := domainmocks.NewServiceOrderService(t)
	mockSvc.EXPECT().AverageExecutionTime(mock.Anything).Return(2.5, nil)

	h := HttpHandler(mockSvc)
	router := gin.New()
	router.GET("/reports/average-execution-time", h.GetAverageExecutionTime)

	req := httptest.NewRequest(http.MethodGet, "/reports/average-execution-time", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp averageExecutionTimeResponse
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, 2.5, resp.AverageExecutionTimeHours)
}

func TestGetAverageExecutionTime_ZeroWhenNoCompletedOrders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := domainmocks.NewServiceOrderService(t)
	mockSvc.EXPECT().AverageExecutionTime(mock.Anything).Return(0.0, nil)

	h := HttpHandler(mockSvc)
	router := gin.New()
	router.GET("/reports/average-execution-time", h.GetAverageExecutionTime)

	req := httptest.NewRequest(http.MethodGet, "/reports/average-execution-time", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp averageExecutionTimeResponse
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, 0.0, resp.AverageExecutionTimeHours)
}

func TestGetAverageExecutionTime_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := domainmocks.NewServiceOrderService(t)
	mockSvc.EXPECT().AverageExecutionTime(mock.Anything).Return(0.0, errors.New("db error"))

	h := HttpHandler(mockSvc)
	router := gin.New()
	router.GET("/reports/average-execution-time", h.GetAverageExecutionTime)

	req := httptest.NewRequest(http.MethodGet, "/reports/average-execution-time", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
