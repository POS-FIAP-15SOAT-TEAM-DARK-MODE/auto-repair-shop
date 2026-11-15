package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	// HTTPRequestDuration records gateway request latency by method, route
	// template, upstream and status. Route is the route pattern, never the raw
	// path, so the label cardinality stays bounded.
	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds, by method, route, upstream and status.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route", "upstream", "status"},
	)
)

func init() {
	prometheus.MustRegister(
		HTTPRequestDuration,
	)
}
