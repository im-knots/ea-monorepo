package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Prometheus metrics
	StepCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ea_agent_manager_handler_steps",
			Help: "Total number of steps executed in the handler function",
		},
		[]string{"path", "step", "type"},
	)

	RequestLatencyHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ea_agent_manager_http_request_duration_seconds",
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

// MetricsMiddleware tracks the latency of HTTP handler functions.
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		path := r.URL.Path
		method := r.Method

		// Use a response wrapper to capture the status code
		rw := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)

		duration := time.Since(start).Seconds()
		RequestLatencyHistogram.WithLabelValues(path, method, http.StatusText(rw.statusCode)).Observe(duration)
	})
}

// responseWriterWrapper wraps http.ResponseWriter to capture the status code.
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriterWrapper) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// MetricsHandler returns an HTTP handler for Prometheus metrics.
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}
