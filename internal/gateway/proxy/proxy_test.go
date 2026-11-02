package proxy_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/gateway/proxy"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newUpstream(t *testing.T, h http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return srv
}

func newProxy(t *testing.T, upstreams map[string]string, timeout time.Duration) *proxy.Proxy {
	t.Helper()
	p, err := proxy.New(upstreams, timeout)
	require.NoError(t, err)
	return p
}

func serve(p *proxy.Proxy, upstream string, req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	p.Forward(rec, req, upstream)
	return rec
}

func TestNew_RejectsInvalidURLWithoutEchoingIt(t *testing.T) {
	_, err := proxy.New(map[string]string{"catalog": "http://user:secret@[::1"}, time.Second)
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "secret")
	assert.Contains(t, err.Error(), "catalog")

	_, err = proxy.New(map[string]string{"catalog": "catalog-service:8082"}, time.Second)
	assert.Error(t, err, "a URL without scheme and host is invalid")
}

func TestNew_SkipsUnconfiguredUpstreams(t *testing.T) {
	p := newProxy(t, map[string]string{"lambda": ""}, time.Second)

	rec := serve(p, "lambda", httptest.NewRequest(http.MethodGet, "/x", nil))
	assert.Equal(t, http.StatusBadGateway, rec.Code)
	assert.JSONEq(t, `{"code":502,"errors":["upstream unavailable"]}`, rec.Body.String())
}

func TestForward_PreservesMethodPathQueryBodyAndHeaders(t *testing.T) {
	var got struct {
		method, path, query, body, auth, forwardedFor, forwardedHost, forwardedProto, requestID string
	}
	up := newUpstream(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		got.method = r.Method
		got.path = r.URL.Path
		got.query = r.URL.RawQuery
		got.body = string(b)
		got.auth = r.Header.Get("Authorization")
		got.forwardedFor = r.Header.Get("X-Forwarded-For")
		got.forwardedHost = r.Header.Get("X-Forwarded-Host")
		got.forwardedProto = r.Header.Get("X-Forwarded-Proto")
		got.requestID = r.Header.Get(web.RequestIDHeader)
		w.WriteHeader(http.StatusCreated)
	})
	p := newProxy(t, map[string]string{"catalog": up.URL}, 5*time.Second)

	req := httptest.NewRequest(http.MethodPost, "http://gateway.local/v1/works?page=2", strings.NewReader(`{"name":"oil"}`))
	req.RemoteAddr = "10.1.2.3:5555"
	req.Header.Set("Authorization", "Bearer abc.def.ghi")
	req.Header.Set(web.RequestIDHeader, "req-123")

	rec := serve(p, "catalog", req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, http.MethodPost, got.method)
	assert.Equal(t, "/v1/works", got.path)
	assert.Equal(t, "page=2", got.query)
	assert.JSONEq(t, `{"name":"oil"}`, got.body)
	assert.Equal(t, "Bearer abc.def.ghi", got.auth, "credentials are forwarded untouched")
	assert.Equal(t, "10.1.2.3", got.forwardedFor)
	assert.Equal(t, "gateway.local", got.forwardedHost)
	assert.Equal(t, "http", got.forwardedProto)
	assert.Equal(t, "req-123", got.requestID, "the request id reaches the upstream")
	assert.Equal(t, []string{"req-123"}, rec.Header().Values(web.RequestIDHeader), "one request id in the response")
}

func TestForward_ResponseRequestIDIsNotDuplicated(t *testing.T) {
	up := newUpstream(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(web.RequestIDHeader, "upstream-value")
		w.WriteHeader(http.StatusOK)
	})
	p := newProxy(t, map[string]string{"catalog": up.URL}, 5*time.Second)

	req := httptest.NewRequest(http.MethodGet, "/v1/works", nil)
	req.Header.Set(web.RequestIDHeader, "gateway-value")
	rec := serve(p, "catalog", req)

	assert.Equal(t, []string{"gateway-value"}, rec.Header().Values(web.RequestIDHeader))
}

func TestForward_UpstreamErrorsArePassedThroughUnchanged(t *testing.T) {
	for _, status := range []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusNotFound, http.StatusConflict, http.StatusInternalServerError} {
		up := newUpstream(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			_, _ = io.WriteString(w, `{"code":`+http.StatusText(status)+`}`)
		})
		p := newProxy(t, map[string]string{"catalog": up.URL}, 5*time.Second)

		rec := serve(p, "catalog", httptest.NewRequest(http.MethodGet, "/v1/works/x", nil))

		assert.Equal(t, status, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.Contains(t, rec.Body.String(), http.StatusText(status))
	}
}

func TestForward_UnavailableUpstreamIsBadGatewayWithoutCause(t *testing.T) {
	up := httptest.NewServer(http.NotFoundHandler())
	address := up.URL
	up.Close() // nothing listens any more: connection refused

	p := newProxy(t, map[string]string{"catalog": address}, 5*time.Second)
	req := httptest.NewRequest(http.MethodGet, "/v1/works", nil)
	req.Header.Set(web.RequestIDHeader, "rid-9")
	rec := serve(p, "catalog", req)

	assert.Equal(t, http.StatusBadGateway, rec.Code)
	var body web.ErrorBody
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, http.StatusBadGateway, body.Code)
	assert.Equal(t, []string{"upstream unavailable"}, body.Errors)
	assert.NotContains(t, rec.Body.String(), "refused")
	assert.NotContains(t, rec.Body.String(), "127.0.0.1")
	assert.Equal(t, []string{"rid-9"}, rec.Header().Values(web.RequestIDHeader))
}

func TestForward_SlowUpstreamIsGatewayTimeoutWithoutCause(t *testing.T) {
	up := newUpstream(t, func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(3 * time.Second):
		}
	})
	p := newProxy(t, map[string]string{"catalog": up.URL}, 50*time.Millisecond)

	rec := serve(p, "catalog", httptest.NewRequest(http.MethodGet, "/v1/works", nil))

	assert.Equal(t, http.StatusGatewayTimeout, rec.Code)
	assert.JSONEq(t, `{"code":504,"errors":["upstream timeout"]}`, rec.Body.String())
}

func TestForward_UnknownUpstreamIsBadGateway(t *testing.T) {
	p := newProxy(t, map[string]string{}, time.Second)
	rec := serve(p, "missing", httptest.NewRequest(http.MethodGet, "/x", nil))
	assert.Equal(t, http.StatusBadGateway, rec.Code)
}

func TestForward_ErrorWithoutRequestIDSetsNoHeader(t *testing.T) {
	up := httptest.NewServer(http.NotFoundHandler())
	address := up.URL
	up.Close()

	p := newProxy(t, map[string]string{"catalog": address}, time.Second)
	rec := serve(p, "catalog", httptest.NewRequest(http.MethodGet, "/x", nil))

	assert.Equal(t, http.StatusBadGateway, rec.Code)
	assert.Empty(t, rec.Header().Values(web.RequestIDHeader))
}
