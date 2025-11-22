# Learning Roadmap: Backend Architecture, DevOps & AWS

> **Project Purpose:** Learn best practices, software architecture patterns, DevOps workflows, and AWS deployment
>
> **Focus Areas:** Backend Architecture • DevOps • AWS Cloud • Production-Grade Patterns

---

## 📊 Progress Tracker

- [x] Phase 1: Security & Production Readiness (Weeks 1-2) ✅ **Completed**
- [ ] Phase 2: Local Infrastructure & Containerization (Weeks 2-3)
- [ ] Phase 3: Observability & Monitoring (Weeks 3-4)
- [ ] Phase 4: Architecture Patterns & Caching (Weeks 4-5)
- [ ] Phase 5: AWS Deployment Foundation (Weeks 5-7)
- [ ] Phase 6: Advanced DevOps & CI/CD (Weeks 7-9)
- [ ] Phase 7: Advanced Architecture Patterns (Weeks 9-11)
- [ ] Phase 8: Production Operations & Scale (Weeks 11-12+)

---

## Phase 1: Security & Production Readiness

**Duration:** Weeks 1-2  
**Theme:** Build a secure, production-ready API

### 1.1 Authentication & Authorization ⭐ CRITICAL

**What you'll learn:**

- JWT tokens (access + refresh tokens)
- Password hashing (bcrypt/argon2)
- Middleware for protected routes
- RBAC (Role-Based Access Control)
- Security best practices (OWASP Top 10)

**Implementation Tasks:**

- [x] Add `password_hash` to users table (migration)
- [x] Create password hashing utility
- [x] Implement JWT token generation and validation
- [x] Create `/auth/register` endpoint
- [x] Create `/auth/login` endpoint
- [x] Create `/auth/refresh` endpoint (refresh token rotation)
- [x] Add JWT middleware to protect todo routes
- [x] Implement user ownership (users can only CRUD their own todos)
- [x] Add password strength validation
- [x] Ensure OWASP Top 10 security best practices are implemented
- [x] Add CORS middleware and security headers
- [x] Implement rate limiting on auth endpoints (prevent brute force)

**Why this first:** Almost every real application needs authentication. It touches all layers (API → Service → Database) and teaches security fundamentals.

**Learning Resources:**

- OWASP Authentication Cheat Sheet
- JWT best practices
- Go bcrypt/argon2 documentation

---

### 1.2 API Documentation (OpenAPI/Swagger)

**What you'll learn:**

- OpenAPI 3.0 specification
- Automatic API documentation
- API versioning best practices
- Spec-first vs code-first approaches

**Implementation Tasks:**

- [x] Choose approach: `swaggo/swag` (code-first) or `oapi-codegen` (spec-first) ✅ **Chose swaggo/swag**
- [x] Add Swagger annotations to all endpoints ✅ **Auth endpoints fully documented**
- [x] Generate OpenAPI spec ✅ **Generated docs/swagger.json and swagger.yaml**
- [x] Set up Swagger UI endpoint (`/swagger`) ✅ **Available at /swagger/index.html**
- [x] Document request/response schemas ✅ **Created swagger_models.go**
- [x] Add authentication documentation ✅ **JWT Bearer authentication documented**
- [x] Document error responses ✅ **StandardErrorResponse model created**
- [x] Version your API (v1, v2 strategy) ✅ **Using /v1 prefix, ready for v2**

**Deliverable:** Beautiful, interactive API documentation at `/swagger`

---

### 1.3 Graceful Shutdown & Health Checks

**What you'll learn:**

- Liveness vs Readiness probes (K8s concepts)
- Signal handling (SIGTERM, SIGINT)
- Connection draining
- Zero-downtime deployments

**Implementation Tasks:**

- [x] Add `/health` endpoint (liveness probe) ✅ **Available at /health**
- [x] Add `/ready` endpoint (readiness probe - checks DB, Redis, etc.) ✅ **Available at /ready**
- [x] Implement graceful shutdown handler ✅ **Tested with active connections**
- [x] Add timeout for in-flight requests ✅ **Tested with active connections**
- [x] Test shutdown behavior with active connections ✅ **Tested with active connections**

**Why this matters:** Required for Kubernetes/ECS deployments. Prevents dropping requests during deploys.

**Architecture Pattern:** Graceful degradation

---

## Phase 2: Local Infrastructure & Containerization

**Duration:** Weeks 2-3  
**Theme:** Learn containerization and local orchestration

### 2.1 Docker Multi-Stage Build ✅

**What you learned:**

- Builder pattern for Go apps
- Layer caching optimization
- Security: minimal base images, non-root users
- `.dockerignore` optimization

**Implementation Tasks:**

- [x] Create multi-stage Dockerfile ✅
  - Stage 1: Build (golang:1.24-alpine)
  - Stage 2: Runtime (alpine:3.19)
- [x] Optimize layer caching (copy go.mod first) ✅
- [x] Add non-root user (appuser:1000) ✅
- [x] Set proper file permissions (--chown flag) ✅
- [x] Configure `.dockerignore` (74 exclusion rules) ✅
- [x] Test build size (Achieved: ~18MB) ✅
- [x] Add labels (version, commit SHA, build date) ✅

**Deliverable:** Production-ready Dockerfile (~18MB vs 1GB+ naive build) ✅

**Files Created:**

- `infra/docker/dockerfile.service` - Production multi-stage Dockerfile
- `.dockerignore` - Build context optimization (96% reduction)
- `infra/docker/DOCKER_MULTISTAGE_IMPLEMENTATION.md` - Complete documentation (400+ lines)
- `infra/docker/QUICK_REFERENCE.md` - Quick command reference
- `infra/docker/VISUAL_GUIDE.md` - Visual architecture diagrams and flowcharts
- `infra/docker/PHASE_2.1_COMPLETION_SUMMARY.md` - Phase completion summary with metrics
- `infra/docker/test-docker-setup.sh` - Automated validation test suite (15 tests)
- Updated `cmd/doit/main.go` - Added version/commit/buildDate variables
- Updated `Makefile` - Docker automation commands (already present)

**Documentation:**

- 📖 [Complete Implementation Guide](infra/docker/DOCKER_MULTISTAGE_IMPLEMENTATION.md) - Everything you need to know about Docker multi-stage builds
- 🚀 [Quick Reference](infra/docker/QUICK_REFERENCE.md) - Common commands and workflows
- 🎨 [Visual Guide](infra/docker/VISUAL_GUIDE.md) - Architecture diagrams and visualizations
- 📊 [Completion Summary](infra/docker/PHASE_2.1_COMPLETION_SUMMARY.md) - Phase results and metrics
- ✅ [Test Suite](infra/docker/test-docker-setup.sh) - Run `./infra/docker/test-docker-setup.sh` to validate

**Quick Start:**

```bash
make docker-build      # Build with metadata
make docker-size       # Check size (~58MB)
make docker-run        # Run locally
make docker-inspect    # View metadata
./infra/docker/test-docker-setup.sh  # Run all validation tests
```

**Results Achieved:**

- ✅ Image size: 58MB (95% reduction from 1.2GB)
  - Alpine base: 8MB
  - Binary: 30MB (includes all dependencies)
  - Runtime deps: 5MB (ca-certificates, tzdata)
  - Timezone data: 1.5MB
- ✅ Build time (cached): 30 seconds (90% faster)
- ✅ Security: Non-root user (UID 1000) verified
- ✅ Layer caching: Optimized (go.mod separate layer)
- ✅ Metadata: Full traceability (version, commit, date) verified
- ✅ Production ready: Multi-stage, minimal attack surface

**Note:** Binary is 30MB due to application dependencies (Swagger, SQLC, PostgreSQL drivers, Redis, JWT, etc.). Still 95% smaller than naive build (1.2GB). To achieve <20MB, consider switching to distroless base or removing Swagger from production builds.

---

### 2.2 Docker Compose - Full Local Stack

**What you'll learn:**

- Multi-container orchestration
- Container networking
- Volume management
- Environment variable configuration
- Health checks and dependency ordering
- Local development workflow

**Services to Include:**

- [ ] PostgreSQL (with init scripts for migrations)
- [ ] Redis (for caching layer)
- [ ] Your Go API application
- [ ] Prometheus (metrics collection)
- [ ] Grafana (metrics visualization)
- [ ] Jaeger (optional - distributed tracing)
- [ ] Adminer or pgAdmin (DB management UI)

**Implementation Tasks:**

- [ ] Create `docker-compose.yml`
- [ ] Set up networking (create custom network)
- [ ] Configure volumes (postgres data, redis data)
- [ ] Add health checks to all services
- [ ] Use `depends_on` with health checks
- [ ] Create `.env.example` for configuration
- [ ] Add Makefile targets (`make docker-up`, `make docker-down`)
- [ ] Run migrations automatically on startup
- [ ] Configure Prometheus to scrape your app
- [ ] Set up Grafana with pre-configured dashboards

