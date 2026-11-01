package handlers_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	rootconfig "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/config"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/app/middleware"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/gateway/config"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/gateway/handlers"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/gateway/match"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/gateway/proxy"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/web"
)

// seen records what an upstream received.
type seen struct {
	mu     sync.Mutex
	paths  []string
	method string
	reqID  string
}

func (s *seen) record(r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.paths = append(s.paths, r.URL.Path)
	s.method = r.Method
	s.reqID = r.Header.Get(web.RequestIDHeader)
}

func (s *seen) called() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.paths) > 0
}

// fakeService answers /ping and /swagger.yaml, and records everything else.
func fakeService(t *testing.T, s *seen, pingStatus int) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.record(r)
		switch r.URL.Path {
		case "/ping":
			w.WriteHeader(pingStatus)
		case "/swagger.yaml":
			_, _ = io.WriteString(w, "openapi: 3.0.3\n")
		default:
			w.Header().Set("X-Request-Id", "upstream-own")
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, `{"from":"service"}`)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

// gatewayFor serves the full gateway router over a real HTTP server, as
// production does: httputil.ReverseProxy needs a CloseNotifier, which
// httptest.ResponseRecorder does not provide.
func gatewayFor(t *testing.T, upstreams map[string]string, timeout time.Duration) string {
	t.Helper()
	gin.SetMode(gin.TestMode)
	t.Setenv("LOG_LEVEL", "PANIC")
	t.Setenv("UPSTREAM_TIMEOUT", timeout.String())
	for _, name := range []string{"customer", "catalog", "billing", "execution", "history", "service_order", "lambda"} {
		t.Setenv(strings.ToUpper(name)+"_UPSTREAM", upstreams[name])
	}

	cfg, err := config.Load(rootconfig.RoutesYAML)
	require.NoError(t, err)
	matcher, err := match.New(cfg.Routes)
	require.NoError(t, err)
	p, err := proxy.New(cfg.Upstreams, cfg.UpstreamTimeout)
	require.NoError(t, err)

	srv := httptest.NewServer(handlers.NewRouter(handlers.New(cfg, matcher, p),
		middleware.Logger(), middleware.Recovery(), middleware.Metrics()))
	t.Cleanup(srv.Close)
	return srv.URL
}

type result struct {
	Code   int
	Header http.Header
	Body   string
}

func do(t *testing.T, base, method, target string, headers ...string) result {
	t.Helper()
	req, err := http.NewRequest(method, base+target, nil)
	require.NoError(t, err)
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return result{Code: resp.StatusCode, Header: resp.Header, Body: string(body)}
}

// allServices returns upstream URLs, all answering /ping with 200, and their recorders.
func allServices(t *testing.T) (map[string]string, map[string]*seen) {
	t.Helper()
	urls := map[string]string{}
	seens := map[string]*seen{}
	for _, name := range config.Services {
		s := &seen{}
		urls[name] = fakeService(t, s, http.StatusOK).URL
		seens[name] = s
	}
	urls[config.LambdaUpstream] = ""
	return urls, seens
}

