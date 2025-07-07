package metrics

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "url_shortener_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "url_shortener_http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// URL shortener specific metrics
	URLsCreatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "url_shortener_urls_created_total",
			Help: "Total number of URLs created",
		},
		[]string{"status"},
	)

	URLClicksTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "url_shortener_clicks_total",
			Help: "Total number of URL clicks",
		},
		[]string{"short_code"},
	)

	URLRedirectsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "url_shortener_redirects_total",
			Help: "Total number of URL redirects",
		},
		[]string{"status"},
	)

	// Database metrics
	DatabaseOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "url_shortener_database_operations_total",
			Help: "Total number of database operations",
		},
		[]string{"operation", "status"},
	)

	DatabaseOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "url_shortener_database_operation_duration_seconds",
			Help:    "Duration of database operations in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	// Cache metrics
	CacheOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "url_shortener_cache_operations_total",
			Help: "Total number of cache operations",
		},
		[]string{"operation", "status"},
	)

	CacheHitRatio = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "url_shortener_cache_hit_ratio",
			Help: "Cache hit ratio",
		},
		[]string{"cache_type"},
	)

	// Rate limiter metrics
	RateLimitedRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "url_shortener_rate_limited_requests_total",
			Help: "Total number of rate limited requests",
		},
		[]string{"ip"},
	)

	ActiveRateLimiters = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "url_shortener_active_rate_limiters",
			Help: "Number of active rate limiters",
		},
	)

	// Worker metrics
	WorkerQueueSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "url_shortener_worker_queue_size",
			Help: "Size of worker queue",
		},
		[]string{"worker_type"},
	)

	WorkerTasksProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "url_shortener_worker_tasks_processed_total",
			Help: "Total number of tasks processed by workers",
		},
		[]string{"worker_type", "status"},
	)

	WorkerTaskDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "url_shortener_worker_task_duration_seconds",
			Help:    "Duration of worker tasks in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"worker_type"},
	)

	// System metrics
	ActiveConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "url_shortener_active_connections",
			Help: "Number of active connections",
		},
	)

	// Error metrics
	ErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "url_shortener_errors_total",
			Help: "Total number of errors",
		},
		[]string{"component", "error_type"},
	)
)

// CacheStats holds cache statistics
type CacheStats struct {
	mu       sync.RWMutex
	hits     int64
	misses   int64
	requests int64
}

// NewCacheStats creates a new cache statistics tracker
func NewCacheStats() *CacheStats {
	return &CacheStats{}
}

// RecordHit records a cache hit
func (cs *CacheStats) RecordHit() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.hits++
	cs.requests++
}

// RecordMiss records a cache miss
func (cs *CacheStats) RecordMiss() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.misses++
	cs.requests++
}

// HitRatio returns the current hit ratio
func (cs *CacheStats) HitRatio() float64 {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	if cs.requests == 0 {
		return 0
	}
	return float64(cs.hits) / float64(cs.requests)
}

// Stats returns current statistics
func (cs *CacheStats) Stats() (hits, misses, requests int64) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.hits, cs.misses, cs.requests
}

// Timer is a helper for timing operations
type Timer struct {
	Start time.Time
}

// NewTimer creates a new timer
func NewTimer() *Timer {
	return &Timer{Start: time.Now()}
}

// ObserveDuration observes the duration since the timer was created
func (t *Timer) ObserveDuration(observer prometheus.Observer) {
	duration := time.Since(t.Start).Seconds()
	observer.Observe(duration)
}

// Duration returns the duration since the timer was created
func (t *Timer) Duration() time.Duration {
	return time.Since(t.Start)
}

// Helper functions for common metric operations

// RecordHTTPRequest records an HTTP request
func RecordHTTPRequest(method, endpoint, statusCode string, duration time.Duration) {
	HTTPRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
	HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

// RecordURLCreated records a URL creation
func RecordURLCreated(status string) {
	URLsCreatedTotal.WithLabelValues(status).Inc()
}

// RecordURLClick records a URL click
func RecordURLClick(shortCode string) {
	URLClicksTotal.WithLabelValues(shortCode).Inc()
}

// RecordURLRedirect records a URL redirect
func RecordURLRedirect(status string) {
	URLRedirectsTotal.WithLabelValues(status).Inc()
}

// RecordDatabaseOperation records a database operation
func RecordDatabaseOperation(operation, status string, duration time.Duration) {
	DatabaseOperationsTotal.WithLabelValues(operation, status).Inc()
	DatabaseOperationDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

// RecordCacheOperation records a cache operation
func RecordCacheOperation(operation, status string) {
	CacheOperationsTotal.WithLabelValues(operation, status).Inc()
}

// UpdateCacheHitRatio updates the cache hit ratio
func UpdateCacheHitRatio(cacheType string, ratio float64) {
	CacheHitRatio.WithLabelValues(cacheType).Set(ratio)
}

// RecordRateLimitedRequest records a rate limited request
func RecordRateLimitedRequest(ip string) {
	RateLimitedRequestsTotal.WithLabelValues(ip).Inc()
}

// SetActiveRateLimiters sets the number of active rate limiters
func SetActiveRateLimiters(count float64) {
	ActiveRateLimiters.Set(count)
}

// SetWorkerQueueSize sets the worker queue size
func SetWorkerQueueSize(workerType string, size float64) {
	WorkerQueueSize.WithLabelValues(workerType).Set(size)
}

// RecordWorkerTask records a worker task
func RecordWorkerTask(workerType, status string, duration time.Duration) {
	WorkerTasksProcessed.WithLabelValues(workerType, status).Inc()
	WorkerTaskDuration.WithLabelValues(workerType).Observe(duration.Seconds())
}

// SetActiveConnections sets the number of active connections
func SetActiveConnections(count float64) {
	ActiveConnections.Set(count)
}

// RecordError records an error
func RecordError(component, errorType string) {
	ErrorsTotal.WithLabelValues(component, errorType).Inc()
}
