# Contributing to NAEOS

Thank you for your interest in contributing to NAEOS!

## Table of Contents

- [Development Setup](#development-setup)
- [Development Workflow](#development-workflow)
- [Code Style](#code-style)
- [Testing Guide](#testing-guide)
- [Project Structure](#project-structure)
- [Documentation](#documentation)
- [Pull Request Process](#pull-request-process)

## Development Setup

### Prerequisites

- Go 1.25 or later
- Git
- golangci-lint (required — runs in CI and blocks PRs)

### Quick Start

```bash
git clone https://github.com/NAEOS-foundation/naeos.git
cd naeos
go mod tidy
go build ./cmd/naeos/
go test ./...
```

## Development Workflow

### Build

```bash
go build ./cmd/naeos/
```

### Test

```bash
# All tests with race detector
go test -race -count=1 -timeout 300s ./...

# Specific package
go test -race -count=1 -timeout 60s ./internal/broker/...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Integration-tagged E2E tests
go test -tags=integration -race -count=1 -timeout 120s ./pkg/pipeline/ -run "TestEndToEnd"
```

### Lint

```bash
golangci-lint run ./...
go vet ./...
```

### Fuzz

```bash
go test -fuzz=FuzzParseVersion -fuzztime=30s ./internal/migration/
go test -fuzz=FuzzParse -fuzztime=30s ./internal/specification/parser/
```

### Benchmarks

```bash
go test -bench=. -benchmem ./internal/compiler/...
go test -bench=. -benchmem ./internal/generation/adapters/...
```

### Generate CLI Docs

```bash
go run ./cmd/naeos/ docsgen
# Output: docs/cli/*.md
```

### Generate Test Skeletons

```bash
go run ./cmd/gentest/ -pkg ./internal/my-package/ -o ./internal/my-package/new_test.go
```

### E2E Test Script

```bash
./scripts/e2e_test.sh
```

## Code Style

### Go Conventions

- Format: `gofmt` + `goimports` with local prefix `github.com/NAEOS-foundation/naeos`
- Import order: stdlib → third-party → internal
- Error handling: return errors, no panics
- Thread safety: use `sync.Mutex`/`sync.RWMutex` + `atomic.Int64`
- Constructor pattern: `New*()` returns struct pointer (not interface)
- Config pattern: constructor takes config struct, validates and stores copy
- Logging: structured via middleware, not inline
- No external logging or assertion libraries

### Naming

| Convention | Example |
|---|---|
| File names | `snake_case.go` |
| Test files | `*_test.go` in same package |
| External test files | `*_ext_test.go` with `package xxx_test` |
| Table-driven loop var | Always `tt` |

## Testing Guide

### Test Patterns

**Basic test:**
```go
func TestFunction(t *testing.T) {
    result := DoSomething()
    if result != expected {
        t.Errorf("got %v, want %v", result, expected)
    }
}
```

**Table-driven (use `tt` as loop variable):**
```go
func TestCompare(t *testing.T) {
    tests := []struct {
        name string
        a, b string
        want int
    }{
        {name: "a < b", a: "a", b: "b", want: -1},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Compare(tt.a, tt.b)
            if got != tt.want {
                t.Errorf("Compare(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
            }
        })
    }
}
```

**Constructor test:**
```go
func TestNewThing(t *testing.T) {
    t.Parallel()
    s := NewThing()
    if s == nil {
        t.Fatal("expected non-nil")
    }
}
```

**Error expected:**
```go
_, err := Parse("invalid")
if err == nil {
    t.Fatal("expected error")
}
```

**Temp directories:**
```go
dir := t.TempDir()
path := filepath.Join(dir, "file.txt")
os.WriteFile(path, []byte("data"), 0644)
```

**Concurrent test:**
```go
func TestConcurrentAccess(t *testing.T) {
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            // concurrent operations
        }()
    }
    wg.Wait()
}
```

### Test Files Added

| Package | File | Tests |
|---|---|---|
| `internal/pluginsdk/scaffold` | `scaffold_test.go` | GoMod, NaeosYAML, PluginGo, WasmMainGo, PluginTestGo, Makefile, CIYML, Readme, WriteAll, WriteAllError, FilesDifferentName, FilesWithoutWASM |
| `internal/migration` | `migration_test.go` | ParseVersion, VersionLess, PlannerPlan, PlannerMigrate, MigrationEngine, FormatMigrationPlan, builtin transforms + fuzz |
| `internal/broker` | `concurrency_test.go` | InMemoryBrokerConcurrentPublishSubscribe, ConnectionPoolConcurrent, MetricsConcurrent, DeadLetterChannelConcurrent |
| `cmd/naeos` | `cmd_test.go` | Version, Status, Health, MigratePlan, MigrateRun, MigrateVersions |
| `cmd/naeos` | `marketplace_cmd_test.go` | Search, Install, ProfileList, PluginList, Help, JSONOutput |
| `pkg/pipeline` | `pipeline_e2e_test.go` | MinimalSpec, WithLanguages, ValidateOnly, FullSpec, ModuleDependencies, Services, EmptySpec |

### Fuzz Targets

| Package | Target |
|---|---|
| `internal/migration` | `FuzzParseVersion` |
| `internal/specification/parser` | `FuzzParse`, `FuzzParseYAMLNode`, `FuzzVariableResolver`, `FuzzSchemaVersionParse`, `FuzzValidateModules` |

### Coverage Configuration

See `.codecov.yml`:
- Project target: 60% (2% threshold)
- Patch target: 80% (5% threshold)

## Project Structure

```
naeos/
├── .github/workflows/   # CI pipeline
├── cmd/
│   ├── naeos/           # Main CLI (package main)
│   └── gentest/         # Test skeleton generator
├── pkg/
│   ├── kernel/          # Core kernel
│   ├── pipeline/        # Pipeline orchestrator
│   └── config/          # Configuration
├── internal/
│   ├── broker/          # Message broker (Redis, RabbitMQ, Kafka)
│   ├── cache/           # In-memory cache
│   ├── compiler/        # NEIR compiler
│   ├── database/        # Database layer (PostgreSQL, MySQL, SQLite)
│   ├── generation/      # Code generation engine + adapters
│   ├── migration/       # Schema migration planner + engine
│   ├── neir/            # NEIR model types + builder + validator
│   ├── pluginsdk/       # Plugin SDK (scaffold, sandbox, wasm)
│   ├── scheduler/       # Task scheduler
│   └── specification/   # Parser, normalizer, resolver
├── docs/
│   ├── cli/             # Auto-generated CLI docs
│   └── adr/             # Architecture Decision Records
├── scripts/             # Utility scripts
└── AGENTS.md            # AI agent instructions
```

## Documentation

### Documentation Map

| Doc | Description |
|---|---|
| `docs/cli/*.md` | Auto-generated CLI reference (`naeos docsgen`) |
| `docs/adr/*.md` | Architecture Decision Records |
| `docs/NES-*.md` | NAEOS Engineering Specifications |
| `CONTRIBUTING.md` | This file |

### Generating CLI Docs

Docs are auto-generated from cobra command definitions:

```bash
go run ./cmd/naeos/ docsgen
git add docs/cli/
git commit -m "docs: update CLI reference"
```

## Pull Request Process

1. **Branch** from `main` with a descriptive name
2. **Commit** with clear messages (imperative mood, <72 chars)
3. **Test** — run full suite: `go test -race -count=1 -timeout 300s ./...`
4. **Lint** — run: `golangci-lint run ./... && go vet ./...`
5. **PR** — describe the change, why it's needed, and how it was tested

### Commit Message Format

```
<type>: <short description>

<optional body>

Closes #<issue>
```

Types: `feat`, `fix`, `test`, `docs`, `refactor`, `chore`, `ci`

### PR Checklist

- [ ] Code compiles: `go build ./cmd/naeos/`
- [ ] All tests pass: `go test -race -count=1 ./...`
- [ ] Linter passes: `golangci-lint run ./...`
- [ ] New code has tests
- [ ] Documentation updated if needed
- [ ] No breaking changes without discussion
