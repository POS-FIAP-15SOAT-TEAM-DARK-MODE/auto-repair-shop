package config_test

import (
	"testing"
	"time"

	rootconfig "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/config"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/gateway/config"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/gateway/match"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleRoutes = `
routes:
  - {methods: ["*"], path: "/internal/*", block: true}
  - {methods: ["GET"], path: "/v1/works*", upstream: catalog}
  - {methods: ["POST"], path: "/auth/customer-login", upstream: lambda, optional: true}
`

func clearUpstreamEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{
		"CUSTOMER_UPSTREAM", "CATALOG_UPSTREAM", "BILLING_UPSTREAM", "EXECUTION_UPSTREAM",
		"HISTORY_UPSTREAM", "SERVICE_ORDER_UPSTREAM", "LAMBDA_UPSTREAM", "UPSTREAM_TIMEOUT", "PORT",
	} {
		t.Setenv(k, "")
	}
}

func TestLoad_DefaultsUseClusterDNSAndDisableLambda(t *testing.T) {
	clearUpstreamEnv(t)

	cfg, err := config.Load([]byte(sampleRoutes))
	require.NoError(t, err)

	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, config.DefaultUpstreamTimeout, cfg.UpstreamTimeout)
	assert.Equal(t, "http://catalog-service:8082", cfg.Upstreams["catalog"])
	assert.Equal(t, "http://service-order-service:8086", cfg.Upstreams["service_order"])
	assert.Empty(t, cfg.Upstreams[config.LambdaUpstream])
	assert.Equal(t, []string{"customer", "catalog", "billing", "execution", "history", "service_order"}, cfg.ConfiguredUpstreams())

	// The optional lambda route is dropped while its upstream is not configured.
	require.Len(t, cfg.Routes, 2)
	assert.Equal(t, "/internal/*", cfg.Routes[0].Path)
	assert.True(t, cfg.Routes[0].Block)
	assert.Equal(t, "catalog", cfg.Routes[1].Upstream)
}

func TestLoad_EnvironmentOverridesValuesAndEnablesOptionalRoute(t *testing.T) {
	clearUpstreamEnv(t)
	t.Setenv("PORT", "9090")
	t.Setenv("UPSTREAM_TIMEOUT", "5s")
	t.Setenv("CATALOG_UPSTREAM", "http://localhost:8082")
	t.Setenv("LAMBDA_UPSTREAM", "https://lambda.local")

	cfg, err := config.Load([]byte(sampleRoutes))
	require.NoError(t, err)

	assert.Equal(t, "9090", cfg.Port)
	assert.Equal(t, 5*time.Second, cfg.UpstreamTimeout)
	assert.Equal(t, "http://localhost:8082", cfg.Upstreams["catalog"])
	assert.Contains(t, cfg.ConfiguredUpstreams(), config.LambdaUpstream)
	require.Len(t, cfg.Routes, 3)
	assert.Equal(t, "lambda", cfg.Routes[2].Upstream)
}

func TestLoad_RejectsInvalidSettings(t *testing.T) {
	cases := map[string]struct {
		env   map[string]string
		yaml  string
		match string
	}{
		"zero timeout":                          {env: map[string]string{"UPSTREAM_TIMEOUT": "0s"}, yaml: sampleRoutes},
		"relative upstream":                     {env: map[string]string{"CATALOG_UPSTREAM": "catalog-service:8082"}, yaml: sampleRoutes},
		"unsupported scheme":                    {env: map[string]string{"CATALOG_UPSTREAM": "ftp://catalog"}, yaml: sampleRoutes},
		"unknown field":                         {yaml: "routes:\n  - {methods: [GET], path: /x, upstrem: catalog}\n"},
		"unknown upstream":                      {yaml: "routes:\n  - {methods: [GET], path: /x, upstream: nope}\n"},
		"required route to unconfigured lambda": {yaml: "routes:\n  - {methods: [POST], path: /x, upstream: lambda}\n"},
		"blocked with upstream":                 {yaml: "routes:\n  - {methods: [GET], path: /x, block: true, upstream: catalog}\n"},
		"blocked and optional":                  {yaml: "routes:\n  - {methods: [GET], path: /x, block: true, optional: true}\n"},
		"malformed yaml":                        {yaml: "routes: [\n"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			clearUpstreamEnv(t)
			for k, v := range tc.env {
				t.Setenv(k, v)
			}
			_, err := config.Load([]byte(tc.yaml))
			assert.Error(t, err)
		})
	}
}

func TestLoad_InvalidURLErrorDoesNotEchoValue(t *testing.T) {
	clearUpstreamEnv(t)
	t.Setenv("CATALOG_UPSTREAM", "http://user:secret@[::1")

	_, err := config.Load([]byte(sampleRoutes))
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "secret")
	assert.Contains(t, err.Error(), "CATALOG_UPSTREAM")
}

func TestConfig_IsService(t *testing.T) {
	clearUpstreamEnv(t)
	cfg, err := config.Load([]byte(sampleRoutes))
	require.NoError(t, err)

	assert.True(t, cfg.IsService("catalog"))
	assert.False(t, cfg.IsService(config.LambdaUpstream), "lambda has no documentation")
	assert.False(t, cfg.IsService("unknown"))
}

func TestRoutesFile_CompilesAndIsWellFormed(t *testing.T) {
	clearUpstreamEnv(t)
	t.Setenv("LAMBDA_UPSTREAM", "http://lambda:8080")
	cfg, err := config.Load(rootconfig.RoutesYAML)
	require.NoError(t, err)

	routes := cfg.Routes
	require.Len(t, routes, 17, "every line of config/routes.yaml is an active route when lambda is configured")

	_, err = match.New(routes)
	require.NoError(t, err)
}
