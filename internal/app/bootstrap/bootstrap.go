// Package bootstrap wires the gateway: configuration, routing table, proxy,
// handlers and HTTP server.
package bootstrap

import (
	"fmt"

	rootconfig "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/config"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/app"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/app/middleware"
	gatewayconfig "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/gateway/config"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/gateway/handlers"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/gateway/match"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/gateway/proxy"
)

// HTTPServer builds the gateway HTTP server from the environment and the
// embedded routing table. It fails when the configuration is invalid.
func HTTPServer() (app.Server, error) {
	cfg, err := gatewayconfig.Load(rootconfig.RoutesYAML)
	if err != nil {
		return nil, fmt.Errorf("gateway configuration: %w", err)
	}

	matcher, err := match.New(cfg.Routes)
	if err != nil {
		return nil, fmt.Errorf("routing table: %w", err)
	}

	reverseProxy, err := proxy.New(cfg.Upstreams, cfg.UpstreamTimeout)
	if err != nil {
		return nil, fmt.Errorf("proxy: %w", err)
	}

	gateway := handlers.New(cfg, matcher, reverseProxy)
	router := handlers.NewRouter(gateway,
		middleware.Logger(),
		middleware.Recovery(),
		middleware.Metrics(),
	)

	return app.NewHTTPServer("gateway", ":"+cfg.Port, router), nil
}