func TestOwnEndpoints_PingAndSwaggerYAML(t *testing.T) {
	urls, _ := allServices(t)
	h := gatewayFor(t, urls, 5*time.Second)

	rec := do(t, h, http.MethodGet, "/ping")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"message":"pong"}`, rec.Body)

	rec = do(t, h, http.MethodGet, "/swagger.yaml")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header.Get("Content-Type"), "yaml")
	assert.Contains(t, rec.Body, "Auto Repair Shop Gateway")
}

func TestSwaggerIndex(t *testing.T) {
	urls, _ := allServices(t)
	h := gatewayFor(t, urls, 5*time.Second)

	for _, target := range []string{"/swagger/index.html", "/swagger/", "/swagger"} {
		rec := do(t, h, http.MethodGet, target)
		if target == "/swagger" {
			// No trailing segment: the index is only served under /swagger/.
			assert.Equal(t, http.StatusNotFound, rec.Code, target)
			continue
		}
		assert.Equal(t, http.StatusOK, rec.Code, target)
		assert.Contains(t, rec.Header.Get("Content-Type"), "text/html")
		assert.Contains(t, rec.Body, "/docs/catalog/swagger.yaml")
	}

	rec := do(t, h, http.MethodGet, "/swagger/other.html")
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.JSONEq(t, `{"code":404,"errors":["route not found"]}`, rec.Body)
}

func TestDocs_ProxiesServiceSwagger(t *testing.T) {
	urls, seens := allServices(t)
	h := gatewayFor(t, urls, 5*time.Second)

	rec := do(t, h, http.MethodGet, "/docs/catalog/swagger.yaml")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "openapi: 3.0.3\n", rec.Body)
	assert.Equal(t, []string{"/swagger.yaml"}, seens["catalog"].paths)
}

func TestDocs_OnlyDocumentedConfiguredServices(t *testing.T) {
	urls, seens := allServices(t)
	h := gatewayFor(t, urls, 5*time.Second)

	for _, target := range []string{
		"/docs/lambda/swagger.yaml",    // the lambda has no documentation
		"/docs/unknown/swagger.yaml",   // not an upstream
		"/docs/catalog/other.yaml",     // only swagger.yaml is served
		"/docs/catalog",                // incomplete
		"/docs/",                       // empty service name
		"/docs/catalog/swagger.yaml/x", // extra segment
	} {
		rec := do(t, h, http.MethodGet, target)
		assert.Equal(t, http.StatusNotFound, rec.Code, target)
	}
	assert.False(t, seens["catalog"].called(), "no request reached the service")
}

func TestReady_AllUpstreamsReady(t *testing.T) {
	urls, _ := allServices(t)
	h := gatewayFor(t, urls, 5*time.Second)

	rec := do(t, h, http.MethodGet, "/ready")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"message":"ready"}`, rec.Body)
}

func TestReady_ListsFailingUpstreamNamesOnly(t *testing.T) {
	urls, _ := allServices(t)
	broken := &seen{}
	urls["catalog"] = fakeService(t, broken, http.StatusInternalServerError).URL

	closed := httptest.NewServer(http.NotFoundHandler())
	closed.Close()
	urls["billing"] = closed.URL

	h := gatewayFor(t, urls, 5*time.Second)
	rec := do(t, h, http.MethodGet, "/ready")

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	var body struct {
		Code      int      `json:"code"`
		Errors    []string `json:"errors"`
		Upstreams []string `json:"upstreams"`
	}
	require.NoError(t, json.Unmarshal([]byte(rec.Body), &body))
	assert.Equal(t, http.StatusServiceUnavailable, body.Code)
	assert.ElementsMatch(t, []string{"catalog", "billing"}, body.Upstreams)
	assert.NotContains(t, rec.Body, "127.0.0.1", "URLs are never listed")
}

func TestReady_LambdaIsCheckedOnlyWhenConfigured(t *testing.T) {
	urls, _ := allServices(t)
	closed := httptest.NewServer(http.NotFoundHandler())
	closed.Close()
	urls[config.LambdaUpstream] = closed.URL

	h := gatewayFor(t, urls, 5*time.Second)
	rec := do(t, h, http.MethodGet, "/ready")
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Contains(t, rec.Body, "lambda")
}

func TestMetricsEndpoint(t *testing.T) {
	urls, _ := allServices(t)
	h := gatewayFor(t, urls, 5*time.Second)

	do(t, h, http.MethodGet, "/ping")
	rec := do(t, h, http.MethodGet, "/metrics")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body, "http_request_duration_seconds")
}

