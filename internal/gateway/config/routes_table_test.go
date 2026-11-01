package config_test

import (
	"testing"

	rootconfig "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/config"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/gateway/config"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/gateway/match"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEmbeddedRoutingTable checks, for each line of config/routes.yaml, that
// a representative request is routed to the expected upstream, plus the
// priority cases where a later line would otherwise match.
func TestEmbeddedRoutingTable(t *testing.T) {
	clearUpstreamEnv(t)
	t.Setenv("LAMBDA_UPSTREAM", "http://lambda:8080")

	cfg, err := config.Load(rootconfig.RoutesYAML)
	require.NoError(t, err)
	m, err := match.New(cfg.Routes)
	require.NoError(t, err)

	cases := []struct {
		name     string
		method   string
		path     string
		upstream string
		block    bool
	}{
		// line 1: /internal/* is never exposed
		{"internal blocked", "GET", "/internal/customers/123", "", true},
		{"internal blocked for any method", "DELETE", "/internal/reservations/1", "", true},
		// lines 2-4: execution commands under service-order
		{"next work", "PUT", "/v1/service-order/so-1/work/w-1/next", "execution", false},
		{"cancel work", "PUT", "/v1/service-order/so-1/work/w-1/cancel", "execution", false},
		{"finish", "PUT", "/v1/service-order/so-1/finish", "execution", false},
		// line 5: execution queue
		{"execution queue", "GET", "/v1/execution/queue", "execution", false},
		// line 6: history of one order, before the generic service-order route
		{"history", "GET", "/v1/service-order/so-1/history", "history", false},
		// line 7: reports
		{"reports", "GET", "/v1/reports/average-execution-time", "history", false},
		// line 8: auth
		{"auth login", "POST", "/v1/auth/login", "customer", false},
		// line 9: users
		{"user role", "PATCH", "/v1/users/u-1/role", "customer", false},
		// line 10: customers
		{"customers list", "GET", "/v1/customers", "customer", false},
		// line 11: vehicles
		{"vehicles", "GET", "/v1/vehicles/v-1", "customer", false},
		// line 12: works
		{"works", "GET", "/v1/works", "catalog", false},
		// line 13: supplies
		{"supplies", "POST", "/v1/supplies", "catalog", false},
		// line 14: budgets
		{"budgets", "GET", "/v1/budgets/so-1", "billing", false},
		// line 15: payments
		{"payments confirm", "POST", "/v1/payments/p-1/confirm", "billing", false},
		// line 16: generic service-order, including the lifecycle routes not listed above
		{"service-order accept", "PUT", "/v1/service-order/so-1/accept", "service_order", false},
		{"service-order status", "GET", "/v1/service-order/so-1/status", "service_order", false},
		{"service-order create", "POST", "/v1/service-order", "service_order", false},
		// line 17: customer login through the lambda
		{"customer login", "POST", "/auth/customer-login", "lambda", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, ok := m.Match(tc.method, tc.path)
			require.True(t, ok, "%s %s must match a route", tc.method, tc.path)
			assert.Equal(t, tc.block, r.Block)
			assert.Equal(t, tc.upstream, r.Upstream)
		})
	}
}

func TestEmbeddedRoutingTable_LambdaRouteIsOffWithoutUpstream(t *testing.T) {
	clearUpstreamEnv(t)

	cfg, err := config.Load(rootconfig.RoutesYAML)
	require.NoError(t, err)
	m, err := match.New(cfg.Routes)
	require.NoError(t, err)

	_, ok := m.Match("POST", "/auth/customer-login")
	assert.False(t, ok, "without LAMBDA_UPSTREAM the login route is not served")
}