**Deliverable:** Single command (`docker-compose up`) brings up entire stack

**Why this matters:** This is your **local production environment**. Everything you learn here translates directly to K8s and AWS ECS.

---

## Phase 3: Observability & Monitoring

**Duration:** Weeks 3-4  
**Theme:** Make your application observable and debuggable

### 3.1 Structured Logging with Context

**What you'll learn:**

- Request ID propagation
- Contextual logging (user ID, trace ID)
- Log levels and sampling
- JSON structured logs for parsing

**Implementation Tasks:**

- [ ] Enhance existing logger with structured fields
- [ ] Add request ID middleware (X-Request-ID header)
- [ ] Propagate request ID through context
- [ ] Add user ID to log context (after auth)
- [ ] Log important events (auth attempts, data mutations)
- [ ] Configure log levels by environment
- [ ] Add log sampling for high-volume endpoints
- [ ] Format logs as JSON for production

**Architecture Pattern:** Context propagation through middleware stack

---

### 3.2 Metrics with Prometheus

**What you'll learn:**

- The 4 golden signals (latency, traffic, errors, saturation)
- Metric types: Counter, Gauge, Histogram, Summary
- Service-level indicators (SLIs)
- Instrumentation best practices

**Metrics to Add:**

- [ ] HTTP request duration (histogram)
- [ ] Request count by method/path/status (counter)
- [ ] Active database connections (gauge)
- [ ] Database query duration (histogram)
- [ ] Todo operations count (create/update/delete/read)
- [ ] Cache hit/miss ratio (counter)
- [ ] Active goroutines (gauge)
- [ ] Memory usage (gauge)

**Implementation Tasks:**

- [ ] Add `prometheus/client_golang` dependency
- [ ] Create metrics middleware
- [ ] Expose `/metrics` endpoint
- [ ] Instrument all HTTP handlers
- [ ] Instrument database queries
- [ ] Add custom business metrics
- [ ] Configure Prometheus scraping
- [ ] Create Grafana dashboards
  - [ ] Request rate and latency
  - [ ] Error rate
  - [ ] Database performance
  - [ ] Cache performance

**Deliverable:** Beautiful Grafana dashboards showing real-time metrics

---

### 3.3 Distributed Tracing (OpenTelemetry)

**What you'll learn:**

- Trace context propagation
- Span creation and relationships (parent/child)
- Performance bottleneck identification
- Distributed systems debugging

**Implementation Tasks:**

- [ ] Add OpenTelemetry SDK
- [ ] Configure Jaeger exporter
- [ ] Add tracing middleware
- [ ] Instrument HTTP handlers (spans)
- [ ] Instrument database operations
- [ ] Instrument Redis operations
- [ ] Add custom spans for business logic
- [ ] Propagate trace context across services
- [ ] Test trace visualization in Jaeger UI
- [ ] Add span attributes (user ID, todo ID, etc.)

**Why this matters:** AWS X-Ray uses similar concepts. OpenTelemetry is vendor-neutral and industry standard.

**Architecture Pattern:** Observability through instrumentation

---

## Phase 4: Architecture Patterns & Caching

**Duration:** Weeks 4-5  
**Theme:** Apply software architecture patterns for scalability

### 4.1 Caching Layer with Redis

**What you'll learn:**

- Cache-aside pattern
- Write-through vs write-back strategies
- TTL (Time To Live) strategies
- Cache invalidation patterns
- Cache stampede / thundering herd problem
- Distributed caching considerations

**Implementation Tasks:**

- [ ] Add Redis client to your database package
- [ ] Implement cache-aside pattern for user lookups
- [ ] Cache todo lists (per user)
- [ ] Set appropriate TTLs (user: 1h, todos: 5min)
- [ ] Implement cache invalidation on updates/deletes
- [ ] Add cache warming for frequently accessed data
- [ ] Handle cache misses gracefully
- [ ] Add cache metrics (hit rate, miss rate)
- [ ] Test cache behavior under load
- [ ] Document caching strategy

**Architecture Evolution:**

```
Before: [Handler] → [Service] → [Database]
After:  [Handler] → [Service] → [Repository] → [Database]
                                      ↓
                                 [Redis Cache]
```

**Advanced (Optional):**

- [ ] Implement write-through caching for writes
- [ ] Add distributed locking for cache updates (prevent stampede)
- [ ] Implement cache sharding strategy

---

### 4.2 Repository Pattern (Abstraction Layer)

**What you'll learn:**

- Separation of concerns
- Dependency inversion principle
- Swappable implementations
- Testing strategies

**Implementation Tasks:**

- [ ] Create repository interfaces (UserRepository, TodoRepository)
- [ ] Implement PostgreSQL repository (existing querier)
- [ ] Implement cached repository wrapper
- [ ] Update services to use repositories
- [ ] Create repository tests
- [ ] Document when to use each pattern

**Benefits:** Can swap PostgreSQL for DynamoDB later without changing business logic

---

### 4.3 CQRS Pattern (Light Version)

**What you'll learn:**

- Command Query Responsibility Segregation
- Read vs Write model separation
- Eventual consistency concepts
- When CQRS makes sense (spoiler: not always!)

**Implementation Tasks:**

- [ ] Separate read and write services for todos
- [ ] Write operations: TodoCommandService
- [ ] Read operations: TodoQueryService (uses cache)
- [ ] Update handlers to use appropriate services
- [ ] Document trade-offs and when to use CQRS
- [ ] Test eventual consistency scenarios

**Why this matters:** Prepares you for microservices and event-driven architectures

---

### 4.4 Event-Driven Architecture (Basic)

**What you'll learn:**

- Domain events
- Event bus pattern
- Pub/Sub with Redis (or NATS)
- Async processing
- Decoupled systems

**Events to Implement:**

- [ ] `UserRegistered` event
- [ ] `TodoCreated` event
- [ ] `TodoCompleted` event
- [ ] `TodoDeleted` event

**Implementation Tasks:**

- [ ] Create event bus interface
- [ ] Implement Redis Pub/Sub event bus
- [ ] Create event publisher
- [ ] Create event subscribers
- [ ] Add event handlers:
  - [ ] Audit log handler (logs all events)
  - [ ] Analytics handler (counts events)
  - [ ] Notification handler (future: send emails)
- [ ] Handle subscriber failures gracefully
- [ ] Add retry logic for failed events
- [ ] Test event flow end-to-end

**Architecture:**

```
User creates todo →
  1. Save to DB
  2. Emit "TodoCreated" event →
     - Analytics service listens
     - Audit log service listens
     - Notification service listens (future)
```

**Why this matters:** Prepares you for AWS EventBridge, SQS, SNS

---

### 4.5 Integration Tests with Real Dependencies

**What you'll learn:**

- Testcontainers (spin up real PostgreSQL)
- Database fixtures and cleanup
- Test isolation strategies
- E2E testing patterns

**Implementation Tasks:**

- [ ] Add `testcontainers-go` dependency
- [ ] Create integration test helpers
- [ ] Write integration tests for auth flow
- [ ] Write integration tests for todo CRUD
- [ ] Test caching behavior
- [ ] Test event publishing
- [ ] Add to CI pipeline
- [ ] Document when to use unit vs integration tests

**Deliverable:** High confidence in your full application stack

---

## Phase 5: Kubernetes & Helm Charts

**Duration:** Weeks 5-7  
**Theme:** Master container orchestration with Kubernetes

**Why Learn This Now:**

- You already understand Docker containers (Phase 2.1)
- You've orchestrated services with Docker Compose (Phase 2.2)
- You have observability in place (Phase 3)
- Now learn production-grade orchestration before cloud deployment

**Learning Path:** Local K8s → Manifests → Helm → Production Patterns

---

### 5.1 Kubernetes Fundamentals

**What you'll learn:**

- Kubernetes architecture (Control Plane, Nodes, Pods)
- Core concepts: Pods, Deployments, Services, ConfigMaps, Secrets
- kubectl CLI and context management
- Declarative vs imperative configuration
- Kubernetes namespaces and resource organization
- Label selectors and annotations

**Setup Tasks:**

- [ ] Install Docker Desktop with Kubernetes enabled (or minikube/kind)
- [ ] Verify installation: `kubectl version`
- [ ] Explore with: `kubectl get nodes`, `kubectl cluster-info`
- [ ] Install k9s (terminal UI for K8s - highly recommended!)
- [ ] Understand kubectl contexts: `kubectl config get-contexts`
- [ ] Create a test namespace: `kubectl create namespace test`

**Learning Exercises:**

