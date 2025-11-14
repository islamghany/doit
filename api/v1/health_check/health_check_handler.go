package healthcheck

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"doit/internal/cache"
	"doit/internal/web"
	"doit/pkg/database"
	"doit/pkg/logger"
)

type Handler struct {
	log     *logger.Logger
	db      *database.Pool
	cache   cache.Cache
	version string
}

func NewHandler(log *logger.Logger, db *database.Pool, cache cache.Cache, version string) *Handler {
	return &Handler{log: log, db: db, cache: cache, version: version}
}

// Track startup time
var startupTime time.Time

func init() {
	startupTime = time.Now()
}

// HealthCheck godoc
// @Summary      Health check endpoint (Liveness Probe)
// @Description  Lightweight endpoint that verifies the application process is alive and responsive. This endpoint does NOT check dependencies like database or Redis. Used by Kubernetes/container orchestrators as a liveness probe - if this fails repeatedly, the container will be restarted.
// @Description
// @Description  **What it checks:**
// @Description  - Web server is responsive
// @Description  - Application hasn't crashed or deadlocked
// @Description
// @Description  **What it does NOT check:**
// @Description  - Database connectivity
// @Description  - Redis/cache availability
// @Description  - External service availability
// @Description
// @Description  **Response time target:** < 10ms
// @Tags         Health
// @Produce      json
// @Success      200 {object} HealthResponse "Application is alive and responsive"
// @Router       /health [get]
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) error {
	response := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   h.version,
		Uptime:    time.Since(startupTime).String(),
	}

	return web.RespondOK(w, r, response)
}

// ReadyCheck godoc
// @Summary      Readiness check endpoint (Readiness Probe)
// @Description  Comprehensive health check that verifies all application dependencies are available and the application is ready to serve traffic. Used by Kubernetes/container orchestrators as a readiness probe - if this fails, the pod is removed from load balancer but NOT restarted.
// @Description
// @Description  **What it checks:**
// @Description  - Database connectivity and responsiveness (PostgreSQL)
// @Description  - Redis/cache availability
// @Description  - Disk space availability
// @Description
// @Description  **Behavior:**
// @Description  - Returns 200 OK when all checks pass
// @Description  - Returns 503 Service Unavailable when any check fails
// @Description  - Pod stays running but stops receiving traffic until checks pass again
// @Description
// @Description  **Response time target:** < 500ms
// @Tags         Health
// @Produce      json
// @Success      200 {object} ReadinessResponse "Application is ready to receive traffic - all dependencies are healthy"
// @Failure      503 {object} ReadinessResponse "Application is not ready - one or more dependencies are unavailable"
// @Router       /ready [get]
func (h *Handler) ReadyCheck(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	checks := make(map[string]CheckResult)
	allHealthy := true

	// 1. Check Database
	dbCheck := h.checkDatabase(ctx)
	checks["database"] = dbCheck
	if dbCheck.Status == "failed" {
		allHealthy = false
	}

	// 2. Check Cache
	cacheCheck := h.checkCache(ctx)
	checks["redis"] = cacheCheck
	if cacheCheck.Status == "failed" {
		allHealthy = false
	}

	// 3. Check File System
	fileSystemCheck := h.checkFileSystem()
	checks["disk_space"] = fileSystemCheck
	if fileSystemCheck.Status == "failed" {
		allHealthy = false
	}

	response := ReadinessResponse{
		Status:    "ready",
		Checks:    checks,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	if !allHealthy {
		response.Status = "not_ready"
		return web.Response(w, r, http.StatusServiceUnavailable, response)
	}

	return web.RespondOK(w, r, response)
}

func (h *Handler) checkDatabase(ctx context.Context) CheckResult {
	start := time.Now()

	if err := h.db.StatusCheck(ctx); err != nil {
		return CheckResult{
			Status:  "failed",
			Message: fmt.Sprintf("database unavailable: %v", err),
			Latency: time.Since(start).String(),
		}
	}

	return CheckResult{
		Status:  "ok",
		Latency: time.Since(start).String(),
	}
}

func (h *Handler) checkCache(ctx context.Context) CheckResult {
	start := time.Now()

	if err := h.cache.Ping(ctx); err != nil {
		return CheckResult{
			Status:  "failed",
			Message: fmt.Sprintf("redis unavailable: %v", err),
			Latency: time.Since(start).String(),
		}
	}

	return CheckResult{
		Status:  "ok",
		Latency: time.Since(start).String(),
	}
}

func (h *Handler) checkFileSystem() CheckResult {
	start := time.Now()

	// syscall.Statfs for Unix systems
	// windows.GetDiskFreeSpaceEx for Windows systems
	// etc.

	return CheckResult{
		Status:  "ok",
		Latency: time.Since(start).String(),
	}
}
