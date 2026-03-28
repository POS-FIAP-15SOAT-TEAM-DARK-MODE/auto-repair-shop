package middleware

import (
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/auth"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var (
	publicRoutes = map[string]struct{}{
		"/v1/auth/login":    {},
		"/v1/auth/register": {},
		"/ping":             {},
	}

	privateRoutes = map[string]map[string][]string{
		"POST": {
			"/v1/customers": {"ADMIN", "ATTENDANT"},
			"/v1/services":  {"ADMIN", "ATTENDANT"},
		},
		"GET": {
			"/v1/services": {"ADMIN", "ATTENDANT"},
		},
		"PUT": {
			"/v1/services/:id": {"ADMIN", "ATTENDANT"},
		},
		"DELETE": {
			"/v1/services/:id": {"ADMIN"},
		},
	}
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if isPublicRoute(c.Request.URL.Path) {
			c.Next()
			return
		}

		token := c.GetHeader("Authorization")
		if token == "" || !strings.HasPrefix(token, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		token = strings.TrimPrefix(token, "Bearer ")

		claims, err := auth.ParseToken(token)
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				c.JSON(http.StatusForbidden, gin.H{"error": "Token expired"})
				c.Abort()
				return
			}
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		if !hasRoleToAccess(c.Request.Method, path, claims.Roles) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			c.Abort()
			return
		}
	}
}

func isPublicRoute(path string) bool {
	if strings.HasPrefix(path, "/swagger") {
		return true
	}

	_, ok := publicRoutes[path]
	return ok
}

func hasRoleToAccess(method, path string, userRoles []string) bool {
	byMethod, ok := privateRoutes[method]
	if !ok {
		return false
	}

	if rolesNeeded, ok := byMethod[path]; ok {
		for _, role := range userRoles {
			if slices.Contains(rolesNeeded, role) {
				return true
			}
		}
		return false
	}

	return false
}
