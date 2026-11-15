package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/web"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestLogger_GeneratesRequestIDWhenMissing(t *testing.T) {
	var seenByHandler string
	r := gin.New()
	r.Use(Logger())
	r.GET("/x", func(c *gin.Context) {
		seenByHandler = c.Request.Header.Get(web.RequestIDHeader)
		c.Status(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))

	id := rec.Header().Get(web.RequestIDHeader)
	require.NotEmpty(t, id)
	assert.Equal(t, id, seenByHandler, "the id is propagated to the upstream request")
}

func TestLogger_KeepsValidIncomingRequestID(t *testing.T) {
	r := gin.New()
	r.Use(Logger())
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set(web.RequestIDHeader, "abc-123_x.y:z")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, "abc-123_x.y:z", rec.Header().Get(web.RequestIDHeader))
}

func TestAcceptedRequestID(t *testing.T) {
	assert.Equal(t, "", acceptedRequestID(""))
	assert.Equal(t, "req-1", acceptedRequestID("req-1"))
	assert.Equal(t, "", acceptedRequestID("line\nbreak"), "control characters are rejected")
	assert.Equal(t, "", acceptedRequestID("has space"))
	assert.Equal(t, "", acceptedRequestID(strings.Repeat("a", maxRequestIDLength+1)), "too long")
	assert.Equal(t, strings.Repeat("a", maxRequestIDLength), acceptedRequestID(strings.Repeat("a", maxRequestIDLength)))
}

func TestUpstreamLabel_DefaultsToGateway(t *testing.T) {
	r := gin.New()
	var labels []string
	r.Use(func(c *gin.Context) {
		c.Next()
		labels = append(labels, upstreamLabel(c))
	})
	r.GET("/own", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/proxied", func(c *gin.Context) {
		c.Set("upstream", "catalog")
		c.Status(http.StatusOK)
	})

	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/own", nil))
	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/proxied", nil))

	assert.Equal(t, []string{"gateway", "catalog"}, labels)
}
