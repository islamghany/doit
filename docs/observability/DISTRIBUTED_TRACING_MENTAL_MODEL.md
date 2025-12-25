# 🔍 Distributed Tracing - Building Your Mental Model

## The Problem: "Where Did My Request Go?"

Imagine you're a detective investigating a crime. You have three types of evidence:

| Evidence Type | Observability Pillar | What It Tells You |
|--------------|---------------------|-------------------|
| **Witness statements** | Logs | "What happened" - detailed events |
| **Statistics** | Metrics | "How much" - counts, rates, percentages |
| **GPS tracking** | Traces | "The complete journey" - path through the system |

**The Gap**: Logs and metrics don't connect events across a request's journey.

```
Without Tracing:
┌─────────────────────────────────────────────────────────────────────┐
│  Log: "User 123 requested todos"                                    │
│  Log: "Cache miss for key todos:123"                                │
│  Log: "Query executed in 450ms"                                     │
│  Log: "Response sent"                                               │
│                                                                     │
│  Question: Are these logs from the SAME request? 🤷                 │
│  Question: Which operation caused the 500ms total latency? 🤷       │
└─────────────────────────────────────────────────────────────────────┘

With Tracing:
┌─────────────────────────────────────────────────────────────────────┐
│  [Trace: abc-123] GET /api/v1/todos                                 │
│  ├── All events connected by the same Trace ID                      │
│  ├── Each operation timed individually                              │
│  └── Parent-child relationships show causality                      │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Core Concepts

### 1️⃣ Trace - The Complete Story

A **Trace** represents the entire journey of a single request through your system.

```
┌─────────────────────────────────────────────────────────────────────┐
│                        TRACE: abc-123-def-456                       │
│                                                                     │
│  "Everything that happened because of this one user action"         │
│                                                                     │
│  • Unique ID that follows the request everywhere                    │
│  • Contains multiple spans (operations)                             │
│  • Shows total duration and all sub-operations                      │
└─────────────────────────────────────────────────────────────────────┘
```

**Analogy**: A trace is like a **case file number** in a police investigation. Every piece of evidence, every interview, every action related to that case gets tagged with that number.

### 2️⃣ Span - A Single Operation

A **Span** represents a single unit of work within a trace.

```
┌─────────────────────────────────────────────────────────────────────┐
│                           SPAN                                      │
├─────────────────────────────────────────────────────────────────────┤
│  Name:        "PostgreSQL: SELECT todos"                            │
│  Trace ID:    abc-123-def-456                                       │
│  Span ID:     span-789                                              │
│  Parent ID:   span-456  (the span that created this one)            │
│  Start Time:  2024-01-15 10:30:00.123                               │
│  Duration:    45ms                                                  │
│  Status:      OK (or ERROR)                                         │
│  Attributes:  {                                                     │
│                 "db.system": "postgresql",                          │
│                 "db.statement": "SELECT * FROM todos WHERE...",     │
│                 "db.rows_affected": 15,                             │
│                 "user.id": "123"                                    │
│               }                                                     │
│  Events:      [                                                     │
│                 { time: ..., name: "connection acquired" },         │
│                 { time: ..., name: "query executed" }               │
│               ]                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

**Analogy**: A span is like a **chapter in the case file**. It has a title (name), timestamps, details (attributes), and notes about what happened (events).

### 3️⃣ Span Relationships - The Tree Structure

Spans form a **tree** showing what caused what:

```
GET /api/v1/todos (Root Span) ─────────────────────────────── 500ms
│
├── Auth Middleware ──────────────────────────────────────── 5ms
│   └── JWT Token Validation ─────────────────────────────── 3ms
│
├── Redis Cache Check ────────────────────────────────────── 2ms
│   └── GET todos:user:123 ───────────────────────────────── 2ms
│
├── PostgreSQL Query ─────────────────────────────────────── 480ms ⚠️
│   ├── Connection Pool Wait ─────────────────────────────── 450ms 🔴
│   └── Execute SELECT ───────────────────────────────────── 30ms
│
└── JSON Serialization ───────────────────────────────────── 8ms
```

**Visual Timeline (Waterfall View)**:
```
Time:   0ms        100ms       200ms       300ms       400ms       500ms
        |-----------|-----------|-----------|-----------|-----------|
Root:   [======================== GET /api/v1/todos ==================]
Auth:   [==]
JWT:     [=]
Redis:     [=]
PG:          [================= PostgreSQL Query ==================]
Wait:        [============== Connection Wait ================]
Query:                                                      [====]
JSON:                                                            [==]
                                                    ↑
                                          BOTTLENECK IDENTIFIED!
```

### 4️⃣ Context Propagation - Passing the Baton

How does the trace ID travel between services?

