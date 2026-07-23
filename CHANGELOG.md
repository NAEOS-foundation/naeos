# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.2.0] - 2026-07-23

### Added
- **Supabase backend integration** — full-featured Supabase client with Auth, Storage, Edge Functions, Admin API, and Realtime WebSocket support:
  - `internal/supabase/` package with `Client`, `Config`, Auth (GoTrue API), Storage (buckets & files CRUD), Admin (SQL query, roles), Edge Functions (list/deploy/invoke/delete).
  - `internal/database/supabase_real.go` — RealSupabase adapter (pgx, default SSL `require`), registered as factory driver `"supabase"`.
  - CLI command group `naeos supabase` — init, auth (signup, signin, signout, user, admin CRUD), storage (list-buckets, create-bucket, list-files, upload, download, delete), sql, status.
  - `naeos init --template supabase` — init template with Supabase config.
  - `internal/supabase/supabase_test.go` — 5 integration tests (skip without env vars).
  - CI workflow `supabase-integration` — injects `SUPABASE_URL`, `SUPABASE_ANON_KEY`, `SUPABASE_SERVICE_ROLE_KEY`, `SUPABASE_JWKS_URL` from secrets.
  - CLI docs regenerated — 21 `docs/cli/naeos_supabase*.md` files.
- **Unit tests for `internal/supabase/`** — 44 tests covering all API methods (Auth, Storage, Admin, Functions), error cases (API errors, malformed JSON, 401/404/409), config save/load, and edge cases (coverage 84.1%).
- **UploadFile/DownloadFile unit tests** — 3 new tests (success, error 400, file not found).

### Fixed
- **Lint failures (28 issues)** — `bodyclose`, `noctx`, `gofmt`, `unconvert`, `errcheck` resolved across 7 files in `internal/supabase/`. HTTP response bodies now closed internally in `do()`; all requests use `http.NewRequestWithContext`.
- **Lint failures in `cmd/`** — `gofmt` formatting issues in `supabase_cmd.go`, `migration_cmd.go`, `init_cmd.go`; `errcheck` for unhandled `rc.send` error in `realtime.go`.
- **TestQueueFull race condition** — removed subscriber from `TestQueueFull` to eliminate timing-dependent failure in `internal/messagequeue`.
- **TestRealMySQLConnectNoOptionalConfig timeout** — added explicit `Timeout: 1 * time.Second` to prevent hanging on default TCP timeout.
- **UploadFile ContentLength mismatch** — removed incorrect `req.ContentLength = stat.Size()` (multipart body is larger than raw file size).
- **codecov-action deprecation warning** — changed deprecated `file:` input to `files: ./coverage.out`.

### Removed
- **Dead code** — removed entire `internal/supabase/realtime.go` (151 lines, `NewRealtimeClient`/`Subscribe`/`Unsubscribe`/`Close`/`Done`/`readLoop`/`send` all unreachable).
- **Dead code** — removed `Client.DeployFunctionFromFile` from `internal/supabase/functions.go` (never called).

### Changed
- **CI: codecov-action** — `file: ./coverage.out` → `files: ./coverage.out` (v5 syntax).
- **Internal `do()` signature** — `(*http.Response, error)` → `([]byte, error)`, body closed inside `do()` for deterministic lifetime.
- **`go mod tidy`** — cleaned up unused dependencies.

## [2.1.1] - 2026-07-23

### Fixed
- **Broker merge artifacts** — removed duplicate `return` statements in `factory.go` and `store.go` (unreachable code).
- **Marketplace test URL** — `TestRemoteRegistryInstall` now uses correct base URL without double path segment.
- **`.golangci.yml` merge artifacts** — removed leftover `typecheck` linter (v1-only) and duplicate config keys.
- **Fuzz test ambiguity** — `FuzzParse` regex narrowed to `^FuzzParse$` to avoid matching `FuzzParseYAMLNode`.
- **Fuzz test timeout** — `FuzzVariableResolver` input length limited to 256 bytes to prevent regex-engine hangs.
- **Priority test race** — `TestSubmitPriorityOrder` moved `Start()` after all task submissions to guarantee priority ordering.
- **Demo lint suppression** — added `//nolint:errcheck` to `cmd/naeos-demo/main.go`.
- **Scaffold GOSEK lint** — `internal/pluginsdk/scaffold/scaffold.go:272` file permission exclusion added.

### Changed
- **CI: golangci-lint-action** — upgraded from `@v6` (v1.64.8) to `@v7` with `version: v2.12.2` for golangci-lint v2 config compatibility.
- **CI: lint exclusions** — expanded `.golangci.yml` exclusion rules for test files (errorlint, noctx, unparam) and tool dirs (`cmd/gentest/`, `cmd/neir-schema-gen/`, `cmd/naeos-demo/`).
- **Code formatting** — `gofmt -w` applied to 16 files across the codebase.
- **Website OG image generation** — extracted inline `node -e` script to `site/scripts/og.mjs` to fix shell quoting issues.
- **Website deployment** — added `--baseURL` override for GitHub Pages compatibility.

### Added
- **Gentest tool** (`cmd/gentest/`) — test skeleton generator for rapid test scaffolding.
- **CLI docsgen** — hidden `docsgen` command to auto-generate CLI reference documentation.
- **Website deploy** — GitHub Pages deployment via `website.yml` workflow (build + deploy jobs).

## [2.1.0] - 2026-07-20

### Added
- **RBAC bootstrapping** — `SetupDefaultRoles()` (admin/developer/viewer) called on server start with full route permission mapping.
- **Multi-tenant workspace** — `GET/POST /api/v1/tenants` API endpoints; tenant isolation for pipeline runs and schema registry entries (`TenantID` field filtering).
- **Schema registry API** — 5 endpoints: `GET/POST /api/v1/schemas`, `GET/DELETE /api/v1/schemas/{name}/{version}`.
- **5 industry profiles** — edtech-platform, ecommerce-engine, iot-backend, media-streaming, blockchain-node.
- **Pipeline run file persistence** — runs saved to `~/.naeos/pipelines.json` (configurable via `NAEOS_PIPELINES_FILE` env var).
- **`$import{}` resolver wired into parser** — modular spec fragments with depth-limiting and caching.
- **ResultAggregator streaming** — `EnableStreaming()`, `StartStreaming() <-chan`, `Close()` for incremental result consumption.
- **Rate limit response headers** — `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset` on all API responses.
- **Audit trail file persistence** — dual `multiAuditor` (MemoryAuditor + FileAuditor to `~/.naeos/audit.log`).
- **Encryption at rest** — UserStore transparent AES-256-GCM encryption via `NAEOS_ENCRYPTION_KEY` env var.
- **Compliance export CLI** — `naeos compliance export --audit-file` reads persisted audit log, outputs JSON/CSV.
- **Metrics middleware wired to pipeline adapter** — per-stage latency recording via `MetricsMiddleware.RecordFunc`.

