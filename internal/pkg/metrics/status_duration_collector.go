package metrics

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	soHistoryDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order_history/domain"
	"github.com/prometheus/client_golang/prometheus"
)

// StatusDurationReader is the read model the collector needs — satisfied by
// interfaces.ServiceOrderStatusDurationReader without importing that package
// here, keeping this package dependency-free of the domain it measures.
type StatusDurationReader interface {
	AverageDurationByStatusInHours(ctx context.Context) ([]soHistoryDomain.StatusDuration, error)
}

var statusDurationDesc = prometheus.NewDesc(
	"service_order_status_avg_duration_hours",
	"Average hours a service order takes to reach a status from whatever status preceded it.",
	[]string{"status"}, nil,
)

type statusDurationCollector struct {
	reader StatusDurationReader
}

// NewStatusDurationCollector queries the database on every Prometheus scrape
// rather than on a timer: the metric is low-cardinality and slow-changing,
// so the extra query per scrape interval is cheap and needs no goroutine
// lifecycle to manage.
func NewStatusDurationCollector(reader StatusDurationReader) prometheus.Collector {
	return &statusDurationCollector{reader: reader}
}

func (c *statusDurationCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- statusDurationDesc
}

func (c *statusDurationCollector) Collect(ch chan<- prometheus.Metric) {
	durations, err := c.reader.AverageDurationByStatusInHours(context.Background())
	if err != nil {
		logger.Global().Error(err)
		return
	}

	for _, d := range durations {
		ch <- prometheus.MustNewConstMetric(statusDurationDesc, prometheus.GaugeValue, d.AverageHours, d.Status.String())
	}
}
