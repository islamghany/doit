# GitHub Actions CI Workflow Explanation

## Overview

This is a comprehensive CI (Continuous Integration) pipeline that automatically validates your Go codebase. It runs on:

- **Pull Requests** - Every time someone opens or updates a PR
- **Pushes to `main`** - When code is merged to the main branch

## The 5 Jobs

The workflow is organized into 5 independent jobs that run in parallel:

### 1. **Build Job**

Verifies that your application compiles successfully.

**Steps:**

- Checks out your code
- Sets up Go 1.24.2 with dependency caching for faster builds
- Downloads all Go module dependencies
- Verifies the integrity of `go.mod` and `go.sum`
- Builds the main application (`cmd/doit`) into `bin/doit`
- Builds the database seeder (`cmd/seed`) into `bin/seed`

**Purpose:** Catches compilation errors early before wasting time on tests.

---

### 2. **Test Job**

Runs your entire test suite with comprehensive checks.

**Steps:**

- Sets up the same Go environment
- Downloads dependencies
- Runs all tests (`./...`) with:
  - `-v` - Verbose output to see each test
  - `-race` - Detects race conditions in concurrent code
  - `-coverprofile=coverage.out` - Generates coverage report
  - `-covermode=atomic` - Precise coverage tracking for concurrent tests
- Displays a coverage summary showing which functions are tested

**Purpose:** Ensures code correctness and identifies untested code paths.

---

### 3. **Code Generation Verification Job**

Ensures generated code is up-to-date and committed.

**Steps:**

- Installs `sqlc` (generates Go code from SQL queries)
- Installs `mockgen` (generates test mocks)
- Regenerates sqlc code from your SQL files
- Regenerates mocks using your script
- Checks if any files changed (`git status --porcelain`)
- **Fails if there are changes** - This means someone forgot to regenerate code after modifying SQL or interfaces

**Purpose:** Prevents bugs from stale generated code. Forces developers to commit regenerated code.

---

### 4. **Security Scan Job**

Analyzes code for security vulnerabilities.

**Steps:**

- Runs **Gosec** (Go Security Checker) which scans for common security issues like:
  - SQL injection vulnerabilities
  - Unsafe crypto usage
  - File permission issues
  - Hardcoded credentials
- Generates a SARIF report (standardized security format)
- Uploads results to GitHub's Security tab (visible in the repo's Security section)
- Uses `-no-fail` so it won't block the build, just reports issues

**Purpose:** Identifies security flaws before they reach production.

---

### 5. **Vulnerability Check Job**

Scans dependencies for known vulnerabilities.

**Steps:**

- Installs **govulncheck** (official Go vulnerability scanner)
- Checks all dependencies against the Go vulnerability database
- Uses `continue-on-error: true` so it won't block the build if vulnerabilities are found

**Purpose:** Alerts you to vulnerable dependencies (e.g., outdated libraries with CVEs).

---

## How They Work Together

```
PR Created/Updated
        ↓
┌───────┴───────────────────────────────────────┐
│  All 5 jobs run in parallel:                 │
│  ✓ Build (2-3 min)                           │
│  ✓ Test (3-5 min)                            │
│  ✓ Codegen verification (2-3 min)            │
│  ✓ Security scan (1-2 min)                   │
│  ✓ Vulnerability check (1-2 min)             │
└───────┬───────────────────────────────────────┘
        ↓
   All must pass ✓
        ↓
   PR can be merged
```

## Key Benefits

1. **Fast Feedback** - Jobs run in parallel, so you get results in ~5 minutes instead of 15+
2. **Comprehensive Checks** - Covers compilation, testing, security, and code generation
3. **Prevents Common Issues** - Catches forgotten code regeneration, race conditions, and security flaws
4. **Caching** - Go dependencies are cached between runs for speed
5. **Non-Blocking Security** - Security checks report issues but don't block development

## When Builds Fail

- **Build fails** → Syntax or compilation error
- **Test fails** → Failed test or race condition detected
- **Codegen fails** → You modified SQL/interfaces but didn't run `make sqlc` or `make generate-mocks`
- **Security/Vuln** → Review the Security tab in GitHub for details (won't block the build)

## Workflow Configuration

The workflow is defined in `go.yaml` and uses:

- **GitHub Actions**: Automation platform
- **Actions used**:
  - `actions/checkout@v4` - Checks out the repository
  - `actions/setup-go@v5` - Sets up Go environment
  - `securego/gosec@master` - Security scanner
  - `github/codeql-action/upload-sarif@v3` - Uploads security results

### Permissions

The workflow declares explicit permissions to follow the principle of least privilege:

```yaml
permissions:
  contents: read # Explicit read access to repository code
  security-events: write  # Required for uploading SARIF files
  actions: read          # Workflow metadata access
```

**Why these permissions are needed:**

- **`contents: read`** - Explicit read access to repository code. This is a standard permission for workflows.

- **`security-events: write`** - Essential for the Security Scan job to upload SARIF results to GitHub's Security tab. Without this, you'll get a "Resource not accessible by integration" error when trying to upload security scan results.

- **`actions: read`** - Allows the workflow to read metadata about itself and other workflow runs. This is a standard permission for workflows.

**Security Note:** By explicitly declaring permissions, we ensure the workflow only has access to what it needs, following security best practices. The default `GITHUB_TOKEN` permissions are restricted to only these specific scopes.

## Running Checks Locally

Before pushing code, you can run these checks locally:

```bash
# Build
make build

# Run tests with race detector
go test -v -race ./...

# Generate code
make sqlc
make generate-mocks

# Security scan
gosec ./...

# Vulnerability check
govulncheck ./...
```

This ensures your code will pass CI before pushing, saving time and CI resources.

---

This is a production-grade CI setup that ensures code quality, security, and consistency! 🚀
