package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/web"
)

// Recovery turns a panic into a 500 with a generic message. The panic value
// and the stack are logged on the server only.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.Of(c.Request.Context()).Error(fmt.Errorf("panic: %v", rec),
					zap.String("behavior", "[PANIC RECOVERED]"),
					zap.String("stack", string(debug.Stack())),
				)
				web.WriteError(c.Writer, http.StatusInternalServerError, "internal server error")
				c.Abort()
			}
		}()

		c.Next()
	}
}
