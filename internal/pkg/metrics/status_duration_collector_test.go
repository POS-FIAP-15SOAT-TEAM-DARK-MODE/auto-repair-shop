package metrics_test

import (
	"context"
	"errors"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/metrics"
	soDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order/domain"
	soHistoryDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order_history/domain"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubStatusDurationReader struct {
	durations []soHistoryDomain.StatusDuration
	err       error
}

func (r stubStatusDurationReader) AverageDurationByStatusInHours(_ context.Context) ([]soHistoryDomain.StatusDuration, error) {
	return r.durations, r.err
}

func TestStatusDurationCollector_Collect(t *testing.T) {
	reader := stubStatusDurationReader{durations: []soHistoryDomain.StatusDuration{
		{Status: soDomain.SERVICE_ORDER_STATUS_IN_DIAGNOSIS, AverageHours: 2.5},
		{Status: soDomain.SERVICE_ORDER_STATUS_COMPLETED, AverageHours: 10},
	}}
	collector := metrics.NewStatusDurationCollector(reader)

	ch := make(chan prometheus.Metric, 10)
	collector.Collect(ch)
	close(ch)

	var got []*dto.Metric
	for m := range ch {
		var d dto.Metric
		require.NoError(t, m.Write(&d))
		got = append(got, &d)
	}
	require.Len(t, got, 2)

	assert.Equal(t, "IN_DIAGNOSIS", got[0].GetLabel()[0].GetValue())
	assert.Equal(t, 2.5, got[0].GetGauge().GetValue())
	assert.Equal(t, "COMPLETED", got[1].GetLabel()[0].GetValue())
	assert.Equal(t, float64(10), got[1].GetGauge().GetValue())
}

func TestStatusDurationCollector_Collect_ReaderErrorEmitsNothing(t *testing.T) {
	reader := stubStatusDurationReader{err: errors.New("db down")}
	collector := metrics.NewStatusDurationCollector(reader)

	ch := make(chan prometheus.Metric, 10)
	collector.Collect(ch)
	close(ch)

	var count int
	for range ch {
		count++
	}
	assert.Equal(t, 0, count)
}

func TestStatusDurationCollector_Describe(t *testing.T) {
	collector := metrics.NewStatusDurationCollector(stubStatusDurationReader{})

	ch := make(chan *prometheus.Desc, 10)
	collector.Describe(ch)
	close(ch)

	var count int
	for range ch {
		count++
	}
	assert.Equal(t, 1, count)
}
