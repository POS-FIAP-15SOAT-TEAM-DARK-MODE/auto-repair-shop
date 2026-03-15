package user

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
)

func TestLoginUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name             string
		requestBody      string
		mockSvc          func(t *testing.T) *mocks.UserService
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
			name:           "Should fail gracefully when the email is empty",
			requestBody:    `{"email": "", "password": "password"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Should fail gracefully when the password is empty",
			requestBody:    `{"email": "test@example.com", "password": ""}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Should fail gracefully when parsing dto to domain fails",
			requestBody:    `{"email": "test@example.com", "password": "password12345678901234567890123456789012345678901234567890123456789012345678901234567890"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Should fail gracefully when service fails",
			requestBody:    `{"email": "test@example.com", "password": "password"}`,
			expectedStatus: http.StatusInternalServerError,
			mockSvc: func(t *testing.T) *mocks.UserService {
				svc := mocks.NewUserService(t)
				svc.On("Login", mock.Anything).Return(nil, errors.New("service failed"))
				return svc
			},
		},
		{
			name:           "Should login successfully",
			requestBody:    `{"email": "test@example.com", "password": "password"}`,
			expectedStatus: http.StatusOK,
			mockSvc: func(t *testing.T) *mocks.UserService {
				svc := mocks.NewUserService(t)
				svc.On("Login", mock.Anything).Return(&domain.LoginResponse{
					Token:     "success-token",
					ExpiresIn: 3600,
				}, nil)
				return svc
			},
			expectedResponse: `{"token":"success-token","expires_in":3600}`,
		},
	}

	for _, tt := range tests {
		loginRoute := "/v1/auth/login"
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mocks.UserService{}
			if tt.mockSvc != nil {
				mockSvc = tt.mockSvc(t)
			}

			h := NewHandler(mockSvc)

			router := gin.New()
			router.POST(loginRoute, h.LoginUser)

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