- [ ] Deploy nginx with `kubectl run` (imperative)
- [ ] Expose nginx with `kubectl expose` (imperative)
- [ ] Delete and recreate with YAML (declarative)
- [ ] Understand the difference: imperative vs declarative

**Architecture Understanding:**

```
Kubernetes Cluster
├── Control Plane
│   ├── API Server (kubectl talks to this)
│   ├── Scheduler (assigns Pods to Nodes)
│   ├── Controller Manager (maintains desired state)
│   └── etcd (cluster state storage)
└── Nodes (Worker machines)
    └── Pods (smallest deployable unit)
        └── Containers (your Docker images)
```

---

### 5.2 Kubernetes Manifests for DoIt API

**What you'll learn:**

- Writing production-ready Kubernetes manifests
- Resource limits and requests
- Liveness and readiness probes
- ConfigMaps for configuration
- Secrets for sensitive data
- Multi-container pods
- Init containers for migrations

**Project Structure:**

```
k8s/
├── base/                       # Base manifests
│   ├── namespace.yaml
│   ├── configmap.yaml
│   ├── secret.yaml
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── postgres-deployment.yaml
│   ├── postgres-service.yaml
│   ├── postgres-pvc.yaml
│   ├── redis-deployment.yaml
│   └── redis-service.yaml
├── overlays/                   # Environment-specific
│   ├── dev/
│   │   └── kustomization.yaml
│   ├── staging/
│   │   └── kustomization.yaml
│   └── prod/
│       └── kustomization.yaml
└── README.md
```

**Implementation Tasks:**

#### **5.2.1 Namespace**

- [ ] Create namespace manifest (`k8s/base/namespace.yaml`)

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: doit
  labels:
    app: doit
    environment: dev
```

#### **5.2.2 ConfigMap**

- [ ] Create ConfigMap for non-sensitive config

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: doit-config
  namespace: doit
data:
  APP_ENVIRONMENT: "production"
  APP_NAME: "doit-api"
  LOG_LEVEL: "info"
  DB_HOST: "postgres-service"
  DB_PORT: "5432"
  DB_NAME: "doit"
  REDIS_ADDR: "redis-service:6379"
```

#### **5.2.3 Secret**

- [ ] Create Secret for sensitive data

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: doit-secrets
  namespace: doit
type: Opaque
stringData:
  DB_USER: "doit"
  DB_PASSWORD: "changeme"
  JWT_SECRET: "your-super-secret-key"
  REDIS_PASSWORD: ""
```

- [ ] Learn about sealed-secrets for GitOps (store secrets safely in Git)

#### **5.2.4 Deployment (Your API)**

- [ ] Create Deployment manifest
- [ ] Define resource requests and limits:
  ```yaml
  resources:
    requests:
      memory: "128Mi"
      cpu: "100m"
    limits:
      memory: "512Mi"
      cpu: "500m"
  ```
- [ ] Add liveness probe (is app alive?)
  ```yaml
  livenessProbe:
    httpGet:
      path: /health/liveness
      port: 8080
    initialDelaySeconds: 10
    periodSeconds: 30
  ```
- [ ] Add readiness probe (is app ready for traffic?)
  ```yaml
  readinessProbe:
    httpGet:
      path: /health/readiness
      port: 8080
    initialDelaySeconds: 5
    periodSeconds: 10
  ```
- [ ] Configure environment variables from ConfigMap and Secret
- [ ] Set replica count: 3 (for high availability)
- [ ] Add pod anti-affinity (spread across nodes)

#### **5.2.5 Service**

- [ ] Create Service to expose API
- [ ] Type: ClusterIP (internal) or LoadBalancer (external)
- [ ] Configure selectors to match Deployment labels
- [ ] Expose port 80 → targetPort 8080

#### **5.2.6 PostgreSQL Deployment**

- [ ] Create PersistentVolumeClaim for database storage
- [ ] Create PostgreSQL Deployment
- [ ] Create PostgreSQL Service (ClusterIP - internal only)
- [ ] Add init container for database initialization
- [ ] Configure resource limits

#### **5.2.7 Redis Deployment**

- [ ] Create Redis Deployment
- [ ] Create Redis Service
- [ ] Configure persistence (if needed)
- [ ] Set resource limits

**Testing:**

- [ ] Apply all manifests: `kubectl apply -f k8s/base/`
- [ ] Check resources: `kubectl get all -n doit`
- [ ] View logs: `kubectl logs -n doit deployment/doit-api`
- [ ] Port-forward to test: `kubectl port-forward -n doit svc/doit-api 8080:80`
- [ ] Test API: `curl http://localhost:8080/health`

---

### 5.3 Advanced Kubernetes Patterns

**What you'll learn:**

- Horizontal Pod Autoscaler (HPA)
- Ingress controllers for routing
- Network Policies for security
- Pod Disruption Budgets (PDB)
- Resource Quotas and Limits
- StatefulSets vs Deployments

#### **5.3.1 Horizontal Pod Autoscaler**

**What you'll learn:**

- Auto-scale based on CPU/memory
- Custom metrics (requests per second)

**Implementation:**

- [ ] Install metrics-server (if not present)

```bash
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
```

- [ ] Create HPA manifest:

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: doit-api-hpa
  namespace: doit
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: doit-api
  minReplicas: 2
  maxReplicas: 10
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: 80
```

- [ ] Test auto-scaling with load (use `hey` or `ab`)
- [ ] Watch scaling: `kubectl get hpa -n doit --watch`

#### **5.3.2 Ingress Controller**

**What you'll learn:**

- L7 load balancing
- Path-based routing
- TLS/SSL termination
- Multiple services behind one IP

**Implementation:**

- [ ] Install ingress-nginx controller

```bash
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.2/deploy/static/provider/cloud/deploy.yaml
```

- [ ] Create Ingress manifest:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: doit-ingress
  namespace: doit
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  ingressClassName: nginx
  tls:
    - hosts:
        - api.doit.example.com
      secretName: doit-tls
  rules:
    - host: api.doit.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: doit-api
                port:
                  number: 80
```

- [ ] Test ingress routing
- [ ] (Optional) Install cert-manager for automatic TLS certificates

#### **5.3.3 Network Policies**

**What you'll learn:**

- Pod-to-pod network security
- Zero-trust networking
- Ingress and egress rules

**Implementation:**

- [ ] Create NetworkPolicy to restrict PostgreSQL access:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: postgres-network-policy
  namespace: doit
spec:
  podSelector:
    matchLabels:
      app: postgres
  policyTypes:
    - Ingress
  ingress:
    - from:
        - podSelector:
            matchLabels:
              app: doit-api # Only API can access Postgres
      ports:
        - protocol: TCP
          port: 5432
```

- [ ] Test that external access is blocked
- [ ] Create similar policy for Redis

#### **5.3.4 Pod Disruption Budget**

**What you'll learn:**

- Ensure availability during voluntary disruptions
- Rolling updates without downtime

**Implementation:**

- [ ] Create PDB manifest:

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: doit-api-pdb
  namespace: doit
spec:
  minAvailable: 2 # Always keep 2 pods running
  selector:
    matchLabels:
      app: doit-api
```

---

### 5.4 Helm Charts - Package Management for Kubernetes

**What you'll learn:**

- Helm architecture (Charts, Releases, Repositories)
- Chart structure and templating
- Values files for different environments
- Helm hooks (pre-install, post-install)
- Chart dependencies
- Helm best practices

**Why Helm:**

- Reusable templates (deploy to dev/staging/prod with different values)
- Version control for releases
- Easy rollbacks
- Share charts with team
- Industry standard for K8s package management

#### **5.4.1 Helm Basics**

**Setup:**

- [ ] Install Helm: `brew install helm` (macOS) or download from helm.sh
- [ ] Verify: `helm version`
- [ ] Add popular repos:

```bash
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
```

**Learning Exercises:**

- [ ] Install PostgreSQL with Helm:

```bash
helm install my-postgres bitnami/postgresql -n doit
```

- [ ] List releases: `helm list -n doit`
- [ ] Get values: `helm get values my-postgres -n doit`
- [ ] Upgrade: `helm upgrade my-postgres bitnami/postgresql --set auth.password=newpass -n doit`
- [ ] Rollback: `helm rollback my-postgres -n doit`
- [ ] Uninstall: `helm uninstall my-postgres -n doit`

#### **5.4.2 Create Your Own Helm Chart**

**Chart Structure:**

