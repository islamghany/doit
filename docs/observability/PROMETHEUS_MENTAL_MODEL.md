# 📊 Prometheus - Building Your Mental Model

## What is Prometheus?

Prometheus is a **time-series database** designed specifically for **metrics**. Think of it as a specialized database that excels at answering questions like:

- "How many requests did we get in the last hour?"
- "What's the 95th percentile response time?"
- "Is our error rate increasing?"

---

## The Core Mental Model: Pull vs Push

```
┌─────────────────────────────────────────────────────────────────────┐
│                    PUSH MODEL (Traditional)                         │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   App ──push──▶ Metrics Server                                      │
│                                                                     │
│   Problems:                                                         │
│   • App needs to know where to send metrics                         │
│   • If metrics server is down, data is lost                         │
│   • Hard to scale (thundering herd)                                 │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────┐
│                    PULL MODEL (Prometheus)                          │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   Prometheus ──scrape──▶ App /metrics endpoint                      │
│                                                                     │
│   Benefits:                                                         │
│   • App just exposes metrics, doesn't care who reads them           │
│   • Prometheus controls the pace (no thundering herd)               │
│   • Easy to add/remove targets                                      │
│   • Can detect if target is down (scrape fails)                     │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

**Analogy**: 
- **Push** = You calling your friends every hour to tell them you're okay
- **Pull** = Your friends checking your social media status when they want to know

---

## The Four Golden Signals

Google's Site Reliability Engineering book defines four key metrics every service should track:

```
┌─────────────────────────────────────────────────────────────────────┐
│                    THE FOUR GOLDEN SIGNALS                          │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  1. LATENCY ⏱️                                                      │
│     └── How long requests take                                      │
│     └── "Response time for successful requests"                     │
│     └── Track separately: success vs error latency                  │
│                                                                     │
│  2. TRAFFIC 📈                                                      │
│     └── How much demand is hitting your system                      │
│     └── "Requests per second"                                       │
│     └── For APIs: HTTP requests/sec                                 │
│                                                                     │
│  3. ERRORS ❌                                                       │
│     └── Rate of failed requests                                     │
│     └── "Percentage of requests that fail"                          │
│     └── HTTP 5xx, explicit errors, implicit (wrong content)         │
│                                                                     │
│  4. SATURATION 🔋                                                   │
│     └── How "full" your service is                                  │
│     └── "CPU at 90%", "Memory at 85%", "Connection pool full"       │
│     └── Predicts future problems                                    │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## The Four Metric Types

Prometheus has exactly four metric types. Understanding these is crucial:

### 1️⃣ Counter - Only Goes Up

```
┌─────────────────────────────────────────────────────────────────────┐
│                         COUNTER                                     │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   Value:  0 → 1 → 2 → 3 → 4 → 5 → ... (never decreases)            │
│                                                                     │
│   Use for:                                                          │
│   • Total requests received                                         │
│   • Total errors occurred                                           │
│   • Total bytes processed                                           │
│                                                                     │
│   Example:                                                          │
│   http_requests_total{method="GET", status="200"} 15234             │
│                                                                     │
│   Query (rate of change):                                           │
│   rate(http_requests_total[5m])  →  "requests per second"           │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

**Analogy**: Car odometer - only goes up, shows total distance traveled.

### 2️⃣ Gauge - Goes Up and Down

```
┌─────────────────────────────────────────────────────────────────────┐
│                          GAUGE                                      │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   Value:  5 → 8 → 3 → 12 → 7 → ... (can increase or decrease)      │
│                                                                     │
│   Use for:                                                          │
│   • Current temperature                                             │
│   • Current memory usage                                            │
│   • Active connections right now                                    │
│   • Queue depth                                                     │
│                                                                     │
│   Example:                                                          │
│   db_connections_active 23                                          │
│   memory_usage_bytes 1073741824                                     │
│                                                                     │
│   Query (current value):                                            │
│   db_connections_active  →  "23 connections right now"              │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

**Analogy**: Speedometer - shows current speed, goes up and down.

### 3️⃣ Histogram - Distribution of Values