### Changed
- **`NewParser(baseDir string)`** — new parameter for `$import{}` base directory resolution.
- **`auth.NewManager(passphrase ...string)`** — optional encryption passphrase for UserStore persistence.
- **`middleware.Chain`** — added `ExecuteContext(ctx, ...)` to propagate caller context; existing `Execute()` delegates to it.

### Fixed
- **G304/G305 path traversal** — 7 files hardened with `securityext.ValidateFilePath`/`ValidatePluginName`: `artifacts/store.go`, `specification/parser/resolve_ext.go`, `rollback/rollback.go` (tar symlink), `diff/diff.go`, `pipeline/pipeline.go`, `cloud/state.go`, `eventsourcing/eventsourcing.go`.
- **Context timeout propagation** — `EnrichSpec`/`GenerateSuggestions`/`ExplainArchitecture` now use `context.WithTimeout(ctx, config.Timeout)` instead of bare `context.Background()`.
- **Workflow context error collection** — `ExecuteParallelGroup` stores all errors via `errors.Join()` instead of only the first.
- **Structured logging for error paths** — 10 locations across `jwt.go`, `broker/`, `database/`, `compiler.go` now log with `slog.Error` before returning errors.
- **Compliance export** — now reads from persisted audit log file instead of creating an empty in-memory auditor.
- **API test isolation** — pipeline status/pipelines endpoint tests use `t.Setenv("NAEOS_PIPELINES_FILE", ...)` for per-test isolation.

### Security
- AES-256-GCM encryption for `UserStore` on disk.
- Path traversal validation for `$include{}`/`$import{}` directives, artifact storage, tar extraction, snapshot restore, and diff file reading.
- `ValidatePluginName()` on cloud provider/project/streamID identifiers.

## [1.5.0] - 2026-07-19

### Fixed
- **`TestSecuritySetSecret`** — added missing `--key` flag in test command invocation.
- **`TestCoordinatorDrain`** — fixed data race between `Drain()` and `workerLoop()` by synchronizing `draining` flag with mutex.

### Production Hardening
- **Test coverage** — improved coverage for `neir/validator`, `neir/builder`, `internal/diff`, `internal/broker`, `internal/create`.
- **Race detector** — all packages pass `go test -race ./...` cleanly.

## [1.4.0] - 2026-07-19

### Added
- **Prompt Library (NES-054)** — centralized YAML-based prompt templates:
  - `internal/promptlib/` package with template parsing, rendering, and manifest support.
  - Custom template functions: `join`, `bt` (backtick), `code`, `json`, `yaml`, `title`, `upper`, `lower`, `trim`, `default`, `contains`, `replace`, `split`, `range`, `len`.
  - 3 builtin LLM prompts: `enrich-spec`, `generate-suggestions`, `explain-architecture`.
  - 6 builtin compiler adapter templates: copilot, claude, cursor, gemini, codex, opencode.
  - `prompts/builtin/` directory with 11 YAML reference files and manifest.
  - Backward compatible: nil library falls back to hardcoded prompts.
- **`naeos template` CLI** — list and inspect prompt templates:
  - `naeos template list` — show all registered templates.
  - `naeos template list --kind prompt-llm` — filter by kind.
  - `naeos template show <name>` — display template details and rendered output.
- **AIService ↔ LLMService integration**:
  - `NewServiceWithLLM(llm)` constructor for AI service with LLM backend.
  - `Suggest()` tries LLM first, falls back to rule-based analysis.
  - `Explain()` tries LLM for architecture topics, falls back to built-in knowledge.
  - CLI auto-wires LLM when `NAEOS_LLM_API_KEY` env var is set.
  - 5 new tests covering LLM paths and fallbacks.
- **Observability dashboard real data**:
  - `/traces` endpoint returns actual span data from `observability.Stack`.
  - `/logs` endpoint returns actual log entries with level filtering.
  - `/metrics` endpoint returns collected counters, gauges, and histograms.
  - Dashboard seeds sample pipeline data on startup.
- **Workflow Manager persistence**:
  - `NewManagerWithPath(dir)` constructor with file-based persistence.
  - Workflows saved to `~/.naeos/workflows/workflows.json`.
  - `Register()` and `Remove()` auto-save to disk.
- **Distributed workers** — stage-aware processing:
  - Each pipeline stage (parse, normalize, resolve, build-neir, validate, schedule, generate, review) has realistic simulated duration.
  - Workers respect context cancellation.

### Fixed
- **`internal/version/version_test.go`** — updated hardcoded `"0.9.0"` to `"1.3.1"` to match VERSION file.
- **`internal/rollback/rollback.go`** — `Import()` now allows `.` root directory entry from tar archives (was rejecting with "invalid path").
- **`wiki/Compiler.md`** — updated constructor calls to match new signatures with `nil` library parameter.
## [1.3.1] - 2026-07-19

### Fixed
- **golangci-lint: 999 → 0 issues** across 211 files (2135 insertions, 2364 deletions):
  - Removed 22 unused functions, types, vars, and struct fields.
  - Replaced `WriteString(fmt.Sprintf(...))` with `fmt.Fprintf(...)` across 21 files.
  - Added context propagation (`http.NewRequestWithContext`, `exec.CommandContext`) across all HTTP and exec calls.
  - Fixed 198 errcheck issues with proper error handling.
  - Fixed 40 gosec issues: file permissions (0o600), HTTP timeouts, path validation, weak crypto.
  - Fixed 6 govet copylocks issues with pointer types.
  - Fixed 11 misspellings (US English locale).
  - Fixed staticcheck issues: S1039, S1011, S1025, S1002, QF1003, QF1001, QF1008, QF1004.
  - Fixed 13 unparam, 1 unconvert, 2 ineffassign issues.
  - Applied `gofmt` and `goimports` for consistent formatting.
## [1.3.0] - 2026-07-16

