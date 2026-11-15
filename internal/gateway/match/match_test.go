package match_test

import (
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/gateway/match"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustNew(t *testing.T, routes ...match.Route) *match.Matcher {
	t.Helper()
	m, err := match.New(routes)
	require.NoError(t, err)
	return m
}

func TestNew_RejectsInvalidRoutes(t *testing.T) {
	cases := map[string]match.Route{
		"path without leading slash": {Methods: []string{"GET"}, Path: "v1/works", Upstream: "catalog"},
		"no methods":                 {Path: "/v1/works", Upstream: "catalog"},
		"blank method":               {Methods: []string{" "}, Path: "/v1/works", Upstream: "catalog"},
		"block with upstream":        {Methods: []string{"*"}, Path: "/internal/*", Upstream: "customer", Block: true},
		"route without upstream":     {Methods: []string{"*"}, Path: "/v1/works*"},
		"star in the middle":         {Methods: []string{"*"}, Path: "/v1/*/works", Upstream: "catalog"},
		"star not last segment":      {Methods: []string{"*"}, Path: "/v1*/works", Upstream: "catalog"},
		"two stars":                  {Methods: []string{"*"}, Path: "/v1/a**", Upstream: "catalog"},
		"parameter with star":        {Methods: []string{"*"}, Path: "/v1/:id*", Upstream: "catalog"},
		"parameter without name":     {Methods: []string{"*"}, Path: "/v1/:", Upstream: "catalog"},
		"empty middle segment":       {Methods: []string{"*"}, Path: "/v1//works", Upstream: "catalog"},
	}
	for name, route := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := match.New([]match.Route{route})
			assert.Error(t, err)
		})
	}
}

func TestMatch_RejectsNonAbsolutePathAndDotSegments(t *testing.T) {
	m := mustNew(t, match.Route{Methods: []string{"*"}, Path: "/v1/works*", Upstream: "catalog"})

	_, ok := m.Match("GET", "v1/works")
	assert.False(t, ok, "path must be absolute")

	for _, p := range []string{"/v1/works/../internal/x", "/v1/works/./x", "/v1/../v1/works"} {
		_, ok = m.Match("GET", p)
		assert.False(t, ok, p)
	}
}

func TestMatch_InternalPathsAreBlockedBeforeAnyUpstream(t *testing.T) {
	m := mustNew(t,
		match.Route{Methods: []string{"*"}, Path: "/internal/*", Block: true},
		match.Route{Methods: []string{"*"}, Path: "/v1/customers*", Upstream: "customer"},
	)

	for _, method := range []string{"GET", "POST", "PUT", "DELETE"} {
		r, ok := m.Match(method, "/internal/resource/123")
		require.True(t, ok)
		assert.True(t, r.Block)
		assert.Empty(t, r.Upstream)
	}

	_, ok := m.Match("GET", "/internal")
	assert.False(t, ok, "the bare prefix is not an internal path and is not routed")
}

func TestMatch_MethodMustMatch(t *testing.T) {
	m := mustNew(t, match.Route{Methods: []string{"put", "POST"}, Path: "/v1/service-order/:id/finish", Upstream: "execution"})

	r, ok := m.Match("put", "/v1/service-order/1/finish")
	require.True(t, ok, "method comparison is case-insensitive")
	assert.Equal(t, "execution", r.Upstream)

	_, ok = m.Match("POST", "/v1/service-order/1/finish")
	assert.True(t, ok)

	_, ok = m.Match("GET", "/v1/service-order/1/finish")
	assert.False(t, ok)
}

func TestMatch_ParameterMatchesOneNonEmptySegment(t *testing.T) {
	m := mustNew(t, match.Route{Methods: []string{"PUT"}, Path: "/v1/service-order/:id/work/:workId/next", Upstream: "execution"})

	r, ok := m.Match("PUT", "/v1/service-order/so-1/work/w-2/next")
	require.True(t, ok)
	assert.Equal(t, "execution", r.Upstream)

	_, ok = m.Match("PUT", "/v1/service-order//work/w-2/next")
	assert.False(t, ok, "empty parameter does not match")

	_, ok = m.Match("PUT", "/v1/service-order/so-1/work/w-2/extra/next")
	assert.False(t, ok, "parameter does not span segments")

	_, ok = m.Match("PUT", "/v1/service-order/so-1/work/w-2/next/")
	assert.False(t, ok, "exact segment count is required without a trailing star")
}

func TestMatch_LiteralSegmentsMustBeEqual(t *testing.T) {
	m := mustNew(t, match.Route{Methods: []string{"GET"}, Path: "/v1/service-order/:id/history", Upstream: "history"})

	_, ok := m.Match("GET", "/v1/service-order/1/histories")
	assert.False(t, ok)
	_, ok = m.Match("GET", "/v1/service-order/1")
	assert.False(t, ok)
	_, ok = m.Match("GET", "/v1/service-order/1/history/extra")
	assert.False(t, ok)
}

func TestMatch_TrailingStarOnSegmentMatchesPrefixAndSuffix(t *testing.T) {
	m := mustNew(t, match.Route{Methods: []string{"*"}, Path: "/v1/customers*", Upstream: "customer"})

	for _, p := range []string{"/v1/customers", "/v1/customers/", "/v1/customers/123", "/v1/customersX"} {
		r, ok := m.Match("GET", p)
		assert.True(t, ok, p)
		assert.Equal(t, "customer", r.Upstream)
	}
	_, ok := m.Match("GET", "/v1/vehicles")
	assert.False(t, ok)
	_, ok = m.Match("GET", "/v1/custom")
	assert.False(t, ok, "shorter than the literal prefix")
}

func TestMatch_TrailingStarAfterSegmentRequiresSegment(t *testing.T) {
	m := mustNew(t, match.Route{Methods: []string{"*"}, Path: "/v1/auth/*", Upstream: "customer"})

	for _, p := range []string{"/v1/auth/", "/v1/auth/login", "/v1/auth/a/b/c"} {
		_, ok := m.Match("POST", p)
		assert.True(t, ok, p)
	}
	_, ok := m.Match("POST", "/v1/auth")
	assert.False(t, ok)
	_, ok = m.Match("POST", "/v1/authx")
	assert.False(t, ok, "a star after a full segment does not match a longer segment")
}

func TestMatch_FirstRouteWinsByPriority(t *testing.T) {
	m := mustNew(t,
		match.Route{Methods: []string{"PUT"}, Path: "/v1/service-order/:id/finish", Upstream: "execution"},
		match.Route{Methods: []string{"*"}, Path: "/v1/service-order*", Upstream: "service_order"},
	)

	r, ok := m.Match("PUT", "/v1/service-order/1/finish")
	require.True(t, ok)
	assert.Equal(t, "execution", r.Upstream)

	r, ok = m.Match("DELETE", "/v1/service-order/1/finish")
	require.True(t, ok)
	assert.Equal(t, "service_order", r.Upstream, "method mismatch falls through to the next route")
}

func TestMatch_UnknownRouteIsNotFound(t *testing.T) {
	m := mustNew(t, match.Route{Methods: []string{"*"}, Path: "/v1/works*", Upstream: "catalog"})

	_, ok := m.Match("GET", "/nope")
	assert.False(t, ok)
	_, ok = m.Match("GET", "/")
	assert.False(t, ok)
}

func TestMatch_EmptyTableMatchesNothing(t *testing.T) {
	m := mustNew(t)
	_, ok := m.Match("GET", "/v1/works")
	assert.False(t, ok)
}
