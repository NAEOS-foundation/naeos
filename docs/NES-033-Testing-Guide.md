# NES-033 — Testing Guide

> Status: Draft
> Last Updated: 2026-07-22

Guide for running and writing tests in the NAEOS codebase.

---

## Running Tests

### All Tests

```bash
go test -race -count=1 -timeout 300s ./...
```

### Specific Package

```bash
go test -race -count=1 -timeout 60s ./internal/broker/...
go test -race -count=1 -timeout 60s ./cmd/naeos/...
```

### With Coverage

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Fuzz Testing

```bash
go test -fuzz=FuzzParseVersion -fuzztime=30s ./internal/migration/
go test -fuzz=FuzzParse -fuzztime=30s ./internal/specification/parser/
```

### Benchmarks

```bash
go test -bench=. -benchmem ./internal/compiler/...
```

### Integration/E2E Tests

Build-tagged with `//go:build integration`:

```bash
go test -tags=integration -race -count=1 -timeout 120s ./pkg/pipeline/ -run "TestEndToEnd"
```

---

## Test Structure

Tests use the standard `testing` package. No external test frameworks.

### Basic Pattern

```go
func TestSomething(t *testing.T) {
    result := DoSomething()
    if result != expected {
        t.Errorf("got %v, want %v", result, expected)
    }
}
```

### Table-Driven Tests

```go
func TestCompare(t *testing.T) {
    tests := []struct {
        name string
        a, b string
        want int
    }{
        {name: "a < b", a: "a", b: "b", want: -1},
        {name: "a > b", a: "b", b: "a", want: 1},
        {name: "equal",  a: "a", b: "a", want: 0},
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

### Constructor Tests

```go
func TestNewThing(t *testing.T) {
    t.Parallel()
    s := NewThing()
    if s == nil {
        t.Fatal("expected non-nil")
    }
}
```

### Temporary Directories

```go
func TestWritesFile(t *testing.T) {
    dir := t.TempDir()
    path := filepath.Join(dir, "output.txt")
    os.WriteFile(path, []byte("content"), 0644)
}
```

### Error Checking

```go
func TestFailsOnInvalidInput(t *testing.T) {
    _, err := Parse("{invalid")
    if err == nil {
        t.Fatal("expected error, got nil")
    }
}
```

### Concurrent Tests

```go
func TestConcurrentAccess(t *testing.T) {
    var wg sync.WaitGroup
    var count atomic.Int64
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            count.Add(1)
        }()
    }
    wg.Wait()
    if count.Load() != 10 {
        t.Errorf("expected 10, got %d", count.Load())
    }
}
```

---

## Test Files by Package

### cmd/naeos

| File | Tests |
|---|---|
| `main_test.go` | All CLI commands, output formats, flags |
| `cmd_test.go` | Version, Status, Health, Migrate commands |
| `marketplace_cmd_test.go` | Search, Install, Profile, Plugin, JSON output |
| `ai_cmd_test.go` | AI suggest, explain, compile, enrich |
| `perf_cmd_test.go` | Pool, Cache operations |
| `init_cmd_test.go` | Init command, custom output |
| `cloud_cmd_test.go` | Deploy, Plan, Status, Export |
| `gateway_cmd_test.go` | Gateway add-backend, status, rate-limit |
| `db_cmd_test.go` | Connect, Disconnect, List, Migrate |
| `security_cmd_test.go` | Audit, Secrets, Password, Sanitize, Validate |
| `workflow_cmd_test.go` | Create, Execute, List, Approve, Reject |
| `observability_cmd_test.go` | Trace, Log, Metrics, Dashboard, Status |
| `artifacts_cmd_test.go` | List, Info, Dedup, Summary |
| `doctor_test.go` | Doctor diagnostics |
| `e2e_test.go` | CLI end-to-end pipeline workflow |

### pkg/pipeline

| File | Tests |
|---|---|
| `pipeline_test.go` | Run produces result, injected parser, output dir, config loading |
| `pipeline_integration_test.go` | Integration tests with real specs |
| `pipeline_e2e_test.go` | E2E: minimal, multi-lang, validate, full-spec, modules, services, empty |
| `pipeline_coverage_test.go` | Pipeline graph, registry, hooks, kernel services |

### pkg/kernel

| File | Tests |
|---|---|
| `kernel_test.go` | Register/resolve, start/stop, event bus, telemetry |

### internal/migration

| File | Tests |
|---|---|
| `migration_test.go` | ParseVersion, Version comparison, Planner Plan/Migrate, Engine Migrate/Plan, FormatMigrationPlan, builtin transforms, edge cases (97.9% coverage) |
| `fuzz_test.go` | `FuzzParseVersion` |

### internal/broker

| File | Tests |
|---|---|
| `broker_test.go` | Redis/RabbitMQ/Kafka stubs, InMemoryBroker, ConnectionPool, DeadLetter, Metrics, Middleware, MessageFilter, Retry |
| `concurrency_test.go` | InMemoryBroker concurrent, ConnectionPool concurrent, Manager concurrent, Metrics concurrent |

### internal/pluginsdk/scaffold

| File | Tests |
|---|---|
| `scaffold_test.go` | GoMod, NaeosYAML, PluginGo, WasmMainGo, PluginTestGo, Makefile, CIYML, Readme, WriteAll, WriteAllError, FilesDifferentName, FilesWithoutWASM |

### internal/specification/parser

| File | Tests |
|---|---|
| `parser_test.go` | JSON/YAML parse, invalid input, full spec |
| `fuzz_test.go` | `FuzzParse`, `FuzzParseYAMLNode`, `FuzzVariableResolver`, `FuzzSchemaVersionParse`, `FuzzValidateModules` |

### internal/cache

| File | Tests |
|---|---|
| `cache_test.go` | Set/Get/Delete, TTL, Clear, concurrent access |
| `cache_new_test.go` | MatchPattern, NewCache, serialization |

### internal/scheduler

| File | Tests |
|---|---|
| `scheduler_test.go` | Schedule intervals, job execution, Start/Stop, multiple jobs |

### internal/distributed

| File | Tests |
|---|---|
| `distributed_test.go` | Worker pool, task distribution, results aggregation, error handling |

### internal/events

| File | Tests |
|---|---|
| `bus_test.go` | Pub/sub, multiple subscribers, unsubscribe, topics |

### internal/workflow

| File | Tests |
|---|---|
| `workflow_test.go` | Create workflow, execute steps, approve/reject, list |
| `workflow_new_test.go` | NewWorkflow, NewStep validation |
| `workflow_ext_test.go` | External package tests |

---

## Writing New Tests

### 1. Create Test File

Place `*_test.go` in the same package:

```go
// internal/myPackage/myfile_test.go
package myPackage

