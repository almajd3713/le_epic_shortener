package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

/**
Metrics that would be tracked:
- Total number of HTTP requests (Counter)
- Response times (Histogram)
- Error rates (Counter)
- Redis current cache size (Gauge)
*/

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	httpRequestErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_errors_total",
			Help: "Total number of HTTP request errors",
		},
		[]string{"method", "endpoint", "status"},
	)

	cacheSizeGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "redis_cache_size",
			Help: "Current size of the Redis cache",
		},
	)
)

func InitializeMetrics(registry *prometheus.Registry) {
	registry.MustRegister(httpRequestsTotal)
	registry.MustRegister(httpRequestDuration)
	registry.MustRegister(httpRequestErrorsTotal)
	registry.MustRegister(cacheSizeGauge)
}

func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		endpoint := c.FullPath()

		// Start timer
		timer := prometheus.NewTimer(httpRequestDuration.WithLabelValues(method, endpoint))
		defer timer.ObserveDuration()

		c.Next()

		status := c.Writer.Status()
		httpRequestsTotal.WithLabelValues(method, endpoint, string(rune(status))).Inc()

		if status >= 400 {
			httpRequestErrorsTotal.WithLabelValues(method, endpoint, string(rune(status))).Inc()
		}
	}
}