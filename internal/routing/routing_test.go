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
)

func TestSetupRouter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockPing := mocks.NewPingHandler(t)
	mockUser := mocks.NewUserHandler(t)
	mockCustomer := mocks.NewCustomerHandler(t)
	mockWork := mocks.NewWorkHandler(t)

	mockPing.EXPECT().Ping().Return(func(c *gin.Context) { c.Status(goHttp.StatusOK) })
	mockUser.EXPECT().Create().Return(func(c *gin.Context) { c.Status(goHttp.StatusCreated) })
	mockCustomer.EXPECT().Create().Return(func(c *gin.Context) { c.Status(goHttp.StatusCreated) })
	mockWork.EXPECT().Create().Return(func(c *gin.Context) { c.Status(goHttp.StatusCreated) })
	mockWork.EXPECT().List().Return(func(c *gin.Context) { c.Status(goHttp.StatusOK) })
	mockWork.EXPECT().Update().Return(func(c *gin.Context) { c.Status(goHttp.StatusOK) })
	mockWork.EXPECT().Delete().Return(func(c *gin.Context) { c.Status(goHttp.StatusNoContent) })

	c := &http.HandlersWrapper{
		PingHandler:     mockPing,
		UserHandler:     mockUser,
		CustomerHandler: mockCustomer,
		WorkHandler:     mockWork,
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
		{goHttp.MethodGet, "/ping", goHttp.StatusOK},
		{goHttp.MethodPost, "/v1/auth/register", goHttp.StatusCreated},
		{goHttp.MethodPost, "/v1/customers", goHttp.StatusCreated},
		{goHttp.MethodPost, "/v1/works", goHttp.StatusCreated},
		{goHttp.MethodGet, "/v1/works", goHttp.StatusOK},
		{goHttp.MethodPut, "/v1/works/1", goHttp.StatusOK},
		{goHttp.MethodDelete, "/v1/works/1", goHttp.StatusNoContent},
		{goHttp.MethodGet, "/swagger.yaml", goHttp.StatusOK},
		{goHttp.MethodGet, "/swagger/index.html", goHttp.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			assert.Equal(t, tt.status, rec.Code)
		})
	}
}
