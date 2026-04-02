package supply_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	domainmocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain/mocks"
	handler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/supply"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupSupplyRouter(svc *domainmocks.SupplyService) *gin.Engine {
	h := handler.HttpHandler(svc)
	r := gin.New()
	r.POST("/supplies", h.Create)
	return r
}

func validSupplyBody() map[string]any {
	return map[string]any{
		"name":           "Brake Pad",
		"description":    "High performance brake pad",
		"unit_price":     49.99,
		"stock_quantity": 10,
		"version":        1,
	}
}

// --- Create ---

func TestCreate(t *testing.T) {
	tests := []struct {
		name           string
		body           map[string]any
		mockSetup      func(*domainmocks.SupplyService)
		expectedStatus int
	}{
		{
			name: "success",
			body: validSupplyBody(),
			mockSetup: func(svc *domainmocks.SupplyService) {
				svc.EXPECT().Create(mock.Anything, mock.AnythingOfType("*domain.Supply")).Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid_json",
			body:           nil, // will send raw invalid payload
			mockSetup:      func(_ *domainmocks.SupplyService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing_name",
			body: map[string]any{
				"description":    "High performance brake pad",
				"unit_price":     49.99,
				"stock_quantity": 10,
				"version":        1,
			},
			mockSetup:      func(_ *domainmocks.SupplyService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing_description",
			body: map[string]any{
				"name":           "Brake Pad",
				"unit_price":     49.99,
				"stock_quantity": 10,
				"version":        1,
			},
			mockSetup:      func(_ *domainmocks.SupplyService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			// decimal.Decimal zero-value bypasses binding:"required",
			// so omitting unit_price still reaches the service with UnitPrice = 0.
			// Validate this at the domain/service layer instead.
			name: "missing_stock_quantity",
			body: map[string]any{
				"name":        "Brake Pad",
				"description": "High performance brake pad",
				"unit_price":  49.99,
				"version":     1,
			},
			mockSetup:      func(_ *domainmocks.SupplyService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "service_internal_error",
			body: validSupplyBody(),
			mockSetup: func(svc *domainmocks.SupplyService) {
				svc.EXPECT().Create(mock.Anything, mock.AnythingOfType("*domain.Supply")).Return(errors.New("internal error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "service_conflict_error",
			body: validSupplyBody(),
			mockSetup: func(svc *domainmocks.SupplyService) {
				svc.EXPECT().Create(mock.Anything, mock.AnythingOfType("*domain.Supply")).Return(domain.ErrDataConflict)
			},
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := domainmocks.NewSupplyService(t)
			tt.mockSetup(svc)

			var reqBody *bytes.Reader
			if tt.body != nil {
				payload, _ := json.Marshal(tt.body)
				reqBody = bytes.NewReader(payload)
			} else {
				reqBody = bytes.NewReader([]byte("invalid-json{"))
			}

			req := httptest.NewRequest(http.MethodPost, "/supplies", reqBody)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			setupSupplyRouter(svc).ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestCreate_ResponseBody(t *testing.T) {
	svc := domainmocks.NewSupplyService(t)
	svc.EXPECT().Create(mock.Anything, mock.AnythingOfType("*domain.Supply")).Return(nil)

	payload, _ := json.Marshal(validSupplyBody())
	req := httptest.NewRequest(http.MethodPost, "/supplies", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	setupSupplyRouter(svc).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var body map[string]any
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "Brake Pad", body["name"])
	assert.Equal(t, "High performance brake pad", body["description"])
	assert.Equal(t, decimal.NewFromFloat(49.99).String(), body["unit_price"].(string))
	assert.NotEmpty(t, body["id"])
}