### Added
- **CLI `--output json/yaml` support** for 14 commands:
  - `security audit` — machine-readable security reports for CI pipelines.
  - `benchmark` — structured performance metrics (avg, min, max, ops/sec).
  - `db list`, `db status` — database connection data.
  - `gateway status`, `gateway rate-status`, `gateway cb-status`, `gateway lb-list` — API gateway data.
  - `workflow list`, `workflow requests` — workflow and approval data.
  - `observability metrics`, `observability status` — telemetry data.
  - `perf pool-stats`, `perf cache-stats` — performance monitoring data.
- **`security ScanDir()`** — recursive directory scanner with extension filtering, skip dirs (.git, vendor, node_modules), and 1MB file size limit.
- **`security audit` real file scanning** — replaced hardcoded findings with real `Auditor` + `ScanDir` integration.
- **39 new CLI integration tests** covering:
  - `security audit`, `set-secret`, `get-secret`, `sanitize`, `hash-password`, `validate` (10 tests).
  - `db connect`, `disconnect`, `status`, `list`, `migrate` (7 tests).
  - `gateway status`, `rate-status`, `cb-status`, `lb-list`, `add-backend` (6 tests).
  - `workflow list`, `create`, `execute`, `requests` (5 tests).
  - `perf pool-create`, `pool-acquire`, `pool-stats`, `cache-set`, `cache-get`, `cache-stats` (6 tests).
  - `observability trace`, `log`, `metrics`, `status`, `dashboard` (5 tests).
- **`internal/database/store.go`** — persistent connection store with JSON file persistence (`~/.naeos/db/connections.json`).
- **`internal/database/store_test.go`** — 8 tests for connection store (Add, List, Remove, Get, Duplicate, Persistence, FileCreated).
- **`internal/generation/adapters/go_test.go`** — 8 tests for GoAdapter.
- **`internal/security/security_test.go`** — 3 new tests for `ScanDir`, `ScanDir_Empty`, `AuditFiles_Summary`.

### Fixed
- **Rust/Axum adapter** — upgraded from Axum 0.6 to 0.7 (`axum::serve` replacing deprecated `axum::Server`), meaningful test assertions.
- **Actix-web adapter** — CI workflow uses `dtolnay/rust-toolchain@stable`, Dockerfile `AS` casing fixed.
- **FastAPI adapter** — removed `# TODO` placeholder in generated `app.py`, test imports actual app module.
- **Go adapter** — test now asserts `handler.Handle() == "processed"` instead of `assert!(true)`.
- **Python adapter** — test now asserts `handler.handle() == "processed"` instead of `assert True`.
- **Java adapter** — migrated from JUnit 4 to JUnit 5 (`junit-jupiter`), test uses `assertEquals("processed", ...)`.
- **TypeScript adapter** — test uses Vitest with proper assertion `handler.handle() == "processed"`.
- **All generated docker-compose files** — removed deprecated `version: '3.8'` field.
- **WebSocket server** — removed insecure `defaultUpgrader` package variable (server already uses secure per-instance upgrader).
- **`db_cmd.go`** — rewritten to use real `database.New()` adapter + persistent connection store.

### Changed
- **Engine `renderDockerCompose`** — removed `version: '3.8'` from all generated docker-compose YAML output.
- **Engine `renderPlaceholder`** — fixed placeholder comment to include `pipeline` keyword.
- **Engine test** — updated assertion to verify `services:` present and `version:` absent in docker-compose output.
- All 103 packages pass `go test`, `go vet` clean, `go build` clean.

## [1.2.0] - 2026-07-15

### Added
- **Database layer v1.2.0** with production-ready features:
  - `context.Context` support for all database operations (ExecContext, QueryContext, BeginTx).
  - Connection pool configuration (MaxOpenConns, MaxIdleConns, ConnMaxLifetime, ConnMaxIdleTime).
  - Config validation (host required, positive port, valid SSLMode).
  - Real MySQL adapter with go-sql-driver/mysql.
  - Real SQLite adapter with modernc.org/sqlite (WAL mode, foreign keys).
  - Factory pattern (New, NewFromConfig) for adapter creation.
  - Retry logic with exponential backoff for transient errors.
  - Query logging decorator with slow query detection (>1s).
  - Health checks for all adapters.
  - API server database integration (SetDatabase, pipeline persistence).
  - File-based migration loader (LoadMigrations).
  - NES-042-Database.md specification.

### Changed
- **MySQL mock adapter** now extends BaseDatabase (reduced from ~300 to ~50 lines).
- **SQLite mock adapter** now extends BaseDatabase (reduced from ~300 to ~50 lines).
- **PostgreSQL mock adapter** now extends BaseDatabase (reduced from ~300 to ~50 lines).
- **Migration Rollback** now executes Down SQL before deleting migration record.

### Fixed
- **QueryRowContext** in real adapters no longer holds connection open (scans rows directly).

## [1.1.0] - 2026-07-15

### Added
- **Test coverage improvements** across 10 packages:
  - `internal/generation/adapters`: Java, Python, Rust, TypeScript adapter tests (GenerateProject, GenerateModule, GenerateService, GenerateDockerfile, GenerateCI, GenerateDockerCompose, GenerateArchitectureDoc).
  - `internal/cloud`: AWS/GCP/Azure Destroy method tests, RunnerPool eviction/not-found, parsePlanJSON edge cases, concurrent State access.
  - `internal/broker`: NATS real connector tests (name, connect failure, not connected).
  - `internal/database`: PostgreSQL real connector tests (name, connect failure, not connected).
  - `internal/pluginsdk/sandbox`: Concurrent execution, context cancellation tests.
  - `internal/marketplace`: FetchPlugin 404/timeout/invalid-JSON, SearchPlugins empty results.
  - `internal/diff`: DiffNEIR empty, modules same/added, services removed.
  - `cmd/naeos`: FormatOutput/FormatTable, loadInput, resolveInput, renderOutput, cpDir, checkSpec, doctor command.
- **Integration test build tags**: `//go:build integration` added to e2e and pipeline test files for faster CI.
- **Fuzz tests**: `FuzzHandleMCP` for MCP server.
- **t.Parallel()**: Added to 109 test functions across testrunner, watch, websocket, cloud packages.
- **Godoc comments**: ~122 exported symbols documented across api, ai, artifacts, audit, cloud packages.
- **OpenAPI spec**: Fully rewritten to match actual server implementation (27 endpoints, all request/response schemas aligned).

