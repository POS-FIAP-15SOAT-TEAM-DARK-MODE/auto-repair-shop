package routing_test

import (
	"io"
	goHttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/http"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/http/mocks"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/auth"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/routing"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSetupRouter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockPing := mocks.NewPingHandler(t)
	mockUser := mocks.NewUserHandler(t)
	mockCustomer := mocks.NewCustomerHandler(t)
	mockWork := mocks.NewWorkHandler(t)
	mockVehicle := mocks.NewVehicleHandler(t)
	mockSupply := mocks.NewSupplyHandler(t)
	mockSO := mocks.NewServiceOrderHandler(t)
	mockSOHistory := mocks.NewServiceOrderHistoryHandler(t)

	// Set up only the handlers/routes that exist in internal/routing/routing.go

	// Ping
	mockPing.EXPECT().Ping(mock.Anything).RunAndReturn(func(c *gin.Context) { c.Status(goHttp.StatusOK) })

	// User
	mockUser.EXPECT().Create(mock.Anything).RunAndReturn(func(c *gin.Context) { c.Status(goHttp.StatusCreated) })
	mockUser.EXPECT().Login(mock.Anything).RunAndReturn(func(c *gin.Context) { c.Status(goHttp.StatusOK) })

	// Customer
	mockCustomer.EXPECT().Create(mock.Anything).RunAndReturn(func(c *gin.Context) { c.Status(goHttp.StatusCreated) })
	mockCustomer.EXPECT().GetByID(mock.Anything).RunAndReturn(func(c *gin.Context) { c.Status(goHttp.StatusOK) })
	mockCustomer.EXPECT().GetByDocument(mock.Anything).RunAndReturn(func(c *gin.Context) { c.Status(goHttp.StatusOK) })
	mockCustomer.EXPECT().Update(mock.Anything).RunAndReturn(func(c *gin.Context) { c.Status(goHttp.StatusOK) })
	mockCustomer.EXPECT().Delete(mock.Anything).RunAndReturn(func(c *gin.Context) { c.Status(goHttp.StatusNoContent) })

	// Work
	mockWork.EXPECT().Create(mock.Anything).RunAndReturn(func(c *gin.Context) { c.Status(goHttp.StatusCreated) })
	mockWork.EXPECT().List(mock.Anything).RunAndReturn(func(c *gin.Context) { c.Status(goHttp.StatusOK) })
	mockWork.EXPECT().Update(mock.Anything).RunAndReturn(func(c *gin.Context) { c.Status(goHttp.StatusOK) })
	mockWork.EXPECT().Delete(mock.Anything).RunAndReturn(func(c *gin.Context) { c.Status(goHttp.StatusNoContent) })

	// Vehicle
	mockVehicle.EXPECT().Create(mock.Anything).RunAndReturn(func(c *gin.Context) { c.Status(goHttp.StatusCreated) })
	mockVehicle.EXPECT().FindByLicensePlate(mock.Anything).RunAndReturn(func(c *gin.Context) { c.Status(goHttp.StatusOK) })
	mockVehicle.EXPECT().Update(mock.Anything).RunAndReturn(func(c *gin.Context) { c.Status(goHttp.StatusOK) })
	mockVehicle.EXPECT().Delete(mock.Anything).RunAndReturn(func(c *gin.Context) { c.Status(goHttp.StatusNoContent) })

	// Supply
	mockSupply.EXPECT().Create(mock.Anything).RunAndReturn(func(c *gin.Context) { c.Status(goHttp.StatusCreated) })
	mockSupply.EXPECT().List(mock.Anything).RunAndReturn(func(c *gin.Context) { c.Status(goHttp.StatusOK) })
	mockSupply.EXPECT().Update(mock.Anything).RunAndReturn(func(c *gin.Context) { c.Status(goHttp.StatusOK) })
	mockSupply.EXPECT().Delete(mock.Anything).RunAndReturn(func(c *gin.Context) { c.Status(goHttp.StatusNoContent) })

	// Service Order
	mockSO.EXPECT().Create(mock.Anything).RunAndReturn(func(c *gin.Context) { c.Status(goHttp.StatusCreated) })
	mockSO.EXPECT().SendToCustomerApproval(mock.Anything).RunAndReturn(func(c *gin.Context) { c.Status(goHttp.StatusAccepted) })

	// Service Order Works
	mockSO.EXPECT().GetWorks(mock.Anything).RunAndReturn(func(c *gin.Context) { c.Status(goHttp.StatusOK) })
	mockSO.EXPECT().AddWork(mock.Anything).RunAndReturn(func(c *gin.Context) { c.Status(goHttp.StatusNoContent) })
	mockSO.EXPECT().DeleteWork(mock.Anything).RunAndReturn(func(c *gin.Context) { c.Status(goHttp.StatusNoContent) })

	// Service Order Supplies
	mockSO.EXPECT().GetSupplies(mock.Anything).RunAndReturn(func(c *gin.Context) { c.Status(goHttp.StatusOK) })
	mockSO.EXPECT().AddSupplies(mock.Anything).RunAndReturn(func(c *gin.Context) { c.Status(goHttp.StatusNoContent) })
	mockSO.EXPECT().DeleteSupply(mock.Anything).RunAndReturn(func(c *gin.Context) { c.Status(goHttp.StatusNoContent) })

	// Service Order History
	mockSOHistory.EXPECT().GetHistoryByID(mock.Anything).RunAndReturn(func(c *gin.Context) { c.Status(goHttp.StatusOK) })

	c := &http.HandlersWrapper{
		PingHandler:                mockPing,
		UserHandler:                mockUser,
		CustomerHandler:            mockCustomer,
		WorkHandler:                mockWork,
		VehicleHandler:             mockVehicle,
		SupplyHandler:              mockSupply,
		ServiceOrderHandler:        mockSO,
		ServiceOrderHistoryHandler: mockSOHistory,
	}

	m := &http.Middlewares{
		"test_mw": func(c *gin.Context) { c.Next() },
	}

	router := routing.SetupRouter(c, m)

	// create a long-lived token that includes all roles so tests can call protected routes
	allRoles := []domain.Role{domain.ADMIN, domain.ATTENDANT, domain.MECHANIC, domain.CUSTOMER}
	testToken, err := auth.GenerateToken("test-user", allRoles, time.Now().Add(100*365*24*time.Hour))
	if err != nil {
		t.Fatalf("failed to generate test token: %v", err)
	}

	tests := []struct {
		method string
		path   string
		status int
		body   string
	}{
		// Ping
		{goHttp.MethodGet, "/ping", goHttp.StatusOK, ""},

		// User
		{goHttp.MethodPost, "/v1/auth/register", goHttp.StatusCreated, ""},
		{goHttp.MethodPost, "/v1/auth/login", goHttp.StatusOK, ""},

		// Customer
		{goHttp.MethodPost, "/v1/customers", goHttp.StatusCreated, ""},
		{goHttp.MethodGet, "/v1/customers/:id", goHttp.StatusOK, ""},
		{goHttp.MethodGet, "/v1/customers", goHttp.StatusOK, ""},
		{goHttp.MethodPut, "/v1/customers/:id", goHttp.StatusOK, ""},
		{goHttp.MethodDelete, "/v1/customers/:id", goHttp.StatusNoContent, ""},

		// Work (catalog)
		{goHttp.MethodPost, "/v1/works", goHttp.StatusCreated, ""},
		{goHttp.MethodGet, "/v1/works", goHttp.StatusOK, ""},
		{goHttp.MethodPut, "/v1/works/:id", goHttp.StatusOK, ""},
		{goHttp.MethodDelete, "/v1/works/:id", goHttp.StatusNoContent, ""},

		// Vehicle
		{goHttp.MethodGet, "/v1/vehicles", goHttp.StatusOK, ""},
		{goHttp.MethodPost, "/v1/vehicles", goHttp.StatusCreated, ""},
		{goHttp.MethodPut, "/v1/vehicles/:id", goHttp.StatusOK, ""},
		{goHttp.MethodDelete, "/v1/vehicles/:id", goHttp.StatusNoContent, ""},

		// Supply
		{goHttp.MethodPost, "/v1/supplies", goHttp.StatusCreated, ""},
		{goHttp.MethodGet, "/v1/supplies", goHttp.StatusOK, ""},
		{goHttp.MethodPut, "/v1/supplies/:id", goHttp.StatusOK, ""},
		{goHttp.MethodGet, "/v1/supplies", goHttp.StatusOK, ""},
		{goHttp.MethodPut, "/v1/supplies/:id", goHttp.StatusOK, ""},
		{goHttp.MethodDelete, "/v1/supplies/:id", goHttp.StatusNoContent, ""},

		// Service Order
		{goHttp.MethodPost, "/v1/service-order", goHttp.StatusCreated, ""},

		// Service Order Works
		{goHttp.MethodGet, "/v1/service-order/:id/services", goHttp.StatusOK, ""},
		{goHttp.MethodPost, "/v1/service-order/:id/services", goHttp.StatusNoContent, `{"services":["work-id-1"]}`},
		{goHttp.MethodPut, "/v1/service-order/:id/send", goHttp.StatusAccepted, ""},
		{goHttp.MethodDelete, "/v1/service-order/:id/services/:serviceId", goHttp.StatusNoContent, ""},

		// Service Order Supplies
		{goHttp.MethodGet, "/v1/service-order/:id/supplies", goHttp.StatusOK, ""},
		{goHttp.MethodPost, "/v1/service-order/:id/supplies", goHttp.StatusNoContent, `{"supplies":["supply-id-1"]}`},
		{goHttp.MethodDelete, "/v1/service-order/:id/supplies/:serviceId", goHttp.StatusNoContent, ""},

		// Service Order History
		{goHttp.MethodGet, "/v1/service-order/:id/history", goHttp.StatusOK, ""},

		// Swagger UI (based on mountSwaggerUI in routing.go)
		{goHttp.MethodGet, "/swagger.yaml", goHttp.StatusOK, ""},
		{goHttp.MethodGet, "/swagger/index.html", goHttp.StatusOK, ""},
	}

	// Replace :id with a concrete value in test URLs for request generation
	for _, tt := range tests {
		path := tt.path
		// Replace :id with "1" for tests
		path = replacePathParamsWithSampleValues(path)
		t.Run(path, func(t *testing.T) {
			var body io.Reader
			if tt.body != "" {
				body = strings.NewReader(tt.body)
			}
			req := httptest.NewRequest(tt.method, path, body)
			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			// add Authorization header with long-lived token so protected routes pass
			req.Header.Set("Authorization", "Bearer "+testToken)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			assert.Equal(t, tt.status, rec.Code)
		})
	}
}

// Helper to replace :id and similar path params with a dummy value.
func replacePathParamsWithSampleValues(path string) string {
	path = replace(path, ":id", "1")
	path = replace(path, ":serviceId", "1")
	return path
}

// Local replacement so we don't require strings.ReplaceAll (Go 1.12+)
func replace(s, old, new string) string {
	for {
		i := index(s, old)
		if i < 0 {
			break
		}
		s = s[:i] + new + s[i+len(old):]
	}
	return s
}

// Local index function so we don't require strings package
func index(s, sep string) int {
	n := len(sep)
	for i := 0; i+n <= len(s); i++ {
		if s[i:i+n] == sep {
			return i
		}
	}
	return -1
}