```
helm/
└── doit-api/
    ├── Chart.yaml           # Chart metadata
    ├── values.yaml          # Default values
    ├── values-dev.yaml      # Dev environment overrides
    ├── values-staging.yaml  # Staging overrides
    ├── values-prod.yaml     # Production overrides
    ├── templates/
    │   ├── NOTES.txt       # Post-install notes
    │   ├── _helpers.tpl    # Template helpers
    │   ├── deployment.yaml
    │   ├── service.yaml
    │   ├── configmap.yaml
    │   ├── secret.yaml
    │   ├── ingress.yaml
    │   ├── hpa.yaml
    │   ├── serviceaccount.yaml
    │   └── tests/
    │       └── test-connection.yaml
    └── .helmignore
```

**Implementation Tasks:**

- [ ] Create chart skeleton:

```bash
helm create helm/doit-api
```

- [ ] Customize `Chart.yaml`:

```yaml
apiVersion: v2
name: doit-api
description: A Helm chart for DoIt REST API
type: application
version: 0.1.0
appVersion: "1.0.0"
keywords:
  - doit
  - api
  - golang
  - rest
maintainers:
  - name: Your Name
    email: your.email@example.com
```

- [ ] Define `values.yaml` with sensible defaults:

```yaml
replicaCount: 3

image:
  repository: doit-api
  pullPolicy: IfNotPresent
  tag: "latest"

service:
  type: ClusterIP
  port: 80
  targetPort: 8080

ingress:
  enabled: true
  className: nginx
  hosts:
    - host: api.doit.local
      paths:
        - path: /
          pathType: Prefix
  tls: []

resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 100m
    memory: 128Mi

autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70

config:
  appEnvironment: "production"
  logLevel: "info"
  dbHost: "postgres-service"
  dbPort: "5432"
  dbName: "doit"
  redisAddr: "redis-service:6379"

secrets:
  dbUser: "doit"
  dbPassword: "changeme"
  jwtSecret: "your-secret-key"

postgresql:
  enabled: true
  auth:
    username: doit
    password: changeme
    database: doit
  primary:
    persistence:
      enabled: true
      size: 8Gi

redis:
  enabled: true
  auth:
    enabled: false
```

- [ ] Create environment-specific values files:
  - `values-dev.yaml`: Lower resources, debug logging
  - `values-staging.yaml`: Medium resources, realistic data
  - `values-prod.yaml`: Full resources, monitoring enabled

#### **5.4.3 Helm Templating**

**What you'll learn:**

- Go templating syntax
- Built-in objects (`.Values`, `.Chart`, `.Release`)
- Template functions (default, required, quote, toYaml)
- Control structures (if, range, with)
- Named templates and helpers

**Example: Templated Deployment**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "doit-api.fullname" . }}
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "doit-api.labels" . | nindent 4 }}
spec:
  {{- if not .Values.autoscaling.enabled }}
  replicas: {{ .Values.replicaCount }}
  {{- end }}
  selector:
    matchLabels:
      {{- include "doit-api.selectorLabels" . | nindent 6 }}
  template:
    metadata:
      labels:
        {{- include "doit-api.selectorLabels" . | nindent 8 }}
    spec:
      containers:
      - name: {{ .Chart.Name }}
        image: "{{ .Values.image.repository }}:{{ .Values.image.tag | default .Chart.AppVersion }}"
        imagePullPolicy: {{ .Values.image.pullPolicy }}
        ports:
        - name: http
          containerPort: {{ .Values.service.targetPort }}
          protocol: TCP
        env:
        - name: APP_ENVIRONMENT
          value: {{ .Values.config.appEnvironment | quote }}
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: {{ include "doit-api.fullname" . }}-secret
              key: dbPassword
        resources:
          {{- toYaml .Values.resources | nindent 10 }}
```

**Tasks:**

- [ ] Template all manifests
- [ ] Create `_helpers.tpl` with reusable functions
- [ ] Use `required` for mandatory values
- [ ] Add conditional blocks (e.g., ingress enabled/disabled)
- [ ] Test rendering: `helm template doit-api helm/doit-api`
- [ ] Lint chart: `helm lint helm/doit-api`

#### **5.4.4 Helm Dependencies**

**What you'll learn:**

- Including other charts as dependencies
- Subchart values override
- Managing external dependencies

**Implementation:**

- [ ] Add dependencies to `Chart.yaml`:

```yaml
dependencies:
  - name: postgresql
    version: "12.x.x"
    repository: https://charts.bitnami.com/bitnami
    condition: postgresql.enabled
  - name: redis
    version: "17.x.x"
    repository: https://charts.bitnami.com/bitnami
    condition: redis.enabled
  - name: prometheus
    version: "25.x.x"
    repository: https://prometheus-community.github.io/helm-charts
    condition: prometheus.enabled
```

- [ ] Update dependencies:

```bash
helm dependency update helm/doit-api
```

- [ ] This downloads subcharts to `charts/` directory
- [ ] Override subchart values in your `values.yaml`

#### **5.4.5 Helm Hooks**

**What you'll learn:**

- Run jobs before/after install, upgrade, delete
- Database migrations as pre-upgrade hooks
- Cleanup jobs as post-delete hooks

**Use Cases:**

- Run database migrations before deploying new version
- Seed initial data on first install
- Clean up resources on uninstall

**Implementation:**

- [ ] Create migration job with hook:

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ include "doit-api.fullname" . }}-migration
  annotations:
    "helm.sh/hook": pre-upgrade,pre-install
    "helm.sh/hook-weight": "-5"
    "helm.sh/hook-delete-policy": before-hook-creation
spec:
  template:
    spec:
      restartPolicy: Never
      containers:
      - name: migration
        image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
        command: ["migrate"]
        args: ["-path", "/migrations", "-database", "$(DB_URL)", "up"]
        env:
        - name: DB_URL
          value: "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable"
```

- [ ] Test hook execution during install/upgrade

#### **5.4.6 Chart Testing**

- [ ] Create test in `templates/tests/test-connection.yaml`:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: "{{ include "doit-api.fullname" . }}-test"
  annotations:
    "helm.sh/hook": test
spec:
  restartPolicy: Never
  containers:
  - name: wget
    image: busybox
    command: ['wget']
    args: ['{{ include "doit-api.fullname" . }}:{{ .Values.service.port }}/health']
```

- [ ] Run tests:

```bash
helm test doit-api -n doit
```

---

### 5.5 Deploying with Helm

**Installation:**

```bash
# Install to dev environment
helm install doit-api helm/doit-api \
  -f helm/doit-api/values-dev.yaml \
  -n doit-dev \
  --create-namespace

# Install to production
helm install doit-api helm/doit-api \
  -f helm/doit-api/values-prod.yaml \
  -n doit-prod \
  --create-namespace
```

**Upgrade:**

```bash
# Upgrade with new values
helm upgrade doit-api helm/doit-api \
  -f helm/doit-api/values-prod.yaml \
  -n doit-prod

# Upgrade with specific image tag
helm upgrade doit-api helm/doit-api \
  --set image.tag=v1.2.3 \
  -n doit-prod
```

**Rollback:**

```bash
# View history
helm history doit-api -n doit-prod

# Rollback to previous version
helm rollback doit-api -n doit-prod

# Rollback to specific revision
helm rollback doit-api 3 -n doit-prod
```

**Uninstall:**

```bash
helm uninstall doit-api -n doit-prod
```

**Tasks:**

- [ ] Document installation procedure
- [ ] Create Makefile targets:
  - `make helm-install-dev`
  - `make helm-install-prod`
  - `make helm-upgrade-dev`
  - `make helm-test`
- [ ] Version your chart (update Chart.yaml on changes)
- [ ] Package chart: `helm package helm/doit-api`
- [ ] (Optional) Publish to chart repository

---

### 5.6 Observability in Kubernetes

**What you'll learn:**

- Prometheus Operator
- Grafana in K8s
- Service Monitors
- Custom dashboards for K8s metrics

**Implementation:**

- [ ] Install kube-prometheus-stack via Helm:

```bash
helm install kube-prometheus prometheus-community/kube-prometheus-stack \
  -n monitoring \
  --create-namespace
```

This installs:

- Prometheus Operator
- Grafana
- Alertmanager
- Node Exporter
- kube-state-metrics

- [ ] Create ServiceMonitor for your API:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: doit-api
  namespace: doit
spec:
  selector:
    matchLabels:
      app: doit-api
  endpoints:
    - port: http
      path: /metrics
      interval: 30s
```

- [ ] Access Grafana:

```bash
kubectl port-forward -n monitoring svc/kube-prometheus-grafana 3000:80
```

- [ ] Import Kubernetes dashboards
- [ ] Create custom dashboard for your API

---

### 5.7 Production Kubernetes Best Practices

**What you'll learn:**

- Resource quotas per namespace
- Limit ranges
- Pod security policies/standards
- RBAC (Role-Based Access Control)
- Service accounts
- Security contexts

#### **5.7.1 Resource Quotas**

