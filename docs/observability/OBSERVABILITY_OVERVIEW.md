# 🔭 Observability Overview - The Three Pillars

## What is Observability?

Observability is the ability to understand the **internal state** of a system by examining its **external outputs**. In other words: Can you answer "Why is this broken?" just by looking at the data your system produces?

---

## The Three Pillars

```
┌─────────────────────────────────────────────────────────────────────┐
│                    THREE PILLARS OF OBSERVABILITY                   │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│        LOGS                 METRICS                TRACES           │
│        ────                 ───────                ──────           │
│         📝                    📊                     🔍             │
│                                                                     │
│   "What happened"       "How much"            "The journey"         │
│                                                                     │
│   ┌─────────────┐    ┌─────────────┐      ┌─────────────┐          │
│   │ Detailed    │    │ Aggregated  │      │ Request     │          │
│   │ events with │    │ numbers     │      │ flow across │          │
│   │ context     │    │ over time   │      │ services    │          │
│   └─────────────┘    └─────────────┘      └─────────────┘          │
│                                                                     │
│   Examples:           Examples:            Examples:                │
│   • Error stack      • Request rate       • Waterfall view         │
│   • Debug info       • Error percentage   • Span timeline          │
│   • User actions     • Latency P99        • Service map            │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## How They Work Together

```
┌─────────────────────────────────────────────────────────────────────┐
│                    INCIDENT INVESTIGATION FLOW                      │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   STEP 1: ALERT (from Metrics)                                      │
│   ┌─────────────────────────────────────────────────────────────┐  │
│   │  🚨 "Error rate exceeded 5% for 5 minutes"                   │  │
│   │     Source: Prometheus + Alertmanager                        │  │
│   └─────────────────────────────────────────────────────────────┘  │
│                           │                                         │
│                           ▼                                         │
│   STEP 2: INVESTIGATE (with Traces)                                 │
│   ┌─────────────────────────────────────────────────────────────┐  │
│   │  🔍 "Errors are happening in database connection step"       │  │
│   │     Source: Jaeger trace analysis                            │  │
│   │     Finding: 450ms wait time for DB connections              │  │
│   └─────────────────────────────────────────────────────────────┘  │
│                           │                                         │
│                           ▼                                         │
│   STEP 3: ROOT CAUSE (from Logs)                                    │
│   ┌─────────────────────────────────────────────────────────────┐  │
│   │  📝 "Connection pool exhausted: max_connections=25"          │  │
│   │     Source: Application logs                                 │  │
│   │     Root cause: Connection leak in new feature               │  │
│   └─────────────────────────────────────────────────────────────┘  │
│                                                                     │
│   RESOLUTION: Fix connection leak, increase pool size temporarily   │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Comparison Table

| Aspect           | Logs                    | Metrics             | Traces               |
| ---------------- | ----------------------- | ------------------- | -------------------- |
| **Data Type**    | Text/JSON events        | Numeric time-series | Structured spans     |
| **Cardinality**  | High (unique per event) | Low (aggregated)    | Medium (per request) |
| **Storage Cost** | High                    | Low                 | Medium               |
| **Query Speed**  | Slow (search)           | Fast (indexed)      | Medium               |
| **Best For**     | Debugging details       | Alerting, trends    | Request flow         |
| **Sampling**     | Optional                | No                  | Often required       |
| **Tool Example** | Loki, ELK               | Prometheus          | Jaeger, Tempo        |

---

## The DoIt API Observability Stack

