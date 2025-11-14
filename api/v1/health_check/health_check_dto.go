// Package healthcheck provides health check endpoints for monitoring application liveness and readiness.
// These endpoints are designed for use with container orchestrators like Kubernetes.
package healthcheck

// HealthResponse represents the liveness probe response
type HealthResponse struct {
	// Status indicates if the application is alive (always "ok" if responding)
	Status string `json:"status" example:"ok"`
	// Timestamp of the health check in RFC3339 format
	Timestamp string `json:"timestamp" example:"2025-11-14T10:30:00Z"`
	// Version of the application
	Version string `json:"version" example:"v1.0.0"`
	// Uptime duration since application started
	Uptime string `json:"uptime,omitempty" example:"2h15m30s"`
	// Stats contains optional runtime statistics
	Stats *RuntimeStats `json:"stats,omitempty"`
}

// RuntimeStats represents runtime statistics of the application
type RuntimeStats struct {
	// Goroutines is the number of active goroutines
	Goroutines int `json:"goroutines" example:"42"`
	// MemoryAlloc is the allocated memory in megabytes
	MemoryAlloc uint64 `json:"memory_alloc_mb" example:"128"`
	// HeapObjects is the number of allocated heap objects
	HeapObjects uint64 `json:"heap_objects" example:"250000"`
}

// CheckResult represents the result of a single dependency check
type CheckResult struct {
	// Status indicates if the check passed ("ok") or failed ("failed")
	Status string `json:"status" example:"ok" enums:"ok,failed"`
	// Message provides additional context when a check fails
	Message string `json:"message,omitempty" example:"database unavailable: connection timeout"`
	// Latency indicates how long the check took to complete
	Latency string `json:"latency,omitempty" example:"15ms"`
}

// ReadinessResponse represents the readiness probe response
type ReadinessResponse struct {
	// Status indicates overall readiness ("ready" or "not_ready")
	Status string `json:"status" example:"ready" enums:"ready,not_ready"`
	// Checks contains the status of each dependency check
	Checks map[string]CheckResult `json:"checks"`
	// Timestamp of the readiness check in RFC3339 format
	Timestamp string `json:"timestamp" example:"2025-11-14T10:30:00Z"`
}