```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: doit-quota
  namespace: doit-prod
spec:
  hard:
    requests.cpu: "10"
    requests.memory: 20Gi
    limits.cpu: "20"
    limits.memory: 40Gi
    persistentvolumeclaims: "5"
    services.loadbalancers: "2"
```

#### **5.7.2 Pod Security Standards**

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: doit-prod
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/warn: restricted
```

#### **5.7.3 RBAC**

- [ ] Create ServiceAccount for your app
- [ ] Create Role with minimal permissions
- [ ] Bind Role to ServiceAccount
- [ ] Use ServiceAccount in Deployment

**Tasks:**

- [ ] Implement all security best practices
- [ ] Document security model
- [ ] Run security scans (kubesec, kube-bench)

---

### 5.8 Local Kubernetes Testing Tools

**Tools to master:**

1. **k9s** - Terminal UI for K8s

   ```bash
   brew install k9s
   k9s
   ```

2. **stern** - Multi-pod log tailing

   ```bash
   brew install stern
   stern doit-api -n doit
   ```

3. **kubectx/kubens** - Context/namespace switching

   ```bash
   brew install kubectx
   kubectx docker-desktop
   kubens doit
   ```

4. **kustomize** - Template-free customization

   ```bash
   kubectl apply -k k8s/overlays/dev/
   ```

5. **helm diff** - Preview changes
   ```bash
   helm plugin install https://github.com/databus23/helm-diff
   helm diff upgrade doit-api helm/doit-api -n doit
   ```

---

### 5.9 CI/CD with Kubernetes & Helm

**What you'll learn:**

- Automated Helm deployments
- Image tagging strategies
- ArgoCD for GitOps (optional)

**GitHub Actions Workflow:**

```yaml
name: Deploy to Kubernetes

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Build Docker image
        run: |
          docker build -t doit-api:${{ github.sha }} .

      - name: Push to registry
        run: |
          echo "${{ secrets.REGISTRY_PASSWORD }}" | docker login -u ${{ secrets.REGISTRY_USERNAME }} --password-stdin
          docker push doit-api:${{ github.sha }}

      - name: Setup kubectl
        uses: azure/setup-kubectl@v3

      - name: Setup Helm
        uses: azure/setup-helm@v3

      - name: Deploy with Helm
        run: |
          helm upgrade --install doit-api helm/doit-api \
            --set image.tag=${{ github.sha }} \
            -f helm/doit-api/values-dev.yaml \
            -n doit-dev \
            --create-namespace
```

**Tasks:**

- [ ] Set up CI/CD pipeline for K8s
- [ ] Implement proper image tagging (git SHA, semver)
- [ ] Add smoke tests after deployment
- [ ] Configure rollback on failure

---

## Phase 6: AWS Deployment Foundation

**Duration:** Weeks 7-9  
**Theme:** Deploy to real cloud infrastructure

**Note:** Now that you understand Kubernetes, you can choose between ECS (simpler, managed) or EKS (Kubernetes on AWS). Both paths are covered below.

### 6.1 AWS Account Setup & Fundamentals

**What you'll learn:**

- AWS account best practices
- IAM users, roles, and policies (least privilege)
- VPC, subnets, security groups
- AWS CLI configuration
- Cost management and billing alerts

**Setup Tasks:**

- [ ] Create AWS account (use free tier)
- [ ] Enable MFA on root account
- [ ] Create IAM admin user (don't use root!)
- [ ] Configure AWS CLI with profiles
- [ ] Set up billing alerts
- [ ] Understand AWS Free Tier limits
- [ ] Create budget alerts ($10, $20, $50)

**Security Tasks:**

- [ ] Set up CloudTrail (audit logging)
- [ ] Enable AWS Config (compliance)
- [ ] Review IAM Access Analyzer

---

### 6.2 Infrastructure as Code (Terraform)

**What you'll learn:**

- Declarative infrastructure
- State management (local, S3 backend)
- Modules and reusability
- Workspaces (dev/staging/prod)
- Terraform best practices

**Project Structure:**

```
infrastructure/
  terraform/
    modules/
      networking/    # VPC, subnets, security groups
      compute/       # ECS, EC2, or EKS
      database/      # RDS, ElastiCache
      monitoring/    # CloudWatch, alarms
      storage/       # S3 buckets
    environments/
      dev/
      staging/
      prod/
```

**Implementation Tasks:**

- [ ] Install Terraform
- [ ] Create S3 bucket for Terraform state
- [ ] Set up DynamoDB table for state locking
- [ ] Create networking module (VPC)
  - [ ] VPC with public/private subnets
  - [ ] Internet Gateway
  - [ ] NAT Gateway (or NAT instance for free tier)
  - [ ] Security groups
- [ ] Create database module (RDS)
  - [ ] PostgreSQL RDS instance
  - [ ] Subnet group
  - [ ] Security group rules
- [ ] Create cache module (ElastiCache)
  - [ ] Redis cluster
  - [ ] Subnet group
- [ ] Create variables and outputs
- [ ] Test `terraform plan` and `terraform apply`
- [ ] Document infrastructure

---

### 6.3 Deployment Path A: ECS Fargate (Simpler, Managed)

**What you'll learn:**

- Container orchestration on AWS
- ECS task definitions
- ECS services and clusters
- Application Load Balancer (ALB)
- Service discovery
- Auto-scaling policies
- CloudWatch integration

**Architecture:**

```
Internet
  ↓
Application Load Balancer (ALB)
  ↓
ECS Fargate Tasks (your Go app - auto-scaled)
  ↓
├─→ RDS PostgreSQL (private subnet)
└─→ ElastiCache Redis (private subnet)
```

**Implementation Tasks:**

- [ ] Create ECR repository for Docker images
- [ ] Create ECS cluster
- [ ] Write ECS task definition (JSON)
  - [ ] Define container specs
  - [ ] Set environment variables
  - [ ] Configure secrets (from Secrets Manager)
  - [ ] Set health check command
- [ ] Create Application Load Balancer
  - [ ] Configure target group
  - [ ] Set up health checks
  - [ ] Configure listeners (HTTP/HTTPS)
- [ ] Create ECS service
  - [ ] Link to task definition
  - [ ] Configure desired count
  - [ ] Set up service discovery
- [ ] Configure auto-scaling
  - [ ] Target tracking scaling (CPU/memory)
  - [ ] Request count per target
- [ ] Set up CloudWatch log groups
- [ ] Test deployment
- [ ] Configure custom domain (Route 53)

**Terraform Modules:**

- [ ] ALB module
- [ ] ECS cluster module
- [ ] ECS task definition module
- [ ] ECS service module

---

### 6.4 Deployment Path B: EKS (Kubernetes on AWS) - Production Grade

**What you'll learn:**

- EKS cluster provisioning with Terraform
- AWS-specific Kubernetes integrations
- AWS Load Balancer Controller
- EKS IAM roles for service accounts (IRSA)
- Amazon EBS CSI driver for storage
- AWS Secrets Manager integration
- EKS managed node groups
- Cluster autoscaler
- Cost optimization strategies

**Why Choose EKS:**

- ✅ You already know Kubernetes (Phase 5)
- ✅ Portable skills (works on any K8s cluster)
- ✅ More control and flexibility
- ✅ Strong ecosystem (Helm, operators, etc.)
- ✅ Multi-cloud strategy possible
- ⚠️ More complex than ECS
- ⚠️ More expensive (control plane + nodes)

**Architecture:**

```
Internet
  ↓
AWS Load Balancer (ALB - created by Ingress)
  ↓
EKS Cluster
  ├─ Control Plane (AWS managed)
  └─ Worker Nodes (EC2 instances - auto-scaled)
      ├─ doit-api Pods (3+ replicas)
      ├─ Ingress Controller Pods
      └─ Monitoring Pods (Prometheus, Grafana)

