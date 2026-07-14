# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
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
