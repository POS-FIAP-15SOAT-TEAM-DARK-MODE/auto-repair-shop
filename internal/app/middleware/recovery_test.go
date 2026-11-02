package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRecovery_PanicBecomesGenericInternalError(t *testing.T) {
	r := gin.New()
	r.Use(Recovery())
	r.GET("/boom", func(c *gin.Context) {
		panic("secret internal detail")
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/boom", nil))

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.JSONEq(t, `{"code":500,"errors":["internal server error"]}`, rec.Body.String())
	assert.NotContains(t, rec.Body.String(), "secret")
}

func TestRecovery_PassesThroughWithoutPanic(t *testing.T) {
	r := gin.New()
	r.Use(Recovery())
	r.GET("/ok", func(c *gin.Context) { c.Status(http.StatusTeapot) })

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ok", nil))

	assert.Equal(t, http.StatusTeapot, rec.Code)
}
