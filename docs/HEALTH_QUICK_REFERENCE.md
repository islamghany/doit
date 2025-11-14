# Health Endpoints Quick Reference

## Endpoints at a Glance

| Endpoint  | Purpose         | Target Response Time | Checks DB/Redis | HTTP Status |
| --------- | --------------- | -------------------- | --------------- | ----------- |
| `/health` | Liveness Probe  | < 10ms               | ❌ No           | 200         |
| `/ready`  | Readiness Probe | < 500ms              | ✅ Yes          | 200 / 503   |

---

## `/health` - Liveness Probe

```bash
curl http://localhost:8080/health
```

**Response (200 OK):**

```json
{
  "status": "ok",
  "timestamp": "2025-11-14T10:30:00Z",
  "version": "v1.0.0",
  "uptime": "2h15m30s"
}
```

**Use Case:** Kubernetes liveness probe - restart if fails  
**Checks:** Process is alive, not frozen  
**Action on Failure:** Container restart

---

## `/ready` - Readiness Probe

```bash
curl http://localhost:8080/ready
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

**Use Case:** Kubernetes readiness probe - remove from load balancer if fails  
**Checks:** Database, Redis, Disk Space  
**Action on Failure:** Remove from service endpoints (no restart)

---

## Kubernetes Probe Configuration

```yaml
# Minimal example
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  periodSeconds: 10
  failureThreshold: 3

readinessProbe:
  httpGet:
    path: /ready
    port: 8080
  periodSeconds: 5
  failureThreshold: 2
```

---

## Quick Test Commands

```bash
# Test health
curl -i http://localhost:8080/health

# Test readiness
curl -i http://localhost:8080/ready

# Test with timing
curl -w "\nTime: %{time_total}s\n" http://localhost:8080/health

# JSON pretty print
curl -s http://localhost:8080/ready | jq .

# Watch continuously (Linux/macOS)
watch -n 2 'curl -s http://localhost:8080/ready | jq .'
```

---

## Common Issues

| Issue          | Symptom            | Check                  | Solution                                          |
| -------------- | ------------------ | ---------------------- | ------------------------------------------------- |
| Pod restarting | `CrashLoopBackOff` | Liveness failing       | Increase `failureThreshold` or add `startupProbe` |
| No traffic     | 0/1 Ready          | Readiness failing      | Check DB/Redis connectivity with `/ready`         |
| Slow response  | Timeouts           | High latency in checks | Optimize dependency checks or increase timeout    |

---

## Decision Tree

```
Is the application process alive and responsive?
  ├─ Yes → /health returns 200 ✅
  └─ No → /health fails → Kubernetes restarts container 🔄

Are all dependencies (DB, Redis) available?
  ├─ Yes → /ready returns 200 ✅ → Traffic flows
  └─ No → /ready returns 503 ⚠️ → Removed from load balancer
```

---

## Monitoring Commands

```bash
# Watch health status
while true; do curl -s http://localhost:8080/health | jq -r '.status'; sleep 2; done

# Watch readiness with checks
while true; do
  echo "=== $(date) ==="
  curl -s http://localhost:8080/ready | jq '.'
  sleep 5
done

# Kubernetes pod status
kubectl get pods -w

# View probe failures
kubectl describe pod <pod-name> | grep -A 10 "Events:"
```

---

**Pro Tip:** In Swagger UI (`/swagger/index.html`), find these endpoints under the "Health" tag!