import "testing"

func TestMyFunction(t *testing.T) {
    // ...
}
```

### 2. Use Test Skeleton Generator

```bash
go run ./cmd/gentest/ -pkg ./internal/my-package/ -o ./internal/my-package/new_test.go
```

### 3. Use Subtests

```go
func TestParse(t *testing.T) {
    t.Run("valid input", func(t *testing.T) {
        // ...
    })
    t.Run("empty input", func(t *testing.T) {
        // ...
    })
}
```

### 4. Use Helper Functions

```go
func setupTest(t *testing.T) *MyStruct {
    t.Helper()
    return NewMyStruct()
}
```

---

## Test Conventions

| Convention | Description |
|---|---|
| File naming | `*_test.go` in same package |
| External tests | `*_ext_test.go` with `package xxx_test` |
| Function naming | `TestFunctionName` |
| Subtest naming | `t.Run("description", ...)` |
| Table-driven var | Always `tt` |
| Temp dirs | `t.TempDir()` for auto-cleanup |
| Error expected | `if err == nil { t.Fatal(...) }` |
| No external deps | Use standard `testing` package only |
| Table-driven | Use `[]struct{...}` pattern for multiple cases |
| Concurrent tests | Use `sync.WaitGroup` + `atomic` |
| t.Parallel | Use for independent tests |
| Race detector | Always run tests with `-race` |

---

## CI Integration

Tests run automatically via GitHub Actions:

```yaml
# .github/workflows/ci.yml
- name: Build
  run: go build ./cmd/naeos/
- name: Test with race detector
  run: go test -race -count=1 -timeout 300s ./...
- name: Fuzz
  run: |
    go test -fuzz=FuzzParseVersion -fuzztime=30s ./internal/migration/
    go test -fuzz=FuzzParse -fuzztime=30s ./internal/specification/parser/
- name: Coverage
  run: go test -race -count=1 -coverprofile=coverage.out -covermode=atomic ./...
- uses: codecov/codecov-action@v5
```

Coverage thresholds: `.codecov.yml` — project 60%, patch 80%.
