package user_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
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
	router.POST("/auth/register", h.Create)

	body := map[string]string{
		"name":            "Test User",
		"email":           "test@example.com",
		"password":        "Secret@123",
		"confirmPassword": "Secret@123",
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
	router.POST("/auth/register", h.Create)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString("invalid json"))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUserHandler_Create_PasswordMismatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handler.HttpHandler(nil)

	router := gin.New()
	router.POST("/auth/register", h.Create)

	body := map[string]string{
		"name":            "Test User",
		"email":           "test@example.com",
		"password":        "Secret@123",
		"confirmPassword": "mismatch",
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
	router.POST("/auth/register", h.Create)

	body := map[string]string{
		"name":            "Test User",
		"email":           "test@example.com",
		"password":        "Secret@123",
		"confirmPassword": "Secret@123",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(jsonBody))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestLoginUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name             string
		requestBody      string
		mockSvc          func(t *testing.T) *domainmocks.UserService
		expectedStatus   int
		expectedResponse string
	}{
		{
			name: "Should fail gracefully when request body is invalid",
			requestBody: `{
				"email": "test@example.com"
				"password": "password"
			}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Should fail gracefully when service fails",
			requestBody:    `{"email": "test@example.com", "password": "password"}`,
			expectedStatus: http.StatusInternalServerError,
			mockSvc: func(t *testing.T) *domainmocks.UserService {
				svc := domainmocks.NewUserService(t)
				svc.On("Login", mock.Anything, mock.Anything).Return(errors.New("service failed"))
				return svc
			},
		},
		{
			name:           "Should login successfully",
			requestBody:    `{"email": "test@example.com", "password": "password"}`,
			expectedStatus: http.StatusOK,
			mockSvc: func(t *testing.T) *domainmocks.UserService {
				svc := domainmocks.NewUserService(t)
				svc.On("Login", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					loggedUser := args[1].(*domain.LoggedUser)
					loggedUser.SessionToken = "success-token"
					loggedUser.SessionExpiresIn = 3600
				}).Return(nil)
				return svc
			},
			expectedResponse: `{"token":"success-token","expires_in":3600}`,
		},
	}

	for _, tt := range tests {
		loginRoute := "/v1/auth/login"
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &domainmocks.UserService{}
			if tt.mockSvc != nil {
				mockSvc = tt.mockSvc(t)
			}

			h := handler.HttpHandler(mockSvc)

			router := gin.New()
			router.POST(loginRoute, h.Login)

			req := httptest.NewRequest(http.MethodPost, loginRoute, bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedResponse != "" {
				if rec.Body.String() != tt.expectedResponse {
					t.Errorf("expected response %s, got %s", tt.expectedResponse, rec.Body.String())
				}
			}
		})
	}
}

func TestUpdateRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		path       string
		body       string
		mockSvc    func(t *testing.T) *domainmocks.UserService
		expectedSC int
	}{
		{
			name:       "Should fail gracefully when request body is invalid",
			path:       "/v1/users/123/role",
			body:       `{"role":`,
			expectedSC: http.StatusBadRequest,
		},
		{
			name:       "Should fail gracefully when role is empty",
			path:       "/v1/users/123/role",
			body:       `{}`,
			expectedSC: http.StatusBadRequest,
		},
		{
			name: "Should fail gracefully when service fails",
			path: "/v1/users/123/role",
			body: `{"role":"MECHANIC"}`,
			mockSvc: func(t *testing.T) *domainmocks.UserService {
				svc := domainmocks.NewUserService(t)
				svc.EXPECT().UpdateRole(mock.Anything, "123", domain.MECHANIC).Return(errors.New("service failed"))
				return svc
			},
			expectedSC: http.StatusInternalServerError,
		},
		{
			name: "Should update role successfully",
			path: "/v1/users/123/role",
			body: `{"role":"MECHANIC"}`,
			mockSvc: func(t *testing.T) *domainmocks.UserService {
				svc := domainmocks.NewUserService(t)
				svc.EXPECT().UpdateRole(mock.Anything, "123", domain.MECHANIC).Return(nil)
				return svc
			},
			expectedSC: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &domainmocks.UserService{}
			if tt.mockSvc != nil {
				mockSvc = tt.mockSvc(t)
			}

			h := handler.HttpHandler(mockSvc)
			router := gin.New()
			router.PATCH("/v1/users/:id/role", h.UpdateRole)

			req := httptest.NewRequest(http.MethodPatch, tt.path, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedSC, rec.Code)
		})
	}
}
