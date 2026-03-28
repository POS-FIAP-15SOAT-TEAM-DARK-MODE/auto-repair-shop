package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/auth"
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

		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			c.Abort()
			return
		}

		claims, err := auth.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		fmt.Println(claims)
	}
}

func isPublicRoute(path string) bool {
	if strings.HasPrefix(path, "/swagger") {
		return true
	}

	_, ok := publicRoutes[path]
	return ok
}
