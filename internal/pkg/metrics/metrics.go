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

	ServiceOrdersCreatedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "service_orders_created_total",
		Help: "Total number of service orders created.",
	})

	ServiceOrderStatusTransitionsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "service_order_status_transitions_total",
			Help: "Total number of service order status transitions, by new status.",
		},
		[]string{"status"},
	)

	ServiceOrderNotificationFailuresTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "service_order_notification_failures_total",
		Help: "Total number of failed customer status-change notifications.",
	})
)

func init() {
	prometheus.MustRegister(
		HTTPRequestDuration,
		ServiceOrdersCreatedTotal,
		ServiceOrderStatusTransitionsTotal,
		ServiceOrderNotificationFailuresTotal,
	)
}
