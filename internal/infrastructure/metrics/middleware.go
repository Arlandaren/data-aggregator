package metrics

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	requestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_requests_total",
			Help: "Total number of requests received",
		},
		[]string{"method", "endpoint"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "app_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
)

type Prometheus struct {
	metrics []prometheus.Collector
}

func NewPrometheus() *Prometheus {
	return &Prometheus{
		metrics: []prometheus.Collector{
			requestCounter,
			requestDuration,
		},
	}
}

func (p *Prometheus) RegisterMetrics() {
	prometheus.MustRegister(p.metrics...)
}

func MetricsMiddleware() gin.HandlerFunc {
	prom := NewPrometheus()
	prom.RegisterMetrics()
	return func(c *gin.Context) {
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}
		timer := prometheus.NewTimer(requestDuration.WithLabelValues(c.Request.Method, path))
		defer timer.ObserveDuration()

		requestCounter.WithLabelValues(c.Request.Method, path).Inc()

		c.Next()
	}
}