Connected to:
├─→ RDS PostgreSQL (private subnet)
├─→ ElastiCache Redis (private subnet)
└─→ AWS Secrets Manager
```

**Implementation Tasks:**

#### **6.4.1 EKS Cluster Creation (Terraform)**

- [ ] Create VPC module for EKS

```hcl
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.0"

  name = "doit-eks-vpc"
  cidr = "10.0.0.0/16"

  azs             = ["us-east-1a", "us-east-1b", "us-east-1c"]
  private_subnets = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
  public_subnets  = ["10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"]

  enable_nat_gateway   = true
  single_nat_gateway   = false  # High availability
  enable_dns_hostnames = true

  # Tags for EKS
  public_subnet_tags = {
    "kubernetes.io/role/elb" = "1"
  }

  private_subnet_tags = {
    "kubernetes.io/role/internal-elb" = "1"
  }
}
```

- [ ] Create EKS cluster module

```hcl
module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 19.0"

  cluster_name    = "doit-eks"
  cluster_version = "1.28"

  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets

  # OIDC provider for IRSA
  enable_irsa = true

  # Managed node groups
  eks_managed_node_groups = {
    general = {
      desired_size = 2
      min_size     = 2
      max_size     = 10

      instance_types = ["t3.medium"]
      capacity_type  = "ON_DEMAND"

      labels = {
        role = "general"
      }

      tags = {
        Environment = "production"
      }
    }
  }

  # Cluster add-ons
  cluster_addons = {
    coredns = {
      most_recent = true
    }
    kube-proxy = {
      most_recent = true
    }
    vpc-cni = {
      most_recent = true
    }
    aws-ebs-csi-driver = {
      most_recent = true
    }
  }
}
```

- [ ] Apply Terraform:

```bash
cd infrastructure/terraform/environments/prod
terraform init
terraform plan
terraform apply
```

- [ ] Configure kubectl:

```bash
aws eks update-kubeconfig --name doit-eks --region us-east-1
kubectl get nodes
```

#### **6.4.2 AWS Load Balancer Controller**

**What it does:** Creates AWS ALB/NLB from Kubernetes Ingress

- [ ] Create IAM role for controller (IRSA):

```hcl
module "aws_load_balancer_controller_irsa_role" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"
  version = "~> 5.0"

  role_name = "aws-load-balancer-controller"

  attach_load_balancer_controller_policy = true

  oidc_providers = {
    ex = {
      provider_arn               = module.eks.oidc_provider_arn
      namespace_service_accounts = ["kube-system:aws-load-balancer-controller"]
    }
  }
}
```

- [ ] Install controller via Helm:

```bash
helm repo add eks https://aws.github.io/eks-charts
helm install aws-load-balancer-controller eks/aws-load-balancer-controller \
  -n kube-system \
  --set clusterName=doit-eks \
  --set serviceAccount.create=true \
  --set serviceAccount.name=aws-load-balancer-controller \
  --set serviceAccount.annotations."eks\.amazonaws\.com/role-arn"="arn:aws:iam::ACCOUNT:role/aws-load-balancer-controller"
```

- [ ] Verify installation:

```bash
kubectl get deployment -n kube-system aws-load-balancer-controller
```

#### **6.4.3 Deploy Your Helm Chart to EKS**

- [ ] Create production values for EKS (`values-eks-prod.yaml`):

```yaml
image:
  repository: ACCOUNT.dkr.ecr.us-east-1.amazonaws.com/doit-api
  tag: "v1.0.0"

replicaCount: 3

ingress:
  enabled: true
  className: alb # Use AWS ALB
  annotations:
    alb.ingress.kubernetes.io/scheme: internet-facing
    alb.ingress.kubernetes.io/target-type: ip
    alb.ingress.kubernetes.io/healthcheck-path: /health
    alb.ingress.kubernetes.io/listen-ports: '[{"HTTP": 80}, {"HTTPS": 443}]'
    alb.ingress.kubernetes.io/certificate-arn: arn:aws:acm:REGION:ACCOUNT:certificate/CERT_ID
  hosts:
    - host: api.doit.example.com
      paths:
        - path: /
          pathType: Prefix

config:
  dbHost: doit-prod.xxxxx.us-east-1.rds.amazonaws.com # RDS endpoint
  dbPort: "5432"
  dbName: doit
  redisAddr: doit-redis.xxxxx.cache.amazonaws.com:6379 # ElastiCache endpoint

# Don't deploy PostgreSQL/Redis in K8s - use AWS managed services
postgresql:
  enabled: false

redis:
  enabled: false

# Use AWS Secrets Manager via External Secrets Operator (see below)
externalSecrets:
  enabled: true
```

- [ ] Deploy:

```bash
helm upgrade --install doit-api helm/doit-api \
  -f helm/doit-api/values-eks-prod.yaml \
  -n doit-prod \
  --create-namespace
```

- [ ] Verify:

```bash
kubectl get all -n doit-prod
kubectl get ingress -n doit-prod
```

#### **6.4.4 AWS Secrets Manager Integration**

**Why:** Store secrets in AWS Secrets Manager, not in K8s Secrets

- [ ] Install External Secrets Operator:

```bash
helm repo add external-secrets https://charts.external-secrets.io
helm install external-secrets external-secrets/external-secrets \
  -n external-secrets-system \
  --create-namespace
```

- [ ] Create IAM role for External Secrets (IRSA)
- [ ] Create SecretStore:

```yaml
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: aws-secrets-manager
  namespace: doit-prod
spec:
  provider:
    aws:
      service: SecretsManager
      region: us-east-1
      auth:
        jwt:
          serviceAccountRef:
            name: external-secrets-sa
```

- [ ] Create ExternalSecret:

```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: doit-secrets
  namespace: doit-prod
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: aws-secrets-manager
    kind: SecretStore
  target:
    name: doit-secrets
    creationPolicy: Owner
  data:
    - secretKey: dbPassword
      remoteRef:
        key: doit/prod/database
        property: password
    - secretKey: jwtSecret
      remoteRef:
        key: doit/prod/jwt
        property: secret
```

#### **6.4.5 Cluster Autoscaler**

**What it does:** Automatically adds/removes nodes based on demand

- [ ] Create IAM role for Cluster Autoscaler (IRSA)
- [ ] Install via Helm:

```bash
helm repo add autoscaler https://kubernetes.github.io/autoscaler
helm install cluster-autoscaler autoscaler/cluster-autoscaler \
  -n kube-system \
  --set autoDiscovery.clusterName=doit-eks \
  --set awsRegion=us-east-1 \
  --set rbac.serviceAccount.annotations."eks\.amazonaws\.com/role-arn"="arn:aws:iam::ACCOUNT:role/cluster-autoscaler"
```

- [ ] Test autoscaling:

```bash
# Scale up workload
kubectl scale deployment doit-api -n doit-prod --replicas=20

# Watch nodes being added
kubectl get nodes --watch
```

#### **6.4.6 Monitoring on EKS**

- [ ] Install kube-prometheus-stack (if not already installed from Phase 5):

```bash
helm install kube-prometheus prometheus-community/kube-prometheus-stack \
  -n monitoring \
  --create-namespace \
  -f values-prometheus-eks.yaml
```

- [ ] Expose Grafana via Ingress:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: grafana
  namespace: monitoring
  annotations:
    alb.ingress.kubernetes.io/scheme: internet-facing
    alb.ingress.kubernetes.io/target-type: ip
spec:
  ingressClassName: alb
  rules:
    - host: grafana.doit.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: kube-prometheus-grafana
                port:
                  number: 80
```

- [ ] Access Grafana at https://grafana.doit.example.com

#### **6.4.7 Cost Optimization for EKS**

- [ ] Use Spot Instances for non-critical workloads:

```hcl
eks_managed_node_groups = {
  spot = {
    desired_size = 2
    min_size     = 1
    max_size     = 10

    instance_types = ["t3.medium", "t3a.medium"]
    capacity_type  = "SPOT"

    labels = {
      role = "spot"
    }

    taints = [{
      key    = "spot"
      value  = "true"
      effect = "NoSchedule"
    }]
  }
}
```

- [ ] Configure pod tolerations for spot:

```yaml
tolerations:
  - key: "spot"
    operator: "Equal"
    value: "true"
    effect: "NoSchedule"
```

- [ ] Use Karpenter for more advanced autoscaling (optional)
- [ ] Set resource requests accurately
- [ ] Use Horizontal Pod Autoscaler (already configured in Phase 5)
- [ ] Use Vertical Pod Autoscaler for right-sizing

#### **6.4.8 EKS Production Checklist**

- [ ] Enable EKS control plane logging
- [ ] Configure pod security standards
- [ ] Set up network policies
- [ ] Enable EKS Secrets encryption (KMS)
- [ ] Configure IAM roles for service accounts (IRSA) for all workloads
- [ ] Set up AWS CloudWatch Container Insights
- [ ] Configure cluster and node backups (Velero)
- [ ] Set up multi-AZ deployment
- [ ] Document incident response procedures
- [ ] Test disaster recovery

**Deliverable:** Production-grade Kubernetes cluster on AWS with your application running! 🚀

**Comparison: ECS vs EKS**

| Aspect             | ECS Fargate (6.3) | EKS (6.4)              |
| ------------------ | ----------------- | ---------------------- |
| **Complexity**     | Lower             | Higher                 |
| **Cost**           | $$                | $$$                    |
| **Learning Curve** | Easier            | Steeper                |
| **Portability**    | AWS only          | Any cloud              |
| **Control**        | Less              | More                   |
| **Ecosystem**      | Limited           | Rich (Helm, Operators) |
| **Best For**       | AWS-first teams   | K8s-first teams        |

**Recommendation:**

