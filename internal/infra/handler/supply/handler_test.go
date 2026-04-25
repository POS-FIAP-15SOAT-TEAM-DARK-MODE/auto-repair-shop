package supply_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	domainmocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain/mocks"
	handler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/supply"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// -------------------------
// Setup helpers
// -------------------------

func setupRouter(svc *domainmocks.SupplyService) *gin.Engine {
	h := handler.HttpHandler(svc)
	r := gin.New()
	r.POST("/supplies", h.Create)
	r.GET("/supplies", h.List)
	r.PUT("/supplies/:id", h.Update)
	return r
}

func validSupplyDomain() *domain.Supply {
	return &domain.Supply{
		ID:            uuid.New().String(),
		Name:          "Brake Pad",
		Description:   "High performance brake pad",
		UnitPrice:     decimal.NewFromFloat(49.99),
		StockQuantity: 10,
		Version:       1,
	}
}

func validSupplyPayload() map[string]any {
	return map[string]any{
		"name":          "Brake Pad",
		"description":   "High performance brake pad",
		"unitPrice":     "49.99",
		"stockQuantity": 10,
	}
}

func toJSON(t *testing.T, v any) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(v)
	assert.NoError(t, err)
	return bytes.NewBuffer(b)
}

// -------------------------
// Create
// -------------------------

