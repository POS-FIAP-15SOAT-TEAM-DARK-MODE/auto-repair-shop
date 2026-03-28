package routing_test

import (
	goHttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/http"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/http/mocks"
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

	// Work (Service Order)
	mockWork.EXPECT().Create(mock.Anything).RunAndReturn(func(c *gin.Context) { c.Status(goHttp.StatusCreated) })
	mockWork.EXPECT().List(mock.Anything).RunAndReturn(func(c *gin.Context) { c.Status(goHttp.StatusOK) })
	mockWork.EXPECT().Update(mock.Anything).RunAndReturn(func(c *gin.Context) { c.Status(goHttp.StatusOK) })
	mockWork.EXPECT().Delete(mock.Anything).RunAndReturn(func(c *gin.Context) { c.Status(goHttp.StatusNoContent) })

	// Vehicle
	mockVehicle.EXPECT().Create(mock.Anything).RunAndReturn(func(c *gin.Context) { c.Status(goHttp.StatusCreated) })

	c := &http.HandlersWrapper{
		PingHandler:     mockPing,
		UserHandler:     mockUser,
		CustomerHandler: mockCustomer,
		WorkHandler:     mockWork,
		VehicleHandler:  mockVehicle,
	}

	m := &http.Middlewares{
		"test_mw": func(c *gin.Context) { c.Next() },
	}

	router := routing.SetupRouter(c, m)

	tests := []struct {
		method string
		path   string
		status int
	}{
		// Ping
		{goHttp.MethodGet, "/ping", goHttp.StatusOK},

		// User
		{goHttp.MethodPost, "/v1/auth/register", goHttp.StatusCreated},
		{goHttp.MethodPost, "/v1/auth/login", goHttp.StatusOK},

		// Customer
		{goHttp.MethodPost, "/v1/customers", goHttp.StatusCreated},
		{goHttp.MethodGet, "/v1/customers/:id", goHttp.StatusOK},
		{goHttp.MethodGet, "/v1/customers", goHttp.StatusOK},
		{goHttp.MethodPut, "/v1/customers/:id", goHttp.StatusOK},
		{goHttp.MethodDelete, "/v1/customers/:id", goHttp.StatusNoContent},

		// Work (Service Order)
		{goHttp.MethodPost, "/v1/works", goHttp.StatusCreated},
		{goHttp.MethodGet, "/v1/works", goHttp.StatusOK},
		{goHttp.MethodPut, "/v1/works/:id", goHttp.StatusOK},
		{goHttp.MethodDelete, "/v1/works/:id", goHttp.StatusNoContent},

		// Vehicle
		{goHttp.MethodPost, "/v1/vehicle", goHttp.StatusCreated},

		// Swagger UI (based on mountSwaggerUI in routing.go)
		{goHttp.MethodGet, "/swagger.yaml", goHttp.StatusOK},
		{goHttp.MethodGet, "/swagger/index.html", goHttp.StatusOK},
	}

	// Replace :id with a concrete value in test URLs for request generation
	for _, tt := range tests {
		path := tt.path
		// Replace :id with "1" for tests
		path = replacePathParamsWithSampleValues(path)
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, path, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			assert.Equal(t, tt.status, rec.Code)
		})
	}
}

// Helper to replace :id and similar path params with a dummy value.
func replacePathParamsWithSampleValues(path string) string {
	// Only replacing ":id" with "1"
	// If more params are needed in the future, extend this.
	return replace(path, ":id", "1")
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
