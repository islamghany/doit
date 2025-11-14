# Health Check Package

This package provides health check endpoints for monitoring the DoIt API application health and readiness.

## Overview

The health check package implements two main endpoints designed for container orchestrators (Kubernetes, Docker Swarm, AWS ECS):

- **`/health`** - Liveness Probe
- **`/ready`** - Readiness Probe

## Architecture

```
┌─────────────────────────────────────────────────┐
│                 Load Balancer                   │
└───────────────────┬─────────────────────────────┘
                    │
        ┌───────────┴───────────┐
        ▼                       ▼
┌───────────────┐       ┌───────────────┐
│  Pod 1        │       │  Pod 2        │
│               │       │               │
│  /health  ✅  │       │  /health  ✅  │
│  /ready   ✅  │       │  /ready   ❌  │ ← Failing readiness
│               │       │               │
│  Gets traffic │       │  No traffic   │
└───────────────┘       └───────────────┘
```

## Files

- **`health_check_handler.go`** - Handler implementation with probe logic
- **`health_check_dto.go`** - Response data structures
- **`health_check_routes.go`** - Route registration
- **`README.md`** - This file

## Handler Structure

```go
type Handler struct {
    log     *logger.Logger    // Structured logger
    db      *database.Pool     // PostgreSQL connection pool
    cache   cache.Cache        // Redis cache interface
    version string             // Application version
}
```

## Endpoints

### 1. Health Check (Liveness Probe)

**Route:** `GET /health`

**Purpose:** Verify the application process is alive and responsive.

**Implementation:**

```go
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) error {
    response := HealthResponse{
        Status:    "ok",
        Timestamp: time.Now().Format(time.RFC3339),
        Version:   h.version,
        Uptime:    time.Since(startupTime).String(),
    }
    return web.RespondOK(w, r, response)
}
```

**Response:**

```json
{
  "status": "ok",
  "timestamp": "2025-11-14T10:30:00Z",
  "version": "v1.0.0",
  "uptime": "2h15m30s"
}
```

**Characteristics:**

- ⚡ Ultra-fast (< 10ms)
- 🚫 No database queries
- 🚫 No Redis checks
- ✅ Always returns 200 OK if process is alive

**Kubernetes Action on Failure:** Restart container

---

### 2. Readiness Check (Readiness Probe)

**Route:** `GET /ready`

**Purpose:** Verify all dependencies are available and application can serve traffic.

**Implementation:**

```go
func (h *Handler) ReadyCheck(w http.ResponseWriter, r *http.Request) error {
    ctx := r.Context()
    checks := make(map[string]CheckResult)
    allHealthy := true

    // Check Database
    dbCheck := h.checkDatabase(ctx)
    checks["database"] = dbCheck
    if dbCheck.Status == "failed" {
        allHealthy = false
    }

    // Check Cache
    cacheCheck := h.checkCache(ctx)
    checks["redis"] = cacheCheck
    if cacheCheck.Status == "failed" {
        allHealthy = false
    }

    // Check Disk Space
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
```

**Success Response (200 OK):**

```json
{
  "status": "ready",
  "checks": {
    "database": { "status": "ok", "latency": "12ms" },
    "redis": { "status": "ok", "latency": "3ms" },
    "disk_space": { "status": "ok", "latency": "1ms" }
  },
  "timestamp": "2025-11-14T10:30:00Z"
}
```

**Failure Response (503 Service Unavailable):**

```json
{
  "status": "not_ready",
  "checks": {
    "database": { "status": "ok", "latency": "15ms" },
    "redis": {
      "status": "failed",
      "message": "redis unavailable: connection timeout",
      "latency": "5.2s"
    },
    "disk_space": { "status": "ok", "latency": "1ms" }
  },
  "timestamp": "2025-11-14T10:30:00Z"
}
```

**Characteristics:**

- 🔍 Comprehensive checks
- ⏱️ Reasonable timeout (< 500ms)
- 📊 Detailed check results
- ✅ Returns 200 when all pass
- ⚠️ Returns 503 when any fail

**Kubernetes Action on Failure:** Remove from service endpoints (no restart)

---

## Dependency Checks

### Database Check

```go
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
```

Uses `database.Pool.StatusCheck()` which performs `PING` to PostgreSQL.

### Cache Check

```go
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
```

Uses `cache.Cache.Ping()` which performs `PING` to Redis.

### File System Check

```go
func (h *Handler) checkFileSystem() CheckResult {
    start := time.Now()

    // Can be extended with syscall.Statfs for Unix systems
    // or windows.GetDiskFreeSpaceEx for Windows

    return CheckResult{
        Status:  "ok",
        Latency: time.Since(start).String(),
    }
}
```

Currently returns `ok`. Can be extended for actual disk space checking.

---

## Data Structures

### HealthResponse

```go
type HealthResponse struct {
    Status    string        `json:"status" example:"ok"`
    Timestamp string        `json:"timestamp" example:"2025-11-14T10:30:00Z"`
    Version   string        `json:"version" example:"v1.0.0"`
    Uptime    string        `json:"uptime,omitempty" example:"2h15m30s"`
    Stats     *RuntimeStats `json:"stats,omitempty"`
}
```