```
┌─────────────────────────────────────────────────────────────────────┐
│                        HISTOGRAM                                    │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   Buckets: How many observations fell into each range               │
│                                                                     │
│   Request Duration Distribution:                                    │
│   ┌─────────────────────────────────────────────────────────────┐  │
│   │  ≤10ms   ████████████████████████████████  150 requests     │  │
│   │  ≤25ms   ████████████████████  100 requests                 │  │
│   │  ≤50ms   ████████████  60 requests                          │  │
│   │  ≤100ms  ████████  40 requests                              │  │
│   │  ≤250ms  ████  20 requests                                  │  │
│   │  ≤500ms  ██  10 requests                                    │  │
│   │  ≤1s     █  5 requests                                      │  │
│   │  +Inf    █  3 requests (> 1s)                               │  │
│   └─────────────────────────────────────────────────────────────┘  │
│                                                                     │
│   Creates three metrics:                                            │
│   • _bucket: cumulative count per bucket                            │
│   • _sum: total sum of all values                                   │
│   • _count: total number of observations                            │
│                                                                     │
│   Query (percentiles):                                              │
│   histogram_quantile(0.95, rate(http_request_duration_bucket[5m]))  │
│   →  "95% of requests complete within X seconds"                    │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

**Analogy**: Sorting mail into bins by weight - you know how many letters are in each weight range.

### 4️⃣ Summary - Pre-calculated Percentiles

```
┌─────────────────────────────────────────────────────────────────────┐
│                         SUMMARY                                     │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   Pre-calculates percentiles on the client side                     │
│                                                                     │
│   Example output:                                                   │
│   http_request_duration{quantile="0.5"}  0.023    (median)         │
│   http_request_duration{quantile="0.9"}  0.089    (90th percentile)│
│   http_request_duration{quantile="0.99"} 0.234    (99th percentile)│
│   http_request_duration_sum   1234.5                                │
│   http_request_duration_count 50000                                 │
│                                                                     │
│   Histogram vs Summary:                                             │
│   ┌─────────────────────┬─────────────────────────────────────────┐│
│   │     Histogram       │           Summary                       ││
│   ├─────────────────────┼─────────────────────────────────────────┤│
│   │ Server-side calc    │ Client-side calc                        ││
│   │ Can aggregate       │ Cannot aggregate across instances       ││
│   │ Fixed buckets       │ Fixed quantiles                         ││
│   │ Preferred for APIs  │ Use when you know exact quantiles needed││
│   └─────────────────────┴─────────────────────────────────────────┘│
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Labels - The Power of Dimensions

Labels add dimensions to your metrics, enabling powerful queries:

```
┌─────────────────────────────────────────────────────────────────────┐
│                          LABELS                                     │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   Without labels:                                                   │
│   http_requests_total 15234                                         │
│   → "We had 15234 requests" (not very useful)                       │
│                                                                     │
│   With labels:                                                      │
│   http_requests_total{method="GET", path="/api/users", status="200"}│
│   http_requests_total{method="POST", path="/api/users", status="201"}│
│   http_requests_total{method="GET", path="/api/users", status="500"}│
│                                                                     │
│   Now you can query:                                                │
│   • All requests: sum(http_requests_total)                          │
│   • Just errors: sum(http_requests_total{status=~"5.."})            │
│   • By endpoint: sum by (path) (http_requests_total)                │
│   • Error rate: sum(rate(http_requests_total{status=~"5.."}[5m]))   │
│                          /                                          │
│                 sum(rate(http_requests_total[5m]))                  │
│                                                                     │
│   ⚠️ WARNING: High cardinality labels (user_id, request_id) can    │
│              explode your metrics storage!                          │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## PromQL - Query Language Basics

```
┌─────────────────────────────────────────────────────────────────────┐
│                    PROMQL ESSENTIALS                                │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  INSTANT VECTOR (current value):                                    │
│  http_requests_total                                                │
│  → Returns current value of all matching series                     │
│                                                                     │
│  RANGE VECTOR (values over time):                                   │
│  http_requests_total[5m]                                            │
│  → Returns all values from last 5 minutes                           │
│                                                                     │
│  RATE (per-second rate of change):                                  │
│  rate(http_requests_total[5m])                                      │
│  → "X requests per second over last 5 minutes"                      │
│                                                                     │
│  INCREASE (total increase over time):                               │
│  increase(http_requests_total[1h])                                  │
│  → "Total new requests in last hour"                                │
│                                                                     │
│  AGGREGATION:                                                       │
│  sum(rate(http_requests_total[5m]))                                 │
│  sum by (method) (rate(http_requests_total[5m]))                    │
│  avg, min, max, count, topk(3, ...)                                 │
│                                                                     │
│  HISTOGRAM PERCENTILES:                                             │
│  histogram_quantile(0.95, rate(http_duration_bucket[5m]))           │
│  → "95th percentile latency"                                        │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                   PROMETHEUS ARCHITECTURE                           │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   ┌─────────────┐     ┌─────────────┐     ┌─────────────┐          │
│   │   Your App  │     │   Your App  │     │   Your App  │          │
│   │  /metrics   │     │  /metrics   │     │  /metrics   │          │
│   └──────┬──────┘     └──────┬──────┘     └──────┬──────┘          │
│          │                   │                   │                  │
│          └───────────────────┼───────────────────┘                  │
│                              │ scrape every 15s                     │
│                              ▼                                      │
│                    ┌─────────────────┐                              │
│                    │   PROMETHEUS    │                              │
│                    │                 │                              │
│                    │  ┌───────────┐  │                              │
│                    │  │  TSDB     │  │  Time Series Database        │
│                    │  │ (storage) │  │                              │
│                    │  └───────────┘  │                              │
│                    │                 │                              │
│                    │  ┌───────────┐  │                              │
│                    │  │  PromQL   │  │  Query Engine                │
│                    │  │  Engine   │  │                              │
│                    │  └───────────┘  │                              │
│                    │                 │                              │
│                    │  ┌───────────┐  │                              │
│                    │  │  Alert    │  │  Alerting Rules              │
│                    │  │  Manager  │  │                              │
│                    │  └───────────┘  │                              │
│                    └────────┬────────┘                              │
│                             │                                       │
│              ┌──────────────┼──────────────┐                        │
│              ▼              ▼              ▼                        │
│        ┌──────────┐  ┌──────────┐  ┌──────────────┐                │
│        │ Grafana  │  │  Alerts  │  │  API Clients │                │
│        │   UI     │  │ (Slack)  │  │              │                │
│        └──────────┘  └──────────┘  └──────────────┘                │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Best Practices