```
┌─────────────────────────────────────────────────────────────────────┐
│                   DoIt API OBSERVABILITY STACK                      │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│                        ┌──────────────┐                             │
│                        │   GRAFANA    │  Visualization              │
│                        │   :3000      │  & Dashboards               │
│                        └──────┬───────┘                             │
│                               │                                     │
│              ┌────────────────┼────────────────┐                    │
│              │                │                │                    │
│              ▼                ▼                ▼                    │
│       ┌──────────┐     ┌──────────┐     ┌──────────┐               │
│       │PROMETHEUS│     │  JAEGER  │     │   LOKI   │               │
│       │  :9090   │     │  :16686  │     │  :3100   │               │
│       │ Metrics  │     │  Traces  │     │   Logs   │               │
│       └────┬─────┘     └────┬─────┘     └────┬─────┘               │
│            │                │                │                      │
│            │    ┌───────────┴───────────┐    │                      │
│            │    │                       │    │                      │
│            ▼    ▼                       ▼    ▼                      │
│       ┌─────────────────────────────────────────┐                  │
│       │              DoIt API                    │                  │
│       │                                          │                  │
│       │  /metrics ──▶ Prometheus scrapes         │                  │
│       │  OTel SDK ──▶ Jaeger receives            │                  │
│       │  stdout   ──▶ Loki collects              │                  │
│       │                                          │                  │
│       └─────────────────────────────────────────┘                  │
│                                                                     │
│   FUTURE (AWS):                                                     │
│   • Prometheus ──▶ CloudWatch Metrics                               │
│   • Jaeger ──▶ AWS X-Ray                                            │
│   • Loki ──▶ CloudWatch Logs                                        │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Key Principles

### 1. Correlation is Key

```
┌─────────────────────────────────────────────────────────────────────┐
│                        CORRELATION                                  │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   All three pillars should share a common identifier:               │
│                                                                     │
│   REQUEST ID / TRACE ID: "abc-123-def-456"                          │
│                                                                     │
│   ┌─────────────┐                                                   │
│   │    LOG      │  {"trace_id": "abc-123", "msg": "Query slow"}    │
│   └─────────────┘                                                   │
│          │                                                          │
│          │  Same ID                                                 │
│          ▼                                                          │
│   ┌─────────────┐                                                   │
│   │   TRACE     │  Trace ID: abc-123, Span: pg.query, 450ms        │
│   └─────────────┘                                                   │
│          │                                                          │
│          │  Same ID in labels                                       │
│          ▼                                                          │
│   ┌─────────────┐                                                   │
│   │   METRIC    │  db_query_duration{trace_id="abc-123"} 0.45      │
│   └─────────────┘  (exemplars link metrics to traces)              │
│                                                                     │
│   Click from metric ──▶ See trace ──▶ Jump to logs                  │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 2. The Right Tool for the Right Question

| Question                      | Best Tool            |
| ----------------------------- | -------------------- |
| "Is something wrong?"         | Metrics (alerts)     |
| "What's the trend?"           | Metrics (graphs)     |
| "Where is the bottleneck?"    | Traces (waterfall)   |
| "What services are affected?" | Traces (service map) |
| "What exactly happened?"      | Logs (search)        |
| "What was the error message?" | Logs (details)       |

### 3. Cost vs. Value Trade-off

```
┌─────────────────────────────────────────────────────────────────────┐
│                    COST CONSIDERATIONS                              │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   METRICS: Low cost, high value                                     │
│   └── Always collect, aggregate aggressively                        │
│                                                                     │
│   TRACES: Medium cost, high value for debugging                     │
│   └── Sample in production (1-10% of requests)                      │
│   └── Always trace errors and slow requests                         │
│                                                                     │
│   LOGS: High cost, essential for debugging                          │
│   └── Use log levels (DEBUG only in dev)                            │
│   └── Structured logging (JSON) for searchability                   │
│   └── Retention policies (delete old logs)                          │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Implementation Phases in DoIt

| Phase | Component           | Status  | Documentation                                                 |
| ----- | ------------------- | ------- | ------------------------------------------------------------- |
| 3.1   | Structured Logging  | Planned | -                                                             |
| 3.2   | Prometheus Metrics  | ✅ Done | [Prometheus Mental Model](./PROMETHEUS_MENTAL_MODEL.md)       |
| 3.2   | Grafana Dashboards  | ✅ Done | [Grafana Mental Model](./GRAFANA_MENTAL_MODEL.md)             |
| 3.3   | Distributed Tracing | Planned | [Tracing Mental Model](./DISTRIBUTED_TRACING_MENTAL_MODEL.md) |

---

## Related Documentation

- [Prometheus Mental Model](./PROMETHEUS_MENTAL_MODEL.md) - Deep dive into metrics
- [Grafana Mental Model](./GRAFANA_MENTAL_MODEL.md) - Visualization and dashboards
- [Distributed Tracing Mental Model](./DISTRIBUTED_TRACING_MENTAL_MODEL.md) - Request flow tracking