func TestProxy_ForwardsRequestAndAddsRequestID(t *testing.T) {
	urls, seens := allServices(t)
	h := gatewayFor(t, urls, 5*time.Second)

	rec := do(t, h, http.MethodGet, "/v1/works?page=1")

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"from":"service"}`, rec.Body)
	assert.Equal(t, []string{"/v1/works"}, seens["catalog"].paths)
	require.NotEmpty(t, rec.Header.Get(web.RequestIDHeader))
	assert.Equal(t, rec.Header.Get(web.RequestIDHeader), seens["catalog"].reqID, "the same id reaches the upstream")
	assert.Equal(t, []string{rec.Header.Get(web.RequestIDHeader)}, rec.Header.Values(web.RequestIDHeader),
		"the upstream's own request id is replaced")
}

func TestProxy_KeepsClientRequestID(t *testing.T) {
	urls, seens := allServices(t)
	h := gatewayFor(t, urls, 5*time.Second)

	rec := do(t, h, http.MethodGet, "/v1/supplies", web.RequestIDHeader, "client-42")

	assert.Equal(t, "client-42", rec.Header.Get(web.RequestIDHeader))
	assert.Equal(t, "client-42", seens["catalog"].reqID)
}

func TestProxy_RoutesToTheExpectedUpstream(t *testing.T) {
	urls, seens := allServices(t)
	h := gatewayFor(t, urls, 5*time.Second)

	cases := []struct {
		method, target, upstream string
	}{
		{http.MethodPut, "/v1/service-order/so-1/finish", "execution"},
		{http.MethodGet, "/v1/service-order/so-1/history", "history"},
		{http.MethodGet, "/v1/service-order/so-1/status", "service_order"},
		{http.MethodPost, "/v1/customers", "customer"},
		{http.MethodGet, "/v1/budgets/so-1", "billing"},
	}
	for _, tc := range cases {
		rec := do(t, h, tc.method, tc.target)
		assert.Equal(t, http.StatusOK, rec.Code, tc.target)
		assert.Equal(t, tc.method, seens[tc.upstream].method, tc.target)
	}
}

func TestProxy_InternalRoutesAreNeverForwarded(t *testing.T) {
	urls, seens := allServices(t)
	h := gatewayFor(t, urls, 5*time.Second)

	for _, target := range []string{"/internal/customers/1", "/internal/works/1"} {
		rec := do(t, h, http.MethodGet, target)
		assert.Equal(t, http.StatusNotFound, rec.Code, target)
		assert.JSONEq(t, `{"code":404,"errors":["route not found"]}`, rec.Body)
	}
	for name, s := range seens {
		assert.False(t, s.called(), "upstream %s must not see internal paths", name)
	}
}

func TestProxy_UnmatchedAndDotSegmentsAreNotFound(t *testing.T) {
	urls, seens := allServices(t)
	h := gatewayFor(t, urls, 5*time.Second)

	for _, c := range []struct{ method, target string }{
		{http.MethodGet, "/nope"},
		{http.MethodDelete, "/v1/execution/queue"},
		{http.MethodPost, "/ping"},
		{http.MethodGet, "/v1/works/../internal/customers/1"},
	} {
		rec := do(t, h, c.method, c.target)
		assert.Equal(t, http.StatusNotFound, rec.Code, c.method+" "+c.target)
	}
	assert.False(t, seens["catalog"].called())
	assert.False(t, seens["execution"].called())
}

func TestProxy_OptionalLambdaRouteOnlyWhenConfigured(t *testing.T) {
	urls, _ := allServices(t)
	h := gatewayFor(t, urls, 5*time.Second)
	assert.Equal(t, http.StatusNotFound, do(t, h, http.MethodPost, "/auth/customer-login").Code)

	lambda := &seen{}
	urls[config.LambdaUpstream] = fakeService(t, lambda, http.StatusOK).URL
	h = gatewayFor(t, urls, 5*time.Second)
	rec := do(t, h, http.MethodPost, "/auth/customer-login")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, []string{"/auth/customer-login"}, lambda.paths)
}

func TestProxy_UpstreamDownIsBadGatewayWithoutCause(t *testing.T) {
	urls, _ := allServices(t)
	closed := httptest.NewServer(http.NotFoundHandler())
	closed.Close()
	urls["catalog"] = closed.URL
	h := gatewayFor(t, urls, 5*time.Second)

	rec := do(t, h, http.MethodGet, "/v1/works")

	assert.Equal(t, http.StatusBadGateway, rec.Code)
	assert.JSONEq(t, `{"code":502,"errors":["upstream unavailable"]}`, rec.Body)
	assert.NotContains(t, rec.Body, "refused")
}

func TestProxy_SlowUpstreamIsGatewayTimeout(t *testing.T) {
	urls, _ := allServices(t)
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(3 * time.Second):
		}
	}))
	t.Cleanup(slow.Close)
	urls["catalog"] = slow.URL
	h := gatewayFor(t, urls, 50*time.Millisecond)

	rec := do(t, h, http.MethodGet, "/v1/works")

	assert.Equal(t, http.StatusGatewayTimeout, rec.Code)
	assert.JSONEq(t, `{"code":504,"errors":["upstream timeout"]}`, rec.Body)
}

func TestProxy_UpstreamErrorsPassThrough(t *testing.T) {
	urls, _ := allServices(t)
	failing := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_, _ = io.WriteString(w, `{"code":409,"errors":["already exists"]}`)
	}))
	t.Cleanup(failing.Close)
	urls["catalog"] = failing.URL
	h := gatewayFor(t, urls, 5*time.Second)

	rec := do(t, h, http.MethodPost, "/v1/works")

	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.JSONEq(t, `{"code":409,"errors":["already exists"]}`, rec.Body)
}
