package middleware

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		ctx, reqId := logger.Request(ctx)
		c.Request = c.Request.WithContext(ctx)

		// Write request ID into response headers
		c.Writer.Header().Set("X-Request-ID", reqId)

		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
		}

		logger.Of(ctx).Info("starting request", fields...)

		c.Next()

		fields = append(fields,
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)))

		if len(c.Errors) > 0 {
			for _, e := range c.Errors.Errors() {
				logger.Of(ctx).Error(errors.New(e), fields...)
			}
			return
		}

		logger.Of(ctx).Info("ending request", fields...)
	}
}
