// Package handlers serves the gateway's own endpoints (health, readiness,
// metrics, documentation) and hands every other request to the matcher and
// the proxy.
package handlers

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	api "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/docs"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/gateway/config"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/gateway/match"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/gateway/proxy"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/web"
)

// readyProbeTimeout bounds each upstream check made by /ready.
const readyProbeTimeout = 2 * time.Second

// Gateway holds the dependencies of the HTTP handlers.
type Gateway struct {
	cfg     *config.Config
	matcher *match.Matcher
	proxy   *proxy.Proxy
	probe   *http.Client
}

// New creates the gateway handlers.
func New(cfg *config.Config, matcher *match.Matcher, p *proxy.Proxy) *Gateway {
	return &Gateway{cfg: cfg, matcher: matcher, proxy: p, probe: &http.Client{}}
}

// Ping answers the liveness and readiness probes of the gateway itself.
func (g *Gateway) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}

// Ready answers 200 when every configured upstream responds to /ping within
// readyProbeTimeout. Otherwise it answers 503 and lists the failing upstream
// names, never their URLs.
func (g *Gateway) Ready(c *gin.Context) {
	names := g.cfg.ConfiguredUpstreams()
	failed := make([]bool, len(names))

	var wg sync.WaitGroup
	for i, name := range names {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := g.probeUpstream(c.Request.Context(), g.cfg.Upstreams[name]); err != nil {
				failed[i] = true
				logger.Of(c.Request.Context()).Warn("upstream not ready", zap.String("upstream", name))
			}
		}()
	}
	wg.Wait()

	down := make([]string, 0)
	for i, name := range names {
		if failed[i] {
			down = append(down, name)
		}
	}
	if len(down) > 0 {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"code":      http.StatusServiceUnavailable,
			"errors":    []string{"upstreams not ready"},
			"upstreams": down,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ready"})
}

// SwaggerYAML serves the gateway's own OpenAPI document.
func (g *Gateway) SwaggerYAML(c *gin.Context) {
	c.Data(http.StatusOK, "application/yaml; charset=utf-8", api.SwaggerBytes)
}

// SwaggerIndex serves the static page that links to the documentation of each service.
func (g *Gateway) SwaggerIndex(c *gin.Context) {
	switch c.Param("any") {
	case "", "/", "/index.html":
		c.Data(http.StatusOK, "text/html; charset=utf-8", api.IndexHTML)
	default:
		g.notFound(c)
	}
}

// Docs proxies /docs/<service>/swagger.yaml to the swagger.yaml of that
// service. Only documented, configured services are served.
func (g *Gateway) Docs(c *gin.Context) {
	name, file, found := strings.Cut(strings.TrimPrefix(c.Param("path"), "/"), "/")
	if !found || file != "swagger.yaml" || !g.cfg.IsService(name) {
		g.notFound(c)
		return
	}

	c.Set("upstream", name)
	req := c.Request.Clone(c.Request.Context())
	req.URL.Path = "/swagger.yaml"
	req.URL.RawPath = ""
	g.proxy.Forward(c.Writer, req, name)
}

// Route handles every request that no gateway endpoint claimed. Blocked and
// unknown routes answer 404; the others are proxied to their upstream.
func (g *Gateway) Route(c *gin.Context) {
	rule, ok := g.matcher.Match(c.Request.Method, c.Request.URL.Path)
	if !ok {
		c.Set("upstream", "none")
		g.notFound(c)
		return
	}

	c.Set("route", rule.Path)
	if rule.Block {
		c.Set("upstream", "none")
		g.notFound(c)
		return
	}

	c.Set("upstream", rule.Upstream)
	g.proxy.Forward(c.Writer, c.Request, rule.Upstream)
}

func (g *Gateway) notFound(c *gin.Context) {
	web.WriteError(c.Writer, http.StatusNotFound, "route not found")
}

func (g *Gateway) probeUpstream(ctx context.Context, baseURL string) error {
	ctx, cancel := context.WithTimeout(ctx, readyProbeTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimSuffix(baseURL, "/")+"/ping", nil)
	if err != nil {
		return err
	}
	resp, err := g.probe.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusOK {
		return errUnexpectedStatus
	}
	return nil
}

var errUnexpectedStatus = errors.New("upstream did not answer 200 to /ping")
