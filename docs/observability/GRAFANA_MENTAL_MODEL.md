# 📈 Grafana - Building Your Mental Model

## What is Grafana?

Grafana is a **visualization and dashboarding platform**. Think of it as the "presentation layer" for your observability data. While Prometheus stores metrics, Grafana makes them **visual and actionable**.

---

## The Core Mental Model

```
┌─────────────────────────────────────────────────────────────────────┐
│                    GRAFANA'S ROLE                                   │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   Data Sources              Grafana              Outputs            │
│   ────────────              ───────              ───────            │
│                                                                     │
│   ┌──────────┐                                                      │
│   │Prometheus│──┐                              ┌──────────────┐     │
│   └──────────┘  │                              │  Dashboards  │     │
│                 │      ┌──────────────┐        └──────────────┘     │
│   ┌──────────┐  │      │              │                             │
│   │  Jaeger  │──┼─────▶│   GRAFANA    │───────▶ ┌──────────────┐   │
│   └──────────┘  │      │              │        │    Alerts     │    │
│                 │      └──────────────┘        └──────────────┘     │
│   ┌──────────┐  │                                                   │
│   │PostgreSQL│──┘                              ┌──────────────┐     │
│   └──────────┘                                 │   Reports    │     │
│                                                └──────────────┘     │
│                                                                     │
│   Grafana doesn't store data - it VISUALIZES data from sources      │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

**Analogy**: 
- **Prometheus** = The filing cabinet (stores all the data)
- **Grafana** = The report generator (creates visual reports from the data)

---

## Core Concepts

### 1️⃣ Data Sources - Where Data Comes From

```
┌─────────────────────────────────────────────────────────────────────┐
│                       DATA SOURCES                                  │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   Grafana connects to MANY types of data sources:                   │
│                                                                     │
│   METRICS:                                                          │
│   ├── Prometheus (most common)                                      │
│   ├── InfluxDB                                                      │
│   ├── CloudWatch                                                    │
│   └── Graphite                                                      │
│                                                                     │
│   LOGS:                                                             │
│   ├── Loki (Grafana's log aggregator)                               │
│   ├── Elasticsearch                                                 │
│   └── CloudWatch Logs                                               │
│                                                                     │
│   TRACES:                                                           │
│   ├── Jaeger                                                        │
│   ├── Tempo (Grafana's trace backend)                               │
│   └── Zipkin                                                        │
│                                                                     │
│   DATABASES:                                                        │
│   ├── PostgreSQL                                                    │
│   ├── MySQL                                                         │
│   └── MongoDB                                                       │
│                                                                     │
│   One dashboard can query MULTIPLE data sources!                    │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 2️⃣ Dashboards - The Big Picture

```
┌─────────────────────────────────────────────────────────────────────┐
│                        DASHBOARD                                    │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   A dashboard is a COLLECTION of panels organized into rows         │
│                                                                     │
│   ┌─────────────────────────────────────────────────────────────┐  │
│   │  🗄️ DoIt API - Overview                          [⟳ 10s]    │  │
│   ├─────────────────────────────────────────────────────────────┤  │
│   │                                                              │  │
│   │  ROW: Key Metrics                                            │  │
│   │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐        │  │
│   │  │ Request  │ │  Error   │ │   P95    │ │ In-Flight│        │  │
│   │  │  Rate    │ │  Rate    │ │ Latency  │ │ Requests │        │  │
│   │  │  150/s   │ │  0.5%    │ │  45ms    │ │    23    │        │  │
│   │  └──────────┘ └──────────┘ └──────────┘ └──────────┘        │  │
│   │                                                              │  │
│   │  ROW: Trends                                                 │  │
│   │  ┌────────────────────────────┐ ┌────────────────────────┐  │  │
│   │  │    Request Rate Over Time  │ │   Latency Percentiles  │  │  │
│   │  │    📈 ~~~~/\~~~~           │ │   📈 P50 P90 P95 P99   │  │  │
│   │  └────────────────────────────┘ └────────────────────────┘  │  │
│   │                                                              │  │
│   │  ROW: Details                                                │  │
│   │  ┌────────────────────────────────────────────────────────┐ │  │
│   │  │              Top Endpoints Table                        │ │  │
│   │  │  Endpoint          │ Requests │ Errors │ P95           │ │  │
│   │  │  /api/v1/todos     │  5000    │  25    │ 34ms          │ │  │
│   │  │  /api/v1/users     │  2000    │  10    │ 45ms          │ │  │
│   │  └────────────────────────────────────────────────────────┘ │  │
│   │                                                              │  │
│   └─────────────────────────────────────────────────────────────┘  │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 3️⃣ Panels - Individual Visualizations

Each panel is a single visualization with its own query:

```
┌─────────────────────────────────────────────────────────────────────┐
│                         PANEL TYPES                                 │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  STAT - Single big number                                           │
│  ┌──────────────┐   Best for: Current values, KPIs                  │
│  │     150      │   Example: Request rate, error count              │
│  │   req/sec    │                                                   │
│  └──────────────┘                                                   │
│                                                                     │
│  GAUGE - Value with thresholds                                      │
│  ┌──────────────┐   Best for: Saturation, capacity                  │
│  │    ◐ 75%     │   Example: CPU usage, pool saturation             │
│  │   [====== ]  │                                                   │
│  └──────────────┘                                                   │
│                                                                     │
│  TIME SERIES - Values over time                                     │
│  ┌──────────────┐   Best for: Trends, patterns                      │
│  │  📈 /\/\/\   │   Example: Request rate, latency                  │
│  │    ──────    │                                                   │
│  └──────────────┘                                                   │
│                                                                     │
│  BAR CHART - Comparing categories                                   │
│  ┌──────────────┐   Best for: Comparisons                           │
│  │  ████        │   Example: Requests by endpoint                   │
│  │  ██████      │                                                   │
│  │  ███         │                                                   │
│  └──────────────┘                                                   │
│                                                                     │
│  TABLE - Detailed data                                              │
│  ┌──────────────┐   Best for: Detailed breakdowns                   │
│  │ Col1 │ Col2  │   Example: Top endpoints, error details           │
│  │──────┼───────│                                                   │
│  │ val  │ val   │                                                   │
│  └──────────────┘                                                   │
│                                                                     │
│  PIE CHART - Distribution                                           │
│  ┌──────────────┐   Best for: Proportions                           │
│  │    ◔◕◑       │   Example: Traffic by status code                 │
│  │   /   \      │                                                   │
│  └──────────────┘                                                   │
│                                                                     │
│  HEATMAP - Density over time                                        │
│  ┌──────────────┐   Best for: Distribution patterns                 │
│  │  ░▒▓█▓▒░     │   Example: Latency distribution                   │
│  │  ░▒▓▓▓▒░     │                                                   │
│  └──────────────┘                                                   │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 4️⃣ Queries - Getting Data

Each panel has one or more queries that fetch data:

```
┌─────────────────────────────────────────────────────────────────────┐
│                        PANEL QUERY                                  │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   Panel: Request Rate                                               │
│   ┌─────────────────────────────────────────────────────────────┐  │
│   │  Data Source: Prometheus                                     │  │
│   │                                                              │  │
│   │  Query A:                                                    │  │
│   │  ┌────────────────────────────────────────────────────────┐ │  │
│   │  │ sum(rate(http_requests_total[5m]))                     │ │  │
│   │  └────────────────────────────────────────────────────────┘ │  │
│   │  Legend: Total Requests                                      │  │
│   │                                                              │  │
│   │  Query B:                                                    │  │
│   │  ┌────────────────────────────────────────────────────────┐ │  │
│   │  │ sum(rate(http_requests_total{status=~"5.."}[5m]))      │ │  │
│   │  └────────────────────────────────────────────────────────┘ │  │
│   │  Legend: Errors                                              │  │
│   │                                                              │  │
│   └─────────────────────────────────────────────────────────────┘  │
│                                                                     │
│   Multiple queries = Multiple lines/series on the same panel        │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 5️⃣ Variables - Dynamic Dashboards

Variables make dashboards reusable and interactive:

```
┌─────────────────────────────────────────────────────────────────────┐
│                        VARIABLES                                    │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   Dashboard: API Overview                                           │
│   ┌─────────────────────────────────────────────────────────────┐  │
│   │  Environment: [Production ▼]  Service: [api-gateway ▼]      │  │
│   └─────────────────────────────────────────────────────────────┘  │
│                                                                     │
│   Variable Definition:                                              │
│   ┌─────────────────────────────────────────────────────────────┐  │
│   │  Name: environment                                           │  │
│   │  Type: Query                                                 │  │
│   │  Query: label_values(http_requests_total, environment)       │  │
│   │  Values: [production, staging, development]                  │  │
│   └─────────────────────────────────────────────────────────────┘  │
│                                                                     │
│   Using in queries:                                                 │
│   ┌─────────────────────────────────────────────────────────────┐  │
│   │  sum(rate(http_requests_total{env="$environment"}[5m]))     │  │
│   │                              ↑                               │  │
│   │                     Variable reference                       │  │
│   └─────────────────────────────────────────────────────────────┘  │
│                                                                     │
│   Benefits:                                                         │
│   • One dashboard for all environments                              │
│   • Users can filter without editing                                │
│   • Dropdown populated automatically from data                      │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 6️⃣ Thresholds & Colors - Visual Alerts

```
┌─────────────────────────────────────────────────────────────────────┐
│                   THRESHOLDS & COLORS                               │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   Thresholds turn numbers into meaning:                             │
│                                                                     │
│   Error Rate Panel:                                                 │
│   ┌─────────────────────────────────────────────────────────────┐  │
│   │  Thresholds:                                                 │  │
│   │  • 0% - 1%   → 🟢 Green (healthy)                            │  │
│   │  • 1% - 5%   → 🟡 Yellow (warning)                           │  │
│   │  • 5%+       → 🔴 Red (critical)                             │  │
│   │                                                              │  │
│   │  Current: 0.5%  →  🟢 Green                                  │  │
│   └─────────────────────────────────────────────────────────────┘  │
│                                                                     │
│   Visual impact:                                                    │
│   ┌──────────┐  ┌──────────┐  ┌──────────┐                         │
│   │  🟢 0.5% │  │  🟡 2.3% │  │  🔴 7.8% │                         │
│   │  Healthy │  │  Warning │  │ Critical │                         │
│   └──────────┘  └──────────┘  └──────────┘                         │
│                                                                     │
│   Instant visual feedback - no need to interpret numbers!           │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Alerting - Proactive Monitoring

```
┌─────────────────────────────────────────────────────────────────────┐
│                      GRAFANA ALERTING                               │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   Alert Rule:                                                       │
│   ┌─────────────────────────────────────────────────────────────┐  │
│   │  Name: High Error Rate                                       │  │
│   │  Condition: error_rate > 5% for 5 minutes                    │  │
│   │  Query: sum(rate(http_requests_total{status=~"5.."}[5m]))    │  │
│   │          / sum(rate(http_requests_total[5m])) * 100          │  │
│   │                                                              │  │
│   │  When triggered:                                             │  │
│   │  ├── Send to Slack #alerts                                   │  │
│   │  ├── Send email to oncall@company.com                        │  │
│   │  └── Create PagerDuty incident                               │  │
│   └─────────────────────────────────────────────────────────────┘  │
│                                                                     │
│   Alert States:                                                     │
│   • OK       → Condition not met                                    │
│   • Pending  → Condition met, waiting for duration                  │
│   • Alerting → Condition met for required duration                  │
│   • No Data  → Query returned no data                               │
│                                                                     │
│   Contact Points:                                                   │
│   ├── Slack                                                         │
│   ├── Email                                                         │
│   ├── PagerDuty                                                     │
│   ├── OpsGenie                                                      │
│   ├── Webhook                                                       │
│   └── Many more...                                                  │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Provisioning - Infrastructure as Code

```
┌─────────────────────────────────────────────────────────────────────┐
│                       PROVISIONING                                  │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   Instead of clicking in the UI, define everything in files:        │
│                                                                     │
│   grafana/                                                          │
│   ├── provisioning/                                                 │
│   │   ├── datasources/                                              │
│   │   │   └── prometheus.yml      # Define data sources             │
│   │   │                                                             │
│   │   ├── dashboards/                                               │
│   │   │   └── default.yml         # Where to find dashboards        │
│   │   │                                                             │
│   │   └── alerting/                                                 │
│   │       └── alerts.yml          # Alert rules                     │
│   │                                                                 │
│   └── dashboards/                                                   │
│       ├── api-overview.json       # Dashboard definitions           │
│       └── database.json                                             │
│                                                                     │
│   Benefits:                                                         │
│   • Version controlled (Git)                                        │
│   • Reproducible across environments                                │
│   • No manual setup needed                                          │
│   • Team collaboration                                              │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### Datasource Provisioning Example

```yaml
# provisioning/datasources/prometheus.yml
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: false
```

### Dashboard Provider Example

```yaml
# provisioning/dashboards/default.yml
apiVersion: 1

providers:
  - name: 'default'
    folder: 'DoIt API'
    type: file
    options:
      path: /var/lib/grafana/dashboards
```

---

## Dashboard Design Best Practices

```
┌─────────────────────────────────────────────────────────────────────┐
│                 DASHBOARD DESIGN PATTERNS                           │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  1. TOP-DOWN LAYOUT                                                 │
│     ┌─────────────────────────────────────────────────────────┐    │
│     │  ROW 1: Key metrics (Stats/Gauges) - "At a glance"      │    │
│     │  ROW 2: Trends (Time series) - "What's changing?"       │    │
│     │  ROW 3: Breakdowns (Tables/Bars) - "Details"            │    │
│     └─────────────────────────────────────────────────────────┘    │
│                                                                     │
│  2. THE "RED METHOD" FOR SERVICES                                   │
│     • Rate - Requests per second                                    │
│     • Errors - Error rate                                           │
│     • Duration - Latency percentiles                                │
│                                                                     │
│  3. THE "USE METHOD" FOR RESOURCES                                  │
│     • Utilization - How busy? (CPU %)                               │
│     • Saturation - How overloaded? (queue depth)                    │
│     • Errors - Error count                                          │
│                                                                     │
│  4. DASHBOARD HIERARCHY                                             │
│     ┌─────────────┐                                                 │
│     │  Overview   │  ← Start here (all services)                    │
│     └──────┬──────┘                                                 │
│            │                                                        │
│     ┌──────┴──────┐                                                 │
│     ▼             ▼                                                 │
│  ┌──────┐    ┌──────┐                                               │
│  │ API  │    │  DB  │  ← Drill down (specific service)              │
│  └──────┘    └──────┘                                               │
│                                                                     │
│  5. CONSISTENT COLORS                                               │
│     • Green = Good/Success                                          │
│     • Yellow = Warning                                              │
│     • Red = Error/Critical                                          │
│     • Blue = Informational                                          │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Common PromQL Patterns for Grafana

```
┌─────────────────────────────────────────────────────────────────────┐
│               COMMON PROMQL FOR DASHBOARDS                          │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  REQUEST RATE:                                                      │
│  sum(rate(http_requests_total[5m]))                                 │
│                                                                     │
│  ERROR RATE (%):                                                    │
│  sum(rate(http_requests_total{status=~"5.."}[5m]))                  │
│  / sum(rate(http_requests_total[5m])) * 100                         │
│                                                                     │
│  LATENCY PERCENTILES:                                               │
│  histogram_quantile(0.50, rate(http_duration_bucket[5m]))  # P50    │
│  histogram_quantile(0.95, rate(http_duration_bucket[5m]))  # P95    │
│  histogram_quantile(0.99, rate(http_duration_bucket[5m]))  # P99    │
│                                                                     │
│  TOP 5 BY RATE:                                                     │
│  topk(5, sum by (path) (rate(http_requests_total[5m])))             │
│                                                                     │
│  AVAILABILITY (%):                                                  │
│  (1 - sum(rate(http_requests_total{status=~"5.."}[5m]))             │
│       / sum(rate(http_requests_total[5m]))) * 100                   │
│                                                                     │
│  SATURATION:                                                        │
│  db_connections_in_use / db_connections_max * 100                   │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Mental Model Summary

```
┌─────────────────────────────────────────────────────────────────────┐
│                    GRAFANA MENTAL MODEL                             │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  GRAFANA = Visualization layer for observability data               │
│            (doesn't store data, just displays it beautifully)       │
│                                                                     │
│  DATA SOURCES = Connections to where data lives                     │
│                 (Prometheus, Jaeger, PostgreSQL, etc.)              │
│                                                                     │
│  DASHBOARDS = Collections of panels organized by topic              │
│               (API Overview, Database Performance, etc.)            │
│                                                                     │
│  PANELS = Individual visualizations with queries                    │
│           (Stat, Gauge, Time Series, Table, etc.)                   │
│                                                                     │
│  VARIABLES = Make dashboards dynamic and reusable                   │
│              ($environment, $service, etc.)                         │
│                                                                     │
│  PROVISIONING = Infrastructure as Code for Grafana                  │
│                 (version control your dashboards!)                  │
│                                                                     │
│  ┌───────────────────────────────────────────────────────────────┐ │
│  │  THE POWER:                                                    │ │
│  │                                                                │ │
│  │  Raw numbers ──▶ Visual insights                               │ │
│  │  "Error rate is 0.023" ──▶ 🟢 "System is healthy"              │ │
│  │  "Latency is 0.450" ──▶ 🔴 "Database is slow!"                 │ │
│  │                                                                │ │
│  │  Humans process visuals 60,000x faster than text.              │ │
│  │                                                                │ │
│  └───────────────────────────────────────────────────────────────┘ │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Related Documentation

- [Prometheus Mental Model](./PROMETHEUS_MENTAL_MODEL.md) - Metrics collection
- [Distributed Tracing Mental Model](./DISTRIBUTED_TRACING_MENTAL_MODEL.md) - Request flow tracking
- [Observability Overview](./OBSERVABILITY_OVERVIEW.md) - The three pillars combined

