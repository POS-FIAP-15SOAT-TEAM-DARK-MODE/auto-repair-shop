// Package config loads the gateway settings from the environment and the
// embedded routing table.
package config

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/gateway/match"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/env"
)

// DefaultUpstreamTimeout bounds the wait for the response headers of an upstream.
const DefaultUpstreamTimeout = 30 * time.Second

// LambdaUpstream is the name of the optional customer login upstream.
const LambdaUpstream = "lambda"

// Services are the upstreams that expose a swagger.yaml, served under
// /docs/<name>/swagger.yaml. The lambda has no API documentation.
var Services = []string{"customer", "catalog", "billing", "execution", "history", "service_order"}

var upstreamNames = append(slices.Clone(Services), LambdaUpstream)

var defaultUpstreams = map[string]string{
	"customer":      "http://customer-service:8081",
	"catalog":       "http://catalog-service:8082",
	"billing":       "http://billing-service:8083",
	"execution":     "http://execution-service:8084",
	"history":       "http://history-service:8085",
	"service_order": "http://service-order-service:8086",
	LambdaUpstream:  "",
}

// Config is the validated runtime configuration of the gateway.
type Config struct {
	Port            string
	UpstreamTimeout time.Duration
	// Upstreams maps an upstream name to its base URL. An empty URL means the
	// upstream is not configured.
	Upstreams map[string]string
	// Routes is the active routing table, in priority order. Optional routes
	// whose upstream is not configured are left out.
	Routes []match.Route
}

// Load reads PORT, UPSTREAM_TIMEOUT and the <NAME>_UPSTREAM variables, and
// parses the routing table. It fails on any invalid value so the gateway never
// starts with a partial configuration.
func Load(routesYAML []byte) (*Config, error) {
	cfg := &Config{
		Port:            env.GetString("PORT", "8080"),
		UpstreamTimeout: env.GetTimeDuration("UPSTREAM_TIMEOUT", DefaultUpstreamTimeout),
		Upstreams:       make(map[string]string, len(upstreamNames)),
	}
	if cfg.UpstreamTimeout <= 0 {
		return nil, errors.New("UPSTREAM_TIMEOUT must be positive")
	}

	for _, name := range upstreamNames {
		key := envKey(name)
		raw := env.GetString(key, defaultUpstreams[name])
		if raw != "" {
			if err := validateURL(raw); err != nil {
				return nil, fmt.Errorf("%s: %w", key, err)
			}
		}
		cfg.Upstreams[name] = raw
	}

	routes, err := parseRoutes(routesYAML, cfg.Upstreams)
	if err != nil {
		return nil, err
	}
	cfg.Routes = routes
	return cfg, nil
}

// ConfiguredUpstreams returns the names of the upstreams that have a URL, in
// the order they are declared.
func (c *Config) ConfiguredUpstreams() []string {
	names := make([]string, 0, len(upstreamNames))
	for _, name := range upstreamNames {
		if c.Upstreams[name] != "" {
			names = append(names, name)
		}
	}
	return names
}

// IsService reports whether name is an upstream with a documented API that is configured.
func (c *Config) IsService(name string) bool {
	return slices.Contains(Services, name) && c.Upstreams[name] != ""
}

type fileRoute struct {
	Methods  []string `yaml:"methods"`
	Path     string   `yaml:"path"`
	Upstream string   `yaml:"upstream"`
	Block    bool     `yaml:"block"`
	Optional bool     `yaml:"optional"`
}

func parseRoutes(data []byte, upstreams map[string]string) ([]match.Route, error) {
	var file struct {
		Routes []fileRoute `yaml:"routes"`
	}
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&file); err != nil {
		return nil, fmt.Errorf("invalid routes file: %w", err)
	}

	routes := make([]match.Route, 0, len(file.Routes))
	for i, fr := range file.Routes {
		if fr.Block {
			if fr.Optional || fr.Upstream != "" {
				return nil, fmt.Errorf("route %d (%s): a blocked route takes no upstream or optional flag", i, fr.Path)
			}
			routes = append(routes, match.Route{Methods: fr.Methods, Path: fr.Path, Block: true})
			continue
		}

		if !slices.Contains(upstreamNames, fr.Upstream) {
			return nil, fmt.Errorf("route %d (%s): unknown upstream %q", i, fr.Path, fr.Upstream)
		}
		if upstreams[fr.Upstream] == "" {
			if fr.Optional {
				continue
			}
			return nil, fmt.Errorf("route %d (%s): upstream %q is not configured", i, fr.Path, fr.Upstream)
		}
		routes = append(routes, match.Route{Methods: fr.Methods, Path: fr.Path, Upstream: fr.Upstream})
	}
	return routes, nil
}

func envKey(name string) string {
	return strings.ToUpper(name) + "_UPSTREAM"
}

// validateURL rejects anything but an absolute http(s) URL. The value itself
// is never echoed in the error, since it may carry credentials.
func validateURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return errors.New("must be an absolute http or https URL")
	}
	return nil
}
