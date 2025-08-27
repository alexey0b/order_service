package monitoring

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type PrometheusMetrics struct {
	requestsTotal   prometheus.Counter
	requestDuration prometheus.Histogram
}

func NewPrometheusMetrics() (*PrometheusMetrics, error) {
	metrics := &PrometheusMetrics{
		requestsTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "app_requests_total",
				Help: "Общее количество запросов в приложение",
			}),

		requestDuration: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "app_request_duration_seconds",
				Help:    "Время обработки запроса",
				Buckets: prometheus.DefBuckets,
			},
		),
	}

	if err := prometheus.Register(metrics.requestsTotal); err != nil {
		return nil, fmt.Errorf("failed to registered metric: %w", err)
	}
	if err := prometheus.Register(metrics.requestDuration); err != nil {
		return nil, fmt.Errorf("failed to registered metric: %w", err)
	}

	return metrics, nil
}

func (m *PrometheusMetrics) IncRequest() {
	m.requestsTotal.Inc()
}

func (m *PrometheusMetrics) ObserveRequest(start time.Time) {
	duration := time.Since(start).Seconds()
	m.requestDuration.Observe(duration)
}