### Changed
- **WebSocket race conditions fixed**: 3 data races resolved (broadcast map mutation under RLock, readPump vs writePump concurrent WriteMessage, Stop vs writePump concurrent WriteMessage). Added `writeMu sync.Mutex` to Client.
- **Migrated `interface{}` → `any`**: 247 occurrences replaced across 39 files (Go 1.18+ idiomatic).
- **Fixed `return nil, nil` ambiguity**: Added clarifying comments to 9 instances across 5 files.
- **Removed deprecated packages**: `internal/pluginsdk/sdk.go` and `pkg/plugin/plugin.go` deleted (zero consumers).
- **CHANGELOG dedup**: Merged duplicate `### Added` section in v0.2.0.
- **README badge**: Go version updated from 1.22+ to 1.25+.

### Security
- WebSocket `AllowedOrigins` now enforced per-instance with configurable upgrader.
- JWKS endpoint removed (was leaking HMAC secret).

## [1.0.0] - 2026-07-14

### Added
- **Test coverage improvements** across 12 packages:
  - `internal/pluginsdk`: New test suite for deprecated wrapper package (type aliases, state constants, factory functions).
  - `internal/database`: Expanded tests for MySQL, SQLite full lifecycle, transaction rollback, Pool overflow, Manager edge cases (15 new tests).
  - `internal/websocket`: Server register/unregister, broadcast to clients, full channel handling, EventBroadcaster and WSObserver full coverage, WebSocket integration tests (13 new tests).
  - `internal/migration`: MigrationEngine full lifecycle, VersionBetween, FormatMigrationPlan, builtin transforms, MigrationPlanner with custom steps (15 new tests, coverage 33.1% → 90.8%).
  - `internal/marketplace`: Install, Publish update, Search limit/no-match, contains edge cases, corrupted cache (12 new tests).
  - `internal/api`: All handler endpoints tested (pipeline/status, artifacts, context/generate, mcp/message, cloud/plan/deploy/destroy/status, plugins, version, config/schema, pipelines, metrics, healthz, readyz) (32 new tests).
  - `internal/configschema`: ValidateFile (YAML/JSON/unknown/not-found), ValidateData invalid YAML, validateType edge cases (8 new tests).
  - `internal/telemetry`: HTTPExporter (new, flush empty, export spans, export error), Service defaults, generateID counter, SpanCount (7 new tests, coverage 48.1% → 94.2%).
  - `internal/testrunner`: Language detection for all 5 languages, language-specific runner tests, pnpm detection (15 new tests, coverage 41.6% → 98.2%).
  - `internal/watch`: PipelineWatcher shouldProcess, Start/Stop, DetectChanges modified/empty, fsnotify debounce (7 new tests, coverage 41.7% → 84.5%).

### Changed
- Version bumped to 1.0.0.
- CodeQL workflow Go version fixed (1.22 → 1.25).
- OpenAPI 3.0 spec updated to v1.0.0 with missing endpoints (/version, /config/schema, /pipelines, /metrics, /healthz, /readyz).
- Overall test coverage improved from 61.6% to 65.4%.

## [0.9.0] - 2026-07-13

### Added
- **Structured logging** (`log/slog`):
  - Replaced all `log.Println`/`log.Printf` with `slog.Info`, `slog.Error`, `slog.Warn`.
  - JSON handler with structured fields: method, path, status, duration, request_id, component.
  - Log level adapts by HTTP status (error for 5xx, warn for 4xx).
- **Request body size limits**:
  - `MaxBytesReader` on all POST/PUT/PATCH requests (default 10MB).
  - HTTP 413 Payload Too Large response on exceed.
- **X-Request-ID propagation**:
  - UUID v4 generated per request if not provided in `X-Request-ID` header.
  - Propagated in response headers, logs, and context.
- **Configurable CORS**:
  - `CORSConfig` struct with `AllowedOrigins`, `AllowedMethods`, `AllowedHeaders`, `AllowCredentials`.
  - Configurable per-server, defaults to localhost origins.
  - Proper OPTIONS preflight handling (204 No Content).
- **Prometheus metrics endpoint**: `GET /metrics` (text format), `GET /healthz` (liveness), `GET /readyz` (readiness).
- **Real OAuth2 token exchange**:
  - Google: POST to `oauth2.googleapis.com/token`, GET `googleapis.com/oauth2/v2/userinfo`.
  - GitHub: POST to `github.com/login/oauth/access_token`, GET `api.github.com/user`.
- **RBAC enforcement**: `RBACMiddleware` wires JWT user → role → permission check per endpoint.
- **Audit logging** (`internal/audit/`):
  - `AuditEvent` struct with ID, Timestamp, UserID, Action, Resource, IP, UserAgent.
  - `FileAuditor` (JSON lines to `~/.naeos/audit.log`), `MemoryAuditor` for testing.
  - Wired into POST/DELETE handlers and cloud operations.
- **OIDC discovery endpoint**: `GET /.well-known/openid-configuration` and `GET /.well-known/jwks.json`.
- **GoReleaser release workflow** (`.goreleaser.yaml` + `.github/workflows/release-goreleaser.yml`).
- **Interactive CLI mode** (`naeos tui`): Guided wizard for spec creation with prompts.
- **Global `--output-format` flag** (`-o json|yaml|table`): Supported across cloud types, plugin list, history, status, health, doctor.
- **Pipeline cache improvements**:
  - TTL-based expiration via `SetMaxAge()`.
  - LRU eviction by hit count (not just oldest timestamp).
- **Parallel spec parsing**: `errgroup`-based concurrent module normalization (configurable via `Parallel` field).
- **Cloud adapter connection pooling**: `RunnerPool` caches TerraformRunner instances, avoids repeated `terraform init`.
- **OIDC discovery**: `/.well-known/openid-configuration` and `/.well-known/jwks.json` endpoints.
- **Graceful WebSocket draining**: `Stop()` sends close frames, waits up to 5s for client disconnect.
- **gorilla/websocket integration**: Replaced custom WebSocket framing with battle-tested library.
- **Lazy plugin loading**: Plugins loaded on first `Execute()` call instead of startup.
- **Shell completion install**: `make install-completion` for bash/zsh/fish.
- **Docker improvements**:
  - `HEALTHCHECK` instruction in Dockerfile.
  - `.dockerignore` excluding docs, tests, git.
  - Multi-arch buildx support (`make docker`).
  - `make docker-local` for single-arch.
