package monitoring

import (
	"net/http"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsMiddleware wraps an http.Handler and records metrics
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a custom response writer to capture the status code
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Call the next handler
		next.ServeHTTP(rw, r)

		// Record request metrics
		duration := time.Since(start)
		RecordRequest(r.URL.Path, http.StatusText(rw.statusCode), duration)

		// Update system metrics periodically (every 100th request)
		if RequestsTotal.WithLabelValues(r.URL.Path, "total").Inc(); true {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			UpdateMemoryUsage(float64(m.HeapInuse), float64(m.StackInuse))
			UpdateGoroutineCount(float64(runtime.NumGoroutine()))
		}
	})
}

// responseWriter wraps http.ResponseWriter to capture the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// PrometheusHandler returns the Prometheus metrics handler
func PrometheusHandler() http.Handler {
	return promhttp.Handler()
}
