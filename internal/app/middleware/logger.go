package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/web"
)

const maxRequestIDLength = 128

// Logger attaches a request-scoped logger tagged with request_id, sets the
// X-Request-Id header on the request (so it reaches the upstream) and on the
// response, and writes one JSON line when the request completes.
//
// The query string is not logged: it can carry customer documents.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, requestID := logger.Request(c.Request.Context(), acceptedRequestID(c.GetHeader(web.RequestIDHeader)))
		c.Request = c.Request.WithContext(ctx)
		c.Request.Header.Set(web.RequestIDHeader, requestID)
		c.Writer.Header().Set(web.RequestIDHeader, requestID)

		start := time.Now()
		c.Next()

		logger.Of(ctx).Info("request completed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("ip", c.ClientIP()),
			zap.String("upstream", upstreamLabel(c)),
			zap.Int("status", c.Writer.Status()),
			zap.Int64("duration_ms", time.Since(start).Milliseconds()),
		)
	}
}

// acceptedRequestID returns the incoming id when it is short and made of safe
// characters, so a client cannot inject line breaks or huge values into the
// logs. Otherwise it returns "" and a new id is generated.
func acceptedRequestID(id string) string {
	if len(id) > maxRequestIDLength {
		return ""
	}
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		case r == '-', r == '_', r == '.', r == ':':
		default:
			return ""
		}
	}
	return id
}

// upstreamLabel names the upstream that handled the request, or "gateway" for
// the gateway's own endpoints.
func upstreamLabel(c *gin.Context) string {
	if upstream := c.GetString("upstream"); upstream != "" {
		return upstream
	}
	return "gateway"
}