```
┌─────────────────────────────────────────────────────────────────────┐
│                     CONTEXT PROPAGATION                             │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  Client Request                                                     │
│       │                                                             │
│       │  HTTP Headers:                                              │
│       │  traceparent: 00-abc123-span456-01                          │
│       │  tracestate: vendor=info                                    │
│       ▼                                                             │
│  ┌─────────────┐     HTTP + Headers     ┌─────────────┐            │
│  │   API       │ ───────────────────▶   │  Service B  │            │
│  │  (Go)       │                        │  (Python)   │            │
│  └─────────────┘                        └─────────────┘            │
│       │                                       │                     │
│       │ Same Trace ID                         │ Same Trace ID       │
│       ▼                                       ▼                     │
│  ┌─────────────┐                        ┌─────────────┐            │
│  │ PostgreSQL  │                        │   Redis     │            │
│  └─────────────┘                        └─────────────┘            │
│                                                                     │
│  ALL operations linked by Trace ID: abc123                          │
└─────────────────────────────────────────────────────────────────────┘
```

**Standard Headers** (W3C Trace Context):
```
traceparent: 00-{trace-id}-{parent-span-id}-{flags}
             │      │            │            │
             │      │            │            └── 01 = sampled
             │      │            └── Current span
             │      └── 32-char hex trace ID
             └── Version (always 00)

Example: traceparent: 00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01
```

---

## The Three Pillars of Observability

```
┌─────────────────────────────────────────────────────────────────────┐
│                    THREE PILLARS OF OBSERVABILITY                   │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   LOGS                    METRICS                  TRACES           │
│   ────                    ───────                  ──────           │
│                                                                     │
│   📝 What happened        📊 How much              🔍 The journey   │
│                                                                     │
│   • Detailed events       • Aggregated data        • Request flow   │
│   • Debug info            • Trends over time       • Causality      │
│   • Error messages        • Alerting               • Latency breakdown│
│   • High cardinality      • Low cardinality        • Sampling       │
│                                                                     │
│   "User 123 failed        "Error rate: 2.5%"       "Request abc123  │
│    to authenticate"                                 failed at DB    │
│                                                     connection step"│
│                                                                     │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   USE TOGETHER:                                                     │
│                                                                     │
│   1. Metrics alert you: "Error rate spike!"                         │
│   2. Traces show you: "Errors happen in DB connection step"         │
│   3. Logs give details: "Connection timeout after 30s"              │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## OpenTelemetry - The Universal Standard

### What is OpenTelemetry?

```
┌─────────────────────────────────────────────────────────────────────┐
│                       OPENTELEMETRY (OTel)                          │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   "A vendor-neutral observability framework"                        │
│                                                                     │
│   ┌─────────────────────────────────────────────────────────────┐  │
│   │                      OTel SDK                                │  │
│   │  ┌──────────┐  ┌──────────┐  ┌──────────┐                   │  │
│   │  │  Traces  │  │  Metrics │  │   Logs   │                   │  │
│   │  │   API    │  │   API    │  │   API    │                   │  │
│   │  └────┬─────┘  └────┬─────┘  └────┬─────┘                   │  │
│   │       │             │             │                          │  │
│   │       └─────────────┼─────────────┘                          │  │
│   │                     │                                        │  │
│   │              ┌──────▼──────┐                                 │  │
│   │              │   OTLP      │  (OpenTelemetry Protocol)       │  │
│   │              │  Exporter   │                                 │  │
│   │              └──────┬──────┘                                 │  │
│   └─────────────────────┼───────────────────────────────────────┘  │
│                         │                                          │
│                         ▼                                          │
│   ┌─────────────────────────────────────────────────────────────┐  │
│   │                  ANY BACKEND                                 │  │
│   │  ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐    │  │
│   │  │ Jaeger │ │ Zipkin │ │ Tempo  │ │AWS X-Ray│ │Datadog │    │  │
│   │  └────────┘ └────────┘ └────────┘ └────────┘ └────────┘    │  │
│   └─────────────────────────────────────────────────────────────┘  │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

**Analogy**:
- **OpenTelemetry** = USB standard (universal connector)
- **Jaeger/Zipkin/X-Ray** = Different devices that accept USB
- **Your Code** = Uses the USB standard, works with any device

### Why OpenTelemetry?

| Before OTel | After OTel |
|-------------|------------|
| Vendor lock-in (Datadog SDK, X-Ray SDK, etc.) | One SDK, any backend |
| Different APIs for traces/metrics/logs | Unified API |
| Inconsistent instrumentation | Standardized conventions |
| Hard to switch vendors | Easy migration |

---

## Jaeger - The Trace Backend

### What is Jaeger?

