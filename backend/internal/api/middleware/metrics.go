package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// RequestsTotal is a counter for total requests by method, path, and status
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "aiproxy_requests_total",
			Help: "Total number of requests by method, path, and status",
		},
		[]string{"method", "path", "status"},
	)

	// RequestDuration is a histogram for request duration in seconds
	RequestDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "aiproxy_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
	)

	// ActiveStreams is a gauge for active streaming connections
	ActiveStreams = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "aiproxy_active_streams",
			Help: "Number of active streaming connections",
		},
	)

	// UpstreamLatency is a histogram for upstream provider latency
	UpstreamLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "aiproxy_upstream_latency_seconds",
			Help:    "Latency of upstream provider requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"provider"},
	)

	// TokensTotal is a counter for token usage by type and provider
	TokensTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "aiproxy_tokens_total",
			Help: "Total tokens processed by type (prompt/completion) and provider",
		},
		[]string{"type", "provider"},
	)

	// RTKBytesSaved is a counter for bytes saved by RTK compression
	RTKBytesSaved = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "aiproxy_rtk_bytes_saved_total",
			Help: "Total bytes saved by RTK compression",
		},
	)
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

// Metrics returns a middleware that records Prometheus metrics
func Metrics() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				written:        false,
			}

			next.ServeHTTP(rw, r)

			duration := time.Since(start).Seconds()

			// Record metrics
			status := strconv.Itoa(rw.statusCode)
			RequestsTotal.WithLabelValues(r.Method, r.URL.Path, status).Inc()
			RequestDuration.Observe(duration)
		})
	}
}

// RecordUpstreamLatency records latency for upstream provider requests
func RecordUpstreamLatency(provider string, duration time.Duration) {
	UpstreamLatency.WithLabelValues(provider).Observe(duration.Seconds())
}

// RecordTokens records token usage
func RecordTokens(tokenType, provider string, count int) {
	TokensTotal.WithLabelValues(tokenType, provider).Add(float64(count))
}

// RecordRTKBytesSaved records bytes saved by RTK compression
func RecordRTKBytesSaved(bytes int) {
	RTKBytesSaved.Add(float64(bytes))
}

// IncActiveStreams increments the active streams gauge
func IncActiveStreams() {
	ActiveStreams.Inc()
}

// DecActiveStreams decrements the active streams gauge
func DecActiveStreams() {
	ActiveStreams.Dec()
}
