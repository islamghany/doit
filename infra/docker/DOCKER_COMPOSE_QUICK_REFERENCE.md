# Docker Compose Quick Reference

> **Quick command reference for working with Docker Compose in the DoIt project**

---

## 🚀 Essential Commands

### Starting Services

```bash
# Start all services in background
make compose-up

# Build and start (after code changes)
make compose-up-build

# Start with logs visible
make compose-up-logs

# Start including optional tools (Adminer)
make compose-up-tools
```

### Stopping Services

```bash
# Stop all services (keeps data)
make compose-down

# Stop and remove volumes (⚠️  deletes data!)
make compose-down-v
```

### Viewing Logs

```bash
# All services
make compose-logs

# Specific service
make compose-logs-api
make compose-logs-db
make compose-logs-redis
make compose-logs-prometheus
make compose-logs-grafana
```

### Managing Services

```bash
# Check status
make compose-ps

# Check health
make compose-health

# Restart services
make compose-restart          # All
make compose-restart-api      # API only
make compose-restart-db       # PostgreSQL only
```

---

## 🔍 Debugging & Inspection

### Get Shell Access

```bash
# API container
make compose-shell-api

# PostgreSQL (psql)
make compose-shell-db

# Redis
make compose-shell-redis
```

### Direct Docker Compose Commands

```bash
# Execute command in service
docker-compose exec api /bin/sh

# Run one-off command
docker-compose run --rm api go test ./...

# View container processes
docker-compose top
```

---

## 🗄️ Database Operations

### Migrations

```bash
# Run migrations
make compose-migrate-up

# Rollback last migration
make compose-migrate-down
```

### Direct Database Access

```bash
# Via make command
make compose-shell-db

# Or direct psql
docker-compose exec postgres psql -U doit -d doit

# Execute SQL from host
docker-compose exec -T postgres psql -U doit -d doit -c "SELECT * FROM users;"
```

---

## 📊 Monitoring & Metrics

### Service URLs

| Service    | URL                                      | Credentials |
| ---------- | ---------------------------------------- | ----------- |
| API        | http://localhost:8080                    | -           |
| Swagger    | http://localhost:8080/swagger/index.html | -           |
| Health     | http://localhost:8080/health             | -           |
| Metrics    | http://localhost:8080/metrics            | -           |
| Grafana    | http://localhost:3000                    | admin/admin |
| Prometheus | http://localhost:9090                    | -           |
| Adminer    | http://localhost:8081                    | (see .env)  |

### Quick Checks

```bash
# API health
curl http://localhost:8080/health | jq

# API metrics
curl http://localhost:8080/metrics

# Prometheus targets
curl http://localhost:9090/api/v1/targets | jq

# Check all services
make compose-health
```

---

## 🛠️ Development Workflow

### Initial Setup

```bash
# 1. Copy environment template
make compose-setup

# 2. Edit .env file with your settings
vim .env

# 3. Start services
make compose-up

# 4. Check health
make compose-health

# 5. Run migrations (if needed)
make compose-migrate-up
```

### Daily Development

```bash
# Morning: Start stack
make compose-up

# Check logs when needed
make compose-logs-api

# After code changes: Rebuild API
make compose-up-build

# Or just restart API (if no dependency changes)
make compose-restart-api

# Evening: Stop stack
make compose-down
```

### Testing

```bash
# Run tests in API container
docker-compose exec api go test ./...

# Run tests with coverage
docker-compose exec api go test -cover ./...

# Run specific test
docker-compose exec api go test -v -run TestUserService ./internal/service
```

---

## 🧹 Maintenance

### Clean Up

```bash
# Stop and remove containers
make compose-down

# Stop and remove volumes (⚠️  data loss!)
make compose-down-v

# Remove unused images/volumes
docker system prune -a --volumes
```

### Update Images

```bash
# Pull latest base images
make compose-pull

# Rebuild with latest
make compose-up-build --no-cache
```

### Resource Usage

```bash
# Check resource usage
make compose-stats

# View disk usage
docker system df

# View logs size
docker-compose logs --tail=0 | wc -l
```

---

## 🔧 Troubleshooting

### Common Issues

#### Service Won't Start

```bash
# Check logs
make compose-logs-api

# Check health
docker-compose ps

# Restart service
make compose-restart-api

# Nuclear option: rebuild
make compose-down
make compose-up-build
```

#### Port Already in Use

```bash
# Find process using port
lsof -i :8080

# Kill process or change port in docker-compose.yml
```

