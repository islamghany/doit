package metrics

import (
	"expvar"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	requests   = expvar.NewInt("requests")   // current running goroutines
	errors     = expvar.NewInt("errors")     // total requests received.
	panics     = expvar.NewInt("panics")     // total errros occurred.
	goroutines = expvar.NewInt("goroutines") // total panics occurred.
	inFlight   = expvar.NewInt("in_flight")  // current in-flight requests.

	requestCount atomic.Int64 // For sampling check

	// Prometheus metrics

	// ===============================
	// HTTP metrics

	// RATE: Request count by method/path/status
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "doit",
			Subsystem: "api",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// In-flight HTTP requests (current load)
	HTTPRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "doit",
			Subsystem: "api",
			Name:      "requests_in_flight",
			Help:      "Current number of HTTP requests being processed",
		},
	)

	// DURATION: Request latency
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "doit",
			Subsystem: "api",
			Name:      "request_duration_seconds",
			Help:      "HTTP request duration in seconds",
			Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path"},
	)

	// ERRORS: Tracked via status label in HTTPRequestsTotal
	// Query: rate(doit_http_requests_total{status=~"5.."}[5m])

	// ===============================
	// Database metrics

	// Active database connections (gauge)
	DBConnectionsTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "doit",
			Subsystem: "db",
			Name:      "connections_total",
			Help:      "Total number of database connections in the pool",
		},
	)

	DBConnectionsInUse = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "doit",
			Subsystem: "db",
			Name:      "connections_in_use",
			Help:      "Number of database connections currently in use",
		},
	)

	DBConnectionsIdle = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "doit",
			Subsystem: "db",
			Name:      "connections_idle",
			Help:      "Number of idle database connections",
		},
	)
	// Database connection pool saturation (0-1, where 1 = fully saturated)
	DBPoolSaturation = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "doit",
			Subsystem: "db",
			Name:      "pool_saturation_ratio",
			Help:      "Database connection pool saturation (acquired/total)",
		},
	)

	// Waiting for connections (critical saturation indicator)
	DBPoolWaitCount = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: "doit",
			Subsystem: "db",
			Name:      "pool_wait_total",
			Help:      "Total number of times a connection was not immediately available",
		},
	)

	// Database query duration (histogram)
	DatabaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "doit",
			Subsystem: "db",
			Name:      "query_duration_seconds",
			Help:      "Database query duration in seconds",
			Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"operation", "table"},
	)

	// Query errors
	DBQueryErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "doit",
			Subsystem: "db",
			Name:      "query_errors_total",
			Help:      "Total number of database query errors",
		},
		[]string{"operation", "table"},
	)

	// ===============================
	// Cache metrics

	// Cache hits/misses
	CacheHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "doit",
			Subsystem: "cache",
			Name:      "hits_total",
			Help:      "Total number of cache hits",
		},
		[]string{"operation", "cache"}, // e.g., "redis", "local"
	)
	// Cache misses
	CacheMisses = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "doit",
			Subsystem: "cache",
			Name:      "misses_total",
			Help:      "Total number of cache misses",
		},
		[]string{"operation", "cache"},
	)

	// ===============================
	// Business Metrics

	// Todo operations count (create/update/delete/read)
	TodoOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "doit",
			Subsystem: "todo",
			Name:      "operations_total",
			Help:      "Total number of todo operations",
		},
		[]string{"operation"},
	)

	// ===============================
	// Runtime metrics

	// Goroutines
	Goroutines = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "doit",
			Subsystem: "runtime",
			Name:      "goroutines",
			Help:      "Current number of goroutines",
		},
	)

	// Go also exposes these automatically via prometheus/client_golang:
	// go_goroutines
	// go_memstats_*
	// go_gc_*
)

// ===============================
// Expvar metrics

func AddRequest() int64 {
	requests.Add(1)
	inFlight.Add(1)
	// Sample goroutines periodically (not every request)
	if requestCount.Add(1)%1000 == 0 {
		goroutines.Set(int64(runtime.NumGoroutine()))
	}
	return requests.Value()
}

func RequestDone() int64 {
	inFlight.Add(-1)
	return inFlight.Value()
}

func AddError() int64 {
	errors.Add(1)
	return errors.Value()
}

func AddPanics() int64 {
	panics.Add(1)
	return panics.Value()
}

// ===============================
// Prometheus metrics

// Update runtime metrics periodically
func StartRuntimeMetricsCollector(interval time.Duration) {
	go func() {
		for range time.Tick(interval) {
			Goroutines.Set(float64(runtime.NumGoroutine()))
		}
	}()
}

// Helper functions for common operations

func RecordHTTPRequest(method, path, status string, duration float64) {
	HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
	HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)
}

func RecordDatabaseQuery(operation, table string, duration float64, err error) {
	DatabaseQueryDuration.WithLabelValues(operation, table).Observe(duration)
	if err != nil {
		DBQueryErrors.WithLabelValues(operation, table).Inc()
	}
}

// UpdateDBPoolMetrics updates database connection pool metrics
// Call this periodically (e.g., every 15 seconds) from api.go
func UpdateDBPoolMetrics(totalConns, acquiredConns, idleConns int32) {
	DBConnectionsTotal.Set(float64(totalConns))
	DBConnectionsInUse.Set(float64(acquiredConns))
	DBConnectionsIdle.Set(float64(idleConns))
}

func RecordCacheOperation(operation, cache string, hit bool) {
	if hit {
		CacheHits.WithLabelValues(operation, cache).Inc()
	} else {
		CacheMisses.WithLabelValues(operation, cache).Inc()
	}
}

func RecordTodoOperation(operation string) {
	TodoOperationsTotal.WithLabelValues(operation).Inc()
}
