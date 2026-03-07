package ping_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	handler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/ping"
	appPing "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/ping"
)

func TestPingHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	srvc := appPing.Service()
	h := handler.NewHandler(srvc)

	router := gin.New()
	router.GET("/ping", h.Ping)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	expected := `{"message":"pong"}`
	if rec.Body.String() != expected {
		t.Errorf("expected body '%s', got '%s'", expected, rec.Body.String())
	}
}
