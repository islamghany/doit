# Health Check Endpoints Documentation

This document provides detailed information about the health check endpoints used for monitoring application health and readiness.

## Table of Contents

- [Overview](#overview)
- [Endpoints](#endpoints)
  - [Health Check (Liveness Probe)](#health-check-liveness-probe)
  - [Readiness Check (Readiness Probe)](#readiness-check-readiness-probe)
- [Kubernetes Configuration](#kubernetes-configuration)
- [Testing](#testing)

---

## Overview

The DoIt API provides two health check endpoints designed for use with container orchestrators like Kubernetes, Docker Swarm, or AWS ECS:

| Endpoint  | Purpose         | Checks Dependencies | Failure Action            |
| --------- | --------------- | ------------------- | ------------------------- |
| `/health` | Liveness Probe  | ❌ No               | Restart container         |
| `/ready`  | Readiness Probe | ✅ Yes              | Remove from load balancer |

**Key Differences:**

- **Liveness (`/health`)**: Verifies the application process is alive and not frozen
- **Readiness (`/ready`)**: Verifies the application can serve traffic (all dependencies healthy)

---

## Endpoints

### Health Check (Liveness Probe)

**Endpoint:** `GET /health`

**Purpose:** Lightweight check to verify the application is alive and responsive.

**Response Time:** < 10ms

**What it checks:**

- ✅ Web server is responsive
- ✅ Application hasn't crashed or deadlocked

**What it does NOT check:**

- ❌ Database connectivity
- ❌ Redis/cache availability
- ❌ External service availability

#### Success Response (200 OK)

```json
{
  "status": "ok",
  "timestamp": "2025-11-14T10:30:00Z",
  "version": "v1.0.0",
  "uptime": "2h15m30s"
}
```

#### Response Fields

| Field       | Type   | Description                                      |
| ----------- | ------ | ------------------------------------------------ |
| `status`    | string | Always "ok" if the application is responding     |
| `timestamp` | string | Current time in RFC3339 format                   |
| `version`   | string | Application version                              |
| `uptime`    | string | Duration since application started               |
| `stats`     | object | Optional runtime statistics (goroutines, memory) |

#### Example Request

```bash
curl -X GET http://localhost:8080/health
```

#### Example with curl (with timing)

```bash
curl -w "\nResponse Time: %{time_total}s\n" http://localhost:8080/health
```

---

### Readiness Check (Readiness Probe)

**Endpoint:** `GET /ready`

**Purpose:** Comprehensive check to verify all dependencies are available and the application is ready to serve traffic.

**Response Time:** < 500ms

**What it checks:**

- ✅ Database connectivity (PostgreSQL ping)
- ✅ Redis/cache availability (Redis PING)
- ✅ Disk space availability

**Behavior:**

- Returns `200 OK` when all checks pass
- Returns `503 Service Unavailable` when any check fails
- Pod stays running but stops receiving traffic until checks pass again

#### Success Response (200 OK)

```json
{
  "status": "ready",
  "checks": {
    "database": {
      "status": "ok",
      "latency": "12ms"
    },
    "redis": {
      "status": "ok",
      "latency": "3ms"
    },
    "disk_space": {
      "status": "ok",
      "latency": "1ms"
    }
  },
  "timestamp": "2025-11-14T10:30:00Z"
}
```

#### Failure Response (503 Service Unavailable)

```json
{
  "status": "not_ready",
  "checks": {
    "database": {
      "status": "ok",
      "latency": "15ms"
    },
    "redis": {
      "status": "failed",
      "message": "redis unavailable: connection timeout",
      "latency": "5.2s"
    },
    "disk_space": {
      "status": "ok",
      "latency": "1ms"
    }
  },
  "timestamp": "2025-11-14T10:30:00Z"
}
```

#### Response Fields

| Field                   | Type   | Description                                   |
| ----------------------- | ------ | --------------------------------------------- |
| `status`                | string | "ready" or "not_ready"                        |
| `checks`                | object | Map of dependency checks and their results    |
| `checks.<name>.status`  | string | "ok" or "failed"                              |
| `checks.<name>.message` | string | Error message (only present when check fails) |
| `checks.<name>.latency` | string | How long the check took                       |
| `timestamp`             | string | Current time in RFC3339 format                |

#### Example Request

```bash
curl -X GET http://localhost:8080/ready
```

#### Example with verbose output

```bash
curl -v http://localhost:8080/ready
```

---

## Kubernetes Configuration

### Deployment YAML Example

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: doit-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: doit-api
  template:
    metadata:
      labels:
        app: doit-api
    spec:
      containers:
        - name: doit
          image: doit:latest
          ports:
            - containerPort: 8080
              name: http

          # Startup probe - runs first, only during startup
          # Gives the app time to initialize before liveness kicks in
          startupProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 0
            periodSeconds: 5
            failureThreshold: 12 # Max 60 seconds startup time
            timeoutSeconds: 2

          # Liveness probe - detects if app is frozen/crashed
          # If this fails, Kubernetes will restart the pod
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 10
            periodSeconds: 10 # Check every 10 seconds
            timeoutSeconds: 2
            failureThreshold: 3 # Restart after 3 consecutive failures

          # Readiness probe - detects if app can serve traffic
          # If this fails, pod is removed from Service endpoints
          readinessProbe:
            httpGet:
              path: /ready
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 5 # Check every 5 seconds
            timeoutSeconds: 2
            successThreshold: 1 # Consider ready after 1 success
            failureThreshold: 2 # Remove from LB after 2 failures

          env:
            - name: DB_HOST
              value: postgres-service
            - name: REDIS_ADDR
              value: redis-service:6379

          resources:
            requests:
              memory: "128Mi"
              cpu: "100m"
            limits:
              memory: "256Mi"
              cpu: "200m"
```

### Service YAML Example

```yaml
apiVersion: v1
kind: Service
metadata:
  name: doit-api-service
spec:
  selector:
    app: doit-api
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8080
  type: LoadBalancer
```

---

## Testing

### Manual Testing

#### 1. Test Health Endpoint

```bash
# Should always return 200 OK if app is running
curl -i http://localhost:8080/health
```

**Expected Response:**

```
HTTP/1.1 200 OK
Content-Type: application/json
...

{"status":"ok","timestamp":"2025-11-14T10:30:00Z","version":"v1.0.0","uptime":"5m30s"}
```

#### 2. Test Readiness Endpoint (Happy Path)

```bash
# Should return 200 OK when all dependencies are healthy
curl -i http://localhost:8080/ready
```

**Expected Response:**

```
HTTP/1.1 200 OK
Content-Type: application/json
...

{"status":"ready","checks":{...},"timestamp":"2025-11-14T10:30:00Z"}
```

#### 3. Test Readiness Endpoint (Failure Scenario)

Stop Redis to simulate a dependency failure:

```bash
# Stop Redis
docker stop <redis-container-id>

# Check readiness
curl -i http://localhost:8080/ready
```

**Expected Response:**

```
HTTP/1.1 503 Service Unavailable
Content-Type: application/json
...

{"status":"not_ready","checks":{"redis":{"status":"failed","message":"redis unavailable: connection refused"}},...}
```

### Automated Testing with Scripts

#### Health Check Monitor Script

```bash
#!/bin/bash
# monitor-health.sh

ENDPOINT="http://localhost:8080/health"
INTERVAL=5

echo "Monitoring health endpoint: $ENDPOINT"
echo "Checking every $INTERVAL seconds..."
echo "Press Ctrl+C to stop"
echo ""

while true; do
    RESPONSE=$(curl -s -w "\n%{http_code}" $ENDPOINT)
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | head -n-1)

    if [ "$HTTP_CODE" -eq 200 ]; then
        echo "✅ [$(date '+%Y-%m-%d %H:%M:%S')] Health: OK"
    else
        echo "❌ [$(date '+%Y-%m-%d %H:%M:%S')] Health: FAILED (HTTP $HTTP_CODE)"
    fi

    sleep $INTERVAL
done
```

#### Readiness Check Monitor Script

```bash
#!/bin/bash
# monitor-readiness.sh

ENDPOINT="http://localhost:8080/ready"
INTERVAL=5

echo "Monitoring readiness endpoint: $ENDPOINT"
echo "Checking every $INTERVAL seconds..."
echo "Press Ctrl+C to stop"
echo ""

while true; do
    RESPONSE=$(curl -s -w "\n%{http_code}" $ENDPOINT)
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | head -n-1)
    STATUS=$(echo "$BODY" | jq -r '.status' 2>/dev/null)

    if [ "$HTTP_CODE" -eq 200 ]; then
        echo "✅ [$(date '+%Y-%m-%d %H:%M:%S')] Readiness: READY"
    else
        echo "❌ [$(date '+%Y-%m-%d %H:%M:%S')] Readiness: NOT READY (HTTP $HTTP_CODE)"

        # Show which checks failed
        FAILED_CHECKS=$(echo "$BODY" | jq -r '.checks | to_entries[] | select(.value.status == "failed") | .key' 2>/dev/null)
        if [ ! -z "$FAILED_CHECKS" ]; then
            echo "   Failed checks: $FAILED_CHECKS"
        fi
    fi

    sleep $INTERVAL
done
```

### Load Testing

Test how health checks perform under load:

```bash
# Install Apache Bench if not available
# brew install apache-bench (macOS)
# apt-get install apache2-utils (Ubuntu)

# Test health endpoint (10000 requests, 100 concurrent)
ab -n 10000 -c 100 http://localhost:8080/health

# Test readiness endpoint (1000 requests, 50 concurrent)
ab -n 1000 -c 50 http://localhost:8080/ready
```

---

## Best Practices

### 1. **Keep Liveness Checks Simple**

- ✅ Return 200 OK immediately
- ✅ Don't check dependencies
- ❌ Don't perform heavy computations
- ❌ Don't make database queries

### 2. **Make Readiness Checks Comprehensive**

- ✅ Check all critical dependencies
- ✅ Keep timeout reasonable (< 500ms)
- ✅ Return detailed check results
- ✅ Use circuit breakers for external services

### 3. **Configure Probes Appropriately**

- ✅ Set `startupProbe` for slow-starting apps
- ✅ Use longer `periodSeconds` for liveness (10-30s)
- ✅ Use shorter `periodSeconds` for readiness (5-10s)
- ✅ Set appropriate `failureThreshold` to avoid flapping

### 4. **Monitor Health Check Failures**

- ✅ Alert on repeated liveness failures
- ✅ Track readiness probe failure rate
- ✅ Monitor check latencies
- ✅ Correlate with deployment events

### 5. **Graceful Shutdown Integration**

- ✅ Mark as not ready on SIGTERM
- ✅ Wait for readiness probes to fail
- ✅ Drain in-flight requests
- ✅ Close connections cleanly

---

## Troubleshooting

### Pod Keeps Restarting

**Symptom:** Pod is in `CrashLoopBackOff` state

**Possible Causes:**

1. Liveness probe is too aggressive (low `failureThreshold` or `periodSeconds`)
2. Application is slow to start (add `startupProbe`)
3. Application is actually crashing (check logs)

**Solution:**

```bash
# Check pod events
kubectl describe pod <pod-name>

# Check application logs
kubectl logs <pod-name>

# Adjust probe settings in deployment
kubectl edit deployment doit-api
```

### Pod Not Receiving Traffic

**Symptom:** Service endpoints are empty or requests timeout

**Possible Causes:**

1. Readiness probe is failing
2. Dependencies are unavailable
3. Wrong port configuration

**Solution:**

```bash
# Check readiness status
kubectl get pods
# Look for "0/1 Ready"

# Check endpoint status
kubectl describe endpoints doit-api-service

# Test readiness manually
kubectl port-forward <pod-name> 8080:8080
curl http://localhost:8080/ready
```

---

## Related Documentation

- [Kubernetes Liveness, Readiness, and Startup Probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
- [Swagger API Documentation](http://localhost:8080/swagger/index.html) (when running)
- [Graceful Shutdown Implementation](../README.md#13-graceful-shutdown--health-checks)

---

## Examples in Other Languages

### Python (requests)

```python
import requests
import time

def check_health():
    try:
        response = requests.get('http://localhost:8080/health', timeout=2)
        return response.status_code == 200
    except:
        return False

def check_readiness():
    try:
        response = requests.get('http://localhost:8080/ready', timeout=2)
        data = response.json()
        return data['status'] == 'ready'
    except:
        return False

# Monitor health
while True:
    if check_health():
        print(f"✅ Health: OK")
    else:
        print(f"❌ Health: FAILED")

    if check_readiness():
        print(f"✅ Readiness: READY")
    else:
        print(f"❌ Readiness: NOT READY")

    time.sleep(5)
```

### JavaScript (fetch)

```javascript
async function checkHealth() {
  try {
    const response = await fetch("http://localhost:8080/health");
    return response.ok;
  } catch {
    return false;
  }
}

async function checkReadiness() {
  try {
    const response = await fetch("http://localhost:8080/ready");
    const data = await response.json();
    return data.status === "ready";
  } catch {
    return false;
  }
}

// Monitor health
setInterval(async () => {
  const healthy = await checkHealth();
  const ready = await checkReadiness();

  console.log(`Health: ${healthy ? "✅ OK" : "❌ FAILED"}`);
  console.log(`Readiness: ${ready ? "✅ READY" : "❌ NOT READY"}`);
}, 5000);
```

---

**Last Updated:** November 14, 2025  
**API Version:** v1.0.0
