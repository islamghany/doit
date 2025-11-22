# Docker Compose Implementation Guide

> **Complete implementation documentation for the DoIt API Docker Compose setup**

---

## 📋 Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Implementation Details](#implementation-details)
- [Configuration Deep Dive](#configuration-deep-dive)
- [Best Practices Applied](#best-practices-applied)
- [Testing & Validation](#testing--validation)
- [Production Considerations](#production-considerations)

---

## Overview

### What We've Built

A complete multi-container development environment for the DoIt API with:

- ✅ **API Service** - Your Go application
- ✅ **PostgreSQL** - Primary database
- ✅ **Redis** - Caching layer
- ✅ **Prometheus** - Metrics collection
- ✅ **Grafana** - Metrics visualization
- ✅ **Adminer** - Database management UI (optional)

### Why This Setup?

**Problem**: Running multiple services locally is complex:

- Database setup
- Redis installation
- Manual configuration management
- Port conflicts
- Environment differences

**Solution**: Docker Compose orchestrates everything with one command.

---

## Architecture

### Service Relationships

```
┌─────────────────────────────────────────────────────────┐
│                     doit_network                        │
│                                                         │
│  ┌──────────┐                                          │
│  │ Grafana  │────────────────┐                         │
│  │  :3000   │                │                         │
│  └──────────┘                │                         │
│                               ▼                         │
│  ┌──────────┐         ┌────────────┐                   │
│  │   API    │────────▶│ Prometheus │                   │
│  │  :8080   │         │   :9090    │                   │
│  └──────────┘         └────────────┘                   │
│       │                                                 │
│       │                                                 │
│       ├────────────▶ ┌────────────┐                    │
│       │             │ PostgreSQL │                     │
│       │             │   :5432    │                     │
│       │             └────────────┘                     │
│       │                                                 │
│       └────────────▶ ┌────────────┐                    │
│                     │   Redis    │                     │
│                     │   :6379    │                     │
│                     └────────────┘                     │
│                                                         │
│  ┌──────────┐                                          │
│  │ Adminer  │──────────────▶ PostgreSQL                │
│  │  :8081   │  (profile: tools)                        │
│  └──────────┘                                          │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

### Startup Sequence

```
1. PostgreSQL starts
   ├─ Waits for healthcheck (pg_isready)
   └─ Status: HEALTHY

2. Redis starts
   ├─ Waits for healthcheck (redis-cli ping)
   └─ Status: HEALTHY

3. API starts (depends on 1 & 2)
   ├─ Waits for healthcheck (wget /health)
   └─ Status: HEALTHY

4. Prometheus starts
   └─ Scrapes API metrics

5. Grafana starts
   └─ Connects to Prometheus

6. Adminer starts (if --profile tools)
   └─ Connects to PostgreSQL
```

---

## Implementation Details

### 1. Service: API

**Purpose**: Your Go application

**Configuration Highlights**:

```yaml
api:
  build:
    context: .
    dockerfile: infra/docker/dockerfile.service
  depends_on:
    postgres:
      condition: service_healthy # 🔑 Waits for DB to be ready
    redis:
      condition: service_healthy # 🔑 Waits for cache to be ready
  volumes:
    - ./:/app # 🔑 Hot reload in development
  healthcheck:
    test: ["CMD", "wget", "--spider", "-q", "http://localhost:8080/health"]
    interval: 30s
    timeout: 3s
    retries: 3
    start_period: 40s # 🔑 Gives app time to start
```

**Why These Choices?**

1. **`condition: service_healthy`**

   - **Problem**: API tries to connect before DB is ready → crash loop
   - **Solution**: Wait for healthcheck to pass
   - **Analogy**: Don't enter restaurant until "OPEN" sign is lit

2. **Volume Mount `./:/app`**

   - **Problem**: Need to rebuild image for every code change
   - **Solution**: Mount source code, changes reflect immediately
   - **Analogy**: Live editing a document vs. printing new copies

3. **`start_period: 40s`**
   - **Problem**: App initialization takes time (migrations, etc.)
   - **Solution**: Don't mark as unhealthy during startup
   - **Analogy**: Grace period for athlete warming up

**Environment Variables**:

```yaml
environment:
  DB_HOST: postgres # 🔑 Service name, not localhost!
  DB_PORT: 5432
  REDIS_ADDR: redis:6379 # 🔑 Service name, not localhost!

  # From .env file (with defaults)
  DB_USER: ${DB_USER:-doit}
  DB_PASSWORD: ${DB_PASSWORD:-doit123}
  JWT_SECRET: ${JWT_SECRET:-dev-secret-key}
```

**Why Service Names?**

- **Docker's DNS**: Service names resolve to container IPs
- **Example**: `postgres` → `172.20.0.2` (automatically)
- **Analogy**: Calling "Mom" instead of remembering phone number

---

### 2. Service: PostgreSQL

**Purpose**: Primary database

**Configuration Highlights**:

```yaml
postgres:
  image: postgres:16-alpine # 🔑 Lightweight Alpine variant
  volumes:
    - postgres_data:/var/lib/postgresql/data # 🔑 Named volume
  healthcheck:
    test: ["CMD-SHELL", "pg_isready -U ${DB_USER:-doit} -d ${DB_NAME:-doit}"]
    interval: 5s
    timeout: 3s
    retries: 5
    start_period: 10s
  restart: unless-stopped # 🔑 Auto-restart on failure
```

**Why These Choices?**

1. **Alpine Variant**

   - **Size**: `postgres:16-alpine` (169MB) vs `postgres:16` (379MB)
   - **Security**: Smaller attack surface
   - **Speed**: Faster downloads

2. **Named Volume**

   - **Problem**: Data lost when container stops
   - **Solution**: Named volume persists data
   - **Analogy**: Saving to cloud (persistent) vs. RAM (temporary)

   ```bash
   # Data survives this:
   docker-compose down
   docker-compose up

   # But not this:
   docker-compose down -v  # Removes volumes!
   ```

3. **Health Check**

   - **`pg_isready`**: PostgreSQL-specific health check
   - **Not enough**: Just `pg_isready` checks if server responds
   - **Better**: Also specify user and database

   ```bash
   # Good
   pg_isready -U doit -d doit

   # Not enough (server might be up but DB not ready)
   pg_isready
   ```

4. **`restart: unless-stopped`**
   - Restart on failure
   - Don't restart if manually stopped
   - **Analogy**: Auto-restart crashed app, but respect manual shutdown

**Environment Variables**:

```yaml
environment:
  POSTGRES_DB: ${DB_NAME:-doit}
  POSTGRES_USER: ${DB_USER:-doit}
  POSTGRES_PASSWORD: ${DB_PASSWORD:-doit123}
  PGDATA: /var/lib/postgresql/data/pgdata # 🔑 Custom data dir
```

**Why `PGDATA`?**

- **Problem**: Volume mount can conflict with PostgreSQL's expected structure
- **Solution**: Use subdirectory for data
- **Technical**: Avoids "lost+found" issues on some filesystems

---

### 3. Service: Redis

**Purpose**: Caching and session storage

**Configuration Highlights**:

```yaml
redis:
  image: redis:7-alpine
  command: redis-server --appendonly yes --requirepass ${REDIS_PASSWORD:-}
  volumes:
    - redis_data:/data # 🔑 Persist cache (optional but recommended)
  healthcheck:
    test: ["CMD", "redis-cli", "--raw", "incr", "ping"]
```

**Why These Choices?**

1. **`--appendonly yes`**

   - **Problem**: Redis is in-memory, data lost on crash
   - **Solution**: AOF (Append-Only File) persists every write
   - **Analogy**: Journaling every transaction vs. just keeping balance in head

   ```
   Normal Redis: RAM only → restart → data gone
   With AOF:     RAM + disk → restart → data restored
   ```

2. **`--requirepass`**

   - **Development**: Empty password (convenience)
   - **Production**: MUST set password
   - **Why empty string works**: `${REDIS_PASSWORD:-}` → if unset, empty

3. **Health Check**

   - **Why `incr ping`?**: Tests read AND write
   - **Not just `ping`**: Server could be up but not accepting writes

   ```bash
   # Better health check
   redis-cli --raw incr ping
   # Returns: 1, 2, 3... (proves writes work)

   # Basic health check (not enough)
   redis-cli ping
   # Returns: PONG (but might be read-only)
   ```

---

### 4. Service: Prometheus

**Purpose**: Metrics collection and alerting

**Configuration Highlights**:

```yaml
prometheus:
  image: prom/prometheus:latest
  volumes:
    - ./infra/docker/prometheus.yml:/etc/prometheus/prometheus.yml
    - prometheus_data:/prometheus
  command:
    - "--config.file=/etc/prometheus/prometheus.yml"
    - "--storage.tsdb.path=/prometheus"
    - "--web.enable-lifecycle" # 🔑 Reload config without restart
    - "--storage.tsdb.retention.time=30d" # 🔑 Keep 30 days of data
```

**Why These Choices?**

1. **`--web.enable-lifecycle`**

   - **Problem**: Config changes require restart
   - **Solution**: Reload via API call

   ```bash
   # Update prometheus.yml
   vim infra/docker/prometheus.yml

   # Reload without restart
   curl -X POST http://localhost:9090/-/reload
   ```

2. **Retention Time**

   - **Default**: 15 days
   - **Our choice**: 30 days
   - **Why**: More historical data for debugging
   - **Trade-off**: More disk space

3. **Bind Mount vs Named Volume**

   - **Config file**: Bind mount (edit on host)
   - **Data**: Named volume (managed by Docker)

   ```yaml
   volumes:
     - ./prometheus.yml:/etc/prometheus/prometheus.yml # Bind mount
     - prometheus_data:/prometheus # Named volume
   ```

**Prometheus Configuration** (`prometheus.yml`):

```yaml
scrape_configs:
  - job_name: "doit-api"
    scrape_interval: 10s
    metrics_path: "/metrics"
    static_configs:
      - targets: ["api:8080"] # 🔑 Service name!
        labels:
          service: "doit-api"
          environment: "development"
```

**Key Points**:

- **`targets: ["api:8080"]`**: Uses Docker DNS
- **`scrape_interval: 10s`**: Collects metrics every 10 seconds
- **Labels**: Add metadata for filtering in Grafana

---

### 5. Service: Grafana

**Purpose**: Metrics visualization and dashboards

**Configuration Highlights**:

```yaml
grafana:
  image: grafana/grafana:latest
  depends_on:
    - prometheus
  environment:
    GF_SECURITY_ADMIN_USER: ${GRAFANA_ADMIN_USER:-admin}
    GF_SECURITY_ADMIN_PASSWORD: ${GRAFANA_ADMIN_PASSWORD:-admin}
    GF_USERS_ALLOW_SIGN_UP: "false" # 🔑 Disable self-registration
    GF_AUTH_ANONYMOUS_ENABLED: "false" # 🔑 Require login
  volumes:
    - grafana_data:/var/lib/grafana
    - ./infra/docker/grafana/provisioning:/etc/grafana/provisioning
    - ./infra/docker/grafana/dashboards:/var/lib/grafana/dashboards
```

**Why These Choices?**

1. **Disable Sign-up**

   - **Why**: Prevent unauthorized access
   - **Development**: OK with default admin/admin
   - **Production**: Use strong password or OAuth

2. **Provisioning**

   - **Problem**: Manual setup (datasource, dashboards) on every restart
   - **Solution**: Provisioning configs auto-setup on start

   **Directory Structure**:

   ```
   grafana/
   ├── provisioning/
   │   ├── datasources/
   │   │   └── prometheus.yml    # Auto-add Prometheus
   │   └── dashboards/
   │       └── default.yml        # Auto-load dashboards
   └── dashboards/
       └── api-overview.json      # Your dashboard
   ```

3. **Datasource Provisioning** (`datasources/prometheus.yml`):

   ```yaml
   datasources:
     - name: Prometheus
       type: prometheus
       url: http://prometheus:9090 # 🔑 Service name!
       isDefault: true
       editable: false # 🔑 Prevent accidental changes
   ```

4. **Dashboard Provisioning** (`dashboards/default.yml`):

   ```yaml
   providers:
     - name: "Default"
       folder: "DoIt API"
       type: file
       options:
         path: /var/lib/grafana/dashboards
       allowUiUpdates: true # 🔑 Allow editing in UI
   ```

**Benefits**:

- Start Grafana → Dashboards already loaded ✅
- No manual configuration ✅
- Team members get same setup ✅
- Easy to version control ✅

---

### 6. Service: Adminer (Optional)

**Purpose**: Web-based database management

**Configuration Highlights**:

```yaml
adminer:
  image: adminer:latest
  ports:
    - "8081:8080"
  profiles:
    - tools # 🔑 Only start with --profile tools
  environment:
    ADMINER_DEFAULT_SERVER: postgres
    ADMINER_DESIGN: dracula # 🔑 Dark theme
```

**Why Optional (Profiles)?**

- **Not always needed**: Most development doesn't need GUI
- **Resource saving**: One less container
- **Start when needed**:

  ```bash
  # Normal start (no Adminer)
  docker-compose up -d

  # Start with Adminer
  docker-compose --profile tools up -d
  ```

**Using Adminer**:

1. Open: http://localhost:8081
2. Login:
   - **System**: PostgreSQL
   - **Server**: postgres
   - **Username**: doit
   - **Password**: doit123
   - **Database**: doit

---

## Configuration Deep Dive

### Volumes Explained

**Named Volumes** (Managed by Docker):

```yaml
volumes:
  postgres_data:
    name: doit_postgres_data # 🔑 Custom name (not random)
    driver: local
```

**Why Named Volumes?**

- **Persistence**: Data survives `docker-compose down`
- **Management**: Easy to inspect, backup, migrate
- **Performance**: Better than bind mounts (especially on Mac/Windows)

**Commands**:

```bash
# List volumes
docker volume ls

# Inspect volume
docker volume inspect doit_postgres_data

# Backup volume
docker run --rm -v doit_postgres_data:/data -v $(pwd):/backup \
  alpine tar czf /backup/postgres-backup.tar.gz -C /data .

# Restore volume
docker run --rm -v doit_postgres_data:/data -v $(pwd):/backup \
  alpine tar xzf /backup/postgres-backup.tar.gz -C /data
```

**Bind Mounts** (Direct host access):

```yaml
volumes:
  - ./:/app # API source code
  - ./infra/docker/prometheus.yml:/etc/prometheus/prometheus.yml
```

**Why Bind Mounts?**

- **Development**: Edit files on host, see changes in container
- **Configuration**: Easy to edit config files
- **Debugging**: Direct access to logs

**When to Use What?**

| Type       | Use For                   | Example            |
| ---------- | ------------------------- | ------------------ |
| Named      | Data persistence          | PostgreSQL data    |
| Named      | Generated data            | Prometheus metrics |
| Bind Mount | Source code (development) | Your Go app        |
| Bind Mount | Configuration files       | prometheus.yml     |
| Bind Mount | Shared assets             | Grafana dashboards |

---

### Networks Explained

```yaml
networks:
  doit_network:
    name: doit_network # 🔑 Custom name
    driver: bridge
```

**Bridge Network**:

- **Default**: Docker uses bridge by default
- **Isolation**: Services can't see containers outside this network
- **DNS**: Service names automatically resolve

**How It Works**:

```
┌─────────────────────────────────────────┐
│        doit_network (bridge)            │
│                                         │
│  API (172.20.0.2)                       │
│   ├─ Can reach: postgres, redis        │
│   └─ Uses: Service names via DNS       │
│                                         │
│  PostgreSQL (172.20.0.3)                │
│   └─ Accessible as: postgres           │
│                                         │
│  Redis (172.20.0.4)                     │
│   └─ Accessible as: redis               │
│                                         │
└─────────────────────────────────────────┘

Host (your computer):
  ├─ Can reach services via: localhost:8080, localhost:5432
  └─ Port mapping: container:8080 → host:8080
```

**Service Discovery Example**:

```go
// ❌ Wrong (localhost)
DB_HOST=localhost

// ✅ Correct (service name)
DB_HOST=postgres

// How Docker resolves it:
// 1. App queries DNS for "postgres"
// 2. Docker DNS returns "172.20.0.3"
// 3. App connects to 172.20.0.3:5432
```

---

### Environment Variables Strategy

**Three Sources**:

1. **`.env` File** (Secrets, environment-specific):

   ```bash
   DB_PASSWORD=secret123
   JWT_SECRET=my-jwt-secret
   ```

2. **`docker-compose.yml`** (Defaults, fixed values):

   ```yaml
   environment:
     DB_HOST: postgres
     DB_PORT: 5432
   ```

3. **Shell** (Override for testing):
   ```bash
   DB_PASSWORD=newpass docker-compose up
   ```

**Priority** (highest to lowest):

```
Shell > docker-compose.yml > .env file > ENV defaults
```

**Example**:

```yaml
environment:
  DB_USER: ${DB_USER:-doit}
  #         ^         ^
  #         |         └─ Default if not set
  #         └─────────── From .env or shell
```

**Scenarios**:

```bash
# Scenario 1: Nothing set
# Result: DB_USER=doit

# Scenario 2: .env has DB_USER=myuser
# Result: DB_USER=myuser

# Scenario 3: Shell override
DB_USER=testuser docker-compose up
# Result: DB_USER=testuser
```

---

### Health Checks Deep Dive

**Why Health Checks Matter**:

- **Problem**: Service starts ≠ Service ready
- **Example**: PostgreSQL container running but still initializing database
- **Solution**: Health check confirms service is truly ready

**Health Check Lifecycle**:

```
Container starts
    ↓
start_period (grace period)
    ↓
First health check (after start_period)
    ↓
┌─────────────────┐
│  Check passes?  │
│  Yes → HEALTHY  │
│  No  → UNHEALTHY│
└─────────────────┘
    ↓
Wait interval
    ↓
Next check...
```

**PostgreSQL Health Check**:

```yaml
healthcheck:
  test: ["CMD-SHELL", "pg_isready -U doit -d doit"]
  interval: 5s # Check every 5 seconds
  timeout: 3s # Command must complete in 3 seconds
  retries: 5 # Fail after 5 failed attempts
  start_period: 10s # Grace period (don't fail during this)
```

**Timeline**:

```
0s:  Container starts
     └─ start_period begins (10s grace)

10s: First health check
     └─ pg_isready -U doit -d doit

     Result: Failed (DB still initializing)
     Status: STARTING (still in grace period)

15s: Second check (5s interval)
     Result: Failed
     Status: STARTING

20s: Third check
     Result: Success ✅
     Status: HEALTHY

     API can now start! (depends_on condition met)
```

**Depends On with Conditions**:

```yaml
api:
  depends_on:
    postgres:
      condition: service_healthy # ⏱️ Wait for HEALTHY
    redis:
      condition: service_healthy
```

**What This Does**:

1. Start postgres and redis first
2. Wait for both to become HEALTHY
3. Only then start API

**Without Conditions** (Basic `depends_on`):

```yaml
api:
  depends_on:
    - postgres # ⚠️ Only waits for container to START, not be ready!
```

**Why This Is Bad**:

```
0s:  PostgreSQL starts
1s:  API starts (depends_on satisfied)
2s:  API tries to connect to DB → fails (DB not ready)
3s:  API crashes 💥
```

---

## Best Practices Applied

### 1. **Explicit Service Names**

❌ **Bad** (Auto-generated names):

```yaml
services:
  postgres:
    container_name: doit_postgres_1 # Random suffix
```

✅ **Good** (Explicit names):

```yaml
services:
  postgres:
    container_name: doit-postgres # Predictable
```

**Why?**

- Easier to identify containers
- Consistent across team members
- Easier to script operations

---

### 2. **Restart Policies**

```yaml
restart: unless-stopped
```

**Options**:

- `no`: Never restart
- `always`: Always restart (even after manual stop)
- `on-failure`: Only restart on error
- `unless-stopped`: Restart unless manually stopped ✅

**Why `unless-stopped`?**

- Resilient to crashes
- Respects manual shutdowns
- Good for development and production

---

### 3. **Security: Disable Unnecessary Features**

```yaml
grafana:
  environment:
    GF_USERS_ALLOW_SIGN_UP: "false"
    GF_AUTH_ANONYMOUS_ENABLED: "false"
```

**Principle**: Secure by default, relax as needed

---

### 4. **Resource Limits** (Optional but recommended)

```yaml
services:
  api:
    deploy:
      resources:
        limits:
          cpus: "0.5"
          memory: 512M
        reservations:
          cpus: "0.25"
          memory: 256M
```

**Why?**

- Prevent one service from hogging resources
- Predictable performance
- Better for production

---

### 5. **Logging Configuration**

```yaml
services:
  api:
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

**Why?**

- Prevents logs from filling disk
- Keeps last 3 files of 10MB each
- Total max: 30MB per service

---

### 6. **Build Arguments for Flexibility**

```yaml
api:
  build:
    context: .
    dockerfile: infra/docker/dockerfile.service
    args:
      VERSION: ${VERSION:-dev}
      GO_VERSION: "1.21"
```

**Allows**:

```bash
# Development
docker-compose build

# Production with version
VERSION=1.0.0 docker-compose build
```

---

## Testing & Validation

### Pre-flight Checklist

Before starting services:

```bash
# 1. Check Docker is running
docker info

# 2. Check .env file exists
make compose-setup

# 3. Validate compose file
docker-compose config --quiet

# 4. Check no port conflicts
lsof -i :8080,5432,6379,9090,3000
```

---

### Start Services

```bash
# Start all services
make compose-up

# Expected output:
Creating network "doit_network" ... done
Creating volume "doit_postgres_data" ... done
Creating volume "doit_redis_data" ... done
Creating doit-postgres ... done
Creating doit-redis    ... done
Creating doit-api      ... done
Creating doit-prometheus ... done
Creating doit-grafana  ... done
```

---

### Verify Health

```bash
# Check all services are healthy
make compose-health

# Expected output:
🏥 Checking service health...

📊 API Health:
{
  "status": "ok",
  "timestamp": "2025-11-22T10:30:00Z"
}

📊 PostgreSQL Health:
postgres:5432 - accepting connections

📊 Redis Health:
PONG

📊 Prometheus Health:
Prometheus is Healthy.

📊 Grafana Health:
{
  "database": "ok",
  "version": "10.0.0"
}
```

---

### Test Each Service

**1. API**:

```bash
# Health check
curl http://localhost:8080/health | jq

# Swagger docs
open http://localhost:8080/swagger/index.html
```

**2. PostgreSQL**:

```bash
# Connect
make compose-shell-db

# List tables
\dt

# Query
SELECT * FROM users LIMIT 5;
```

**3. Redis**:

```bash
# Connect
make compose-shell-redis

# Test set/get
SET test "Hello Docker Compose"
GET test
```

**4. Prometheus**:

```bash
# Check targets
open http://localhost:9090/targets

# Query metrics
curl 'http://localhost:9090/api/v1/query?query=up' | jq
```

**5. Grafana**:

```bash
# Login
open http://localhost:3000

# Credentials: admin/admin

# Check datasource
curl -u admin:admin http://localhost:3000/api/datasources | jq
```

---

### Common Issues & Fixes

#### Issue: Services not starting

```bash
# Check logs
make compose-logs

# Look for errors in specific service
make compose-logs-api
```

#### Issue: Port already in use

```bash
# Find process
lsof -i :8080

# Kill process
kill -9 <PID>

# Or change port in docker-compose.yml
```

#### Issue: Database connection refused

```bash
# Verify PostgreSQL is healthy
docker-compose ps

# Check logs
make compose-logs-db

# Verify health check
docker-compose exec postgres pg_isready -U doit -d doit
```

#### Issue: Environment variables not working

```bash
# Verify .env is loaded
docker-compose config | grep DB_PASSWORD

# Check service environment
docker-compose exec api env | grep DB
```

---

## Production Considerations

### What to Change for Production

#### 1. **Secrets Management**

❌ **Development**:

```yaml
environment:
  DB_PASSWORD: doit123
```

✅ **Production**:

```yaml
environment:
  DB_PASSWORD: ${DB_PASSWORD} # From secrets manager
```

Use:

- Docker Secrets (Swarm)
- AWS Secrets Manager
- HashiCorp Vault
- Environment variables from CI/CD

---

#### 2. **TLS/SSL**

```yaml
postgres:
  environment:
    DB_SSL_MODE: require # Not disable!
```

---

#### 3. **Remove Development Tools**

```yaml
# Remove these in production
services:
  adminer: # ❌ Remove entire service

  api:
    volumes:
      - ./:/app # ❌ Remove source code mount
```

---

#### 4. **Resource Limits**

```yaml
api:
  deploy:
    resources:
      limits:
        cpus: "1.0"
        memory: 1G
      reservations:
        cpus: "0.5"
        memory: 512M
    restart_policy:
      condition: on-failure
      max_attempts: 3
```

---

#### 5. **Logging**

```yaml
logging:
  driver: "json-file"
  options:
    max-size: "10m"
    max-file: "5"
```

Or use centralized logging:

- Fluentd
- ELK Stack
- CloudWatch

---

#### 6. **Health Check Endpoints**

Ensure health checks are meaningful:

```yaml
healthcheck:
  test: ["CMD", "wget", "-q", "--spider", "http://localhost:8080/health/ready"]
```

Where `/health/ready` checks:

- Database connectivity ✅
- Redis connectivity ✅
- Critical dependencies ✅

---

#### 7. **Use Specific Image Tags**

❌ **Bad**:

```yaml
image: postgres:latest # Can break unexpectedly
```

✅ **Good**:

```yaml
image: postgres:16.1-alpine # Specific, reproducible
```

---

#### 8. **Network Security**

```yaml
networks:
  frontend:
    internal: false # Exposed to outside
  backend:
    internal: true # Isolated (DB, Redis)

services:
  api:
    networks:
      - frontend
      - backend

  postgres:
    networks:
      - backend # Only accessible internally
```

---

### Production Alternatives

Consider these instead of Docker Compose:

1. **Kubernetes** (orchestration at scale)
2. **Docker Swarm** (simpler than K8s)
3. **Managed Services**:
   - AWS ECS/Fargate
   - Google Cloud Run
   - Azure Container Instances

**When to use Docker Compose in production?**

✅ Good for:

- Small applications
- Single server deployments
- Internal tools
- Staging environments

❌ Not ideal for:

- Multi-server orchestration
- Auto-scaling needs
- High availability requirements
- Large-scale applications

---

## Summary

### What We Achieved

✅ Complete development environment with one command  
✅ Service orchestration with proper dependencies  
✅ Data persistence with volumes  
✅ Monitoring and metrics (Prometheus + Grafana)  
✅ Development conveniences (hot reload, easy logs)  
✅ Production-ready patterns (health checks, restart policies)  
✅ Comprehensive documentation

### Next Steps

1. **Phase 2.3**: Run migrations automatically on startup
2. **Phase 3**: API testing with Postman/curl
3. **Phase 5**: Kubernetes and Helm charts
4. **Phase 6**: AWS deployment

---

## Related Documentation

- [Docker Compose Mental Model](./DOCKER_COMPOSE_MENTAL_MODEL.md)
- [Docker Compose Quick Reference](./DOCKER_COMPOSE_QUICK_REFERENCE.md)
- [Docker Multi-Stage Build](./DOCKER_MULTISTAGE_IMPLEMENTATION.md)
- [Main README](../../README.md)

---

**Congratulations!** 🎉

You now have a professional-grade multi-container development environment with monitoring, caching, and database management—all orchestrated with Docker Compose!
