package supply_test

import (
	"encoding/json"
	"errors"
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

func setupSupplyRouterWithList(svc *domainmocks.SupplyService) *gin.Engine {
	h := handler.HttpHandler(svc)
	r := gin.New()
	r.GET("/supplies", h.List)
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

// --- List ---

func TestList(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*domainmocks.SupplyService)
		expectedStatus int
	}{
		{
			name: "success_returns_supplies",
			mockSetup: func(svc *domainmocks.SupplyService) {
				svc.EXPECT().
					List(mock.Anything).
					Return([]*domain.Supply{validSupplyDomain(), validSupplyDomain()}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "success_returns_empty_list",
			mockSetup: func(svc *domainmocks.SupplyService) {
				svc.EXPECT().
					List(mock.Anything).
					Return([]*domain.Supply{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "service_internal_error",
			mockSetup: func(svc *domainmocks.SupplyService) {
				svc.EXPECT().
					List(mock.Anything).
					Return(nil, errors.New("internal error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := domainmocks.NewSupplyService(t)
			tt.mockSetup(svc)

			req := httptest.NewRequest(http.MethodGet, "/supplies", nil)
			rec := httptest.NewRecorder()
			setupSupplyRouterWithList(svc).ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestList_ResponseBody(t *testing.T) {
	supply1 := validSupplyDomain()
	supply2 := validSupplyDomain()

	svc := domainmocks.NewSupplyService(t)
	svc.EXPECT().
		List(mock.Anything).
		Return([]*domain.Supply{supply1, supply2}, nil)

	req := httptest.NewRequest(http.MethodGet, "/supplies", nil)
	rec := httptest.NewRecorder()
	setupSupplyRouterWithList(svc).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var body []map[string]any
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Len(t, body, 2)

	for _, item := range body {
		assert.NotEmpty(t, item["id"])
		assert.Equal(t, "Brake Pad", item["name"])
		assert.Equal(t, "High performance brake pad", item["description"])
		assert.Equal(t, decimal.NewFromFloat(49.99).String(), item["unit_price"].(string))
	}
}

func TestList_ResponseBody_Empty(t *testing.T) {
	svc := domainmocks.NewSupplyService(t)
	svc.EXPECT().
		List(mock.Anything).
		Return([]*domain.Supply{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/supplies", nil)
	rec := httptest.NewRecorder()
	setupSupplyRouterWithList(svc).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var body []map[string]any
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Empty(t, body)
}