### Naming Conventions

```
┌─────────────────────────────────────────────────────────────────────┐
│                   METRIC NAMING CONVENTIONS                         │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  Format: <namespace>_<name>_<unit>                                  │
│                                                                     │
│  Good examples:                                                     │
│  • http_requests_total                                              │
│  • http_request_duration_seconds                                    │
│  • process_cpu_seconds_total                                        │
│  • db_connections_active                                            │
│                                                                     │
│  Bad examples:                                                      │
│  • requests (too vague)                                             │
│  • http_request_duration_ms (use base units: seconds)               │
│  • HTTPRequestsTotal (use snake_case)                               │
│                                                                     │
│  Units (always use base units):                                     │
│  • seconds (not milliseconds)                                       │
│  • bytes (not kilobytes)                                            │
│  • meters (not kilometers)                                          │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### What to Measure

```
┌─────────────────────────────────────────────────────────────────────┐
│                    WHAT TO INSTRUMENT                               │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  FOR EVERY SERVICE:                                                 │
│  ├── Request rate (Counter)                                         │
│  ├── Request duration (Histogram)                                   │
│  ├── Error rate (Counter with status label)                         │
│  └── In-flight requests (Gauge)                                     │
│                                                                     │
│  FOR DATABASES:                                                     │
│  ├── Connection pool size (Gauge)                                   │
│  ├── Active connections (Gauge)                                     │
│  ├── Query duration (Histogram)                                     │
│  └── Query errors (Counter)                                         │
│                                                                     │
│  FOR CACHES:                                                        │
│  ├── Hit/miss ratio (Counter)                                       │
│  ├── Operation duration (Histogram)                                 │
│  └── Keys count (Gauge)                                             │
│                                                                     │
│  FOR RUNTIME:                                                       │
│  ├── Goroutines count (Gauge)                                       │
│  ├── Memory usage (Gauge)                                           │
│  ├── GC pause duration (Summary)                                    │
│  └── Open file descriptors (Gauge)                                  │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Mental Model Summary

```
┌─────────────────────────────────────────────────────────────────────┐
│                  PROMETHEUS MENTAL MODEL                            │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  PROMETHEUS = Time-series database for metrics                      │
│               (specialized for "how much" questions)                │
│                                                                     │
│  PULL MODEL = Prometheus fetches metrics from your app              │
│               (you expose /metrics, Prometheus scrapes)             │
│                                                                     │
│  FOUR TYPES:                                                        │
│  • Counter = Total count (only goes up)                             │
│  • Gauge = Current value (goes up and down)                         │
│  • Histogram = Distribution in buckets                              │
│  • Summary = Pre-calculated percentiles                             │
│                                                                     │
│  LABELS = Dimensions that enable powerful queries                   │
│           (but beware of high cardinality!)                         │
│                                                                     │
│  PROMQL = Query language for aggregating and analyzing              │
│           (rate, sum, histogram_quantile, etc.)                     │
│                                                                     │
│  ┌───────────────────────────────────────────────────────────────┐ │
│  │  THE POWER:                                                    │ │
│  │                                                                │ │
│  │  "Is something wrong?" ──▶ Metrics tell you YES/NO             │ │
│  │  "How bad is it?" ──▶ Metrics quantify the impact              │ │
│  │  "Is it getting worse?" ──▶ Metrics show trends                │ │
│  │                                                                │ │
│  └───────────────────────────────────────────────────────────────┘ │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Related Documentation

- [Grafana Mental Model](./GRAFANA_MENTAL_MODEL.md) - Visualization and dashboards
- [Distributed Tracing Mental Model](./DISTRIBUTED_TRACING_MENTAL_MODEL.md) - Request flow tracking
- [Observability Overview](./OBSERVABILITY_OVERVIEW.md) - The three pillars combined