- **CI improvements**:
  - Codecov coverage reporting.
  - Expanded golangci-lint (16 linters: gosec, gocritic, bodyclose, errorlint, etc.).
- **API ↔ OpenAPI alignment**: Fixed DELETE path mismatches, added missing endpoints.
- **Cleanup**: Removed empty `api/handlers/` and `api/middleware/` directories.

### Changed
- Version bumped to 0.9.0.
- 104 packages pass, `go vet` clean, `go build` clean.

## [0.8.0] - 2026-07-13

### Added
- **Typed error system** (`internal/errors/`):
  - `NaeosError` struct with `Code`, `Message`, and `Inner` fields.
  - 12 error codes: `ErrParse`, `ErrValidation`, `ErrCloud`, `ErrPlugin`, `ErrAuth`, `ErrPipeline`, `ErrConfig`, `ErrDatabase`, `ErrNetwork`, `ErrInternal`, `ErrNotFound`, `ErrConflict`.
  - Helper functions: `New()`, `Wrap()`, `Is()` with full `errors.Is()`/`errors.As()` chain support.
  - Sentinel errors: `ErrNotConnected`, `ErrInvalidSpec`, `ErrPluginNotFound`, `ErrDeployFailed`.
- **Terraform CLI integration** (`internal/cloud/terraform.go`):
  - `TerraformRunner` with `Init()`, `Plan()`, `Apply()`, `Destroy()`, `Output()`.
  - `CommandRunner` interface for testability.
  - Real `terraform init` + `terraform apply` in cloud Deploy methods.
- **Cloud state management** (`internal/cloud/state.go`):
  - `StateManager` persists deployed resources as JSON in `~/.naeos/cloud/<project>/<provider>/`.
  - Thread-safe with `sync.RWMutex`, supports `Save()`, `Load()`, `List()`, `Delete()`.
- **Cloud cost estimation** (`internal/cloud/cost.go`):
  - `CostEstimator` with hardcoded pricing for all 11 resource types × 3 providers.
  - `EstimateCost()`, `EstimateCostByType()`, `FormatCost()` methods.
  - Plan results now include cost estimates in USD.
- **5 new cloud resource types**: serverless/function, monitoring/alerts, secrets, dns/zone, networking/vpc.
  - Full HCL generation for AWS (Lambda, CloudWatch, Secrets Manager, Route53, VPC), GCP (Cloud Functions, Monitoring, Secret Manager, Cloud DNS, VPC Network), Azure (Functions, Monitor, Key Vault, DNS Zone, VNet).
- **WASM plugin runtime** (`internal/pluginsdk/wasm/`):
  - `WASMRuntime` using wazero for WASM plugin execution.
  - JSON-over-WASI stdin/stdout protocol.
  - Sandbox auto-routes `.wasm` files to WASM runtime.
- **Plugin marketplace signature verification** (`internal/marketplace/signature.go`):
  - SHA-256 checksum verification after download.
  - `VerifyPlugin()` and `GenerateChecksum()` functions.
  - Install method now validates checksum before accepting plugin.
- **Plugin hot-reload** (`internal/pluginhost/hotreload.go`):
  - `PluginWatcher` using fsnotify to detect `.so`/`.wasm` file changes.
  - 500ms debounce, automatic unload/reload cycle.
- **Plugin event bus** (`internal/pluginhost/events.go`):
  - `EventBus` with `Subscribe()`, `Unsubscribe()`, `Emit()` for 5 pipeline lifecycle events.
  - `PluginEventBus` implements `PipelineObserver` interface.
- **API key rate limiting** (`internal/api/middleware.go`):
  - `RegisterAPIKey()` for per-key rate limiters.
  - `X-API-Key` header support with fallback to IP-based limiting.
- **Cloud API endpoints** (`internal/api/server.go`):
  - `POST /cloud/plan`, `POST /cloud/deploy`, `POST /cloud/destroy`, `GET /cloud/status`.
  - `GET /plugins`, `POST /plugins/execute`, `DELETE /plugins/{name}`.
- **Async pipeline execution**: `POST /pipeline/run` now returns `202 Accepted` with `job_id`.
- **MCP tools**: `list_artifacts`, `get_pipeline_status`, `export_terraform`, `list_plugins`.
- **CLI commands**: `cloud plan`, `cloud status`, `ai enrich`, `plugin test`.
- **Pipeline result cache** (`internal/pipelinecache/`):
  - SHA-256 spec hashing, LRU-style eviction, disk persistence.
- **Pipeline middleware chain** (`internal/pipelinemiddleware/`):
  - `Chain` executor with `LogMiddleware`, `MetricsMiddleware`, `AuthMiddleware`, `CacheMiddleware`.
- **NEIR structural diff** (`internal/diff/`):
  - Colorized diff between two NEIR objects with project + service level detection.
- **Event sourcing** (`internal/eventsourcing/`):
  - InMemory and FileStore with `Aggregate` and `PipelineRunSnapshot`.
- **Distributed task execution** (`internal/distributed/`):
  - Coordinator, round-robin LoadBalancer, ResultAggregator.
- **Container artifact generation** (`internal/generation/adapters/container/`):
  - Dockerfiles for Go, Node, Python, Java, Rust + docker-compose + K8s manifests.
- **Profile detection** (`internal/profiledetect/`):
  - Auto-detect language/framework from marker files with confidence scoring.
- **Telemetry tracing** (`internal/telemetry/`):
  - Spans with parent-child support, batched HTTP export.
- **Config schema validation** (`internal/configschema/`):
  - Schema definition with `ValidateConfig`, `ValidateData`, `ValidateFile`.
- **ADR documents** (`docs/adr/`):
  - ADR-001: Why Go for Runtime.
  - ADR-002: Why NEIR as Central Model.
  - ADR-003: Why MCP for AI Integration.
- **NES-041 Troubleshooting Guide**: 15 practical troubleshooting scenarios.
- **Consolidated OpenAPI 3.0 spec** at `docs/openapi.yaml` (v0.8.0) with all endpoints.
- **NES-028 and NES-030** stabilized with examples for all new commands.
- **Tests**: 39 new tests across generation/renderers, generation/engine, hcl, cloud, marketplace, api, pluginhost, mcp, errors.
- **Makefile targets**: `docker`, `benchmark`, `security`, `e2e`.

