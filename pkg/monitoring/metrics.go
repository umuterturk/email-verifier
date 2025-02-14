// Package monitoring provides metrics collection and monitoring functionality for the email validator service.
// It includes Prometheus metrics for tracking request rates, latencies, and various operational metrics.
package monitoring

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// RequestsTotal tracks the total number of requests
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "email_validator_requests_total",
			Help: "Total number of email validation requests",
		},
		[]string{"endpoint", "status"},
	)

	// RequestDuration tracks request duration
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "email_validator_request_duration_seconds",
			Help:    "Request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint"},
	)

	// ValidationScores tracks the distribution of validation scores
	ValidationScores = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "email_validator_scores",
			Help:    "Distribution of email validation scores",
			Buckets: []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
		},
		[]string{"validation_type"},
	)

	// CacheOperations tracks cache hits and misses
	CacheOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "email_validator_cache_operations_total",
			Help: "Total number of cache operations",
		},
		[]string{"operation", "result"},
	)

	// DNSLookupDuration tracks DNS lookup times
	DNSLookupDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "email_validator_dns_lookup_duration_seconds",
			Help:    "DNS lookup duration in seconds",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
		},
		[]string{"lookup_type"},
	)

	// ActiveGoroutines tracks the number of active goroutines
	ActiveGoroutines = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "email_validator_active_goroutines",
			Help: "Current number of active goroutines",
		},
	)

	// MemoryUsage tracks the memory usage
	MemoryUsage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "email_validator_memory_usage_bytes",
			Help: "Current memory usage in bytes",
		},
		[]string{"type"},
	)
)

// RecordRequest records metrics for an API request
func RecordRequest(endpoint, status string, duration time.Duration) {
	RequestsTotal.WithLabelValues(endpoint, status).Inc()
	RequestDuration.WithLabelValues(endpoint).Observe(duration.Seconds())
}

// RecordValidationScore records a validation score
func RecordValidationScore(validationType string, score float64) {
	ValidationScores.WithLabelValues(validationType).Observe(score)
}

// RecordCacheOperation records a cache hit or miss
func RecordCacheOperation(operation, result string) {
	CacheOperations.WithLabelValues(operation, result).Inc()
}

// RecordDNSLookup records DNS lookup duration
func RecordDNSLookup(lookupType string, duration time.Duration) {
	DNSLookupDuration.WithLabelValues(lookupType).Observe(duration.Seconds())
}

// UpdateGoroutineCount updates the active goroutine count
func UpdateGoroutineCount(count float64) {
	ActiveGoroutines.Set(count)
}

// UpdateMemoryUsage updates memory usage metrics
func UpdateMemoryUsage(heapInUse, stackInUse float64) {
	MemoryUsage.WithLabelValues("heap").Set(heapInUse)
	MemoryUsage.WithLabelValues("stack").Set(stackInUse)
}
