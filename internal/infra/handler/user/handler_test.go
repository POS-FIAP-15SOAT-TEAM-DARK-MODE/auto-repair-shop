package user_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	domainmocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain/mocks"
	handler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/user"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserHandler_Create_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := domainmocks.NewUserService(t)
	h := handler.HttpHandler(mockService)

	mockService.EXPECT().Create(mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)

	router := gin.New()
	router.POST("/auth/register", h.Create())

	body := map[string]string{
		"name":             "Test User",
		"email":            "test@example.com",
		"password":         "Secret@123",
		"confirm_password": "Secret@123",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(jsonBody))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestUserHandler_Create_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handler.HttpHandler(nil)

	router := gin.New()
	router.POST("/auth/register", h.Create())

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString("invalid json"))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUserHandler_Create_PasswordMismatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handler.HttpHandler(nil)

	router := gin.New()
	router.POST("/auth/register", h.Create())

	body := map[string]string{
		"name":             "Test User",
		"email":            "test@example.com",
		"password":         "Secret@123",
		"confirm_password": "mismatch",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(jsonBody))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "passwords do not match")
}

func TestUserHandler_Create_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := domainmocks.NewUserService(t)
	h := handler.HttpHandler(mockService)

	mockService.EXPECT().Create(mock.Anything, mock.AnythingOfType("*domain.User")).Return(errors.New("internal error"))

	router := gin.New()
	router.POST("/auth/register", h.Create())

	body := map[string]string{
		"name":             "Test User",
		"email":            "test@example.com",
		"password":         "Secret@123",
		"confirm_password": "Secret@123",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(jsonBody))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
