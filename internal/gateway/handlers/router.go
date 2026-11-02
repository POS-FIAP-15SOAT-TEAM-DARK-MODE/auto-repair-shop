package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// NewRouter registers the gateway's own endpoints explicitly and sends every
// other request to Route. The proxied routes are not registered in Gin: they
// are resolved by the matcher, so Gin's wildcard conflicts never apply to them.
func NewRouter(g *Gateway, middlewares ...gin.HandlerFunc) *gin.Engine {
	router := gin.New()
	// Paths are matched exactly; redirects would rewrite proxied requests.
	router.RedirectTrailingSlash = false
	router.RedirectFixedPath = false
	router.Use(middlewares...)

	router.GET("/ping", g.Ping)
	router.GET("/ready", g.Ready)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.GET("/swagger.yaml", g.SwaggerYAML)
	router.GET("/swagger/*any", g.SwaggerIndex)
	router.GET("/docs/*path", g.Docs)
	router.NoRoute(g.Route)

	return router
}