- Choose **ECS** if you want simpler ops and AWS lock-in is OK
- Choose **EKS** if you want Kubernetes skills and multi-cloud portability

---

### 6.5 AWS Services Integration

#### Database (RDS)

- [ ] Create PostgreSQL RDS instance
- [ ] Configure automated backups
- [ ] Set up Multi-AZ for high availability
- [ ] Configure parameter groups
- [ ] Set up read replica (optional, costs extra)
- [ ] Connect app to RDS

#### Caching (ElastiCache)

- [ ] Create Redis cluster
- [ ] Configure cluster mode (disabled for free tier)
- [ ] Set up parameter groups
- [ ] Connect app to ElastiCache

#### Secrets Management

- [ ] Create secrets in AWS Secrets Manager:
  - [ ] Database credentials
  - [ ] JWT secret
  - [ ] Redis password
- [ ] Configure ECS task to fetch secrets
- [ ] Update app to use secrets from env vars

#### Monitoring (CloudWatch)

- [ ] Configure log groups for ECS tasks
- [ ] Set up log retention policies
- [ ] Create CloudWatch dashboards
- [ ] Set up alarms:
  - [ ] High CPU usage
  - [ ] High memory usage
  - [ ] HTTP 5xx errors
  - [ ] Database connection errors
- [ ] Configure SNS for alarm notifications

#### Security

- [ ] Configure security groups (least privilege)
- [ ] Set up AWS WAF (Web Application Firewall)
- [ ] Enable VPC Flow Logs
- [ ] Configure AWS Shield (DDoS protection)
- [ ] Set up AWS Config rules

**Deliverable:** Fully deployed, production-ready app on AWS! 🎉

---

## Phase 7: Advanced DevOps & CI/CD

**Duration:** Weeks 9-11  
**Theme:** Automate everything

### 7.1 CI/CD Pipeline Enhancement

**Current State:** CI only (testing, security scanning)  
**Goal:** Full CI/CD with automated deployments

**Pipeline Flow:**

```
Code Push to GitHub
  ↓
GitHub Actions CI:
  1. Run tests ✅
  2. Security scan ✅
  3. Code generation verification ✅
  4. Build Docker image
  5. Push to ECR
  6. Update ECS task definition (or K8s manifests)
  7. Deploy to dev environment
  8. Run smoke tests
  9. (Manual approval for prod)
  10. Deploy to production
  11. Run smoke tests
  12. Rollback if failed
```

**Implementation Tasks:**

- [ ] Add Docker build step to CI
- [ ] Configure AWS credentials in GitHub secrets
- [ ] Add ECR push step
- [ ] Create deployment job (separate from CI)
- [ ] Add environment-specific workflows (dev, staging, prod)
- [ ] Implement smoke tests (health check after deploy)
- [ ] Add rollback automation
- [ ] Set up deployment approvals for prod
- [ ] Add deployment notifications (Slack, email)
- [ ] Create deployment dashboards

**Advanced:**

- [ ] Blue-green deployments
- [ ] Canary deployments (10% → 50% → 100%)
- [ ] Feature flags for gradual rollouts

---

### 7.2 Database Migration Strategy

**What you'll learn:**

- Running migrations in production safely
- Zero-downtime migration patterns
- Rollback strategies
- Migration automation

**Decision:** Use `golang-migrate` CLI in CD pipeline

**Implementation Tasks:**

- [ ] Add migration step to CD pipeline
- [ ] Run migrations before deploying new app version
- [ ] Implement safe migration patterns:
  - [ ] Backward compatible migrations
  - [ ] Separate data from schema changes
- [ ] Test rollback scenarios
- [ ] Add migration health checks
- [ ] Document migration process

**Migration Patterns:**

- [ ] Additive changes (add column with default)
- [ ] Expanding then contracting (multi-step changes)
- [ ] Data migrations in separate steps

---

### 7.3 Environment Management

**What you'll learn:**

- Multi-environment strategy (dev, staging, prod)
- Configuration management
- Secrets per environment
- Environment parity

**Environments to Set Up:**

- [ ] Dev (development, auto-deploy from main)
- [ ] Staging (pre-production, auto-deploy from releases)
- [ ] Production (manual approval required)

**Configuration:**

- [ ] Use Terraform workspaces or separate state files
- [ ] Environment-specific variables (AWS SSM Parameter Store)
- [ ] Separate databases per environment
- [ ] Separate AWS accounts (best practice) or VPCs
- [ ] Document promotion process (dev → staging → prod)

---

### 7.4 Disaster Recovery & Backups

**What you'll learn:**

- RTO (Recovery Time Objective) and RPO (Recovery Point Objective)
- Backup strategies
- Point-in-time recovery
- Multi-AZ and Multi-Region

**Implementation Tasks:**

- [ ] Enable automated RDS backups (daily)
- [ ] Test RDS restore from backup
- [ ] Set up RDS snapshots before major changes
- [ ] Configure Redis persistence (AOF or RDB)
- [ ] Document disaster recovery procedures
- [ ] Test failover scenarios
- [ ] Set up Multi-AZ for RDS (high availability)
- [ ] (Optional) Set up cross-region replication

**Recovery Testing:**

- [ ] Test database restore
- [ ] Test application recovery
- [ ] Measure actual RTO and RPO
- [ ] Document lessons learned

---

## Phase 8: Advanced Architecture Patterns

**Duration:** Weeks 11-13  
**Theme:** Scale and resilience patterns

### 8.1 API Gateway Pattern

**What you'll learn:**

- Gateway as single entry point
- Request routing and transformation
- Rate limiting at edge
- Authentication at gateway

**Options:**

- AWS API Gateway (managed service)
- Build your own simple gateway (learning exercise)

**Implementation Tasks:**

- [ ] Create gateway service
- [ ] Implement request routing
- [ ] Add rate limiting at gateway level
- [ ] Move JWT validation to gateway
- [ ] Add request/response transformation
- [ ] Implement API versioning (v1, v2)
- [ ] Add CORS handling
- [ ] Test gateway under load

---

### 8.2 Background Jobs & Queues

**What you'll learn:**

- Asynchronous processing
- Message queues (SQS)
- Worker patterns
- Dead letter queues
- Retry strategies with exponential backoff

**Use Cases:**

