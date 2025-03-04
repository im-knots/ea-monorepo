package metrics

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

// Prometheus metrics
var (
	StepCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ea_job_utils_handler_steps",
			Help: "Total number of steps executed in the handler function",
		},
		[]string{"path", "step", "type"},
	)

	RequestLatencyHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ea_job_utils_http_request_duration_seconds",
			Help:    "Histogram of latencies for HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method", "status_code"},
	)
)

func init() {
	// Register Prometheus metrics
	prometheus.MustRegister(StepCounter)
	prometheus.MustRegister(RequestLatencyHistogram)
}

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next() // Process request

		duration := time.Since(start).Seconds()
		path := c.Request.URL.Path
		method := c.Request.Method
		statusCode := c.Writer.Status()

		RequestLatencyHistogram.WithLabelValues(path, method, http.StatusText(statusCode)).Observe(duration)
	}
}