#### Database Connection Issues

```bash
# Check PostgreSQL is running
make compose-ps

# Check PostgreSQL logs
make compose-logs-db

# Test connection
docker-compose exec postgres pg_isready -U doit -d doit

# Get into psql
make compose-shell-db
```

#### Redis Connection Issues

```bash
# Check Redis is running
make compose-ps

# Check Redis logs
make compose-logs-redis

# Test connection
docker-compose exec redis redis-cli ping
```

#### Migration Issues

```bash
# Check migration version
docker-compose exec postgres psql -U doit -d doit -c "SELECT * FROM schema_migrations;"

# Force migration version
docker-compose exec api migrate -path /app/internal/data/migrations -database "postgresql://doit:doit123@postgres:5432/doit?sslmode=disable" force <version>
```

### Clean Slate (Nuclear Option)

```bash
# Stop everything
make compose-down-v

# Remove all containers, networks, volumes
docker-compose down -v --remove-orphans

# Clean Docker system
docker system prune -a --volumes

# Start fresh
make compose-up-build
```

---

## 📝 Environment Variables

### Required Variables (.env)

```bash
# Database
DB_NAME=doit
DB_USER=doit
DB_PASSWORD=doit123

# Redis
REDIS_PASSWORD=

# JWT
JWT_SECRET=dev-secret-key-change-in-production

# Grafana
GRAFANA_ADMIN_USER=admin
GRAFANA_ADMIN_PASSWORD=admin
```

### Override Variables

```bash
# Override on command line
DB_PASSWORD=newpass docker-compose up -d

# Or export in shell
export DB_PASSWORD=newpass
docker-compose up -d
```

---

## 🎯 Pro Tips

### 1. **Service Dependencies**

Services start in dependency order thanks to `depends_on` + `healthcheck`:

```
PostgreSQL (healthy) → Redis (healthy) → API → Prometheus/Grafana
```

### 2. **Named Volumes**

Data persists even after `make compose-down`:

- `doit_postgres_data` - Database data
- `doit_redis_data` - Cache data
- `doit_prometheus_data` - Metrics data
- `doit_grafana_data` - Dashboards

### 3. **Service Discovery**

Services can reach each other by name:

```go
// In your Go code:
DB_HOST=postgres      // Not localhost!
REDIS_ADDR=redis:6379 // Not localhost:6379!
```

### 4. **Development Volume Mount**

API source code is mounted for hot reload:

```yaml
volumes:
  - ./:/app
```

Change code → Save → Container sees changes

### 5. **Profiles**

Optional services with profiles:

```bash
# Start without Adminer
make compose-up

# Start with Adminer
make compose-up-tools
```

### 6. **Health Checks Matter**

API won't start until DB is truly ready (not just running):

```yaml
depends_on:
  postgres:
    condition: service_healthy # Waits for healthcheck!
```

---

## 🔍 Advanced Usage

### Run Commands in Services

```bash
# Go commands
docker-compose exec api go version
docker-compose exec api go mod tidy
docker-compose exec api go generate ./...

# Database backup
docker-compose exec -T postgres pg_dump -U doit doit > backup.sql

# Database restore
docker-compose exec -T postgres psql -U doit doit < backup.sql

# Redis commands
docker-compose exec redis redis-cli KEYS "*"
docker-compose exec redis redis-cli FLUSHALL
```

### Scale Services (if stateless)

```bash
# Start 3 API instances (requires load balancer)
docker-compose up -d --scale api=3
```

### View Resource Limits

```bash
# Check limits
docker-compose config

# Set limits in docker-compose.yml
services:
  api:
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 512M
```

---

## 📚 Related Documentation

- [Docker Compose Mental Model](./DOCKER_COMPOSE_MENTAL_MODEL.md)
- [Docker Multi-Stage Build](./DOCKER_MULTISTAGE_IMPLEMENTATION.md)
- [Main README](../../README.md)

---

## 🆘 Getting Help

```bash
# Docker Compose help
docker-compose --help
docker-compose up --help

# View configuration
docker-compose config

# Validate configuration
docker-compose config --quiet

# List all make targets
make help
```

---

**Quick Checklist When Things Break:**

- [ ] Check logs: `make compose-logs-api`
- [ ] Check health: `make compose-health`
- [ ] Check status: `make compose-ps`
- [ ] Check .env file exists and is correct
- [ ] Try restart: `make compose-restart`
- [ ] Nuclear option: `make compose-down && make compose-up-build`
