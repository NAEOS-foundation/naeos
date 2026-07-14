# NAEOS Database Development Plan

**15 tasks across 5 phases** — production-ready database layer

---

## Current State

| Aspect | Status |
|--------|--------|
| PostgreSQL (real) | ✅ pgx/v5, transactions, migrations |
| MySQL | ❌ Stub only (no driver, no real SQL) |
| SQLite | ❌ Stub only (no driver, no real SQL) |
| Context support | ❌ Interface methods lack `context.Context` |
| Migration `Down` | ❌ Never executes `Down` SQL |
| Connection pool | ⚠️ Hardcoded (25 open, 5 idle) |
| Config | ⚠️ Minimal (7 fields, no pool/tls) |
| Retry/health | ❌ None |
| Logging/observability | ❌ None |
| API integration | ❌ Database not used by API server |
| Test coverage | 98.7% (mostly stubs) |

---

## Phase 1: Core Interface & PostgreSQL Hardening (4 tasks)

### Task 1: Add Context Support to Database Interface
**Files:** `internal/database/database.go`, `internal/database/postgres_real.go`

Current interface methods don't accept `context.Context`. This is a critical gap for production use (timeouts, cancellation, tracing).

**Changes:**
1. Update `Database` interface — add `Context` variants:
   ```go
   ExecContext(ctx context.Context, query string, args ...any) (Result, error)
   QueryContext(ctx context.Context, query string, args ...any) ([]Row, error)
   QueryRowContext(ctx context.Context, query string, args ...any) (Row, error)
   BeginTx(ctx context.Context, opts *sql.TxOptions) (Transaction, error)
   MigrateContext(ctx context.Context, migrations []Migration) error
   RollbackContext(ctx context.Context, version int) error
   ```
2. Update `Transaction` interface — add context variants
3. Update `RealPostgreSQL` to use `ExecContext`, `QueryContext`, etc. instead of non-context versions
4. Update mock adapters to accept (and ignore) context
5. Keep old methods as wrappers calling context versions with `context.Background()` for backward compat
6. Update tests

### Task 2: Fix Migration Rollback (Execute Down SQL)
**Files:** `internal/database/postgres_real.go`, `internal/database/database.go`

**Bug:** `Rollback(version)` deletes `_migrations` records but never executes `Down` SQL.

**Changes:**
1. In `RealPostgreSQL.Rollback()` (postgres_real.go:186-225):
   - After deleting migration record, execute `m.Down` SQL in a transaction
   - Add error handling — if `Down` fails, rollback the transaction and keep the record
2. Add `down` column to `_migrations` table (or store in separate table)
3. Update mock adapter `Rollback()` to also execute `Down` SQL
4. Add tests: `TestRollbackExecutesDownSQL`

### Task 3: Make Connection Pool Configurable
**Files:** `internal/database/database.go`, `internal/database/postgres_real.go`

Pool sizes are hardcoded (MaxOpenConns=25, MaxIdleConns=5).

**Changes:**
1. Extend `Config` struct:
   ```go
   type Config struct {
       // existing fields...
       MaxOpenConns    int
       MaxIdleConns    int
       ConnMaxLifetime time.Duration
       ConnMaxIdleTime time.Duration
   }
   ```
2. In `RealPostgreSQL.Connect()`, use `Config` values with defaults:
   - Default MaxOpenConns: 25
   - Default MaxIdleConns: 5
   - Default ConnMaxLifetime: 5 minutes
3. Update CLI flags in `db_cmd.go` to expose pool settings
4. Update tests

### Task 4: Add Config Validation
**Files:** `internal/database/database.go`

`Config` has no validation — empty Host, invalid Port, etc. are silently accepted.

**Changes:**
1. Add `func (c *Config) Validate() error` method:
   - Required fields: Host, Port (>0), User, Database
   - SSLMode validation: must be one of `disable`, `require`, `verify-ca`, `verify-full`
   - Timeout must be positive if set
2. Call `Validate()` in `Connect()` before opening connection
3. Add tests: `TestConfigValidation`

---

## Phase 2: MySQL & SQLite Real Adapters (4 tasks)

### Task 5: Implement Real MySQL Adapter
**Files:** `internal/database/mysql_real.go` (new)

