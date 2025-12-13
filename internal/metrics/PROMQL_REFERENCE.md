# PromQL Reference for DoIt API Metrics

This document contains all PromQL queries for monitoring the DoIt API.
Use these queries in Prometheus or Grafana dashboards.

---

## Table of Contents

1. [The 4 Golden Signals](#the-4-golden-signals)
2. [HTTP Metrics](#http-metrics)
3. [Database Metrics](#database-metrics)
4. [Cache Metrics](#cache-metrics)
5. [Business Metrics](#business-metrics)
6. [Runtime Metrics](#runtime-metrics)
7. [Go Runtime (Automatic)](#go-runtime-automatic)
8. [Alerting Rules](#alerting-rules)

---

## The 4 Golden Signals

Quick reference for the most important queries:

| Signal         | Query                                                                                                  | Description         |
| -------------- | ------------------------------------------------------------------------------------------------------ | ------------------- |
| **Latency**    | `histogram_quantile(0.95, rate(doit_api_request_duration_seconds_bucket[5m]))`                         | P95 response time   |
| **Traffic**    | `sum(rate(doit_api_requests_total[5m]))`                                                               | Requests per second |
| **Errors**     | `sum(rate(doit_api_requests_total{status=~"5.."}[5m])) / sum(rate(doit_api_requests_total[5m])) * 100` | Error percentage    |
| **Saturation** | `doit_db_pool_saturation_ratio`                                                                        | DB pool fullness    |

---

## HTTP Metrics

### Request Rate (Traffic)

```promql
# Total requests per second (all endpoints)
sum(rate(doit_api_requests_total[5m]))
```

**What it tells you:** Overall API throughput - how many requests per second your API handles.

```promql
# Requests per second by endpoint
sum(rate(doit_api_requests_total[5m])) by (path)
```

**What it tells you:** Which endpoints are most popular. Helps identify hot paths.

```promql
# Requests per second by method
sum(rate(doit_api_requests_total[5m])) by (method)
```

**What it tells you:** Read vs write ratio (GET vs POST/PUT/DELETE).

```promql
# Requests per second by status code
sum(rate(doit_api_requests_total[5m])) by (status)
```

**What it tells you:** Distribution of response codes. Healthy APIs have mostly 2xx.

```promql
# Top 5 busiest endpoints
topk(5, sum(rate(doit_api_requests_total[5m])) by (path))
```

**What it tells you:** Your most heavily used endpoints - prioritize optimization here.

---

### Error Rate (Errors)

```promql
# Server error rate percentage (5xx errors)
sum(rate(doit_api_requests_total{status=~"5.."}[5m]))
/
sum(rate(doit_api_requests_total[5m])) * 100
```

**What it tells you:** Percentage of requests failing due to server errors. Target: < 1%.

```promql
# Client error rate percentage (4xx errors)
sum(rate(doit_api_requests_total{status=~"4.."}[5m]))
/
sum(rate(doit_api_requests_total[5m])) * 100
```

**What it tells you:** Percentage of bad requests (validation errors, not found, unauthorized).

```promql
# Error rate by endpoint
sum(rate(doit_api_requests_total{status=~"5.."}[5m])) by (path)
/
sum(rate(doit_api_requests_total[5m])) by (path) * 100
```

**What it tells you:** Which endpoints are failing the most. Helps pinpoint problematic code.

```promql
# Absolute error count per second
sum(rate(doit_api_requests_total{status=~"5.."}[5m]))
```

**What it tells you:** Raw number of errors per second. Useful for low-traffic APIs where percentages can be misleading.

```promql
# Success rate percentage (availability)
sum(rate(doit_api_requests_total{status=~"2.."}[5m]))
/
sum(rate(doit_api_requests_total[5m])) * 100
```

**What it tells you:** Your API's availability. Target: > 99.9%.

---

### Latency (Duration)

```promql
# P50 (median) latency - 50% of requests are faster than this
histogram_quantile(0.50, rate(doit_api_request_duration_seconds_bucket[5m]))
```

**What it tells you:** Typical user experience. Half of your users see this latency or better.

```promql
# P90 latency - 90% of requests are faster than this
histogram_quantile(0.90, rate(doit_api_request_duration_seconds_bucket[5m]))
```

**What it tells you:** Most users' experience. Good for SLO targets.

```promql
# P95 latency - 95% of requests are faster than this
histogram_quantile(0.95, rate(doit_api_request_duration_seconds_bucket[5m]))
```

**What it tells you:** Common SLO target. "95% of requests complete within X seconds."

```promql
# P99 latency - 99% of requests are faster than this
histogram_quantile(0.99, rate(doit_api_request_duration_seconds_bucket[5m]))
```

**What it tells you:** Worst-case experience for most users. High P99 indicates tail latency issues.

```promql
# P95 latency by endpoint
histogram_quantile(0.95,
  sum(rate(doit_api_request_duration_seconds_bucket[5m])) by (path, le)
)
```

**What it tells you:** Which endpoints are slowest. Prioritize optimization here.

```promql
# Average latency (less useful than percentiles, but simple)
rate(doit_api_request_duration_seconds_sum[5m])
/
rate(doit_api_request_duration_seconds_count[5m])
```

**What it tells you:** Mean response time. Can be skewed by outliers - prefer percentiles.

```promql
# Latency histogram - requests per bucket
sum(rate(doit_api_request_duration_seconds_bucket[5m])) by (le)
```

**What it tells you:** Distribution of latencies across your defined buckets.

---

### In-Flight Requests (Saturation)

```promql
# Current in-flight requests
doit_api_requests_in_flight
```

**What it tells you:** How many requests are being processed right now. High values indicate load.

```promql
# Average in-flight over time
avg_over_time(doit_api_requests_in_flight[5m])
```

**What it tells you:** Typical concurrent load on your server.

```promql
# Max in-flight over time
max_over_time(doit_api_requests_in_flight[1h])
```

**What it tells you:** Peak concurrent requests in the last hour. Helps with capacity planning.

---

## Database Metrics

### Connection Pool

```promql
# Current pool saturation (0-1)
doit_db_pool_saturation_ratio
```

**What it tells you:** How full your connection pool is. > 0.8 is concerning, 1.0 means exhausted.

```promql
# Total connections in pool
doit_db_connections_total
```

**What it tells you:** Size of your connection pool.

```promql
# Connections currently in use
doit_db_connections_in_use
```

**What it tells you:** How many connections are actively running queries.

```promql
# Idle connections
doit_db_connections_idle
```

**What it tells you:** Connections waiting for work. Low idle + high saturation = need more connections.

```promql
# Connection utilization percentage
(doit_db_connections_in_use / doit_db_connections_total) * 100
```

**What it tells you:** Percentage of pool being used. Same as saturation but as percentage.

```promql
# Rate of connection waits (had to wait for a connection)
rate(doit_db_pool_wait_total[5m])
```

**What it tells you:** How often requests wait for a database connection. Should be 0 ideally.

---

### Query Performance

```promql
# P95 query duration (all queries)
histogram_quantile(0.95, rate(doit_db_query_duration_seconds_bucket[5m]))
```

**What it tells you:** 95% of database queries complete within this time.

```promql
# P95 query duration by operation type
histogram_quantile(0.95,
  sum(rate(doit_db_query_duration_seconds_bucket[5m])) by (operation, le)
)
```

**What it tells you:** Performance breakdown by operation (select, insert, update, delete).

```promql
# P95 query duration by table
histogram_quantile(0.95,
  sum(rate(doit_db_query_duration_seconds_bucket[5m])) by (table, le)
)
```

**What it tells you:** Which tables have the slowest queries. May need indexing.

```promql
# Average query duration by operation and table
sum(rate(doit_db_query_duration_seconds_sum[5m])) by (operation, table)
/
sum(rate(doit_db_query_duration_seconds_count[5m])) by (operation, table)
```

**What it tells you:** Mean query time per operation/table combination.

```promql
# Queries per second by operation
sum(rate(doit_db_query_duration_seconds_count[5m])) by (operation)
```

**What it tells you:** Query throughput. High SELECT rate is normal; high UPDATE rate may indicate issues.

---

### Query Errors

```promql
# Database error rate
sum(rate(doit_db_query_errors_total[5m]))
```

**What it tells you:** How many database errors per second. Should be close to 0.

```promql
# Database errors by operation and table
sum(rate(doit_db_query_errors_total[5m])) by (operation, table)
```

**What it tells you:** Which operations/tables are failing. Helps identify problematic queries.

```promql
# Database error percentage
sum(rate(doit_db_query_errors_total[5m]))
/
sum(rate(doit_db_query_duration_seconds_count[5m])) * 100
```

**What it tells you:** What percentage of database queries fail.

---

## Cache Metrics

### Hit Rate

```promql
# Cache hit rate percentage
sum(rate(doit_cache_hits_total[5m]))
/
(sum(rate(doit_cache_hits_total[5m])) + sum(rate(doit_cache_misses_total[5m]))) * 100
```

**What it tells you:** Percentage of cache lookups that found data. Target: > 80% for effective caching.

```promql
# Cache hit rate by operation
sum(rate(doit_cache_hits_total[5m])) by (operation)
/
(sum(rate(doit_cache_hits_total[5m])) by (operation) + sum(rate(doit_cache_misses_total[5m])) by (operation)) * 100
```

**What it tells you:** Which operations benefit most from caching.

```promql
# Cache hits per second
sum(rate(doit_cache_hits_total[5m]))
```

**What it tells you:** How many cache hits per second. Higher is better.

```promql
# Cache misses per second
sum(rate(doit_cache_misses_total[5m]))
```

**What it tells you:** How many cache misses per second. These result in database queries.

```promql
# Cache miss rate percentage
sum(rate(doit_cache_misses_total[5m]))
/
(sum(rate(doit_cache_hits_total[5m])) + sum(rate(doit_cache_misses_total[5m]))) * 100
```

**What it tells you:** Percentage of lookups that miss cache. High miss rate = cache not effective.

---

## Business Metrics

### Todo Operations

```promql
# Todo operations per second (all types)
sum(rate(doit_todo_operations_total[5m]))
```

**What it tells you:** Overall business activity - how many todo operations per second.

```promql
# Todo operations by type
sum(rate(doit_todo_operations_total[5m])) by (operation)
```

**What it tells you:** Breakdown by operation type (create, read, update, delete, complete).

```promql
# Todo create rate
rate(doit_todo_operations_total{operation="create"}[5m])
```

**What it tells you:** How many new todos are being created per second.

```promql
# Todo completion rate
rate(doit_todo_operations_total{operation="complete"}[5m])
```

**What it tells you:** How many todos are being completed per second.

```promql
# Read to write ratio
sum(rate(doit_todo_operations_total{operation="read"}[5m]))
/
sum(rate(doit_todo_operations_total{operation=~"create|update|delete"}[5m]))
```

**What it tells you:** How read-heavy your workload is. High ratio = good for caching.

---

## Runtime Metrics

### Goroutines

```promql
# Current goroutine count
doit_runtime_goroutines
```

**What it tells you:** Number of goroutines running. Should be stable under consistent load.

```promql
# Goroutine growth rate (leak detection)
deriv(doit_runtime_goroutines[1h])
```

**What it tells you:** Rate of goroutine growth. Positive trend over time = possible leak.

```promql
# Goroutine count over time
avg_over_time(doit_runtime_goroutines[1h])
```

**What it tells you:** Average goroutine count. Compare with baseline to detect anomalies.

```promql
# Max goroutines in last hour
max_over_time(doit_runtime_goroutines[1h])
```

**What it tells you:** Peak goroutine count. Helps identify burst patterns.

---

## Go Runtime (Automatic)

These metrics are automatically exposed by the Prometheus Go client library.

### Memory

```promql
# Heap memory in use (bytes)
go_memstats_heap_inuse_bytes
```

**What it tells you:** Current heap memory usage. Monitor for memory leaks.

```promql
# Heap memory in use (MB) - more readable
go_memstats_heap_inuse_bytes / 1024 / 1024
```

**What it tells you:** Same as above but in megabytes.

```promql
# Total allocated memory (includes freed)
go_memstats_alloc_bytes
```

**What it tells you:** Total bytes allocated (even if later freed).

```promql
# Stack memory in use
go_memstats_stack_inuse_bytes
```

**What it tells you:** Memory used by goroutine stacks.

```promql
# Memory growth rate
deriv(go_memstats_heap_inuse_bytes[1h])
```

**What it tells you:** Rate of memory growth. Sustained positive = memory leak.

---

### Garbage Collection

```promql
# GC pause duration (P95)
histogram_quantile(0.95, rate(go_gc_duration_seconds_bucket[5m]))
```

**What it tells you:** How long GC pauses last. High values impact latency.

```promql
# GC runs per second
rate(go_gc_duration_seconds_count[5m])
```

**What it tells you:** How often GC runs. Very frequent = high allocation rate.

```promql
# Time spent in GC (percentage)
sum(rate(go_gc_duration_seconds_sum[5m])) * 100
```

**What it tells you:** Percentage of time spent in GC. > 5% is concerning.

---

### Goroutines (Built-in)

```promql
# Go runtime goroutine count (alternative to custom metric)
go_goroutines
```

**What it tells you:** Same as `doit_runtime_goroutines` but from Go runtime directly.

---

## Alerting Rules

Example Prometheus alerting rules for your metrics:

```yaml
groups:
  - name: doit-api-alerts
    rules:
      # High Error Rate
      - alert: HighErrorRate
        expr: |
          sum(rate(doit_api_requests_total{status=~"5.."}[5m])) 
          / sum(rate(doit_api_requests_total[5m])) * 100 > 5
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"
          description: 'Error rate is {{ $value | printf "%.2f" }}% (threshold: 5%)'

      # High Latency
      - alert: HighLatency
        expr: |
          histogram_quantile(0.95, rate(doit_api_request_duration_seconds_bucket[5m])) > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High P95 latency detected"
          description: 'P95 latency is {{ $value | printf "%.2f" }}s (threshold: 1s)'

      # Database Pool Saturated
      - alert: DatabasePoolSaturated
        expr: doit_db_pool_saturation_ratio > 0.8
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Database connection pool is saturated"
          description: 'Pool saturation is {{ $value | printf "%.0f" }}% (threshold: 80%)'

      # Database Pool Exhausted
      - alert: DatabasePoolExhausted
        expr: doit_db_pool_saturation_ratio >= 1
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Database connection pool is exhausted"
          description: "All database connections are in use"

      # Low Cache Hit Rate
      - alert: LowCacheHitRate
        expr: |
          sum(rate(doit_cache_hits_total[5m])) 
          / (sum(rate(doit_cache_hits_total[5m])) + sum(rate(doit_cache_misses_total[5m]))) * 100 < 50
        for: 15m
        labels:
          severity: warning
        annotations:
          summary: "Cache hit rate is low"
          description: 'Cache hit rate is {{ $value | printf "%.0f" }}% (threshold: 50%)'

      # Goroutine Leak
      - alert: GoroutineLeak
        expr: deriv(doit_runtime_goroutines[1h]) > 10
        for: 30m
        labels:
          severity: warning
        annotations:
          summary: "Possible goroutine leak detected"
          description: 'Goroutines growing at {{ $value | printf "%.0f" }}/hour'

      # High Memory Usage
      - alert: HighMemoryUsage
        expr: go_memstats_heap_inuse_bytes / 1024 / 1024 > 500
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage"
          description: 'Heap memory usage is {{ $value | printf "%.0f" }}MB (threshold: 500MB)'

      # Service Down
      - alert: ServiceDown
        expr: up{job="doit-api"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "DoIt API is down"
          description: "Prometheus cannot scrape the DoIt API metrics endpoint"

      # High In-Flight Requests
      - alert: HighInFlightRequests
        expr: doit_api_requests_in_flight > 100
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High number of in-flight requests"
          description: "{{ $value }} requests currently being processed"

      # Database Errors
      - alert: DatabaseErrors
        expr: sum(rate(doit_db_query_errors_total[5m])) > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Database errors detected"
          description: '{{ $value | printf "%.2f" }} database errors per second'
```

---

## Grafana Dashboard Panels

Recommended panels for a Grafana dashboard:

### Row 1: Overview (Single Stats)

| Panel              | Query                                    | Thresholds                                 |
| ------------------ | ---------------------------------------- | ------------------------------------------ |
| Request Rate       | `sum(rate(doit_api_requests_total[5m]))` | -                                          |
| Error Rate         | Error rate query                         | Green < 1%, Yellow < 5%, Red > 5%          |
| P95 Latency        | P95 latency query                        | Green < 100ms, Yellow < 500ms, Red > 500ms |
| DB Pool Saturation | `doit_db_pool_saturation_ratio * 100`    | Green < 50%, Yellow < 80%, Red > 80%       |

### Row 2: Traffic

| Panel                          | Query                                                        |
| ------------------------------ | ------------------------------------------------------------ |
| Requests/sec (Graph)           | `sum(rate(doit_api_requests_total[5m]))`                     |
| Requests by Endpoint (Table)   | `topk(10, sum(rate(doit_api_requests_total[5m])) by (path))` |
| Status Code Distribution (Pie) | `sum(rate(doit_api_requests_total[5m])) by (status)`         |

### Row 3: Latency

| Panel                         | Query                                                                   |
| ----------------------------- | ----------------------------------------------------------------------- |
| Latency Percentiles (Graph)   | P50, P90, P95, P99 on same graph                                        |
| Latency by Endpoint (Heatmap) | `sum(rate(doit_api_request_duration_seconds_bucket[5m])) by (path, le)` |

### Row 4: Database

| Panel                      | Query                                                      |
| -------------------------- | ---------------------------------------------------------- |
| Connection Pool (Gauge)    | `doit_db_pool_saturation_ratio * 100`                      |
| Query Duration P95 (Graph) | DB P95 query                                               |
| Query Errors (Graph)       | `sum(rate(doit_db_query_errors_total[5m])) by (operation)` |

### Row 5: Cache & Business

| Panel                   | Query                                                      |
| ----------------------- | ---------------------------------------------------------- |
| Cache Hit Rate (Gauge)  | Cache hit rate query                                       |
| Todo Operations (Graph) | `sum(rate(doit_todo_operations_total[5m])) by (operation)` |

### Row 6: Runtime

| Panel                     | Query                                        |
| ------------------------- | -------------------------------------------- |
| Goroutines (Graph)        | `doit_runtime_goroutines`                    |
| Memory Usage (Graph)      | `go_memstats_heap_inuse_bytes / 1024 / 1024` |
| GC Pause Duration (Graph) | GC P95 query                                 |

---

## Quick Reference Card

```
┌─────────────────────────────────────────────────────────────────┐
│                    PROMQL QUICK REFERENCE                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  FUNCTIONS                                                       │
│  ─────────                                                       │
│  rate(metric[5m])         → Per-second rate over 5 minutes      │
│  sum(metric) by (label)   → Aggregate by label                  │
│  histogram_quantile(0.95, ...) → Calculate percentile           │
│  topk(5, metric)          → Top 5 values                        │
│  deriv(metric[1h])        → Rate of change (for gauges)         │
│  avg_over_time(metric[1h]) → Average over time window           │
│  max_over_time(metric[1h]) → Maximum over time window           │
│                                                                  │
│  SELECTORS                                                       │
│  ─────────                                                       │
│  {label="value"}          → Exact match                         │
│  {label=~"regex"}         → Regex match                         │
│  {label!="value"}         → Not equal                           │
│  {label=~"5.."}           → Match 5xx status codes              │
│                                                                  │
│  TIME RANGES                                                     │
│  ───────────                                                     │
│  [5m]  → Last 5 minutes                                         │
│  [1h]  → Last 1 hour                                            │
│  [1d]  → Last 1 day                                             │
│                                                                  │
│  COMMON PATTERNS                                                 │
│  ───────────────                                                 │
│  Error %:  sum(rate(errors[5m])) / sum(rate(total[5m])) * 100  │
│  P95:      histogram_quantile(0.95, rate(bucket[5m]))          │
│  Growth:   deriv(gauge[1h])                                     │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Metric Naming Convention

All DoIt metrics follow this pattern:

```
doit_<subsystem>_<name>_<unit>
```

| Subsystem | Metrics              |
| --------- | -------------------- |
| `api`     | HTTP request metrics |
| `db`      | Database metrics     |
| `cache`   | Cache metrics        |
| `todo`    | Business metrics     |
| `runtime` | Go runtime metrics   |

| Suffix     | Meaning                     |
| ---------- | --------------------------- |
| `_total`   | Counter (always increasing) |
| `_seconds` | Duration in seconds         |
| `_bytes`   | Size in bytes               |
| `_ratio`   | Value between 0 and 1       |

---

## Further Reading

- [Prometheus Query Basics](https://prometheus.io/docs/prometheus/latest/querying/basics/)
- [PromQL Functions](https://prometheus.io/docs/prometheus/latest/querying/functions/)
- [Histogram and Summaries](https://prometheus.io/docs/practices/histograms/)
- [Alerting Rules](https://prometheus.io/docs/prometheus/latest/configuration/alerting_rules/)
