package prometheus

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	totalRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "backend", "code"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of response time for handler",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"backend"},
	)
)

func Observe(method, backend, code string, start time.Time) {
	duration := time.Since(start).Seconds()
	totalRequests.WithLabelValues(method, backend, code).Inc()
	requestDuration.WithLabelValues(backend).Observe(duration)
}

func MustLoad() {
	prometheus.MustRegister(totalRequests, requestDuration)
}