**Changes:**
1. Add MySQL driver dependency: `go get github.com/go-sql-driver/mysql`
2. Create `mysql_real.go` with build tag `//go:build !nosql`
3. Implement `RealMySQL` struct wrapping `*sql.DB`:
   - `Connect()` — DSN format: `user:password@tcp(host:port)/database?parseTime=true`
   - `Exec()`, `Query()`, `QueryRow()` — use context variants
   - `Begin()`, `Migrate()`, `Rollback()` — same patterns as PostgreSQL
   - Migration tracking table: `_migrations` (same schema)
   - `Rollback()` — execute `Down` SQL (fix from Task 2)
4. Add `mysql_real_test.go` with connection failure tests
5. Register in `Manager` factory

### Task 6: Implement Real SQLite Adapter
**Files:** `internal/database/sqlite_real.go` (new)

**Changes:**
1. Add SQLite driver: `go get modernc.org/sqlite` (pure Go, no CGO)
2. Create `sqlite_real.go` with build tag `//go:build !nosql`
3. Implement `RealSQLite` struct wrapping `*sql.DB`:
   - `Connect()` — DSN format: `file:path?mode=rw`
   - `Exec()`, `Query()`, `QueryRow()` — use context variants
   - `Begin()`, `Migrate()`, `Rollback()` — same patterns
   - Migration tracking table: `_migrations`
   - SQLite-specific: WAL mode, foreign keys, busy timeout
4. Add `sqlite_real_test.go`
5. Register in `Manager` factory

### Task 7: Create Shared Mock Base (Reduce Duplication)
**Files:** `internal/database/database.go`

Current state: PostgreSQL/MySQL/SQLite stubs are ~100 lines each, nearly identical.

**Changes:**
1. Create `BaseDatabase` struct with shared mock logic:
   ```go
   type BaseDatabase struct {
       mu        sync.RWMutex
       connected bool
       tables    map[string][]Row
       migrations []Migration
       lastVersion int
   }
   ```
2. Embed `BaseDatabase` in PostgreSQL/MySQL/SQLite stubs
3. Each stub only overrides `Name()` and `New()` (constructor)
4. Reduce code from ~300 lines (3 adapters) to ~50 lines (base + 3 thin wrappers)
5. Update tests (should be mostly unchanged since interface is preserved)

### Task 8: Add Database Factory Function
**Files:** `internal/database/database.go`

**Changes:**
1. Add factory function:
   ```go
   func New(driver string) Database {
       switch driver {
       case "postgresql", "postgres":
           return NewRealPostgreSQL()
       case "mysql":
           return NewRealMySQL()
       case "sqlite":
           return NewRealSQLite()
       case "mock-postgresql":
           return NewPostgreSQL()
       // etc.
       default:
           return nil
       }
   }
   ```
2. Add `func NewFromConfig(driver string, config *Config) (Database, error)` that creates + connects
3. Update CLI `db_cmd.go` to use factory
4. Add tests

---

## Phase 3: Production Hardening (3 tasks)

### Task 9: Add Retry Logic with Exponential Backoff
**Files:** `internal/database/retry.go` (new), `internal/database/postgres_real.go`

**Changes:**
1. Create `retry.go` with retry helper:
   ```go
   func WithRetry(ctx context.Context, maxRetries int, baseDelay time.Duration, fn func(ctx context.Context) error) error
   ```
2. Exponential backoff: 100ms, 200ms, 400ms, ... up to max
3. Only retry on transient errors (connection refused, timeout, EOF)
4. Apply to `Connect()` and `Ping()`
5. Add tests: `TestWithRetry`, `TestWithRetryMaxRetries`, `TestWithRetryContextCancelled`

### Task 10: Add Query Logging & Metrics
**Files:** `internal/database/logging.go` (new), `internal/database/postgres_real.go`

**Changes:**
1. Create `loggingDatabase` wrapper (decorator pattern):
   ```go
   type loggingDatabase struct {
       inner Database
       logger *slog.Logger
   }
   ```
2. Log all `Exec`, `Query`, `Begin` calls with duration, query (truncated), args count
3. Add slow query logging (> 1 second)
4. Integrate with existing `internal/telemetry/` for metrics:
   - `db_query_duration_seconds` histogram
   - `db_query_total` counter by operation type
   - `db_connections_active` gauge