- Send email notifications (don't block HTTP requests)
- Update analytics (eventual consistency is fine)
- Trigger webhooks
- Image processing (if you add file uploads)

**Implementation Tasks:**

- [ ] Set up AWS SQS queues:
  - [ ] Main queue (email notifications)
  - [ ] Dead letter queue (failed jobs)
- [ ] Create worker service
- [ ] Implement job processors:
  - [ ] Email sender
  - [ ] Analytics updater
  - [ ] Webhook dispatcher
- [ ] Add retry logic with exponential backoff
- [ ] Monitor queue depth (CloudWatch)
- [ ] Set up auto-scaling for workers (based on queue depth)
- [ ] Test failure scenarios

**Architecture:**

```
HTTP Request (Create Todo)
  ↓
API: Save to DB, return 201
  ↓
Publish to SQS queue
  ↓
Worker: Process async (send email, update analytics)
```

**Advanced (Optional):**

- [ ] Use AWS SNS for pub/sub (fan-out pattern)
- [ ] Implement priority queues
- [ ] Add scheduled jobs (cron-like)

---

### 7.3 Rate Limiting & Circuit Breakers

**What you'll learn:**

- Protecting services from overload
- Circuit breaker pattern (prevent cascading failures)
- Bulkhead pattern (isolate failures)
- Retry with exponential backoff (you have this in `pkg/retry`)

**Rate Limiting:**

- [ ] Implement rate limiting per user (100 req/min)
- [ ] Implement rate limiting per IP (1000 req/min)
- [ ] Use Redis for distributed rate limiting
- [ ] Add rate limit headers (X-RateLimit-Remaining)
- [ ] Return 429 Too Many Requests with Retry-After

**Circuit Breaker:**

- [ ] Add circuit breaker for database calls
- [ ] Add circuit breaker for Redis calls
- [ ] Add circuit breaker for external APIs (if you add any)
- [ ] Configure thresholds (fail 5 times → open circuit for 30s)
- [ ] Add health checks that report circuit state
- [ ] Test behavior during failures

**Library:** Use `sony/gobreaker` or build your own

---

### 7.4 Feature Flags

**What you'll learn:**

- Deploy code without releasing features
- A/B testing
- Gradual rollouts
- Kill switches for problematic features

**Implementation Tasks:**

- [ ] Create feature flag service
- [ ] Store flags in database or AWS AppConfig
- [ ] Implement flag evaluation
- [ ] Add flags to key features:
  - [ ] New caching layer (toggle on/off)
  - [ ] Event publishing (toggle on/off)
  - [ ] New API endpoints (beta access)
- [ ] Create admin API to toggle flags
- [ ] Add flag status to health check
- [ ] Document flag lifecycle

**Use Cases:**

- Beta features for specific users
- Gradual rollout (5% → 25% → 50% → 100%)
- Kill switch for buggy features

---

### 7.5 Multi-Region Architecture (Theory + Planning)

**What you'll learn:**

- Active-active vs active-passive
- Data replication strategies
- Latency-based routing (Route 53)
- Conflict resolution (last-write-wins, CRDTs)
- Global load balancing

**Planning Tasks:**

- [ ] Document multi-region strategy
- [ ] Identify stateless vs stateful components
- [ ] Plan database replication (RDS cross-region read replica)
- [ ] Plan cache replication (Redis Global Datastore)
- [ ] Document trade-offs (consistency vs availability)
- [ ] Design conflict resolution strategy
- [ ] Calculate costs for multi-region

**Optional Implementation:**

- [ ] Deploy to second AWS region (e.g., us-west-2)
- [ ] Set up Route 53 latency-based routing
- [ ] Configure cross-region RDS replica
- [ ] Test failover scenarios

---

## Phase 8: Production Operations & Scale

**Duration:** Weeks 11-12+  
**Theme:** Operating at scale

### 8.1 Performance Optimization

**What you'll learn:**

- Profiling Go applications
- Database query optimization
- Connection pooling tuning
- Memory optimization

**Tasks:**

- [ ] Set up Go profiling (pprof)
- [ ] Profile CPU usage under load
- [ ] Profile memory allocations
- [ ] Identify slow database queries (pg_stat_statements)
- [ ] Add database indexes where needed
- [ ] Optimize connection pool settings
- [ ] Reduce allocations in hot paths
- [ ] Benchmark improvements

---

### 8.2 Load Testing & Capacity Planning

**What you'll learn:**

- Load testing tools (k6, Gatling)
- Identifying bottlenecks
- Capacity planning
- Auto-scaling tuning

**Tasks:**

- [ ] Install k6 or similar tool
- [ ] Create load test scenarios:
  - [ ] Steady load (100 RPS)
  - [ ] Spike test (0 → 1000 RPS)
  - [ ] Soak test (sustained load for 1 hour)
- [ ] Run tests against staging
- [ ] Analyze results (latency, error rate, throughput)
- [ ] Identify bottlenecks
- [ ] Tune auto-scaling policies
- [ ] Test again, iterate

---

### 8.3 Cost Optimization

**What you'll learn:**

- Right-sizing instances
- Spot instances for non-critical workloads
- Reserved capacity planning
- Cost monitoring and alerts

**Tasks:**

- [ ] Analyze AWS Cost Explorer
- [ ] Identify biggest cost drivers
- [ ] Right-size RDS instances (don't over-provision)
- [ ] Use Spot instances for workers
- [ ] Consider Reserved Instances (if usage is stable)
- [ ] Set up cost anomaly detection
- [ ] Implement cost allocation tags
- [ ] Document cost optimization strategies

---

### 8.4 Security Hardening

**What you'll learn:**

- Penetration testing basics
- OWASP Top 10 mitigation
- Security scanning automation
- Compliance frameworks

**Tasks:**

- [ ] Run OWASP ZAP security scan
- [ ] Fix any found vulnerabilities
- [ ] Implement security headers:
  - [ ] X-Content-Type-Options
  - [ ] X-Frame-Options
  - [ ] Strict-Transport-Security
  - [ ] Content-Security-Policy
- [ ] Enable AWS GuardDuty (threat detection)
- [ ] Set up AWS Security Hub
- [ ] Review IAM policies (principle of least privilege)
- [ ] Rotate secrets regularly (automate with Lambda)
- [ ] Document security practices

---

### 8.5 Compliance & Audit

**What you'll learn:**

- Audit logging
- Compliance frameworks (SOC2, GDPR concepts)
- Data retention policies
- Access controls

**Tasks:**

- [ ] Implement comprehensive audit logging
- [ ] Log all data mutations (who, what, when)
- [ ] Set up log retention policies
- [ ] Implement GDPR-style data export
- [ ] Implement data deletion (right to be forgotten)
- [ ] Document data handling procedures
- [ ] Review access controls
- [ ] Create compliance documentation

---

## 📚 Learning Resources

### Books

- **"Designing Data-Intensive Applications"** by Martin Kleppmann (architecture patterns - MUST READ)
- **"The Phoenix Project"** by Gene Kim (DevOps culture and practices)
- **"Site Reliability Engineering"** by Google (SRE practices)
- **"Release It!"** by Michael Nygard (production-ready software)
- **"Building Microservices"** by Sam Newman (distributed systems)
- **"Domain-Driven Design"** by Eric Evans (software architecture)

### AWS

- AWS Skill Builder (free courses)
- AWS Well-Architected Framework (read this!)
- AWS Solutions Library (reference architectures)
- AWS Whitepapers (security, performance, cost optimization)
- AWS re:Invent videos on YouTube

### Go & Architecture

- Go official blog (concurrency patterns)
- Effective Go (official guide)
- Practical Go (Dave Cheney's blog)
- The Twelve-Factor App (methodology)

### DevOps & Infrastructure

- Terraform documentation and tutorials
- Docker documentation
- Kubernetes documentation (kubernetes.io)
- CNCF landscape (cloud native tools)

### Monitoring & Observability

- Prometheus documentation
- Grafana tutorials
- OpenTelemetry documentation
- Google's SRE books (free online)

---

## 🎯 Recommended Starting Point

Since you want to learn DevOps, Backend Architecture, and AWS, here's the optimal path:

### **Weeks 1-2: Quick Wins**

1. Authentication (Phase 1.1) - 4 days
2. Health checks (Phase 1.3) - 1 day
3. Docker + Docker Compose (Phase 2) - 3 days
4. Observability basics (Phase 3.1-3.2) - 4 days

**Result:** Secure API with monitoring, running in containers

### **Weeks 3-4: Architecture**

1. Redis caching (Phase 4.1) - 3 days
2. Repository pattern (Phase 4.2) - 2 days
3. Events (Phase 4.4) - 3 days
4. API docs (Phase 1.2) - 2 days

**Result:** Well-architected, documented API with caching

### **Weeks 5-7: AWS Deployment**

1. Terraform basics (Phase 5.2) - 5 days
2. Deploy to ECS (Phase 5.3) - 7 days
3. Monitoring on AWS (Phase 5.5) - 2 days

**Result:** Production app running on AWS!

### **Weeks 8+: Advanced Topics**

Pick what interests you most from Phases 6-8

---

## 💰 Budget Considerations

### AWS Free Tier (12 months)

- ECS: 50 GB/month free
- RDS: 750 hours/month t2.micro or t3.micro
- ElastiCache: 750 hours/month t2.micro or t3.micro
- ALB: 15 LCUs per month
- CloudWatch: 10 custom metrics

### Estimated Monthly Cost (After Free Tier)

- **Minimal:** $20-30/month (single small instance)
- **Dev environment:** $50-75/month
- **Production-like:** $150-200/month (multi-AZ, monitoring, etc.)

### Cost Saving Tips

- Use Terraform to destroy environments when not in use
- Use AWS Budgets and alerts
- Start with smallest instance sizes
- Use Spot instances for workers

---

## ✅ Success Criteria

By the end of this roadmap, you will have:

1. ✅ Production-ready Go REST API with authentication
2. ✅ Comprehensive observability (logs, metrics, traces)
3. ✅ Full Docker and Docker Compose setup
4. ✅ Applied multiple architecture patterns (CQRS, events, caching, repository)
5. ✅ Deployed to AWS with IaC (Terraform)
6. ✅ CI/CD pipeline with automated testing and deployment
7. ✅ Understanding of 10+ AWS services
8. ✅ Real-world DevOps experience
9. ✅ Portfolio project to show employers
10. ✅ Deep understanding of production systems

---

## 📝 Tracking Your Progress

**Update this file as you go!**

- Mark checkboxes as you complete tasks
- Add notes on what you learned
- Document challenges and solutions
- Track time spent on each phase
- Celebrate wins! 🎉

**Additional Tracking:**

- Keep a learning journal (daily or weekly)
- Take notes on problems you solved
- Document architecture decisions (ADRs)
- Build a portfolio README showcasing what you built

---

## 🚀 Ready to Start?

Pick a phase and dive in! I recommend starting with **Phase 1.1 (Authentication)** - it's immediately useful and touches all layers of your application.

Good luck on your learning journey! 🎓

---

**Last Updated:** October 24, 2025  
**Project:** doit (Go REST API with PostgreSQL)  
**Focus:** Backend Architecture • DevOps • AWS
