package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// durationSeries returns the label sets recorded for http_request_duration_seconds.
func durationSeries(t *testing.T) []map[string]string {
	t.Helper()
	families, err := prometheus.DefaultGatherer.Gather()
	require.NoError(t, err)

	var series []map[string]string
	for _, fam := range families {
		if fam.GetName() != "http_request_duration_seconds" {
			continue
		}
		for _, m := range fam.GetMetric() {
			series = append(series, labelMap(m))
		}
	}
	return series
}

func labelMap(m *dto.Metric) map[string]string {
	out := make(map[string]string, len(m.GetLabel()))
	for _, l := range m.GetLabel() {
		out[l.GetName()] = l.GetValue()
	}
	return out
}

func hasSeries(series []map[string]string, want map[string]string) bool {
	for _, s := range series {
		match := true
		for k, v := range want {
			if s[k] != v {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func TestMetrics_LabelsUseRouteTemplateAndUpstream(t *testing.T) {
	r := gin.New()
	r.Use(Metrics())
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/docs/*path", func(c *gin.Context) {
		c.Set("upstream", "catalog")
		c.Status(http.StatusNotFound)
	})
	r.NoRoute(func(c *gin.Context) {
		if c.Request.URL.Path == "/v1/works/1" {
			c.Set("route", "/v1/works*")
			c.Set("upstream", "catalog")
			c.Status(http.StatusBadGateway)
			return
		}
		c.Status(http.StatusNotFound)
	})

	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/ping", nil))
	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/docs/catalog/swagger.yaml", nil))
	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/v1/works/1", nil))
	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/unknown", nil))
	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("FROBNICATE", "/unknown", nil))

	series := durationSeries(t)
	assert.True(t, hasSeries(series, map[string]string{"method": "GET", "route": "/ping", "upstream": "gateway", "status": "200"}))
	assert.True(t, hasSeries(series, map[string]string{"method": "GET", "route": "/docs/*path", "upstream": "catalog", "status": "404"}))
	assert.True(t, hasSeries(series, map[string]string{"method": "GET", "route": "/v1/works*", "upstream": "catalog", "status": "502"}))
	assert.True(t, hasSeries(series, map[string]string{"method": "GET", "route": "unmatched", "upstream": "gateway", "status": "404"}))
	assert.True(t, hasSeries(series, map[string]string{"method": "OTHER", "route": "unmatched", "upstream": "gateway", "status": "404"}),
		"unknown methods are reported as OTHER to bound the label cardinality")
}