5. Wrap real adapters: `NewLoggingDatabase(inner Database) Database`

### Task 11: Add Health Check & Reconnection
**Files:** `internal/database/health.go` (new), `internal/database/postgres_real.go`

**Changes:**
1. Add `HealthCheck() error` method to `Database` interface
2. Implement in `RealPostgreSQL`: `Ping()` + check `db.Stats()` (open connections, in-use, idle)
3. Add periodic health check goroutine:
   - Configurable interval (default 30s)
   - Automatic reconnection on failure
   - Circuit breaker: after 3 failures, stop trying for 60s
4. Add `Stats() DBStats` method returning connection pool statistics
5. Expose health check in API: `/api/v1/health` includes `database: healthy/degraded/unhealthy`
6. Add tests

---

## Phase 4: Integration & Migration (2 tasks)

### Task 12: Integrate Database with API Server
**Files:** `internal/api/server.go`, `cmd/naeos/db_cmd.go`

Currently the API server stores everything in memory. Database integration enables persistence.

**Changes:**
1. Add `db Database` field to API `Server` struct (optional, nil if no DB configured)
2. Store pipeline runs in DB instead of `s.pipelines []pipelineRun`
3. Store deployments in DB instead of `s.deployments []cloudDeployment`
4. Add `--database` flag to `naeos api` command
5. Create schema migrations for pipeline_runs and deployments tables
6. Update API handlers to use DB when available, fall back to in-memory
7. Add DB health check to `/api/v1/health`

### Task 13: Add Migration File System
**Files:** `internal/database/migrations.go` (new)

Currently migrations are passed as in-memory structs. Need file-based system.

**Changes:**
1. Create migration file format: `000001_create_tables.up.sql` / `.down.sql`
2. Add `LoadMigrations(dir string) ([]Migration, error)` function
3. Parse version number from filename prefix
4. Add checksum calculation for integrity verification
5. Add `--dry-run` flag to `db migrate` CLI command
6. Add `--target` flag to migrate to specific version
7. Add migration status command: show applied vs pending
8. Add tests

---

## Phase 5: Documentation & Polish (2 tasks)

### Task 14: Update Documentation
**Files:** `docs/database.md` (new), `docs/openapi.yaml`

**Changes:**
1. Create `docs/database.md`:
   - Architecture overview (interface, adapters, pool)
   - Configuration reference (all Config fields)
   - Migration guide (file format, dry-run, rollback)
   - CLI commands reference
   - Connection pooling best practices
   - Troubleshooting guide
2. Update OpenAPI spec: add DB health check to `/api/v1/health` response
3. Update `README.md`: mention database support

### Task 15: Final Verification & Coverage
**Files:** All test files

**Changes:**
1. Run full test suite: `go test -race ./...`
2. Verify coverage: `go test -coverprofile=cover.out ./internal/database/...`
3. Run `golangci-lint` on database package
4. Verify all new adapters compile: `go build ./...`
5. Verify build tags work: `go test -tags nosql ./internal/database/...`

---

## Execution Order

```
Phase 1 (Core):     Task 1 → Task 2 → Task 3 → Task 4
Phase 2 (Adapters): Task 7 → Task 8 → Task 5 → Task 6
Phase 3 (Hardening): Task 9 → Task 10 → Task 11
Phase 4 (Integration): Task 12 → Task 13
Phase 5 (Docs):     Task 14 → Task 15
```

## Expected Impact

| Metric | Current | v1.2.0 Target |
|--------|---------|---------------|
| Real adapters | 1 (PostgreSQL) | 3 (PG + MySQL + SQLite) |
| Context support | ❌ | ✅ All methods |
| Migration Down | ❌ Never executes | ✅ Executes with rollback |
| Config options | 7 fields | 11 fields |
| Retry logic | ❌ | ✅ Exponential backoff |
| Query logging | ❌ | ✅ With slow query detection |
| Health checks | ❌ | ✅ With circuit breaker |
| API integration | ❌ In-memory | ✅ DB-backed (optional) |
| Migration files | In-memory only | File-based with checksums |
| Code duplication | ~300 lines (3 stubs) | ~50 lines (base + wrappers) |