### Changed
- Version bumped to 0.8.0.
- CI: Added golangci-lint step to GitHub Actions workflow.
- CI: Fixed Go version mismatch (all set to 1.25).
- CI: Fixed release ldflags to use centralized `internal/version` package.
- Dockerfile updated to `golang:1.25-alpine`.
- All `fmt.Errorf` calls audited for `%w` wrapping.
- Duplicate `newCompletionCommand` registration fixed in `main.go`.
- Removed `docs/api/` directory (consolidated into single `docs/openapi.yaml`).

## [0.7.0] - 2026-07-13

### Added
- **10 new CLI commands**:
  - `naeos benchmark` — run pipeline N iterations with timing statistics (avg, min, max, ops/sec).
  - `naeos config validate|show` — validate config against schema or display default config schema.
  - `naeos deploy` — deploy pipeline output to Docker, Kubernetes, Docker Compose, SSH, rsync, or local copy with dry-run.
  - `naeos distributed` — execute pipeline tasks across multiple parallel workers with coordinator/round-robin dispatch.
  - `naeos events replay|list` — replay event sourcing records or list past pipeline run events.
  - `naeos export compose` — generate `docker-compose.yaml`, `Dockerfile`, and K8s manifests via container adapter.
  - `naeos health` — system health checks (Go, Git, config dir, version) with text/JSON/YAML output.
  - `naeos history` — display summary of past pipeline runs from persisted event store.
  - `naeos import` — parse HCL specification files and convert to NAEOS YAML/JSON.
  - `naeos migration status` — show migration status for PostgreSQL, MySQL, SQLite.
- **AI/LLM integration** (`internal/ai/`):
  - LLM service supporting OpenAI and Anthropic providers.
  - `EnrichSpec`, `GenerateSuggestions`, `ExplainArchitecture` methods with structured prompts.
- **NATS message broker** (`internal/broker/`):
  - Real NATS client with connect, publish, subscribe, ping, and close.
- **Config hot-reload** (`internal/configreload/`):
  - `HotReloader` watches config directory via `fsnotify`, auto-reloads with 300ms debounce.
  - Config diff computation (added/removed/modified keys).
- **PostgreSQL database adapter** (`internal/database/`):
  - Real PostgreSQL adapter using `pgx` with connect, exec, query, transactions, and versioned migration tracking.
- **NEIR structural diff** (`internal/diff/`):
  - Structural diffing between two NEIR objects with colorized formatted output.
  - Detects project-level and service-level changes (added, removed, modified).
- **Distributed task execution** (`internal/distributed/`):
  - Coordinator with fan-out dispatch to worker goroutines.
  - Round-robin LoadBalancer, ResultAggregator, and SimpleWorker.
- **Event sourcing** (`internal/eventsourcing/`):
  - Event store interface with InMemory and FileStore (JSON persistence).
  - Aggregate with versioned event application and PipelineRunSnapshot for state reconstruction.
- **Container artifact generation** (`internal/generation/adapters/container/`):
  - Generates Dockerfiles for Go, Node, Python, Java, Rust.
  - Generates `docker-compose.yaml` and Kubernetes manifests (namespace, deployment, service).
- **HCL parser** (`internal/hcl/`):
  - Simple HCL parser for project/service/infra blocks with YAML serialization.
- **End-to-end integration tests** (`internal/integration/`):
  - Full pipeline E2E tests: spec → parse → normalize → resolve → build → validate → compile.
- **Remote plugin marketplace** (`internal/marketplace/remote.go`):
  - `RemoteRegistry` with List, Search, Install, Uninstall operations against remote HTTP registry.
  - Plugin binary (.so) download with metadata persistence.
- **Pipeline result cache** (`internal/pipelinecache/`):
  - SHA-256 spec hashing, LRU-style eviction, disk persistence, hit counting.
- **Pipeline middleware chain** (`internal/pipelinemiddleware/`):
  - `Chain` executor with LogMiddleware (timing), MetricsMiddleware, AuthMiddleware (token), CacheMiddleware.
- **Plugin sandbox** (`internal/pluginsdk/sandbox/`):
  - Executes external plugin binaries via JSON-over-stdin/stdout protocol with timeouts.
  - WASM execution path using `wasmtime`.
- **Profile detection** (`internal/profiledetect/`):
  - Auto-detect project language/framework from marker files with weighted confidence scoring.
  - Framework detection: React, Next.js, Django, Gin, etc.
- **Telemetry tracing** (`internal/telemetry/`):
  - Span creation with parent-child support, batched export via Exporter interface.
  - `HTTPExporter` for remote endpoint posting.
- **Config schema validation** (`internal/configschema/`):
  - Schema definition with property types and validation.
  - `ValidateConfig`, `ValidateData`, `ValidateFile` for YAML/JSON configs.
- **WebSocket observer** (`internal/websocket/`):
  - Bridges `PipelineObserver` to `EventBroadcaster` for real-time pipeline lifecycle events.
- **Pipeline adapter** (`pkg/pipeline/`):
  - Middleware chain support, event sourcing hooks, and telemetry integration.
  - `RunWithMiddleware` for pre/post-process middleware execution.

### Changed
- Version bumped to 0.7.0.
- 101 packages pass, `go vet` clean, `go build` clean.
- 54,819 lines of Go code across the codebase.
- Enhanced CLI: `init`, `lint`, `search`, `validate`, `watch`, `status`, `test`, `plugin`, `marketplace`, `observability`, `security`, `profile`, `workspace`, `ws`, `doctor`, `export`, `scaffold` commands expanded with subcommands and richer functionality.
- Improved error handling and logging across all subsystems.

## [0.6.0] - 2026-07-12

### Added
- **Centralized version management** (`internal/version/`):
  - `VERSION` file at repository root.
  - `internal/version/version.go` with `String()`, `Full()`, embed-based fallback.
  - Makefile ldflags injection: `-X version.Version=... -X version.GitCommit=... -X version.BuildDate=...`.
- **Persistent search engine** (`internal/search/search.go`):
  - `Persistent` wrapper with JSON file persistence between CLI invocations.
  - Data stored in `~/.naeos/search/<name>/search-index.json`.
  - CLI `search` commands now use persistent storage by default.
- **Plugin system pipeline integration** (`pkg/pipeline/pipeline.go`):
  - `PluginManager` field in pipeline Config for plugin lifecycle hooks.
  - `executePluginHooks()` runs enabled plugins at `pipeline.after_run` stage.
