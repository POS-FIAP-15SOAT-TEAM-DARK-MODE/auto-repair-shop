package customer_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	domainmocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain/mocks"
	handler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/customer"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCustomerHandler_Create_Individual_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := domainmocks.NewCustomerService(t)
	h := handler.NewHandler(mockService)

	mockService.EXPECT().Create(mock.Anything, mock.AnythingOfType("domain.Customer")).Return(nil)

	router := gin.New()
	router.POST("/customers", h.Create())

	body := map[string]string{
		"name":     "João Silva",
		"email":    "joao@example.com",
		"password": "Secret@123",
		"type":     "INDIVIDUAL",
		"document": "111.444.777-35",
		"phone":    "11999999999",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/customers", bytes.NewBuffer(jsonBody))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestCustomerHandler_Create_Company_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := domainmocks.NewCustomerService(t)
	h := handler.NewHandler(mockService)

	mockService.EXPECT().Create(mock.Anything, mock.AnythingOfType("domain.Customer")).Return(nil)

	router := gin.New()
	router.POST("/customers", h.Create())

	body := map[string]string{
		"name":         "Empresa SA",
		"email":        "empresa@example.com",
		"password":     "Secret@123",
		"type":         "COMPANY",
		"document":     "11.222.333/0001-81",
		"company_name": "Empresa SA",
		"phone":        "11999999999",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/customers", bytes.NewBuffer(jsonBody))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestCustomerHandler_Create_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handler.NewHandler(nil)

	router := gin.New()
	router.POST("/customers", h.Create())

	req := httptest.NewRequest(http.MethodPost, "/customers", bytes.NewBufferString("invalid json"))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCustomerHandler_Create_ValidationFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handler.NewHandler(nil)

	router := gin.New()
	router.POST("/customers", h.Create())

	body := map[string]string{
		"name":     "João Silva",
		"email":    "joao@example.com",
		"password": "Secret@123",
		"type":     "INDIVIDUAL",
		"document": "", // missing
		"phone":    "11999999999",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/customers", bytes.NewBuffer(jsonBody))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "document is required")
}

func TestCustomerHandler_Create_DomainError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handler.NewHandler(nil)

	router := gin.New()
	router.POST("/customers", h.Create())

	body := map[string]string{
		"name":     "João Silva",
		"email":    "joao@example.com",
		"password": "Secret@123",
		"type":     "INDIVIDUAL",
		"document": "111.444.777-36", // invalid CPF check digit
		"phone":    "11999999999",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/customers", bytes.NewBuffer(jsonBody))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCustomerHandler_Create_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := domainmocks.NewCustomerService(t)
	h := handler.NewHandler(mockService)

	mockService.EXPECT().Create(mock.Anything, mock.AnythingOfType("domain.Customer")).Return(errors.New("internal error"))

	router := gin.New()
	router.POST("/customers", h.Create())

	body := map[string]string{
		"name":     "João Silva",
		"email":    "joao@example.com",
		"password": "Secret@123",
		"type":     "INDIVIDUAL",
		"document": "111.444.777-35",
		"phone":    "11999999999",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/customers", bytes.NewBuffer(jsonBody))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
