package middleware

import (
	"github.com/gin-gonic/gin"
)

// TODO: Remove loose string and convert them into constants

var publicRoutes = map[string]struct{}{
	"/v1/auth/login":    {},
	"/v1/auth/register": {},
	"/ping":             {},
}

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if isPublicRoute(c.Request.URL.Path) {
			c.Next()
			return
		}
	}
}

func isPublicRoute(path string) bool {
	if len(path) >= 8 && path[:8] == "/swagger" {
		return true
	}

	_, ok := publicRoutes[path]
	return ok
}
