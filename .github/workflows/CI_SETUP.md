g# 🚀 CI/CD Setup Guide

## Overview

This project uses **GitHub Actions** for continuous integration with comprehensive checks including linting, testing, security scanning, and code generation verification.

## 📋 What Gets Checked

### 1. **Linting** (`lint` job)

- **Tool**: `golangci-lint` with 20+ linters
- **What it does**: Checks code style, potential bugs, performance issues
- **Config**: `.golangci.yml`
- **Run locally**: `make lint` or `make lint-fix`

### 2. **Build** (`build` job)

- Verifies the application compiles successfully
- Builds both main app (`cmd/doit`) and seeder (`cmd/seed`)
- Validates Go modules integrity
- **Run locally**: `make build`

### 3. **Testing** (`test` job)

- Runs all unit tests with **race detector**
- Generates **code coverage** report
- Uploads coverage to Codecov (optional)
- Uses `-race` flag to detect race conditions
- **Run locally**: `make test`

### 4. **Code Generation** (`codegen` job)

- Verifies sqlc-generated code is up-to-date
- Checks mock files are current
- **Fails if**: Generated code differs from committed code
- **Run locally**: `make sqlc` and `make generate-mocks`

### 5. **Security Scanning** (`security` job)

- **Tool**: Gosec
- Scans for security vulnerabilities
- Uploads results to GitHub Security tab
- **Run locally**: `make security`

### 6. **Vulnerability Check** (`vuln-check` job)

- **Tool**: govulncheck
- Checks dependencies for known vulnerabilities
- Scans Go standard library and third-party packages
- **Run locally**: `make vuln-check`

---

## 🔧 Local Development Workflow

### First Time Setup

```bash
# Install all required tools
make install-lint      # Install golangci-lint
make install-sqlc      # Install sqlc
make install-mockgen   # Install mockgen
```

### Before Committing

```bash
# Run all CI checks locally
make ci

# Or run individual checks:
make lint           # Run linter
make test           # Run tests
make security       # Security scan
make vuln-check     # Check vulnerabilities
```

### Fixing Issues

```bash
# Auto-fix linting issues
make lint-fix

# Regenerate code if out of sync
make sqlc
make generate-mocks
```

---

## 📊 CI Workflow Triggers

The CI pipeline runs on:

- ✅ **All Pull Requests** (to any branch)
- ✅ **Push to `main`** branch

---

## 🛠 Continuous Integration Jobs

### Job Parallelization

All 6 jobs run **in parallel** for faster feedback (typical runtime: 2-3 minutes)

```
┌─────────┐  ┌─────────┐  ┌─────────┐
│  Lint   │  │  Build  │  │  Test   │
└─────────┘  └─────────┘  └─────────┘
┌─────────┐  ┌─────────┐  ┌─────────┐
│ Codegen │  │Security │  │Vuln-Chk │
└─────────┘  └─────────┘  └─────────┘
```

---

## 📝 Configuration Files

### `.golangci.yml`

Configures which linters to run and their settings:

- **Enabled linters**: errcheck, gosimple, govet, staticcheck, gosec, etc.
- **Excludes**: Generated code (`internal/data/db/`), mocks, test files
- **Customizations**: See file for full list

### `.github/workflows/go.yaml`

Defines the CI pipeline:

- Job definitions
- Trigger conditions
- Tool installations
- Check commands

---

## 🚨 Common CI Failures & Fixes

### ❌ Linting Errors

**Problem**: `golangci-lint` reports issues

**Fix**:

```bash
# View issues
make lint

# Auto-fix (when possible)
make lint-fix

# Manual fixes required for:
# - Unused variables
# - Error handling
# - Code complexity
```

### ❌ Code Generation Out of Sync

**Problem**: "Generated code is out of sync"

**Fix**:

```bash
# Regenerate and commit
make sqlc
make generate-mocks
git add internal/data/db/ internal/service/mocks/
git commit -m "chore: regenerate code"
```

### ❌ Test Failures

**Problem**: Tests fail in CI but pass locally

**Common causes**:

1. **Race conditions**: CI uses `-race` flag
2. **Database state**: Tests might not clean up properly
3. **Time-dependent tests**: Use fixed times in tests

**Fix**:

```bash
# Run tests with race detector locally
go test -race ./...

# Run specific test
go test -v -race ./internal/service/...
```

### ❌ Security Vulnerabilities

**Problem**: `gosec` or `govulncheck` finds issues

**Fix**:

```bash
# View detailed report
make security
cat gosec-report.json

# Update vulnerable dependencies
go get -u ./...
go mod tidy
```

---

## 📈 Code Coverage

Coverage reports are:

- Generated on every test run
- Displayed in CI logs
- Uploaded as artifacts (retained for 7 days)
- Optionally sent to Codecov

**View coverage locally**:

```bash
make test
go tool cover -html=coverage.out
```

**Set coverage targets** (optional):
Add to CI workflow:

```yaml
- name: Check coverage threshold
  run: |
    coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    if (( $(echo "$coverage < 80" | bc -l) )); then
      echo "Coverage $coverage% is below 80%"
      exit 1
    fi
```

---

## 🎯 Best Practices

### 1. **Run Checks Before Pushing**

```bash
# Quick check
make ci

# If all pass, push your code
git push
```

### 2. **Keep Generated Code in Sync**

- Run `make sqlc` after changing SQL queries
- Run `make generate-mocks` after changing interfaces
- Commit generated code

### 3. **Write Tests with Coverage**

- Aim for >80% coverage on business logic
- Use table-driven tests for multiple scenarios
- Test error cases

### 4. **Follow Linting Rules**

- Fix linting errors immediately
- Don't disable linters without good reason
- Keep code simple and readable

### 5. **Security First**

- Update dependencies regularly
- Fix security issues ASAP
- Review Gosec findings

---

## 🔄 Adding New Checks

### Add a New Job

Edit `.github/workflows/go.yaml`:

```yaml
new-check:
  name: My New Check
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v5
      with:
        go-version: '1.24.2'
        cache: true
    - name: Run check
      run: go run ./scripts/my-check.go
```

### Add a New Linter

Edit `.golangci.yml`:

```yaml
linters:
  enable:
    - mynewlinter  # Add here
```

---

## 📚 References

- [golangci-lint documentation](https://golangci-lint.run/)
- [GitHub Actions Go setup](https://github.com/actions/setup-go)
- [Gosec security scanner](https://github.com/securego/gosec)
- [Go vulnerability database](https://pkg.go.dev/vuln)
- [sqlc documentation](https://sqlc.dev/)

---

## 💡 Tips

1. **Use `make help`** to see all available commands
2. **Install pre-commit hooks** to run checks before committing
3. **Monitor CI run times** - optimize slow jobs
4. **Review security findings** regularly
5. **Keep Go version up-to-date** in CI and locally

---

## 🆘 Getting Help

If CI fails and you're stuck:

1. **Check the logs** in GitHub Actions tab
2. **Run locally** with `make ci` to reproduce
3. **Read error messages** carefully
4. **Check this guide** for common issues
5. **Ask the team** if still stuck

---

## 🎉 Success!

When all checks pass, you'll see:

- ✅ Green checkmarks on your PR
- 📊 Coverage reports in artifacts
- 🔒 No security vulnerabilities
- 🎨 Clean, formatted code

Happy coding! 🚀
