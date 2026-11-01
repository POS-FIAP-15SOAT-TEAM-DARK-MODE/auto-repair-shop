// Package proxy forwards requests to the upstream services with
// httputil.ReverseProxy. The proxy does not inspect credentials: every service
// validates its own tokens.
package proxy

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"go.uber.org/zap"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/web"
)

// Proxy routes requests to one reverse proxy per configured upstream.
type Proxy struct {
	targets map[string]*httputil.ReverseProxy
}

// New builds a reverse proxy for every upstream with a URL. Upstreams with an
// empty URL are skipped. timeout bounds the wait for the response headers.
func New(upstreams map[string]string, timeout time.Duration) (*Proxy, error) {
	p := &Proxy{targets: make(map[string]*httputil.ReverseProxy, len(upstreams))}
	for name, raw := range upstreams {
		if raw == "" {
			continue
		}
		target, err := url.Parse(raw)
		if err != nil || target.Scheme == "" || target.Host == "" {
			// The raw value is not echoed: it may carry credentials.
			return nil, fmt.Errorf("invalid URL for upstream %s", name)
		}
		p.targets[name] = newReverseProxy(name, target, timeout)
	}
	return p, nil
}

// Forward proxies r to the named upstream. The response carries the request's
// X-Request-Id once, whatever the upstream answered.
func (p *Proxy) Forward(w http.ResponseWriter, r *http.Request, upstream string) {
	rp, ok := p.targets[upstream]
	if !ok {
		web.WriteError(w, http.StatusBadGateway, "upstream unavailable")
		return
	}
	w.Header().Del(web.RequestIDHeader)
	rp.ServeHTTP(w, r)
}

func newReverseProxy(name string, target *url.URL, timeout time.Duration) *httputil.ReverseProxy {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.ResponseHeaderTimeout = timeout

	return &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(target)
			pr.SetXForwarded()
		},
		Transport:      transport,
		ModifyResponse: keepRequestID,
		ErrorHandler:   errorHandler(name),
		ErrorLog:       zap.NewStdLog(logger.Global().Logger),
	}
}

// keepRequestID replaces any X-Request-Id sent by the upstream with the
// gateway's value, so the client never sees two.
func keepRequestID(resp *http.Response) error {
	if resp.Request == nil {
		return nil
	}
	if id := resp.Request.Header.Get(web.RequestIDHeader); id != "" {
		resp.Header.Set(web.RequestIDHeader, id)
	}
	return nil
}

// errorHandler maps transport failures to 504 when the upstream timed out and
// to 502 otherwise. The cause is logged on the server and never sent back.
func errorHandler(name string) func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		status, message := http.StatusBadGateway, "upstream unavailable"
		if isTimeout(err) {
			status, message = http.StatusGatewayTimeout, "upstream timeout"
		}

		logger.Of(r.Context()).Error(err,
			zap.String("upstream", name),
			zap.Int("status", status),
		)

		if id := r.Header.Get(web.RequestIDHeader); id != "" {
			w.Header().Set(web.RequestIDHeader, id)
		}
		web.WriteError(w, status, message)
	}
}

func isTimeout(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}
