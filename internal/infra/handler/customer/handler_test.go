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
	r.PUT("/customers/:id", h.Update)
	r.DELETE("/customers/:id", h.Delete)
	return r
}

func strPtr(s string) *string { return &s }

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
		{
			name: "service_error_returns_500",
			body: map[string]any{
				"name": "João Silva", "email": "joao@example.com",
				"password": "Senha@123", "type": "INDIVIDUAL",
				"document": "111.444.777-35", "phone": "11999999999",
			},
			mockSetup: func(svc *domainmocks.CustomerService) {
				svc.EXPECT().Create(mock.Anything, mock.AnythingOfType("domain.Customer")).Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
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
			name: "not_found_returns_404",
			id:   "unknown",
			mockSetup: func(svc *domainmocks.CustomerService) {
				svc.EXPECT().GetByID(mock.Anything, "unknown").Return(domain.Customer{}, domain.ErrCustomerNotFound)
			},
			expectedStatus: http.StatusNotFound,
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

// --- Update ---

func TestUpdate(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		body           map[string]any
		mockSetup      func(*domainmocks.CustomerService)
		expectedStatus int
		assertBody     func(t *testing.T, body map[string]any)
	}{
		{
			name: "success_all_fields",
			id:   "uuid-individual",
			body: map[string]any{"name": "Novo Nome", "email": "novo@example.com", "phone": "11888888888"},
			mockSetup: func(svc *domainmocks.CustomerService) {
				svc.EXPECT().Update(mock.Anything, "uuid-individual", strPtr("Novo Nome"), strPtr("novo@example.com"), strPtr("11888888888")).
					Return(validCustomer(), nil)
			},
			expectedStatus: http.StatusOK,
			assertBody: func(t *testing.T, body map[string]any) {
				assert.Equal(t, "uuid-individual", body["id"])
			},
		},
		{
			name: "success_partial_only_phone",
			id:   "uuid-individual",
			body: map[string]any{"phone": "11777777777"},
			mockSetup: func(svc *domainmocks.CustomerService) {
				svc.EXPECT().Update(mock.Anything, "uuid-individual", (*string)(nil), (*string)(nil), strPtr("11777777777")).
					Return(validCustomer(), nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "not_found_returns_204",
			id:   "unknown",
			body: map[string]any{"name": "Novo Nome"},
			mockSetup: func(svc *domainmocks.CustomerService) {
				svc.EXPECT().Update(mock.Anything, "unknown", strPtr("Novo Nome"), (*string)(nil), (*string)(nil)).
					Return(domain.Customer{}, domain.ErrCustomerNotFound)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "invalid_name_returns_400",
			id:             "uuid-individual",
			body:           map[string]any{"name": "A"},
			mockSetup:      func(_ *domainmocks.CustomerService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty_phone_returns_400",
			id:             "uuid-individual",
			body:           map[string]any{"phone": ""},
			mockSetup:      func(_ *domainmocks.CustomerService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid_json_returns_400",
			id:             "uuid-individual",
			body:           nil,
			mockSetup:      func(_ *domainmocks.CustomerService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "duplicate_email_returns_409",
			id:   "uuid-individual",
			body: map[string]any{"email": "duplicate@example.com"},
			mockSetup: func(svc *domainmocks.CustomerService) {
				svc.EXPECT().Update(mock.Anything, "uuid-individual", (*string)(nil), strPtr("duplicate@example.com"), (*string)(nil)).
					Return(domain.Customer{}, domain.ErrDataConflict)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "invalid_email_returns_400",
			id:             "uuid-individual",
			body:           map[string]any{"email": "not-an-email"},
			mockSetup:      func(_ *domainmocks.CustomerService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "service_error_returns_500",
			id:   "uuid-individual",
			body: map[string]any{"phone": "11888888888"},
			mockSetup: func(svc *domainmocks.CustomerService) {
				svc.EXPECT().Update(mock.Anything, "uuid-individual", (*string)(nil), (*string)(nil), strPtr("11888888888")).
					Return(domain.Customer{}, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
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

			req := httptest.NewRequest(http.MethodPut, "/customers/"+tt.id, reqBody)
			req.Header.Set("Content-Type", "application/json")
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

// --- Delete ---

func TestDelete(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		mockSetup      func(*domainmocks.CustomerService)
		expectedStatus int
	}{
		{
			name: "success_returns_204",
			id:   "uuid-individual",
			mockSetup: func(svc *domainmocks.CustomerService) {
				svc.EXPECT().Delete(mock.Anything, "uuid-individual").Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name: "not_found_returns_204",
			id:   "unknown",
			mockSetup: func(svc *domainmocks.CustomerService) {
				svc.EXPECT().Delete(mock.Anything, "unknown").Return(domain.ErrCustomerNotFound)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name: "has_service_orders_returns_409",
			id:   "uuid-individual",
			mockSetup: func(svc *domainmocks.CustomerService) {
				svc.EXPECT().Delete(mock.Anything, "uuid-individual").Return(domain.ErrCustomerHasServiceOrders)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "service_error_returns_500",
			id:   "uuid-individual",
			mockSetup: func(svc *domainmocks.CustomerService) {
				svc.EXPECT().Delete(mock.Anything, "uuid-individual").Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := domainmocks.NewCustomerService(t)
			tt.mockSetup(svc)

			req := httptest.NewRequest(http.MethodDelete, "/customers/"+tt.id, nil)
			rec := httptest.NewRecorder()
			setupRouter(svc).ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
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
			name:       "not_found_returns_404",
			queryParam: "document=111.444.777-35",
			mockSetup: func(svc *domainmocks.CustomerService) {
				svc.EXPECT().GetByDocument(mock.Anything, mock.Anything).Return(domain.Customer{}, domain.ErrCustomerNotFound)
			},
			expectedStatus: http.StatusNotFound,
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
		{
			name:       "service_error_returns_500",
			queryParam: "document=111.444.777-35",
			mockSetup: func(svc *domainmocks.CustomerService) {
				svc.EXPECT().GetByDocument(mock.Anything, mock.Anything).Return(domain.Customer{}, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
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

func TestHandler_EdgeCases(t *testing.T) {
	svc := domainmocks.NewCustomerService(t)
	router := setupRouter(svc)

	t.Run("GetByID empty", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/customers/%20", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Update empty id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/customers/%20", bytes.NewReader([]byte(`{"name":"Test"}`)))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Delete empty id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/customers/%20", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}
