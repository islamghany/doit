# Docker Multi-Stage Build - Quick Reference

## 🚀 Common Commands

### Building

```bash
# Build with automatic metadata
make docker-build

# Build without cache (clean build)
make docker-build-no-cache

# Check image size
make docker-size
```

### Running

```bash
# Run locally with environment variables
make docker-run

# Run in background
docker run -d --name doit-api -p 8080:8080 --env-file .env doit-api:latest

# View logs
docker logs -f doit-api
```

### Debugging

```bash
# Get shell access
make docker-shell

# Check user
docker run --rm doit-api:latest id
# Expected: uid=1000(appuser)

# View metadata
make docker-inspect
```

### Testing

```bash
# Build and test
make docker-build
curl http://localhost:8080/health

# Security scan
trivy image doit-api:latest
```

---

## 📁 File Structure

```
doit/
├── infra/docker/
│   ├── dockerfile.service              # Production Dockerfile
│   ├── DOCKER_MULTISTAGE_IMPLEMENTATION.md  # Full documentation
│   └── QUICK_REFERENCE.md             # This file
├── .dockerignore                       # Build optimization
└── Makefile                           # Automation commands
```

---

## 🎯 Key Metrics

- **Image Size**: ~15-20MB (98.5% reduction)
- **Build Time** (cached): ~30 seconds
- **Security**: Non-root user (UID 1000)
- **Base Image**: Alpine 3.19 (minimal)

---

## 🔍 Quick Troubleshooting

| Issue               | Solution                     |
| ------------------- | ---------------------------- |
| Large build context | Check `.dockerignore`        |
| Permission denied   | Verify `--chown` in COPY     |
| Cache not working   | Ensure `go.mod` copied first |
| SSL errors          | Install `ca-certificates`    |

---

## 📊 Build Stages

1. **Builder Stage** (golang:1.24-alpine)

   - Compiles Go binary
   - Downloads dependencies
   - ~300MB (discarded)

2. **Runtime Stage** (alpine:3.19)
   - Runs application
   - Non-root user
   - ~18MB (final image)

---

## 🏷️ Automatic Metadata

Built images include:

- Version (git tag)
- Commit SHA
- Build date
- OCI standard labels

```bash
make docker-inspect
# Shows all metadata
```

---

## ✅ Validation Checklist

```bash
# 1. Size check
make docker-size  # Should be ~15-20MB

# 2. Security check
docker run --rm doit-api:latest id  # Should be uid=1000

# 3. Metadata check
make docker-inspect  # Should show version, commit, date

# 4. Functionality check
make docker-run & sleep 5 && curl localhost:8080/health
```

---

## 📚 Full Documentation

For complete details, see:

- [DOCKER_MULTISTAGE_IMPLEMENTATION.md](./DOCKER_MULTISTAGE_IMPLEMENTATION.md)
- [Makefile](../../Makefile) (Docker section)
- [.dockerignore](../../.dockerignore)

---

**Status**: ✅ Production Ready  
**Last Updated**: 2025-11-15