func TestCreate(t *testing.T) {
	tests := []struct {
		name           string
		payload        any
		mockSetup      func(*domainmocks.SupplyService)
		expectedStatus int
	}{
		{
			name:    "success_creates_supply",
			payload: validSupplyPayload(),
			mockSetup: func(svc *domainmocks.SupplyService) {
				svc.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*domain.Supply")).
					Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:    "service_error_on_first_call",
			payload: validSupplyPayload(),
			mockSetup: func(svc *domainmocks.SupplyService) {
				svc.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*domain.Supply")).
					Return(errors.New("internal error")).
					Once()
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "invalid_payload_returns_bad_request",
			payload:        "invalid-json",
			mockSetup:      func(svc *domainmocks.SupplyService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := domainmocks.NewSupplyService(t)
			tt.mockSetup(svc)

			req := httptest.NewRequest(http.MethodPost, "/supplies", toJSON(t, tt.payload))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			setupRouter(svc).ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestCreate_ResponseBody(t *testing.T) {
	svc := domainmocks.NewSupplyService(t)
	svc.EXPECT().
		Create(mock.Anything, mock.AnythingOfType("*domain.Supply")).
		Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/supplies", toJSON(t, validSupplyPayload()))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	setupRouter(svc).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var body map[string]any
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "Brake Pad", body["name"])
	assert.Equal(t, "High performance brake pad", body["description"])
}

// -------------------------
// List
// -------------------------

func TestList(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		mockSetup      func(*domainmocks.SupplyService)
		expectedStatus int
	}{
		{
			name:  "success_returns_supplies",
			query: "?page=1&pageSize=10",
			mockSetup: func(svc *domainmocks.SupplyService) {
				svc.EXPECT().
					List(mock.Anything, mock.AnythingOfType("*domain.ListSupplyParams")).
					Return(&domain.PaginatorResponse[domain.Supply]{
						Items: []domain.Supply{*validSupplyDomain(), *validSupplyDomain()},
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "success_returns_empty_list",
			query: "?page=1&pageSize=10",
			mockSetup: func(svc *domainmocks.SupplyService) {
				svc.EXPECT().
					List(mock.Anything, mock.AnythingOfType("*domain.ListSupplyParams")).
					Return(&domain.PaginatorResponse[domain.Supply]{
						Items: []domain.Supply{},
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "success_uses_default_params_when_no_query",
			query: "",
			mockSetup: func(svc *domainmocks.SupplyService) {
				svc.EXPECT().
					List(mock.Anything, mock.AnythingOfType("*domain.ListSupplyParams")).
					Return(nil, errors.New("internal error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := domainmocks.NewSupplyService(t)
			tt.mockSetup(svc)

			req := httptest.NewRequest(http.MethodGet, "/supplies"+tt.query, nil)
			rec := httptest.NewRecorder()
			setupRouter(svc).ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestList_ResponseBody(t *testing.T) {
	supply1 := validSupplyDomain()
	supply2 := validSupplyDomain()

	svc := domainmocks.NewSupplyService(t)
	svc.EXPECT().
		List(mock.Anything, mock.AnythingOfType("*domain.ListSupplyParams")).
		Return(&domain.PaginatorResponse[domain.Supply]{
			Items: []domain.Supply{*supply1, *supply2},
		}, nil)

	req := httptest.NewRequest(http.MethodGet, "/supplies?page=1&pageSize=10", nil)
	rec := httptest.NewRecorder()
	setupRouter(svc).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]any
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))

	items, ok := body["items"].([]any)
	assert.True(t, ok)
	assert.Len(t, items, 2)

	for _, item := range items {
		entry := item.(map[string]any)
		assert.NotEmpty(t, entry["id"])
		assert.Equal(t, "Brake Pad", entry["name"])
		assert.Equal(t, "High performance brake pad", entry["description"])
		assert.Equal(t, decimal.NewFromFloat(49.99).String(), entry["unit_price"].(string))
	}
}

func TestList_ResponseBody_Empty(t *testing.T) {
	svc := domainmocks.NewSupplyService(t)
	svc.EXPECT().
		List(mock.Anything, mock.AnythingOfType("*domain.ListSupplyParams")).
		Return(&domain.PaginatorResponse[domain.Supply]{
			Items: []domain.Supply{},
		}, nil)

	req := httptest.NewRequest(http.MethodGet, "/supplies?page=1&pageSize=10", nil)
	rec := httptest.NewRecorder()
	setupRouter(svc).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]any
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))

	items, ok := body["items"].([]any)
	assert.True(t, ok)
	assert.Empty(t, items)
}

// -------------------------
// Update
// -------------------------

func TestUpdate(t *testing.T) {
	validID := uuid.New().String()

	tests := []struct {
		name           string
		id             string
		payload        any
		mockSetup      func(*domainmocks.SupplyService)
		expectedStatus int
	}{
		{
			name:    "success_updates_supply",
			id:      validID,
			payload: validSupplyPayload(),
			mockSetup: func(svc *domainmocks.SupplyService) {
				svc.EXPECT().
					Update(mock.Anything, mock.AnythingOfType("*domain.Supply")).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "service_error_returns_internal_error",
			id:      validID,
			payload: validSupplyPayload(),
			mockSetup: func(svc *domainmocks.SupplyService) {
				svc.EXPECT().
					Update(mock.Anything, mock.AnythingOfType("*domain.Supply")).
					Return(errors.New("internal error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "invalid_payload_returns_bad_request",
			id:             validID,
			payload:        "invalid-json",
			mockSetup:      func(svc *domainmocks.SupplyService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := domainmocks.NewSupplyService(t)
			tt.mockSetup(svc)

			url := fmt.Sprintf("/supplies/%s", tt.id)
			req := httptest.NewRequest(http.MethodPut, url, toJSON(t, tt.payload))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			setupRouter(svc).ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestUpdate_ResponseBody(t *testing.T) {
	validID := uuid.New().String()

	svc := domainmocks.NewSupplyService(t)
	svc.EXPECT().
		Update(mock.Anything, mock.AnythingOfType("*domain.Supply")).
		Return(nil)

	url := fmt.Sprintf("/supplies/%s", validID)
	req := httptest.NewRequest(http.MethodPut, url, toJSON(t, validSupplyPayload()))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	setupRouter(svc).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]any
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, validID, body["id"])
	assert.Equal(t, "Brake Pad", body["name"])
}

func TestDeleteWorkHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := domainmocks.NewSupplyService(t)
	h := handler.HttpHandler(mockService)

	mockService.EXPECT().Delete(mock.Anything, "supply-id").Return(nil)

	router := gin.New()
	router.DELETE("/supplies/:id", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/supplies/supply-id", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestDeleteWorkHandler_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := domainmocks.NewSupplyService(t)
	h := handler.HttpHandler(mockService)

	mockService.EXPECT().Delete(mock.Anything, "work-id").Return(errors.New("internal error"))

	router := gin.New()
	router.DELETE("/works/:id", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/works/work-id", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