### ReadinessResponse

```go
type ReadinessResponse struct {
    Status    string                 `json:"status" example:"ready" enums:"ready,not_ready"`
    Checks    map[string]CheckResult `json:"checks"`
    Timestamp string                 `json:"timestamp" example:"2025-11-14T10:30:00Z"`
}
```

### CheckResult

```go
type CheckResult struct {
    Status  string `json:"status" example:"ok" enums:"ok,failed"`
    Message string `json:"message,omitempty" example:"database unavailable: connection timeout"`
    Latency string `json:"latency,omitempty" example:"15ms"`
}
```

---

## Integration

### Route Registration

```go
// In api/server.go
func NewServer(logger *logger.Logger, cfg *config.Config, dbPool *database.Pool, cache cache.Cache) (http.Handler, error) {
    // ... initialization ...

    healthcheckHandler := healthcheck.NewHandler(logger, dbPool, cache, cfg.App.Version)
    healthcheck.RegisterRoutes(app, healthcheckHandler)

    // ... other routes ...
}
```

### Routes File

```go
// health_check_routes.go
func RegisterRoutes(app *web.WebApp, handler *Handler) {
    app.Handle("GET", "/health", handler.HealthCheck)
    app.Handle("GET", "/ready", handler.ReadyCheck)
}
```

**Note:** Health endpoints have **NO authentication** middleware - they must be publicly accessible for orchestrators.

---

## OpenAPI Documentation

Both endpoints are fully documented with OpenAPI/Swagger annotations:

- View in Swagger UI: `http://localhost:8080/swagger/index.html`
- JSON schema: `http://localhost:8080/swagger/doc.json`
- YAML schema: Available in `docs/swagger.yaml`

**Generate docs:**

```bash
make swagger
```

---

## Testing

### Manual Testing

```bash
# Test health endpoint
curl http://localhost:8080/health

# Test readiness endpoint
curl http://localhost:8080/ready

# Test with pretty JSON
curl -s http://localhost:8080/ready | jq .

# Test with timing
curl -w "\nResponse Time: %{time_total}s\n" http://localhost:8080/health
```

### Unit Testing Example

```go
func TestHealthCheck(t *testing.T) {
    logger := logger.New(os.Stdout, logger.LevelInfo, "test", nil, logger.Events{})
    handler := healthcheck.NewHandler(logger, nil, nil, "v1.0.0")

    req := httptest.NewRequest("GET", "/health", nil)
    w := httptest.NewRecorder()

    err := handler.HealthCheck(w, req)
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, w.Code)

    var response healthcheck.HealthResponse
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.Equal(t, "ok", response.Status)
    assert.Equal(t, "v1.0.0", response.Version)
}
```

---

## Kubernetes Configuration Example

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: doit-api
spec:
  replicas: 3
  template:
    spec:
      containers:
        - name: doit
          image: doit:latest
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 10
            periodSeconds: 10
            failureThreshold: 3
          readinessProbe:
            httpGet:
              path: /ready
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 5
            failureThreshold: 2
```

---

## Best Practices

### ✅ Do

- Keep liveness checks lightweight (< 10ms)
- Check all critical dependencies in readiness
- Return detailed error messages in readiness failures
- Set appropriate timeouts for checks
- Monitor probe failure rates
- Use circuit breakers for external dependencies

### ❌ Don't

- Check database in liveness probe
- Make external API calls in liveness
- Set too aggressive probe intervals
- Return 200 when dependencies are down
- Skip health checks in production
- Forget to configure timeouts

---

## Monitoring

### Metrics to Track

1. **Health endpoint response time** (should be < 10ms)
2. **Readiness endpoint response time** (should be < 500ms)
3. **Individual check latencies** (database, redis, disk)
4. **Probe failure rate** (% of checks that fail)
5. **Time to recover** (how long until readiness passes again)

### Alerting

- Alert on repeated liveness failures (pod restart loops)
- Alert on high readiness failure rate (> 5%)
- Alert on degraded check latencies (database > 100ms)
- Alert on sustained not-ready state (> 1 minute)

---

## Related Documentation

- [Full Health Endpoints Guide](../../../docs/HEALTH_ENDPOINTS.md)
- [Quick Reference Card](../../../docs/HEALTH_QUICK_REFERENCE.md)
- [Kubernetes Documentation](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)

---

## Dependencies

- `doit/pkg/database` - PostgreSQL connection pool with `StatusCheck()` method
- `doit/internal/cache` - Redis cache interface with `Ping()` method
- `doit/internal/web` - Web framework with response helpers
- `doit/pkg/logger` - Structured logging

---

## Version History

- **v1.0.0** (2025-11-14) - Initial implementation with liveness and readiness probes
  - Health endpoint with uptime and version
  - Readiness endpoint with database, Redis, and disk checks
  - Full OpenAPI documentation
  - Kubernetes-ready configuration

---

**Package:** `doit/api/v1/health_check`  
**Maintainer:** DoIt Team  
**Last Updated:** November 14, 2025
