package customer_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	domainmocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain/mocks"
	handler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/customer"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupRouter(svc *domainmocks.CustomerService) *gin.Engine {
	h := handler.NewHandler(svc)
	r := gin.New()
	r.POST("/customers", h.Create)
	r.GET("/customers/:id", h.GetByID)
	r.GET("/customers", h.GetByDocument)
	return r
}

func validCustomer() domain.Customer {
	return domain.Customer{
		ID:    "uuid-individual",
		Type:  domain.IndividualCustomerType,
		CPF:   "11144477735",
		Phone: "11999999999",
		User: &domain.User{
			Name:  "João Silva",
			Email: "joao@example.com",
		},
	}
}

// --- Create ---

func TestCreate(t *testing.T) {
	tests := []struct {
		name           string
		body           map[string]any
		mockSetup      func(*domainmocks.CustomerService)
		expectedStatus int
	}{
		{
			name: "success_individual",
			body: map[string]any{
				"name": "João Silva", "email": "joao@example.com",
				"password": "Senha@123", "type": "INDIVIDUAL",
				"document": "111.444.777-35", "phone": "11999999999",
			},
			mockSetup: func(svc *domainmocks.CustomerService) {
				svc.EXPECT().Create(mock.Anything, mock.AnythingOfType("domain.Customer")).Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "success_company",
			body: map[string]any{
				"name": "Empresa SA", "email": "empresa@example.com",
				"password": "Senha@123", "type": "COMPANY",
				"document": "11.222.333/0001-81", "company_name": "Empresa SA", "phone": "1133334444",
			},
			mockSetup: func(svc *domainmocks.CustomerService) {
				svc.EXPECT().Create(mock.Anything, mock.AnythingOfType("domain.Customer")).Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing_document",
			body: map[string]any{
				"name": "João Silva", "email": "joao@example.com",
				"password": "Senha@123", "type": "INDIVIDUAL", "phone": "11999999999",
			},
			mockSetup:      func(_ *domainmocks.CustomerService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid_cpf",
			body: map[string]any{
				"name": "João Silva", "email": "joao@example.com",
				"password": "Senha@123", "type": "INDIVIDUAL",
				"document": "000.000.000-00", "phone": "11999999999",
			},
			mockSetup:      func(_ *domainmocks.CustomerService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid_json",
			body:           nil, // will send raw invalid payload
			mockSetup:      func(_ *domainmocks.CustomerService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "service_conflict",
			body: map[string]any{
				"name": "João Silva", "email": "joao@example.com",
				"password": "Senha@123", "type": "INDIVIDUAL",
				"document": "111.444.777-35", "phone": "11999999999",
			},
			mockSetup: func(svc *domainmocks.CustomerService) {
				svc.EXPECT().Create(mock.Anything, mock.AnythingOfType("domain.Customer")).Return(domain.ErrDataConflict)
			},
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := domainmocks.NewCustomerService(t)
			tt.mockSetup(svc)

			var reqBody *bytes.Reader
			if tt.body != nil {
				payload, _ := json.Marshal(tt.body)
				reqBody = bytes.NewReader(payload)
			} else {
				reqBody = bytes.NewReader([]byte("invalid-json{"))
			}

			req := httptest.NewRequest(http.MethodPost, "/customers", reqBody)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			setupRouter(svc).ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

// --- GetByID ---

func TestGetByID(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		mockSetup      func(*domainmocks.CustomerService)
		expectedStatus int
		assertBody     func(t *testing.T, body map[string]any)
	}{
		{
			name: "success",
			id:   "uuid-individual",
			mockSetup: func(svc *domainmocks.CustomerService) {
				svc.EXPECT().GetByID(mock.Anything, "uuid-individual").Return(validCustomer(), nil)
			},
			expectedStatus: http.StatusOK,
			assertBody: func(t *testing.T, body map[string]any) {
				assert.Equal(t, "uuid-individual", body["id"])
			},
		},
		{
			name: "not_found_returns_204",
			id:   "unknown",
			mockSetup: func(svc *domainmocks.CustomerService) {
				svc.EXPECT().GetByID(mock.Anything, "unknown").Return(domain.Customer{}, domain.ErrCustomerNotFound)
			},
			expectedStatus: http.StatusNoContent,
			assertBody:     nil,
		},
		{
			name: "service_error_returns_500",
			id:   "any-id",
			mockSetup: func(svc *domainmocks.CustomerService) {
				svc.EXPECT().GetByID(mock.Anything, "any-id").Return(domain.Customer{}, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			assertBody:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := domainmocks.NewCustomerService(t)
			tt.mockSetup(svc)

			req := httptest.NewRequest(http.MethodGet, "/customers/"+tt.id, nil)
			rec := httptest.NewRecorder()
			setupRouter(svc).ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.assertBody != nil {
				var body map[string]any
				assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				tt.assertBody(t, body)
			}
		})
	}
}

// --- GetByDocument ---

func TestGetByDocument(t *testing.T) {
	tests := []struct {
		name           string
		queryParam     string
		mockSetup      func(*domainmocks.CustomerService)
		expectedStatus int
		assertBody     func(t *testing.T, body map[string]any)
	}{
		{
			name:       "success",
			queryParam: "document=111.444.777-35",
			mockSetup: func(svc *domainmocks.CustomerService) {
				svc.EXPECT().GetByDocument(mock.Anything, "111.444.777-35").Return(validCustomer(), nil)
			},
			expectedStatus: http.StatusOK,
			assertBody: func(t *testing.T, body map[string]any) {
				assert.Equal(t, "uuid-individual", body["id"])
			},
		},
		{
			name:       "not_found_returns_204",
			queryParam: "document=111.444.777-35",
			mockSetup: func(svc *domainmocks.CustomerService) {
				svc.EXPECT().GetByDocument(mock.Anything, mock.Anything).Return(domain.Customer{}, domain.ErrCustomerNotFound)
			},
			expectedStatus: http.StatusNoContent,
			assertBody:     nil,
		},
		{
			name:           "missing_param_returns_400",
			queryParam:     "",
			mockSetup:      func(_ *domainmocks.CustomerService) {},
			expectedStatus: http.StatusBadRequest,
			assertBody:     nil,
		},
		{
			name:       "invalid_document_format_returns_400",
			queryParam: "document=123",
			mockSetup: func(svc *domainmocks.CustomerService) {
				svc.EXPECT().GetByDocument(mock.Anything, mock.Anything).Return(domain.Customer{}, domain.ErrInvalidDocumentFormat)
			},
			expectedStatus: http.StatusBadRequest,
			assertBody:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := domainmocks.NewCustomerService(t)
			tt.mockSetup(svc)

			url := "/customers"
			if tt.queryParam != "" {
				url += "?" + tt.queryParam
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			rec := httptest.NewRecorder()
			setupRouter(svc).ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.assertBody != nil {
				var body map[string]any
				assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				tt.assertBody(t, body)
			}
		})
	}
}