```
┌─────────────────────────────────────────────────────────────────────┐
│                           JAEGER                                    │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   "Open-source distributed tracing platform"                        │
│                                                                     │
│   ┌─────────────────────────────────────────────────────────────┐  │
│   │                     JAEGER ARCHITECTURE                      │  │
│   │                                                              │  │
│   │   Your App ──▶ Collector ──▶ Storage ──▶ Query ──▶ UI       │  │
│   │                    │            │                            │  │
│   │                    │            └── Elasticsearch/Cassandra  │  │
│   │                    │                 or in-memory            │  │
│   │                    └── Receives traces via OTLP              │  │
│   └─────────────────────────────────────────────────────────────┘  │
│                                                                     │
│   FEATURES:                                                         │
│   • Trace visualization (waterfall view)                            │
│   • Service dependency graphs                                       │
│   • Performance analysis                                            │
│   • Root cause analysis                                             │
│   • Trace comparison                                                │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### Jaeger UI - What You'll See

```
┌─────────────────────────────────────────────────────────────────────┐
│  JAEGER UI                                              [Search]    │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  Service: doit-api          Operation: GET /api/v1/todos            │
│  Time: Last 1 hour          Limit: 20 traces                        │
│                                                                     │
│  ┌───────────────────────────────────────────────────────────────┐ │
│  │ TRACE LIST                                                     │ │
│  │                                                                │ │
│  │ ● abc123  GET /api/v1/todos     500ms   12 spans   3 errors   │ │
│  │ ○ def456  POST /api/v1/todos    45ms    8 spans    0 errors   │ │
│  │ ○ ghi789  GET /api/v1/todos/1   23ms    6 spans    0 errors   │ │
│  └───────────────────────────────────────────────────────────────┘ │
│                                                                     │
│  ┌───────────────────────────────────────────────────────────────┐ │
│  │ TRACE DETAIL: abc123                                          │ │
│  │                                                                │ │
│  │ Timeline:  0ms    100ms   200ms   300ms   400ms   500ms       │ │
│  │            |-------|-------|-------|-------|-------|          │ │
│  │                                                                │ │
│  │ doit-api                                                       │ │
│  │ └─ GET /api/v1/todos [=================================] 500ms │ │
│  │    ├─ auth.middleware [==] 5ms                                │ │
│  │    ├─ redis.get [=] 2ms                                       │ │
│  │    ├─ pg.query [===========================] 480ms ⚠️         │ │
│  │    │  └─ connection.wait [======================] 450ms 🔴    │ │
│  │    └─ json.marshal [=] 8ms                                    │ │
│  │                                                                │ │
│  │ SPAN DETAILS: pg.query                                        │ │
│  │ ┌────────────────────────────────────────────────────────┐   │ │
│  │ │ db.system: postgresql                                   │   │ │
│  │ │ db.name: doit                                           │   │ │
│  │ │ db.statement: SELECT * FROM todos WHERE user_id = $1    │   │ │
│  │ │ db.rows_affected: 15                                    │   │ │
│  │ │ error: false                                            │   │ │
│  │ └────────────────────────────────────────────────────────┘   │ │
│  └───────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────┘
```

---

## How It Applies to DoIt API

### Current Request Flow (Without Tracing)

```
User Request ──▶ API ──▶ ??? ──▶ Response

"Something is slow, but what?"
```

### With Tracing Instrumented

```
┌─────────────────────────────────────────────────────────────────────┐
│                    TRACED REQUEST FLOW                              │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  User: GET /api/v1/todos                                            │
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  SPAN: HTTP Handler                                          │   │
│  │  ├── Attributes: method=GET, path=/api/v1/todos              │   │
│  │  │                                                           │   │
│  │  │  ┌─────────────────────────────────────────────────────┐ │   │
│  │  │  │  SPAN: Auth Middleware                              │ │   │
│  │  │  │  └── Attributes: user_id=123, role=user             │ │   │
│  │  │  └─────────────────────────────────────────────────────┘ │   │
│  │  │                                                           │   │
│  │  │  ┌─────────────────────────────────────────────────────┐ │   │
│  │  │  │  SPAN: Redis Cache                                  │ │   │
│  │  │  │  └── Attributes: operation=GET, key=todos:123       │ │   │
│  │  │  │      Events: [cache_miss]                           │ │   │
│  │  │  └─────────────────────────────────────────────────────┘ │   │
│  │  │                                                           │   │
│  │  │  ┌─────────────────────────────────────────────────────┐ │   │
│  │  │  │  SPAN: PostgreSQL Query                             │ │   │
│  │  │  │  ├── Attributes: operation=SELECT, table=todos      │ │   │
│  │  │  │  │               rows=15                            │ │   │
│  │  │  │  │                                                  │ │   │
│  │  │  │  │  ┌───────────────────────────────────────────┐  │ │   │
│  │  │  │  │  │  SPAN: Connection Acquire                 │  │ │   │
│  │  │  │  │  │  └── Attributes: pool_size=25, in_use=24  │  │ │   │
│  │  │  │  │  └───────────────────────────────────────────┘  │ │   │
│  │  │  │  └─────────────────────────────────────────────────┘ │   │
│  │  │  └─────────────────────────────────────────────────────┘ │   │
│  │  │                                                           │   │
│  │  │  ┌─────────────────────────────────────────────────────┐ │   │
│  │  │  │  SPAN: Response Serialization                       │ │   │
│  │  │  │  └── Attributes: format=json, size=2048             │ │   │
│  │  │  └─────────────────────────────────────────────────────┘ │   │
│  │  └───────────────────────────────────────────────────────────┘   │
│  └─────────────────────────────────────────────────────────────────┘│
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### What You'll Instrument in DoIt