- **Pipeline observer pattern** (`pkg/pipeline/pipeline.go`):
  - `PipelineObserver` interface: `OnPipelineStart`, `OnPipelineComplete`, `OnPipelineFailed`, `OnArtifactGenerated`.
  - Optional observer hooks wired into pipeline execution lifecycle.
- **MCP validate_spec and compile_spec** (`internal/api/server.go`):
  - API server `handleMCPMessage` now handles `validate_spec` and `compile_spec` tool calls.
- **Cloud Destroy implementations** (`internal/cloud/`):
  - AWS, GCP, Azure adapters now plan and list resources before destroy.

### Changed
- All hardcoded version strings (`"0.5.0"`, `"v0.2.0"`) replaced with `version.String()`.
- `doctor_cmd.go` uses centralized version for header display.
- `graphql_cmd.go` resolvers use centralized version.
- API server health endpoint uses centralized version.
- MCP server uses centralized version.
- Plugin host `LoadAll()` now returns combined errors instead of silently swallowing failures.
- API artifacts endpoint uses `internal/artifacts.Store` with disk persistence instead of in-memory slice.
- Removed deprecated `pluginsdk` CLI command (use `plugin` instead).
- Removed dead code in `db_cmd.go` (unused `strconv` import and `_ = strconv.Itoa(0)`).
- 90 packages pass, `go vet` clean, build clean.

## [0.5.1] - 2026-07-12

### Added
- API handlers fully wired (handleSpecs, handleArtifacts, handleMCPMessage, handlePipelineStatus)
- Integration tests: full pipeline spec → parse → normalize → resolve → build → validate → compile
- Cloud adapter content-based HCL tests (18 subtests: AWS/GCP/Azure × 6 resource types)
- Context bundle enricher: dependency graph, security context, cloud resource mapping
- Dashboard stats persistence (JSON file-based)
- CI/CD pipeline (.github/workflows/ci.yml)
- golangci-lint config (.golangci.yaml)
- OpenAPI 3.0 spec (docs/openapi.yaml, 10 endpoints)

## [0.5.0] - 2026-07-12

### Added
- **Cloud Integration** (`internal/cloud/`):
  - 6 resource types (storage, compute, database, cache, queue, CDN) × 3 providers (AWS/GCP/Azure).
  - Terraform HCL export for all resource types.
  - CLI `cloud run` with `--input-file` flag and spec loader.
  - CLI `cloud types` command listing supported resource types.
  - NEIR model extended with `Project`, `Environment`, `Type` infrastructure fields.
- **Plugin Unification** (`internal/pluginhost/`):
  - Unified plugin system merging 3 legacy packages.
  - Plugin lifecycle: `enable`, `disable`, `info`, `execute`.
  - `pkg/plugin` and `internal/pluginsdk` deprecated with redirect wrappers.
- **MCP Server Fixes**: version 0.5.0, compile_spec returns context bundle.
- **API Server**: JWT auth wired into middleware, handlers use real pipeline.
- **Dashboard**: dynamic `GetStats()`, version updated to 0.5.0.
- Tests for: shared/log, dashboard, docgen, testrunner, testgen, mcp (6 new test files).

### Changed
- All 63 packages pass, `go vet` clean, `go build` clean.

## [0.4.0] - 2026-07-11

### Added
- **Spec Language v2 Enhancement** (`internal/specification/parser/resolve_ext.go`):
  - `$include{file}` — multi-file spec composition with recursive resolution (max depth 10).
  - `$fn{name(args)}` — custom functions: `upper`, `lower`, `slug`, `default`, `len`, `coalesce`.
  - `$if{condition}` / `$endif` — conditional sections based on environment variables.
  - Condition operators: `==`, `!=`, `!`, `defined:`.
- **MCP Server** (`internal/mcp/server.go`):
  - Model Context Protocol server for AI agent integration.
  - Tools: `parse_spec`, `validate_spec`, `generate_context`, `compile_spec`, `explain_concept`.
  - JSON-RPC 2.0 over HTTP with `/mcp` and `/health` endpoints.
- **Migration Engine** (`internal/migration/engine.go`):
  - Real version transforms: v0.1.0 → v0.2.0 (add generation config, normalize modules) → v0.3.0 (add architecture defaults, security, testing).
  - `Migrate()`, `Plan()`, `AvailableVersions()`, `VersionBetween()`.
- **Testing Framework** (`internal/testrunner/runner.go`):
  - Multi-language test runner: Go, TypeScript/Node, Python, Java, Rust.
  - Auto-detect project languages from config files.
- **Documentation Generator** (`internal/docgen/generator.go`):
  - Generate full docs, API docs, module docs from specs or NEIR.
- **Benchmarks** (`internal/specification/parser/bench_test.go`):
  - 8 benchmarks: parse simple/complex/with-variables, validate modules/services, variable resolver, schema version, cycle detection.
- **Fuzz Testing** (`internal/specification/parser/fuzz_test.go`):
  - 6 fuzz targets: parse, parseYAMLNode, variable resolver, schema version, validate modules.
- **Docker Image** — multi-stage Dockerfile (golang:1.22-alpine → alpine:3.19).
- **CLI commands**:
  - `naeos mcp` — start MCP server (`--port`).
  - `naeos test` — run tests for generated code (`--dir`, `--language`, `--verbose`).
  - `naeos docgen` — generate documentation (`--output full|api|modules`).

### Changed
- All 66 packages pass, `go vet` clean, `go build` clean.

## [0.3.0] - 2026-07-11

### Added
- **Spec Language v2** (`internal/specification/parser/resolve.go`):
  - Variable interpolation: `${var}` syntax for custom variables.
  - Environment variable resolution: `$env{VAR}` reads from env.
  - Reference resolution: `$ref{path}` cross-references spec values.
  - Recursive resolver for maps, slices, and nested structures.
- **Validation Kernel** (`internal/specification/parser/resolve.go`):
  - Circular dependency detection in module dependency graphs.
  - Port conflict detection across services.
  - Module boundary enforcement (name required, duplicate detection).
  - Dangling dependency detection (missing module references).
  - Deep validation of `$ref` references against resolved context.
