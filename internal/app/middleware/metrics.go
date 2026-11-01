package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/metrics"
)

// knownMethods bounds the method label: any other method is reported as OTHER.
var knownMethods = map[string]bool{
	"GET": true, "POST": true, "PUT": true, "PATCH": true, "DELETE": true,
	"HEAD": true, "OPTIONS": true,
}

// Metrics records request latency. The route label is the matched route
// template ("/v1/works*", "/docs/*path"), or the one the gateway set for a
// proxied request, and "unmatched" otherwise.
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		route := c.FullPath()
		if route == "" {
			route = c.GetString("route")
		}
		if route == "" {
			route = "unmatched"
		}

		method := c.Request.Method
		if !knownMethods[method] {
			method = "OTHER"
		}

		metrics.HTTPRequestDuration.WithLabelValues(
			method,
			route,
			upstreamLabel(c),
			strconv.Itoa(c.Writer.Status()),
		).Observe(time.Since(start).Seconds())
	}
}
