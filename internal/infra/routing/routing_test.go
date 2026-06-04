package routing_test

import (
	"io"
	goHttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	authAdapters "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/auth/adapters"
	authDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/auth/domain"
	authMocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/auth/interfaces/mocks"
	customerAdapters "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/customer/adapters"
	customerMocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/customer/interfaces/mocks"
	newHttp "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/http"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/http/mocks"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/routing"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/auth"
	soAdapters "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order/adapters"
	soMocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order/interfaces/mocks"
	soHistoryAdapters "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order_history/adapters"
	soHistoryMocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order_history/interfaces/mocks"
	supplyAdapters "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/supply/adapters"
	suppliesMocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/supply/interfaces/mocks"
	vehicleAdapters "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/vehicle/adapters"
	vehiclesMocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/vehicle/interfaces/mocks"
	workAdapters "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/work/adapters"
	workMocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/work/interfaces/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSetupRouter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockPing := mocks.NewPingHandler(t)
	mockCustomer := customerMocks.NewCustomerHTTPController(t)
	mockWork := workMocks.NewWorkHTTPController(t)
	mockSO := soMocks.NewServiceOrderHTTPController(t)
	mockSOHistory := soHistoryMocks.NewServiceOrderHistoryHTTPController(t)

	// MIGRATED
	mockVehicle := vehiclesMocks.NewVehicleHTTPController(t)
	mockSupply := suppliesMocks.NewSupplyHTTPController(t)
	mockAuth := authMocks.NewAuthHTTPController(t)

	// Set up only the handlers/routes that exist in internal/infra/routing/routing.go

	// Ping
	mockPing.EXPECT().Ping(mock.Anything).RunAndReturn(func(c *gin.Context) { c.Status(goHttp.StatusOK) })

	// Auth
	mockAuth.EXPECT().Register(mock.Anything, mock.AnythingOfType("*http.Request")).Return(authAdapters.UserResponse{}, nil)
	mockAuth.EXPECT().Login(mock.Anything, mock.AnythingOfType("*http.Request")).Return(authAdapters.LoginResponse{}, nil)
	mockAuth.EXPECT().ChangeRole(mock.Anything, mock.AnythingOfType("*http.Request")).Return(nil)

	// Customer
	mockCustomer.EXPECT().Create(mock.Anything, mock.AnythingOfType("*http.Request")).Return(customerAdapters.CustomerResponse{}, nil)
	mockCustomer.EXPECT().GetByID(mock.Anything, mock.AnythingOfType("*http.Request")).Return(customerAdapters.CustomerResponse{}, nil)
	mockCustomer.EXPECT().GetByDocument(mock.Anything, mock.AnythingOfType("*http.Request")).Return(customerAdapters.CustomerResponse{}, nil)
	mockCustomer.EXPECT().Update(mock.Anything, mock.AnythingOfType("*http.Request")).Return(customerAdapters.CustomerResponse{}, nil)
	mockCustomer.EXPECT().Delete(mock.Anything, mock.AnythingOfType("*http.Request")).Return(nil)

	// Work
	mockWork.EXPECT().Create(mock.Anything, mock.AnythingOfType("*http.Request")).Return(workAdapters.WorkResponse{}, nil)
	mockWork.EXPECT().List(mock.Anything, mock.AnythingOfType("*http.Request")).Return(workAdapters.PaginatedWorkResponse{}, nil)
	mockWork.EXPECT().Update(mock.Anything, mock.AnythingOfType("*http.Request")).Return(workAdapters.WorkResponse{}, nil)
	mockWork.EXPECT().Delete(mock.Anything, mock.AnythingOfType("*http.Request")).Return(nil)

	// Vehicle
	mockVehicle.EXPECT().Create(mock.Anything, mock.AnythingOfType("*http.Request")).Return(vehicleAdapters.VehicleResponse{}, nil)
	mockVehicle.EXPECT().Edit(mock.Anything, mock.AnythingOfType("*http.Request")).Return(vehicleAdapters.VehicleResponse{}, nil)
	mockVehicle.EXPECT().Delete(mock.Anything, mock.AnythingOfType("*http.Request")).Return(nil)
	mockVehicle.EXPECT().List(mock.Anything, mock.AnythingOfType("*http.Request")).Return(vehicleAdapters.PaginatedVehicleResponse{}, nil)

	// Supply
	mockSupply.EXPECT().Create(mock.Anything, mock.AnythingOfType("*http.Request")).Return(supplyAdapters.SupplyResponse{}, nil)
	mockSupply.EXPECT().Update(mock.Anything, mock.AnythingOfType("*http.Request")).Return(supplyAdapters.SupplyResponse{}, nil)
	mockSupply.EXPECT().Delete(mock.Anything, mock.AnythingOfType("*http.Request")).Return(nil)
	mockSupply.EXPECT().List(mock.Anything, mock.AnythingOfType("*http.Request")).Return(supplyAdapters.PaginatedSupplyResponse{}, nil)

	// Service Order
	mockSO.EXPECT().Create(mock.Anything, mock.AnythingOfType("*http.Request")).Return(soAdapters.SOResponse{}, nil)
	mockSO.EXPECT().List(mock.Anything, mock.AnythingOfType("*http.Request")).Return(soAdapters.PaginatedSOResponse{}, nil)
	mockSO.EXPECT().Finish(mock.Anything, mock.AnythingOfType("*http.Request")).Return(nil)
	mockSO.EXPECT().SendToCustomerApproval(mock.Anything, mock.AnythingOfType("*http.Request")).Return(nil)
	mockSO.EXPECT().SendToDiagnosis(mock.Anything, mock.AnythingOfType("*http.Request")).Return(nil)
	mockSO.EXPECT().Accept(mock.Anything, mock.AnythingOfType("*http.Request")).Return(nil)
	mockSO.EXPECT().Reject(mock.Anything, mock.AnythingOfType("*http.Request")).Return(nil)
	mockSO.EXPECT().Deliver(mock.Anything, mock.AnythingOfType("*http.Request")).Return(nil)
	mockSO.EXPECT().Cancel(mock.Anything, mock.AnythingOfType("*http.Request")).Return(nil)
	mockSO.EXPECT().GetFullByID(mock.Anything, mock.AnythingOfType("*http.Request")).Return(soAdapters.SODetailResponse{}, nil)
	mockSO.EXPECT().GetStatus(mock.Anything, mock.AnythingOfType("*http.Request")).Return(soAdapters.SOResponse{}, nil)

	// Service Order Works
	mockSO.EXPECT().GetWorks(mock.Anything, mock.AnythingOfType("*http.Request")).Return(soAdapters.ListResponse[soAdapters.WorkItemResponse]{}, nil)
	mockSO.EXPECT().AddWork(mock.Anything, mock.AnythingOfType("*http.Request")).Return(nil)
	mockSO.EXPECT().DeleteWork(mock.Anything, mock.AnythingOfType("*http.Request")).Return(nil)

	// Service Order Supplies
	mockSO.EXPECT().GetSupplies(mock.Anything, mock.AnythingOfType("*http.Request")).Return(soAdapters.ListResponse[soAdapters.SupplyItemResponse]{}, nil)
	mockSO.EXPECT().AddSupplies(mock.Anything, mock.AnythingOfType("*http.Request")).Return(nil)
	mockSO.EXPECT().DeleteSupply(mock.Anything, mock.AnythingOfType("*http.Request")).Return(nil)
	mockSO.EXPECT().GetAverageExecutionTime(mock.Anything, mock.AnythingOfType("*http.Request")).Return([]soAdapters.WorkExecutionTimeResponse{}, nil)

	// Service Order History
	mockSOHistory.EXPECT().GetHistoryByID(mock.Anything, mock.AnythingOfType("*http.Request")).Return([]soHistoryAdapters.SOHistoryResponse{}, nil)

	c := &newHttp.HandlersWrapper{
		PingHandler:                mockPing,
		UserHandler:                mockAuth,
		CustomerHandler:            mockCustomer,
		WorkHandler:                mockWork,
		VehicleHandler:             mockVehicle,
		SupplyHandler:              mockSupply,
		ServiceOrderHandler:        mockSO,
		ServiceOrderHistoryHandler: mockSOHistory,
	}

	m := &newHttp.Middlewares{
		"test_mw": func(c *gin.Context) { c.Next() },
	}

	router := routing.SetupRouter(c, m)

	// create a long-lived token that includes all roles so tests can call protected routes
	allRoles := []authDomain.Role{authDomain.ADMIN, authDomain.ATTENDANT, authDomain.MECHANIC, authDomain.CUSTOMER}
	testToken, err := auth.GenerateToken("test-user", allRoles, time.Now().Add(100*365*24*time.Hour))
	if err != nil {
		t.Fatalf("failed to generate test token: %v", err)
	}

	tests := []struct {
		method       string
		path         string
		status       int
		body         string
		useOnlyError bool
		useHandler   bool
	}{
		// Ping
		{method: goHttp.MethodGet, path: "/ping", status: goHttp.StatusOK, body: ""},

		// User
		{method: goHttp.MethodPost, path: "/v1/auth/register", status: goHttp.StatusCreated, body: ""},
		{method: goHttp.MethodPost, path: "/v1/auth/login", status: goHttp.StatusOK, body: ""},
		{method: goHttp.MethodPatch, path: "/v1/users/:id/role", status: goHttp.StatusNoContent, body: `{"role":"MECHANIC"}`},

		// Customer
		{method: goHttp.MethodPost, path: "/v1/customers", status: goHttp.StatusCreated, body: ""},
		{method: goHttp.MethodGet, path: "/v1/customers/:id", status: goHttp.StatusOK, body: ""},
		{method: goHttp.MethodGet, path: "/v1/customers", status: goHttp.StatusOK, body: ""},
		{method: goHttp.MethodPut, path: "/v1/customers/:id", status: goHttp.StatusOK, body: ""},
		{method: goHttp.MethodDelete, path: "/v1/customers/:id", status: goHttp.StatusNoContent, body: ""},

		// Work (catalog)
		{method: goHttp.MethodPost, path: "/v1/works", status: goHttp.StatusCreated, body: ""},
		{method: goHttp.MethodGet, path: "/v1/works", status: goHttp.StatusOK, body: ""},
		{method: goHttp.MethodPut, path: "/v1/works/:id", status: goHttp.StatusOK, body: ""},
		{method: goHttp.MethodDelete, path: "/v1/works/:id", status: goHttp.StatusNoContent, body: ""},

		// Vehicle
		{method: goHttp.MethodGet, path: "/v1/vehicles", status: goHttp.StatusOK, body: "", useHandler: true},
		{method: goHttp.MethodPost, path: "/v1/vehicles", status: goHttp.StatusCreated, body: "", useHandler: true},
		{method: goHttp.MethodPut, path: "/v1/vehicles/:id", status: goHttp.StatusOK, body: "", useHandler: true},
		{method: goHttp.MethodDelete, path: "/v1/vehicles/:id", status: goHttp.StatusNoContent, body: "", useHandler: true, useOnlyError: true},

		// Supply
		{method: goHttp.MethodPost, path: "/v1/supplies", status: goHttp.StatusCreated, body: ""},
		{method: goHttp.MethodGet, path: "/v1/supplies", status: goHttp.StatusOK, body: ""},
		{method: goHttp.MethodPut, path: "/v1/supplies/:id", status: goHttp.StatusOK, body: ""},
		{method: goHttp.MethodDelete, path: "/v1/supplies/:id", status: goHttp.StatusNoContent, body: ""},

		// Service Order
		{method: goHttp.MethodPost, path: "/v1/service-order", status: goHttp.StatusCreated, body: ""},
		{method: goHttp.MethodGet, path: "/v1/service-order", status: goHttp.StatusOK, body: ""},
		{method: goHttp.MethodGet, path: "/v1/service-order/:id", status: goHttp.StatusOK, body: `{"id":"123"}`},
		{method: goHttp.MethodPut, path: "/v1/service-order/:id/finish", status: goHttp.StatusNoContent, body: ""},
		{method: goHttp.MethodGet, path: "/v1/service-order/:id/status", status: goHttp.StatusOK, body: `{"id":"123". "status":"in_progress"}`},
		{method: goHttp.MethodPut, path: "/v1/service-order/:id/start-diagnosis", status: goHttp.StatusNoContent, body: ""},
		{method: goHttp.MethodPut, path: "/v1/service-order/:id/send", status: goHttp.StatusNoContent, body: ""},
		{method: goHttp.MethodPut, path: "/v1/service-order/:id/accept", status: goHttp.StatusNoContent, body: ""},
		{method: goHttp.MethodPut, path: "/v1/service-order/:id/reject", status: goHttp.StatusNoContent, body: ""},
		{method: goHttp.MethodPut, path: "/v1/service-order/:id/deliver", status: goHttp.StatusNoContent, body: ""},
		{method: goHttp.MethodPut, path: "/v1/service-order/:id/cancel", status: goHttp.StatusOK, body: ""},

		// Service Order Works
		{method: goHttp.MethodGet, path: "/v1/service-order/:id/works", status: goHttp.StatusOK, body: ""},
		{method: goHttp.MethodPost, path: "/v1/service-order/:id/works", status: goHttp.StatusNoContent, body: `{"services":["work-id-1"]}`},
		{method: goHttp.MethodDelete, path: "/v1/service-order/:id/works/:serviceId", status: goHttp.StatusNoContent, body: ""},

		// Service Order Supplies
		{method: goHttp.MethodGet, path: "/v1/service-order/:id/supplies", status: goHttp.StatusOK, body: ""},
		{method: goHttp.MethodPost, path: "/v1/service-order/:id/supplies", status: goHttp.StatusNoContent, body: `{"supplies":["supply-id-1"]}`},
		{method: goHttp.MethodDelete, path: "/v1/service-order/:id/supplies/:serviceId", status: goHttp.StatusNoContent, body: ""},

		// Service Order History
		{method: goHttp.MethodGet, path: "/v1/service-order/:id/history", status: goHttp.StatusOK, body: ""},

		// Reports
		{method: goHttp.MethodGet, path: "/v1/reports/average-execution-time", status: goHttp.StatusOK, body: ""},

		// Swagger UI (based on mountSwaggerUI in routing.go)
		{method: goHttp.MethodGet, path: "/swagger.yaml", status: goHttp.StatusOK, body: ""},
		{method: goHttp.MethodGet, path: "/swagger/index.html", status: goHttp.StatusOK, body: ""},
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