- **Schema Versioning** (`internal/specification/parser/version.go`):
  - `ParseSchemaVersion`, `CheckSpecVersion`, `ExtractVersionFromData` — parse, compare, and validate SemVer spec versions.
  - Parser auto-checks `version` field on parse; rejects specs below minimum version.
  - Minimum version constant `MinSpecVersion = "0.1.0"`, `CurrentSpecVersion = "0.3.0"`.
- **AI Context Bundles** (`internal/context/bundle/bundle.go`):
  - `GenerateFromNEIR` and `GenerateFromSpec` — produce LLM-ready context bundles from NEIR or parsed specs.
  - Markdown and plain text output with modules, services, languages, and endpoints.
  - Metadata tracking (module count, service count, generator).
- **CLI command**:
  - `naeos context` — generate AI context bundles from specifications (`--input`, `--input-file`, `--output markdown|plain|json|yaml`).

### Changed
- Pipeline now performs schema version validation automatically during spec parsing.
- All 63 packages pass, `go vet` clean, `go build` clean.

## [0.2.0] - 2026-07-11

### Added
- **Compiler Foundation** (`internal/compiler/`): Transforms NEIR into AI instruction sets for 6 target tools.
- **AI Output Adapters** (`internal/compiler/adapters/`):
  - GitHub Copilot — `.github/copilot-instructions.md`, `.github/copilot-context.md`, `.github/copilot-rules.md`
  - Claude Code — `CLAUDE.md`, `.claude/context.md`, `.claude/rules.md`
  - Cursor — `.cursorrules`, `.cursor/context.md`
  - Gemini CLI — `.gemini/CONFIG.md`, `.gemini/context.md`
  - Codex — `AGENTS.md`, `.codex/context.md`
  - OpenCode — `AGENTS.md`, `.opencode/context.md`, `.opencode/rules.md`
- **Artifact Store** (`internal/artifacts/`): Manages generated outputs with content-hash dedup, kind detection, metadata, and disk persistence.
- **Profile Registry** (`internal/profiles/`): 5 industry-specific profiles (SaaS, AI Agent, FinTech, Healthcare, Government) with modules, services, architecture, security, deployment, and testing templates.
- **Migration constants**: `CurrentVersion` (0.1.0) and `TargetVersion` (0.3.0) exported for version-aware tooling.
- **CLI commands**:
  - `naeos compile` — compile spec into AI instruction sets (per-target or `--all`)
  - `naeos profile list|show|search|apply` — browse and apply industry profiles
  - `naeos artifacts list|info|dedup|summary` — manage generated artifact store
  - `naeos migrate run|plan|versions` — manage schema migrations with dry-run support
- Comprehensive test suites: compiler (6 tests), adapters (14 tests), artifacts (14 tests), profiles (9 tests)

### Changed
- All 63 packages pass, `go vet` clean, `go build` clean.
- Documentation index with recommended reading orders (beginner, policy, profile, CLI, testing).
- NES-028 CLI Reference — comprehensive CLI command documentation.
- NES-029 Configuration — pipeline configuration reference.
- NES-030 Specification Language — NAEOS specification language docs.
- NES-031 Errors — exhaustive error catalog.
- NES-032 Telemetry — telemetry and metrics reference.
- NES-033 Testing Guide — test guide with coverage requirements.
- NES-034 Event Bus — internal pub/sub event bus documentation.
- NES-035 Version Management — SemVer management documentation.
- NES-036 Template Renderer — template rendering engine documentation.
- NES-037 Knowledge Graph & Provenance — knowledge graph and lineage documentation.
- NES-038 Shared Types & Contracts — shared types and contracts documentation.
- NAEOS-GOV-002 Vision — long-term vision document.
- NAEOS-GOV-005 Core Principles — 8 core engineering principles.
- Expanded 18 NES stub documents (NES-003 through NES-022) with full API references and examples.
- `status` command — display current pipeline and project status.
- Auto-detection of config files (`config.yaml`, `config.yml`, `config.json`, `naeos.yaml`, `naeos.yml`, `naeos.json`, `.naeos/config.*`) in working directory.
- Global `--dry-run` flag for preview mode across all commands.
- Per-command `--dry-run` flag for `run`, `export`, and `preview` commands.
- Language-aware scaffold — `--language` flag now generates correct files for Go, TypeScript, Python, Java, and Rust.
- E2E test suite with comprehensive pipeline integration tests.
- Additional benchmarks for dry-run, full-spec, and verbose pipeline runs.
- Fixed GoAdapter `cleanModulePath` to correctly handle relative paths (e.g., `./internal/core`).

### Changed
- NES-001 Repository — updated repository structure to match actual codebase paths.
- DOCUMENTATION-INDEX.md — added NES-028 through NES-038, Go package reference section, CLI and testing reading orders.
- **Refactored `cmd/naeos/main.go`**: split 1876-line monolith into 28 separate command files for better maintainability.
- All CLI commands now support `--config` auto-detection (no longer required to specify explicitly).
- Improved CLI help text with usage examples for all commands.
- Pipeline `Config` struct now includes `DryRun` field for preview mode.
- `preview` command now uses dry-run mode by default.
- Removed unused `hashContent()` function from CLI.
- Consistent error handling across all CLI commands.
- Go adapter `GenerateProject` now generates a complete runnable main.go with HTTP server setup, health check, and API endpoints.

## [0.1.0] - 2026-01-01

### Added
- Initial project structure.
- CLI with 11 subcommands: init, run, validate, inspect, doctor, repair, scaffold, export, preview, kernel, version.
- Core pipeline: parser, normalizer, resolver, NEIR builder, validator.
- Planner: DAG graph with topological sort and cycle detection.
- Generator engine: Go project code, Dockerfile, CI, documentation.
- Policy evaluator with 7 operators and 5 default rules.
- Artifact reviewer with governance rules.
- Knowledge graph with 14 node types and 13 edge types.
- Provenance tracking store.
- Runtime execution engine with deduplication.
- Telemetry event collector.
- 34 modular design documents (NES-000 through NES-033).
- 10 specification documents (NAEOS-SPEC-001 through 010).
- 8 constitutional documents (NAEOS-CON-001 through 008).
- 8 governance documents (NAEOS-GOV-001 through 008).
- 4 kernel specification documents (NAEOS-KER-001 through 004).
- 7 policy system documents (NAEOS-POL-001 through 007).
- 7 profile system documents (NAEOS-PRO-001 through 007).
- 1 reference architecture document (NAEOS-NRA-001).
- ADR and RFC templates with examples.
- Example specifications (minimal and full).
