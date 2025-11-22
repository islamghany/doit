# 🧠 Docker Compose - Complete Mental Model

## 📋 Table of Contents

1. [The Core Concept](#the-core-concept)
2. [Mental Model: Orchestra Conductor](#mental-model-orchestra-conductor)
3. [Core Concepts (Building Blocks)](#core-concepts-building-blocks)
4. [Real-World Scenario: Your DoIt Application](#real-world-scenario-your-doit-application)
5. [The Complete Lifecycle](#the-complete-lifecycle)
6. [Common Operations](#common-operations-mental-models)
7. [Design Patterns & Best Practices](#design-patterns--best-practices)
8. [Relationship to Other Technologies](#how-docker-compose-relates-to-what-you-already-know)
9. [Your DoIt Stack](#your-doit-stack---what-well-build)
10. [Validation Checklist](#validation-do-you-understand)

---

## 🎯 The Core Concept

### "Infrastructure as Code for Local Development"

**What Problem Does Docker Compose Solve?**

Imagine you're a new developer joining a team. Without Docker Compose:

```bash
# You'd have to do this manually:
1. Install PostgreSQL locally
2. Create database and user
3. Install Redis
4. Configure Redis
5. Set up environment variables
6. Run migrations
7. Start your API
8. Hope everything connects properly
9. Debug connection issues for hours...
```

**With Docker Compose:**

```bash
docker-compose up
# ✨ Everything works! (30 seconds)
```

**Key Insight:** Docker Compose lets you define your entire development environment in a single file, and recreate it anywhere with one command.

---

## 🏗️ Mental Model: Orchestra Conductor

Think of Docker Compose as a **conductor** leading an orchestra:

```
Docker Compose (Conductor)
    ↓
Reads the "Sheet Music" (docker-compose.yml)
    ↓
Coordinates Multiple "Musicians" (Containers)
    ↓
Ensures They Play Together in Harmony
```

### The Orchestra Analogy

| Orchestra           | Docker Compose                         |
| ------------------- | -------------------------------------- |
| **Conductor**       | Docker Compose CLI                     |
| **Sheet Music**     | `docker-compose.yml` file              |
| **Musicians**       | Individual containers (API, DB, Redis) |
| **Playing in sync** | Services can communicate               |
| **Same tempo**      | Containers start in order              |
| **Practice room**   | Your local machine                     |

**Why this analogy works:**

- A conductor doesn't play instruments (Compose doesn't run code)
- A conductor ensures timing and coordination (Compose manages dependencies)
- Musicians need the right instruments (Compose provisions resources)
- The sheet music defines everything (YAML is declarative)

---

## 📚 Core Concepts (Building Blocks)

### 1. **Services** - The "What"

A **service** is a container definition. It's not the container itself, but the _blueprint_ for creating containers.

**Mental Model:** Think of a service as a **job description**:

- "We need a PostgreSQL database"
- "We need a Redis cache"
- "We need our Go API"

```yaml
services:
  api:# Service name: "api"
    # This is the job description for the API container

  postgres:# Service name: "postgres"
    # This is the job description for the database

  redis:# Service name: "redis"
    # This is the job description for cache
```

**🔑 Key Insight:** Service names become **DNS hostnames**!

```yaml
services:
  api:
    environment:
      DB_HOST: postgres # ← Use service name
      DB_PORT: 5432
      REDIS_ADDR: redis:6379 # ← Use service name
```

Inside the `api` container:

- You can reach postgres at `postgres:5432`
- You can reach redis at `redis:6379`
- No need for `localhost` or IP addresses
- Docker Compose sets up networking automatically

**Real-World Example:**

```yaml
services:
  api:
    build: .
    ports:
      - "8080:8080"
    environment:
      DB_HOST: postgres # Magic! This resolves to the postgres container
```

---

### 2. **Networks** - The "How They Talk"

**Mental Model:** Networks are like **phone networks** connecting containers.

```
Without Network:
Container A: "Hello?"
Container B: "Who dis?" (can't hear each other)

With Network:
Container A: "postgres, you there?"
Postgres: "Yes! I'm at postgres:5432"
```

#### Default Behavior

Docker Compose creates a **default network** automatically:

```
Project: doit
Default Network: doit_default
All services join this network by default
```

**What this means:**

- All containers can talk to each other
- Using service names as hostnames
- No manual network configuration needed

#### Custom Networks (Advanced)

You can create multiple networks for isolation:

```yaml
networks:
  frontend: # Public-facing services
  backend: # Internal services only

services:
  api:
    networks:
      - frontend # Can access internet
      - backend # Can access database

  postgres:
    networks:
      - backend # NOT accessible from frontend
```

**Mental Model:** Like office floors in a building:

```
┌─────────────────────────────────┐
│  Frontend Network (Public Floor)│
│  ┌────────┐                     │
│  │  API   │                     │
│  └───┬────┘                     │
│      │                          │
│      │ Has keycard to both      │
└──────┼──────────────────────────┘
       │
┌──────┼──────────────────────────┐
│  Backend Network (Server Room)  │
│      │                          │
│  ┌───▼────┐    ┌──────────┐    │
│  │  API   │───→│ Postgres │    │
│  └────────┘    └──────────┘    │
│                                 │
│  (Postgres locked in here)      │
└─────────────────────────────────┘
```

**Benefits:**

- ✅ Security: Database not exposed to frontend
- ✅ Isolation: Services only see what they need
- ✅ Organization: Clear service boundaries

---

### 3. **Volumes** - The "Where Data Lives"

**Mental Model:** Volumes are like **external hard drives** that persist even when containers are deleted.

#### The Problem Without Volumes

```
Start Postgres Container
    ↓
Add data to database
    ↓
Stop container (docker-compose down)
    ↓
Start container again (docker-compose up)
    ↓
💥 All data is GONE! (containers are ephemeral)
```

**Why?** Containers are designed to be temporary. When removed, everything inside is deleted.

#### The Solution: Volumes

```
Start Postgres Container
    ↓
Attach volume: postgres_data
    ↓
Add data → Saved to volume (on your host machine)
    ↓
Stop container (docker-compose down)
    ↓
Start container again (docker-compose up)
    ↓
Attach same volume
    ↓
✅ Data is STILL THERE!
```

#### Types of Volumes

##### 1. **Named Volumes** (Docker manages)

```yaml
volumes:
  postgres_data: # Docker manages this

services:
  postgres:
    volumes:
      - postgres_data:/var/lib/postgresql/data
      #   ↑              ↑
      #   |              └─ Path inside container
      #   └──────────────── Volume name
```

**Mental Model:** Like a safe deposit box at a bank

- Docker manages it (you don't know exact location)
- It persists across container restarts
- Survives `docker-compose down`
- Only deleted with `docker-compose down -v` or `docker volume rm`

**Where is it stored?**

```bash
# Docker stores it somewhere like:
/var/lib/docker/volumes/doit_postgres_data/_data
```

You don't need to know the exact path!

##### 2. **Bind Mounts** (You manage)

```yaml
services:
  api:
    volumes:
      - ./api:/app
      #   ↑     ↑
      #   |     └─ Path inside container
      #   └─────── Path on YOUR laptop
```

**Mental Model:** Like a shared folder between your laptop and container

```
Your Laptop                Container
───────────                ─────────
./api/handler.go    ←──→   /app/handler.go
./api/service.go    ←──→   /app/service.go

Edit on laptop → Changes appear in container instantly!
```

**Perfect for development:**

- Edit code on your laptop with your favorite editor
- Changes appear instantly in container
- Hot reload works (if your app supports it)

**Example:**

```yaml
services:
  api:
    volumes:
      - ./:/app # Mount entire project
      - ./logs:/app/logs # Mount logs directory
      - ~/.ssh:/root/.ssh:ro # Mount SSH keys (read-only)
```

#### Volume Persistence Behavior

```bash
# Start services
docker-compose up

# Stop services (volumes remain)
docker-compose down

# Start again (data still there!)
docker-compose up

# Stop AND delete volumes
docker-compose down -v  # ⚠️ All data deleted!
```

**Mental Model:**

- `docker-compose down` = Turn off computers (data on external drives remains)
- `docker-compose down -v` = Turn off computers AND destroy external drives

---

### 4. **Environment Variables** - The "Configuration"

**Mental Model:** Environment variables are like **settings** or **configuration knobs**.

#### Three Ways to Set Them

##### 1. **Inline in docker-compose.yml**

```yaml
services:
  api:
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
      LOG_LEVEL: info
```

**Good for:**

- Non-sensitive values
- Values shared across environments
- Static configuration

**Bad for:**

- Secrets (passwords, API keys)
- Environment-specific values

##### 2. **From .env file**

```yaml
# .env file
DB_NAME=doit
DB_USER=doit
LOG_LEVEL=debug

# docker-compose.yml
services:
  api:
    env_file:
      - .env
```

**Good for:**

- Team-shared configs
- Non-sensitive defaults
- Can be checked into git (if no secrets)

**Pattern:**

```bash
.env.example     # Template (checked into git)
.env             # Actual values (gitignored)
```

##### 3. **Variable Substitution**

```yaml
services:
  api:
    environment:
      DB_PASSWORD: ${DB_PASSWORD} # Reads from shell or .env
      JWT_SECRET: ${JWT_SECRET}
```

**How it works:**

```bash
# Option 1: Export in shell
export DB_PASSWORD=secret123
docker-compose up

# Option 2: In .env file
DB_PASSWORD=secret123
docker-compose up

# Option 3: Inline
DB_PASSWORD=secret123 docker-compose up
```

**Good for:**

- Sensitive values (not in git)
- CI/CD pipelines (secrets from vault)
- Production deployments

#### Combining All Three

```yaml
services:
  api:
    env_file:
      - .env # Load defaults
    environment:
      DB_HOST: postgres # Override with static value
      DB_PASSWORD: ${DB_PASSWORD} # Override with variable
```

**Mental Model:** Three levels of configuration

```
Priority (lowest to highest):
1. env_file (.env)            → Defaults
2. environment (inline)       → Static overrides
3. environment (${VAR})       → Dynamic overrides
4. Shell environment          → Highest priority
```

---

### 5. **Depends_on** - The "Startup Order"

**Mental Model:** Like a **dependency graph** or **prerequisite courses**.

#### The Problem

```
All containers start simultaneously
    ↓
API starts immediately
    ↓
Tries to connect to postgres
    ↓
💥 Postgres isn't ready yet!
    ↓
API crashes with "connection refused"
```

#### The Solution: depends_on

```yaml
services:
  api:
    depends_on:
      - postgres
      - redis

  postgres:
    # No dependencies (starts first)

  redis:
    # No dependencies (starts first)
```

**What happens:**

```
1. Docker Compose reads dependency graph
2. Starts postgres and redis (in parallel)
3. Waits for them to start
4. Then starts api
```

**Mental Model:** Like making a cake

```
Can't add frosting (API)
until cake is baked (Database)
```

#### Basic depends_on Limitation

⚠️ **Important:** Basic `depends_on` only waits for containers to **start**, not be **ready**.

```
Postgres container started ✅
    ↓
API starts immediately
    ↓
But Postgres is still initializing database...
    ↓
💥 API still fails to connect!
```

#### Advanced: Health Checks

The solution is to use `condition: service_healthy`:

```yaml
services:
  api:
    depends_on:
      postgres:
        condition: service_healthy # Wait until READY
      redis:
        condition: service_started # Just wait for start

  postgres:
    healthcheck:
      test: ["CMD", "pg_isready", "-U", "doit"]
      interval: 5s # Check every 5 seconds
      timeout: 3s # Each check has 3 seconds
      retries: 5 # Try 5 times before giving up
      start_period: 10s # Grace period before checks start
```

**How it works:**

```
1. Postgres container starts
2. Docker waits 10 seconds (start_period)
3. Docker runs: pg_isready -U doit
4. If fails, wait 5 seconds (interval)
5. Try again (up to 5 retries)
6. If succeeds, mark as "healthy"
7. Now API can start
```

**Mental Model:**

```
Basic depends_on:
"Wait for oven to turn on" 🔥
    ↓
Still cold inside!

With health check:
"Wait for oven to reach 350°F" 🌡️
    ↓
Actually ready to bake!
```

#### Complete Example

```yaml
services:
  postgres:
    image: postgres:16
    healthcheck:
      test: ["CMD", "pg_isready", "-U", "${DB_USER}"]
      interval: 5s
      timeout: 3s
      retries: 5

  redis:
    image: redis:7
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5

  api:
    build: .
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    # API won't start until both are HEALTHY ✅
```

---

### 6. **Ports** - The "How You Access Services"

**Mental Model:** Ports are like **apartment numbers** in a building.

```
Your laptop = Building
Container = Apartment
Port = Apartment number
```

#### Port Mapping Syntax

```yaml
services:
  api:
    ports:
      - "8080:8080"
      #   ↑    ↑
      #   |    └─ Container's port (inside)
      #   └────── Your laptop's port (outside)
```

**What this means:**

```
You open browser: http://localhost:8080
    ↓
Request arrives at your laptop's port 8080
    ↓
Docker routes it to container's port 8080
    ↓
Your API receives the request
```

#### Different Port Mapping

```yaml
services:
  api:
    ports:
      - "3000:8080" # External:Internal
```

**What this means:**

```
Browser: http://localhost:3000
    ↓
Docker: Routes to container port 8080
    ↓
API: Listening on port 8080 inside container
```

**Mental Model:** Like a phone extension system

```
Call main number: localhost
Dial extension 3000 → Forwarded to internal line 8080
```

#### Example Multi-Service Setup

```yaml
services:
  api:
    ports:
      - "8080:8080" # API at localhost:8080

  postgres:
    ports:
      - "5432:5432" # DB at localhost:5432

  redis:
    ports:
      - "6379:6379" # Redis at localhost:6379

  grafana:
    ports:
      - "3000:3000" # Grafana at localhost:3000

  prometheus:
    ports:
      - "9090:9090" # Prometheus at localhost:9090
```

#### Internal vs External Access

**Key Concept:** Containers can talk to each other WITHOUT port mapping!

```yaml
services:
  api:
    ports:
      - "8080:8080" # Exposed to YOUR laptop
    environment:
      DB_HOST: postgres # Internal communication
      DB_PORT: 5432 # No mapping needed!

  postgres:
    # NO ports section!
    # Only accessible from other containers
    # NOT accessible from your laptop
```

**Visual Explanation:**

```
┌─────────────────────────────────────────────────┐
│  YOUR LAPTOP (Host)                             │
│                                                 │
│  Browser → localhost:8080 → API Container ✅   │
│  Browser → localhost:5432 → ❌ Not exposed     │
│                                                 │
│  ┌──────────────────────────────────────────┐  │
│  │  Docker Compose Network                  │  │
│  │                                          │  │
│  │  API Container                           │  │
│  │    - Reachable at: localhost:8080       │  │
│  │    - Can reach: postgres:5432 ✅        │  │
│  │    - Can reach: redis:6379 ✅           │  │
│  │                                          │  │
│  │  Postgres Container                      │  │
│  │    - Listening on: 5432 (internal)      │  │
│  │    - NOT exposed to host                 │  │
│  │    - Only containers can reach it ✅     │  │
│  │                                          │  │
│  └──────────────────────────────────────────┘  │
└─────────────────────────────────────────────────┘
```

**When to expose ports:**

- ✅ API/web servers (need browser access)
- ✅ Monitoring UIs (Grafana, Prometheus)
- ✅ Databases (for debugging with tools like pgAdmin)
- ❌ Internal services (they talk via service names)

---

## 🎭 Real-World Scenario: Your DoIt Application

Let's walk through what happens when you run `docker-compose up` with your complete stack.

### Your docker-compose.yml Structure

```yaml
version: "3.8"

services:
  # 1. Database
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: doit
      POSTGRES_USER: doit
      POSTGRES_PASSWORD: doit123
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD", "pg_isready", "-U", "doit"]
      interval: 5s
      timeout: 3s
      retries: 5

  # 2. Cache
  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5

  # 3. Your API
  api:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: doit
      DB_PASSWORD: doit123
      DB_NAME: doit
      REDIS_ADDR: redis:6379
    volumes:
      - ./:/app # Hot reload in development

  # 4. Metrics Collection
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    command:
      - "--config.file=/etc/prometheus/prometheus.yml"

  # 5. Dashboards
  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    depends_on:
      - prometheus
    volumes:
      - grafana_data:/var/lib/grafana
    environment:
      GF_SECURITY_ADMIN_PASSWORD: admin

volumes:
  postgres_data:
  redis_data:
  prometheus_data:
  grafana_data:

networks:
  default:
    name: doit_network
```

---

## 🔄 The Complete Lifecycle

### What Happens During `docker-compose up`

Let's trace **exactly** what happens, step by step:

#### **Phase 1: Planning** (Docker Compose reads the YAML)

```
Docker Compose starts
    ↓
Reads docker-compose.yml
    ↓
Parses all services
    ↓
Builds dependency graph:

    Level 0 (no dependencies):
    ├─ postgres
    ├─ redis
    └─ prometheus

    Level 1 (depends on Level 0):
    ├─ api (depends on postgres, redis)
    └─ grafana (depends on prometheus)
    ↓
Creates execution plan
```

**Mental Model:** Like a project manager creating a Gantt chart before starting work.

#### **Phase 2: Network Creation**

```
Check if network "doit_network" exists
    ↓
Network doesn't exist → Create it
    ↓
Network created: doit_network
    Subnet: 172.20.0.0/16
    Gateway: 172.20.0.1
    ↓
Ready to assign IP addresses to containers
```

**What this provides:**

- Internal DNS server
- IP address pool
- Isolated communication channel

#### **Phase 3: Volume Creation**

```
Check volumes:

postgres_data:
    ├─ Doesn't exist → Create it
    └─ Location: /var/lib/docker/volumes/doit_postgres_data

redis_data:
    ├─ Doesn't exist → Create it
    └─ Location: /var/lib/docker/volumes/doit_redis_data

prometheus_data:
    ├─ Doesn't exist → Create it
    └─ Location: /var/lib/docker/volumes/doit_prometheus_data

grafana_data:
    ├─ Doesn't exist → Create it
    └─ Location: /var/lib/docker/volumes/doit_grafana_data
```

**Mental Model:** Installing external hard drives before booting computers.

#### **Phase 4: Image Preparation**

```
For each service:

postgres:
    ├─ Image: postgres:16-alpine
    ├─ Check local cache → Not found
    ├─ Pull from Docker Hub
    └─ ✅ Image ready

redis:
    ├─ Image: redis:7-alpine
    ├─ Check local cache → Not found
    ├─ Pull from Docker Hub
    └─ ✅ Image ready

api:
    ├─ Build from Dockerfile
    ├─ Check for cached layers
    ├─ Use cached layers (from Phase 2.1!)
    ├─ Build only changed layers
    └─ ✅ Image ready

prometheus:
    ├─ Image: prom/prometheus:latest
    └─ ✅ Image ready

grafana:
    ├─ Image: grafana/grafana:latest
    └─ ✅ Image ready
```

#### **Phase 5: Container Creation & Startup**

**Group 1 starts** (no dependencies):

```
[postgres]
  ├─ Create container: doit-postgres-1
  ├─ Assign IP: 172.20.0.2
  ├─ Mount volume: postgres_data
  ├─ Inject environment variables
  ├─ Start container
  ├─ Initialize PostgreSQL...
  │   ├─ Creating database "doit"
  │   ├─ Creating user "doit"
  │   └─ Setting password
  ├─ Health check: Running pg_isready...
  ├─ Health check: ❌ Not ready yet
  ├─ Wait 5 seconds...
  ├─ Health check: ✅ Ready!
  └─ Status: HEALTHY ✅

[redis]
  ├─ Create container: doit-redis-1
  ├─ Assign IP: 172.20.0.3
  ├─ Mount volume: redis_data
  ├─ Start container
  ├─ Redis server starting...
  ├─ Health check: redis-cli ping
  ├─ Health check: ✅ PONG
  └─ Status: HEALTHY ✅

[prometheus]
  ├─ Create container: doit-prometheus-1
  ├─ Assign IP: 172.20.0.4
  ├─ Mount prometheus.yml
  ├─ Mount prometheus_data
  ├─ Start container
  ├─ Loading config...
  └─ Status: RUNNING ✅
```

**Group 2 starts** (after dependencies are healthy):

```
[api]
  ├─ Waiting for dependencies...
  │   ├─ postgres: ✅ HEALTHY
  │   └─ redis: ✅ HEALTHY
  ├─ All dependencies ready!
  ├─ Create container: doit-api-1
  ├─ Assign IP: 172.20.0.5
  ├─ Mount source code: ./:/app
  ├─ Inject environment variables:
  │   ├─ DB_HOST=postgres
  │   ├─ DB_PORT=5432
  │   ├─ DB_USER=doit
  │   ├─ DB_PASSWORD=doit123
  │   ├─ REDIS_ADDR=redis:6379
  │   └─ ... (more vars)
  ├─ Start container
  ├─ Application starting...
  │   ├─ Connecting to postgres:5432
  │   │   └─ DNS resolves to 172.20.0.2 ✅
  │   ├─ Connected to database ✅
  │   ├─ Running migrations...
  │   ├─ Migrations complete ✅
  │   ├─ Connecting to redis:6379
  │   │   └─ DNS resolves to 172.20.0.3 ✅
  │   ├─ Connected to Redis ✅
  │   └─ Server listening on :8080
  └─ Status: RUNNING ✅

[grafana]
  ├─ Waiting for dependencies...
  │   └─ prometheus: ✅ RUNNING
  ├─ Create container: doit-grafana-1
  ├─ Assign IP: 172.20.0.6
  ├─ Mount grafana_data
  ├─ Start container
  ├─ Initializing database...
  ├─ Starting web server...
  └─ Status: RUNNING ✅
```

#### **Phase 6: Service Discovery in Action**

Now your API container boots and connects:

```go
// Your Go code running inside the API container
dbHost := os.Getenv("DB_HOST")  // "postgres"
dbPort := os.Getenv("DB_PORT")  // "5432"

connString := fmt.Sprintf(
    "host=%s port=%s user=%s password=%s dbname=%s",
    dbHost,  // "postgres"
    dbPort,  // "5432"
    // ...
)

// What actually happens under the hood:
// 1. API tries to connect to "postgres:5432"
// 2. Docker's internal DNS server receives query
// 3. DNS returns: "postgres = 172.20.0.2"
// 4. API connects to 172.20.0.2:5432
// 5. Connection established! ✅
```

**Mental Model:** Like calling a colleague by name:

- You say: "Call John"
- Phone system translates: "John = extension 5432"
- Call connects automatically

#### **Phase 7: Final State**

```
All services running! ✅

Your laptop can access:
  ├─ API:        http://localhost:8080
  ├─ Grafana:    http://localhost:3000
  ├─ Prometheus: http://localhost:9090
  └─ Postgres:   localhost:5432 (if exposed)

Internal communication:
  ├─ api → postgres:5432 ✅
  ├─ api → redis:6379 ✅
  ├─ prometheus → api:8080/metrics ✅
  └─ grafana → prometheus:9090 ✅

Data persistence:
  ├─ postgres_data (database)
  ├─ redis_data (cache)
  ├─ prometheus_data (metrics)
  └─ grafana_data (dashboards)

Everything is connected and ready! 🎉
```

---

### Visual Representation: What's Running

```
┌──────────────────────────────────────────────────────────┐
│  YOUR LAPTOP: localhost                                  │
│                                                          │
│  Browser/curl Access:                                    │
│  ├─ localhost:8080  → API                               │
│  ├─ localhost:3000  → Grafana Dashboard                 │
│  ├─ localhost:9090  → Prometheus UI                     │
│  └─ localhost:5432  → PostgreSQL (direct access)        │
│                                                          │
│  ┌─────────────────────────────────────────────────┐   │
│  │  Docker Network: doit_network (172.20.0.0/16)   │   │
│  │                                                 │   │
│  │  ┌──────────┐                                   │   │
│  │  │   API    │                                   │   │
│  │  │ :8080    │                                   │   │
│  │  │172.20.0.5│                                   │   │
│  │  └────┬─────┘                                   │   │
│  │       │                                         │   │
│  │       ├──────────→ postgres:5432               │   │
│  │       │            (DNS → 172.20.0.2)           │   │
│  │       │                                         │   │
│  │       ├──────────→ redis:6379                  │   │
│  │       │            (DNS → 172.20.0.3)           │   │
│  │       │                                         │   │
│  │       └──────────→ Exposes /metrics             │   │
│  │                                                 │   │
│  │  ┌──────────┐    ┌──────────┐                 │   │
│  │  │ Postgres │    │  Redis   │                 │   │
│  │  │  :5432   │    │  :6379   │                 │   │
│  │  │172.20.0.2│    │172.20.0.3│                 │   │
│  │  └────┬─────┘    └──────────┘                 │   │
│  │       │                                         │   │
│  │       ▼                                         │   │
│  │  [postgres_data] [redis_data]                  │   │
│  │                                                 │   │
│  │  ┌──────────┐      ┌─────────┐               │   │
│  │  │Prometheus│◄─────│ Grafana │               │   │
│  │  │  :9090   │      │  :3000  │               │   │
│  │  │172.20.0.4│      │172.20.0.6│               │   │
│  │  └────┬─────┘      └────┬────┘               │   │
│  │       │                 │                      │   │
│  │       ▼                 ▼                      │   │
│  │  [prometheus_data] [grafana_data]             │   │
│  │                                                │   │
│  └─────────────────────────────────────────────────┘   │
│                                                          │
│  Persistent Volumes (survive restarts):                 │
│  ├─ postgres_data   → Database files                    │
│  ├─ redis_data      → Cache snapshots                   │
│  ├─ prometheus_data → Time-series metrics               │
│  └─ grafana_data    → Dashboard configs                 │
└──────────────────────────────────────────────────────────┘
```

---

## 🔄 Common Operations (Mental Models)

### 1. **`docker-compose up`**

**Mental Model:** "Start the entire band practice"

```bash
docker-compose up

# What happens:
1. Reads sheet music (docker-compose.yml)
2. Ensures all musicians are present (pull/build images)
3. Assigns practice room (network)
4. Sets up equipment (volumes)
5. Musicians take their places (create containers)
6. Starts playing in order (respects depends_on)
7. You hear the music (logs stream to console)
```

**Options:**

```bash
# Start in background (detached)
docker-compose up -d

# Start and rebuild images
docker-compose up --build

# Start specific services only
docker-compose up api postgres

# Start with verbose output
docker-compose up --verbose
```

---

### 2. **`docker-compose down`**

**Mental Model:** "End practice, pack up (but keep instruments safe)"

```bash
docker-compose down

# What happens:
1. Stop all containers gracefully (SIGTERM → SIGKILL after timeout)
2. Remove containers
3. Remove network
4. Keep volumes (data persists!)
```

**Options:**

```bash
# Remove volumes too (⚠️ deletes all data!)
docker-compose down -v

# Remove images too
docker-compose down --rmi all

# Just stop, don't remove
docker-compose stop
```

**Mental Model:**

- `down` = End practice, clean up the room
- `down -v` = End practice, AND throw away all equipment
- `stop` = End practice, leave everything set up

---

### 3. **`docker-compose logs`**

**Mental Model:** "Listen to all conversations"

```bash
# View all logs
docker-compose logs

# Follow logs in real-time
docker-compose logs -f

# Logs from specific service
docker-compose logs api

# Follow logs from specific service
docker-compose logs -f api

# Last 100 lines
docker-compose logs --tail=100

# Logs with timestamps
docker-compose logs -t
```

**What you see:**

```
api_1       | 2024-01-15 10:30:00 INFO  Server starting on :8080
postgres_1  | 2024-01-15 10:30:01 LOG   database system is ready
redis_1     | 2024-01-15 10:30:01 Ready to accept connections
api_1       | 2024-01-15 10:30:02 INFO  Connected to database
```

**Mental Model:** Like a conference call where everyone's talking and you hear all conversations at once (color-coded by speaker).

---

### 4. **`docker-compose exec`**

**Mental Model:** "Walk into a container's room and talk directly"

```bash
# Get a shell in the API container
docker-compose exec api sh

# Run a one-off command
docker-compose exec api ls -la

# Connect to database
docker-compose exec postgres psql -U doit

# Check Redis
docker-compose exec redis redis-cli ping
```

**Difference from `docker exec`:**

```bash
# docker-compose: Use service name
docker-compose exec api sh

# docker: Use container ID/name
docker exec -it doit-api-1 sh
```

**Mental Model:** Like SSH-ing into a server, but it's a container.

---

### 5. **`docker-compose ps`**

**Mental Model:** "Check who's in the office"

```bash
docker-compose ps

# Output:
Name                   State    Ports
─────────────────────────────────────────
doit-api-1           Up     0.0.0.0:8080->8080/tcp
doit-postgres-1      Up     5432/tcp
doit-redis-1         Up     6379/tcp
doit-grafana-1       Up     0.0.0.0:3000->3000/tcp
doit-prometheus-1    Up     0.0.0.0:9090->9090/tcp
```

---

### 6. **`docker-compose restart`**

**Mental Model:** "Just restart specific service, keep others running"

```bash
# Restart one service
docker-compose restart api

# Restart multiple services
docker-compose restart api redis

# Restart everything
docker-compose restart
```

**When to use:**

- After changing environment variables
- After code changes (if not using hot reload)
- To test startup behavior

---

### 7. **`docker-compose pull`**

**Mental Model:** "Update all software to latest versions"

```bash
# Pull latest images
docker-compose pull

# Pull specific service
docker-compose pull postgres
```

**When to use:**

- Before starting to ensure latest versions
- After updating image tags in docker-compose.yml

---

### 8. **`docker-compose build`**

**Mental Model:** "Rebuild your custom images"

```bash
# Build all services with 'build:' directive
docker-compose build

# Build specific service
docker-compose build api

# Build without cache
docker-compose build --no-cache api
```

---

## 🎨 Design Patterns & Best Practices

### Pattern 1: Development vs Production

**Mental Model:** Like having a **demo car** and a **race car** - same base, different tuning.

#### File Structure:

```
docker-compose.yml           # Base (shared config)
docker-compose.override.yml  # Dev overrides (auto-loaded)
docker-compose.prod.yml      # Production overrides (explicit)
```

#### Base Configuration (docker-compose.yml)

```yaml
version: "3.8"

services:
  api:
    build: .
    ports:
      - "8080:8080"
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
    depends_on:
      - postgres

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: doit
```

#### Development Overrides (docker-compose.override.yml)

```yaml
version: "3.8"

services:
  api:
    volumes:
      - ./:/app # Mount source for hot reload
    environment:
      LOG_LEVEL: debug # Verbose logging
      HOT_RELOAD: "true"
    command: air # Hot reload tool

  postgres:
    ports:
      - "5432:5432" # Expose for debugging
    volumes:
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql
```

#### Production Configuration (docker-compose.prod.yml)

```yaml
version: "3.8"

services:
  api:
    image: registry.example.com/doit-api:${VERSION}
    environment:
      LOG_LEVEL: info
    deploy:
      replicas: 3
      resources:
        limits:
          cpus: "1.0"
          memory: 512M

  postgres:
    environment:
      POSTGRES_PASSWORD: ${DB_PASSWORD} # From secrets
    volumes:
      - /mnt/data/postgres:/var/lib/postgresql/data
```

#### Usage:

```bash
# Development (auto-loads override)
docker-compose up

# Production (explicit)
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up
```

**Mental Model:** Layered configuration like Photoshop layers:

```
Base layer (docker-compose.yml)
    ↓
+ Dev overrides (auto)
    ↓
Final dev configuration

OR

Base layer (docker-compose.yml)
    ↓
+ Prod overrides (explicit)
    ↓
Final prod configuration
```

---

### Pattern 2: Init Containers / Setup Scripts

**Mental Model:** "Set up the stage before the show"

#### Database Initialization

```yaml
services:
  postgres:
    image: postgres:16-alpine
    volumes:
      # Any SQL files here run on first start
      - ./init-scripts/001_create_tables.sql:/docker-entrypoint-initdb.d/001_create_tables.sql
      - ./init-scripts/002_seed_data.sql:/docker-entrypoint-initdb.d/002_seed_data.sql
    environment:
      POSTGRES_DB: doit
```

**What happens:**

```
1. Postgres starts for first time
2. Checks /docker-entrypoint-initdb.d/
3. Runs all .sql and .sh files in order (001, 002, etc.)
4. Marks database as initialized
5. On subsequent starts, skips initialization
```

#### Application Migrations

```yaml
services:
  api:
    depends_on:
      postgres:
        condition: service_healthy
    command: >
      sh -c "
        echo 'Waiting for database...'
        sleep 5
        echo 'Running migrations...'
        migrate -path /migrations -database $${DB_URL} up
        echo 'Starting API...'
        ./app
      "
```

**Mental Model:** Like a checklist before takeoff:

```
☑ Database ready
☑ Migrations run
☑ Start application
```

---

### Pattern 3: Shared Configuration with .env

**Mental Model:** "Everyone reads from the same playbook"

#### .env.example (Template - checked into git)

```bash
# Database
DB_NAME=doit
DB_USER=doit
DB_PASSWORD=changeme

# Redis
REDIS_PASSWORD=

# Application
LOG_LEVEL=info
JWT_SECRET=your-secret-here

# Monitoring
GRAFANA_ADMIN_PASSWORD=admin
```

#### .env (Actual values - gitignored)

```bash
# Database
DB_NAME=doit
DB_USER=doit
DB_PASSWORD=super-secret-password

# Redis
REDIS_PASSWORD=another-secret

# Application
LOG_LEVEL=debug
JWT_SECRET=actual-jwt-secret

# Monitoring
GRAFANA_ADMIN_PASSWORD=actual-admin-pass
```

#### docker-compose.yml

```yaml
services:
  api:
    env_file:
      - .env
    environment:
      # Can override specific values
      APP_NAME: doit-api

  postgres:
    env_file:
      - .env
    environment:
      POSTGRES_DB: ${DB_NAME}
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
```

#### .gitignore

```
.env
```

**Workflow for new developer:**

```bash
# 1. Clone repo
git clone ...

# 2. Copy template
cp .env.example .env

# 3. Edit with real values
nano .env

# 4. Start services
docker-compose up
```

---

### Pattern 4: Health Checks for All Services

**Mental Model:** "Make sure everyone's actually ready before starting the show"

```yaml
services:
  postgres:
    healthcheck:
      test: ["CMD", "pg_isready", "-U", "${DB_USER}"]
      interval: 5s
      timeout: 3s
      retries: 5
      start_period: 10s

  redis:
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5

  api:
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:8080/health"]
      interval: 30s
      timeout: 3s
      retries: 3
      start_period: 40s
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
```

**Why this matters:**

- Prevents "connection refused" errors
- Ensures stable startups
- Makes debugging easier
- Prepares you for Kubernetes (same pattern!)

---

### Pattern 5: Service Profiles (Optional Services)

**Mental Model:** "Choose which musicians to include in the performance"

```yaml
services:
  api:
    # Always included

  postgres:
    # Always included

  adminer:
    image: adminer
    profiles:
      - debug # Only start with --profile debug
    ports:
      - "8081:8080"

  jaeger:
    image: jaegertracing/all-in-one
    profiles:
      - tracing # Only start with --profile tracing
    ports:
      - "16686:16686"
```

**Usage:**

```bash
# Start only core services
docker-compose up

# Start with debugging tools
docker-compose --profile debug up

# Start with tracing
docker-compose --profile tracing up

# Start with both
docker-compose --profile debug --profile tracing up
```

---

## 🔗 How Docker Compose Relates to What You Already Know

### From Phase 2.1 (Docker) → Phase 2.2 (Docker Compose)

```
Phase 2.1: Single Dockerfile
    ├─ How to build ONE container
    ├─ Image optimization
    ├─ Security hardening
    └─ Multi-stage builds

Phase 2.2: Docker Compose
    ├─ How to run MANY containers
    ├─ How they communicate
    ├─ How to manage them together
    └─ Local development workflow
```

**Analogy:**

- Phase 2.1: Learning to build a car (single unit)
- Phase 2.2: Learning to coordinate a fleet of cars (system)

### From Docker Compose → Kubernetes (Phase 5)

**Everything translates!** You're learning the concepts now that you'll use in production.

| Docker Compose Concept | Kubernetes Equivalent           | Same Idea?               |
| ---------------------- | ------------------------------- | ------------------------ |
| `services:`            | `Deployments`                   | ✅ Define what runs      |
| `image:` or `build:`   | `spec.template.spec.containers` | ✅ What container to run |
| `ports:`               | `Service` (type: LoadBalancer)  | ✅ Expose to outside     |
| `networks:`            | `Service` (ClusterIP)           | ✅ Internal networking   |
| `volumes:`             | `PersistentVolumeClaims`        | ✅ Persistent storage    |
| `depends_on:`          | `InitContainers`                | ✅ Startup order         |
| `healthcheck:`         | `livenessProbe/readinessProbe`  | ✅ Health monitoring     |
| `environment:`         | `ConfigMaps/Secrets`            | ✅ Configuration         |
| `deploy.replicas:`     | `spec.replicas`                 | ✅ Scaling               |

**Mental Model:** Docker Compose is **Kubernetes training wheels**

- Same concepts, simpler syntax
- Runs on your laptop
- Perfect for learning
- When you learn K8s, you'll say "Oh! This is just like docker-compose!"

---

## 🎯 Your DoIt Stack - What We'll Build

### Complete docker-compose.yml

```yaml
version: "3.8"

services:
  # ============================================
  # Your Go API
  # ============================================
  api:
    build:
      context: .
      dockerfile: infra/docker/dockerfile.service
    container_name: doit-api
    ports:
      - "8080:8080"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    environment:
      # Database
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: ${DB_USER:-doit}
      DB_PASSWORD: ${DB_PASSWORD:-doit123}
      DB_NAME: ${DB_NAME:-doit}

      # Redis
      REDIS_ADDR: redis:6379

      # Application
      APP_ENVIRONMENT: development
      LOG_LEVEL: debug
      JWT_SECRET: ${JWT_SECRET:-dev-secret}
    volumes:
      - ./:/app # Mount for hot reload
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:8080/health"]
      interval: 30s
      timeout: 3s
      retries: 3
    networks:
      - doit_network

  # ============================================
  # PostgreSQL Database
  # ============================================
  postgres:
    image: postgres:16-alpine
    container_name: doit-postgres
    ports:
      - "5432:5432"
    environment:
      POSTGRES_DB: ${DB_NAME:-doit}
      POSTGRES_USER: ${DB_USER:-doit}
      POSTGRES_PASSWORD: ${DB_PASSWORD:-doit123}
      POSTGRES_INITDB_ARGS: "-E UTF8"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./internal/data/migrations:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD", "pg_isready", "-U", "${DB_USER:-doit}"]
      interval: 5s
      timeout: 3s
      retries: 5
      start_period: 10s
    networks:
      - doit_network

  # ============================================
  # Redis Cache
  # ============================================
  redis:
    image: redis:7-alpine
    container_name: doit-redis
    ports:
      - "6379:6379"
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5
    networks:
      - doit_network

  # ============================================
  # Prometheus (Metrics)
  # ============================================
  prometheus:
    image: prom/prometheus:latest
    container_name: doit-prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./infra/docker/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    command:
      - "--config.file=/etc/prometheus/prometheus.yml"
      - "--storage.tsdb.path=/prometheus"
      - "--web.console.libraries=/usr/share/prometheus/console_libraries"
      - "--web.console.templates=/usr/share/prometheus/consoles"
    networks:
      - doit_network

  # ============================================
  # Grafana (Dashboards)
  # ============================================
  grafana:
    image: grafana/grafana:latest
    container_name: doit-grafana
    ports:
      - "3000:3000"
    depends_on:
      - prometheus
    environment:
      GF_SECURITY_ADMIN_USER: admin
      GF_SECURITY_ADMIN_PASSWORD: ${GRAFANA_PASSWORD:-admin}
      GF_USERS_ALLOW_SIGN_UP: "false"
    volumes:
      - grafana_data:/var/lib/grafana
      - ./infra/docker/grafana/provisioning:/etc/grafana/provisioning
    networks:
      - doit_network

  # ============================================
  # Adminer (Database UI)
  # ============================================
  adminer:
    image: adminer:latest
    container_name: doit-adminer
    ports:
      - "8081:8080"
    depends_on:
      - postgres
    environment:
      ADMINER_DEFAULT_SERVER: postgres
    networks:
      - doit_network

# ============================================
# Volumes (Persistent Data)
# ============================================
volumes:
  postgres_data:
    name: doit_postgres_data
  redis_data:
    name: doit_redis_data
  prometheus_data:
    name: doit_prometheus_data
  grafana_data:
    name: doit_grafana_data

# ============================================
# Networks
# ============================================
networks:
  doit_network:
    name: doit_network
    driver: bridge
```

### One Command to Rule Them All

```bash
docker-compose up
```

### What You Get

```
✅ API running on http://localhost:8080
✅ PostgreSQL on localhost:5432
✅ Redis on localhost:6379
✅ Prometheus on http://localhost:9090
✅ Grafana on http://localhost:3000
✅ Adminer on http://localhost:8081

All connected, all working, all monitored! 🎉
```

---

## ✅ Validation: Do You Understand?

### Self-Check Questions

Ask yourself these questions. If you can answer them confidently, you have a solid mental model:

1. **Services**

   - ✅ What is a "service" in Docker Compose?
   - ✅ How do services find each other?
   - ✅ Can I change a service name? What happens?

2. **Networks**

   - ✅ Do I need to manually configure networking?
   - ✅ How does `api` connect to `postgres`?
   - ✅ What's the difference between internal and external access?

3. **Volumes**

   - ✅ What happens to data when I run `docker-compose down`?
   - ✅ When should I use named volumes vs bind mounts?
   - ✅ How do I delete all data?

4. **Dependencies**

   - ✅ What's the difference between `depends_on` and health checks?
   - ✅ Why do services sometimes fail to connect even with `depends_on`?
   - ✅ How do I ensure a service is truly ready?

5. **Lifecycle**

   - ✅ What happens during `docker-compose up`?
   - ✅ In what order do containers start?
   - ✅ How does service discovery work?

6. **Configuration**
   - ✅ What are the three ways to set environment variables?
   - ✅ Which takes precedence?
   - ✅ Where should I store secrets?

### If You Can Explain...

- ✅ How containers communicate with each other
- ✅ Why data persists even after `docker-compose down`
- ✅ What happens during the complete `docker-compose up` lifecycle
- ✅ The difference between internal service names and external ports
- ✅ How to debug a service that fails to start

**Then you're ready to build!** 🎉

---

## 🚀 Ready for Implementation?

Now that you have a complete mental model of Docker Compose:

### You Understand:

- ✅ **What** it is (orchestrator for multiple containers)
- ✅ **Why** we use it (local development parity with production)
- ✅ **How** it works (services, networks, volumes, dependencies)
- ✅ **When** things happen (lifecycle, startup order)
- ✅ **Where** it fits (between single Docker and Kubernetes)

### Next Steps:

1. Create `docker-compose.yml` for your DoIt stack
2. Create `prometheus.yml` configuration
3. Create Grafana dashboard provisioning
4. Create `.env.example` template
5. Add Makefile targets for convenience
6. Test everything with `docker-compose up`

**Ready to start implementing Phase 2.2?** 🚀

---

## 📚 Additional Resources

### Official Documentation

- [Docker Compose Overview](https://docs.docker.com/compose/)
- [Compose File Reference](https://docs.docker.com/compose/compose-file/)
- [Networking in Compose](https://docs.docker.com/compose/networking/)

### Best Practices

- [Compose Best Practices](https://docs.docker.com/compose/production/)
- [12 Factor App](https://12factor.net/) (philosophy behind compose patterns)

### Tools

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) (includes Compose)
- [lazydocker](https://github.com/jesseduffield/lazydocker) (TUI for Docker)
- [ctop](https://github.com/bcicen/ctop) (container metrics)

---

**Document Version:** 1.0  
**Created:** 2025-11-15  
**Status:** ✅ Complete Reference Guide  
**Next Phase:** Implementation of docker-compose.yml
