# Learning Roadmap: Backend Architecture, DevOps & AWS

> **Project Purpose:** Learn best practices, software architecture patterns, DevOps workflows, and AWS deployment
>
> **Focus Areas:** Backend Architecture • DevOps • AWS Cloud • Production-Grade Patterns

---

## 📊 Progress Tracker

- [ ] Phase 1: Security & Production Readiness (Weeks 1-2)
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
- [ ] Implement rate limiting on auth endpoints (prevent brute force)

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

- [ ] Choose approach: `swaggo/swag` (code-first) or `oapi-codegen` (spec-first)
- [ ] Add Swagger annotations to all endpoints
- [ ] Generate OpenAPI spec
- [ ] Set up Swagger UI endpoint (`/swagger`)
- [ ] Document request/response schemas
- [ ] Add authentication documentation
- [ ] Document error responses
- [ ] Version your API (v1, v2 strategy)

**Deliverable:** Beautiful, interactive API documentation at `/swagger`

---

### 1.3 Graceful Shutdown & Health Checks

**What you'll learn:**

- Liveness vs Readiness probes (K8s concepts)
- Signal handling (SIGTERM, SIGINT)
- Connection draining
- Zero-downtime deployments

**Implementation Tasks:**

- [ ] Add `/health` endpoint (liveness probe)
- [ ] Add `/ready` endpoint (readiness probe - checks DB, Redis, etc.)
- [x] Implement graceful shutdown handler
- [x] Add timeout for in-flight requests
- [ ] Test shutdown behavior with active connections
- [ ] Add startup probe logic

**Why this matters:** Required for Kubernetes/ECS deployments. Prevents dropping requests during deploys.

**Architecture Pattern:** Graceful degradation

---

## Phase 2: Local Infrastructure & Containerization

**Duration:** Weeks 2-3  
**Theme:** Learn containerization and local orchestration

### 2.1 Docker Multi-Stage Build

**What you'll learn:**

- Builder pattern for Go apps
- Layer caching optimization
- Security: minimal base images, non-root users
- `.dockerignore` optimization

**Implementation Tasks:**

- [ ] Create multi-stage Dockerfile
  - Stage 1: Build (golang:1.24-alpine)
  - Stage 2: Runtime (alpine or distroless)
- [ ] Optimize layer caching (copy go.mod first)
- [ ] Add non-root user
- [ ] Set proper file permissions
- [ ] Configure `.dockerignore`
- [ ] Test build size (target: <20MB)
- [ ] Add labels (version, commit SHA, build date)

**Deliverable:** Production-ready Dockerfile (~15MB vs 1GB+ naive build)

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

## Phase 5: AWS Deployment Foundation

**Duration:** Weeks 5-7  
**Theme:** Deploy to real cloud infrastructure

### 5.1 AWS Account Setup & Fundamentals

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

### 5.2 Infrastructure as Code (Terraform)

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

### 5.3 Deployment Strategy: ECS Fargate ⭐ Recommended

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

### 5.4 Alternative: EKS (Kubernetes on AWS) - Advanced

**What you'll learn:**

- Kubernetes on cloud
- EKS cluster management
- kubectl and Helm
- Kubernetes manifests (Deployments, Services, ConfigMaps, Secrets)
- Ingress controllers
- Pod autoscaling (HPA)

**Implementation Tasks:**

- [ ] Create EKS cluster (Terraform)
- [ ] Configure kubectl
- [ ] Create Kubernetes manifests:
  - [ ] Deployment for your app
  - [ ] Service (LoadBalancer or ClusterIP)
  - [ ] ConfigMap for configuration
  - [ ] Secret for sensitive data
  - [ ] HorizontalPodAutoscaler
- [ ] Install ingress controller (AWS ALB Ingress Controller)
- [ ] Deploy with `kubectl apply`
- [ ] Set up Helm chart (optional)
- [ ] Configure monitoring (Prometheus Operator)

**Note:** More expensive than ECS, but more powerful and transferable skills

---

### 5.5 AWS Services Integration

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

## Phase 6: Advanced DevOps & CI/CD

**Duration:** Weeks 7-9  
**Theme:** Automate everything

### 6.1 CI/CD Pipeline Enhancement

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

### 6.2 Database Migration Strategy

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

### 6.3 Environment Management

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

### 6.4 Disaster Recovery & Backups

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

## Phase 7: Advanced Architecture Patterns

**Duration:** Weeks 9-11  
**Theme:** Scale and resilience patterns

### 7.1 API Gateway Pattern

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

### 7.2 Background Jobs & Queues

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