| Layer | What to Trace | Span Name Example | Key Attributes |
|-------|---------------|-------------------|----------------|
| **HTTP** | Incoming requests | `HTTP GET /api/v1/todos` | method, path, status_code, user_id |
| **Middleware** | Auth, rate limiting | `auth.jwt_validation` | user_id, role, token_exp |
| **Service** | Business logic | `todo.list_user_todos` | user_id, count |
| **Database** | SQL queries | `pg.query.list_todos` | db.statement, rows, table |
| **Cache** | Redis operations | `redis.get` | key, hit/miss |

---

## Sampling - Don't Trace Everything

At high traffic, tracing every request is expensive. **Sampling** decides which requests to trace:

```
┌─────────────────────────────────────────────────────────────────────┐
│                      SAMPLING STRATEGIES                            │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  1. ALWAYS SAMPLE (Development)                                     │
│     └── Trace 100% of requests                                      │
│                                                                     │
│  2. PROBABILISTIC (Production)                                      │
│     └── Trace X% of requests (e.g., 10%)                            │
│                                                                     │
│  3. RATE LIMITING                                                   │
│     └── Trace N requests per second                                 │
│                                                                     │
│  4. TAIL-BASED (Advanced)                                           │
│     └── Decide AFTER seeing the whole trace                         │
│     └── Always keep errors, slow requests                           │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## AWS X-Ray Connection

Everything you learn with OpenTelemetry + Jaeger applies to AWS X-Ray:

| OpenTelemetry Concept | AWS X-Ray Equivalent |
|----------------------|---------------------|
| Trace | Trace |
| Span | Segment / Subsegment |
| Trace ID | Trace ID (different format) |
| Span Attributes | Annotations / Metadata |
| Context Propagation | X-Amzn-Trace-Id header |

```
┌─────────────────────────────────────────────────────────────────────┐
│                    MIGRATION PATH                                   │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  Development:  OTel SDK ──▶ Jaeger (local Docker)                   │
│                                                                     │
│  Production:   OTel SDK ──▶ AWS X-Ray (via OTel Collector)          │
│                                                                     │
│  Same code, different exporter!                                     │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Mental Model Summary

```
┌─────────────────────────────────────────────────────────────────────┐
│                   DISTRIBUTED TRACING MENTAL MODEL                  │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  TRACE = Complete story of one request                              │
│          (like a case file number)                                  │
│                                                                     │
│  SPAN = One chapter in the story                                    │
│         (a single operation with timing)                            │
│                                                                     │
│  CONTEXT = How the story travels between services                   │
│            (HTTP headers carrying trace ID)                         │
│                                                                     │
│  OPENTELEMETRY = Universal language for telling the story           │
│                  (vendor-neutral standard)                          │
│                                                                     │
│  JAEGER = Library where stories are stored and read                 │
│           (visualization and analysis)                              │
│                                                                     │
│  ┌───────────────────────────────────────────────────────────────┐ │
│  │  THE POWER:                                                    │ │
│  │                                                                │ │
│  │  "Request X is slow" ──▶ "Request X spent 450ms waiting       │ │
│  │                           for a database connection because    │ │
│  │                           the pool was saturated"              │ │
│  │                                                                │ │
│  │  From SYMPTOM to ROOT CAUSE in seconds.                        │ │
│  └───────────────────────────────────────────────────────────────┘ │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Questions to Solidify Understanding

1. **What's the difference between a Trace and a Span?**
2. **Why do we need context propagation?**
3. **How does tracing complement logs and metrics?**
4. **Why is OpenTelemetry vendor-neutral important?**
5. **When would you NOT trace a request (sampling)?**

---

## Related Documentation

- [Prometheus Mental Model](./PROMETHEUS_MENTAL_MODEL.md) - Metrics collection
- [Grafana Mental Model](./GRAFANA_MENTAL_MODEL.md) - Visualization and dashboards
- [Observability Overview](./OBSERVABILITY_OVERVIEW.md) - The three pillars combined

