# Roadmap

Roadmap ini memberikan arah pengembangan dokumentasi dan ekosistem NAEOS.

## Fase 1 — Fondasi
- menyempurnakan dokumen inti,
- memastikan konsistensi terminologi,
- menambahkan panduan kontribusi dan onboarding.

## Fase 2 — Tooling dan Validasi
- menyiapkan template untuk ADR dan RFC,
- memperjelas mekanisme review,
- mengembangkan aturan validasi dokumen.

## Fase 3 — Referensi Implementasi
- menyediakan contoh implementasi referensi,
- memperjelas alur kerja dari requirement ke deployment,
- menyiapkan profil untuk skenario industri tertentu.

## Fase 4 — Ekosistem
- memperluas interoperabilitas dengan AI agent dan toolchain,
- memperkuat dokumentasi publik,
- mendukung adopsi lintas organisasi.

### v0.10.0 — Code Quality & Lint Compliance
- [x] golangci-lint: resolved all 999 issues (999 → 0)
- [x] Removed 22 unused functions, types, vars, and struct fields
- [x] Replaced WriteString(fmt.Sprintf(...)) with fmt.Fprintf(...) across 21 files
- [x] Added context propagation (http.NewRequestWithContext, exec.CommandContext) across all HTTP and exec calls
- [x] Fixed errcheck issues with proper error handling
- [x] Fixed gosec issues: file permissions, HTTP timeouts, path validation, weak crypto
- [x] Fixed govet copylocks issues with pointer types
- [x] Fixed misspellings (US English locale)
- [x] Fixed staticcheck issues: S1039, S1011, S1025, S1002, QF1003, QF1001, QF1008, QF1004
- [x] Fixed unparam, unconvert, ineffassign issues
- [x] Applied gofmt and goimports for consistent formatting

## Prinsip roadmap
Prioritas utama adalah menjaga kualitas, konsistensi, dan keterpahaman dokumen bagi komunitas serta implementer.

---

## Implementasi Teknis (Completed)

### Core Improvements
- [x] Fix `FindByContentSubstring` bug (was hardcoded false)
- [x] Resolver cross-reference: dependency filtering, endpoint normalization, defaults
- [x] Wire `--verbose` CLI flag to pipeline
- [x] Integrate `renderers.Renderer` into pipeline kernel service
- [x] Implement `GenerateForLanguage` with per-language code generation
- [x] Add `ParallelGroups()` to scheduler for priority-based execution
- [x] Add `extractDeployment()` and `extractTesting()` to NEIR builder
- [x] Add `SetOutputDir()` and file write to RuntimeEngine
- [x] 180+ tests passing with race detector
- [x] Clean up duplicate governance files

### v0.5.0 — Cloud Integration & Plugin Unification
- [x] Cloud resource types (storage, compute, database, cache, queue, CDN) for AWS/GCP/Azure
- [x] Terraform HCL export for all 6 resource types × 3 providers (21 adapter tests)
- [x] CLI `cloud run` with `--input-file` flag and spec loader
- [x] CLI `cloud types` command listing supported resource types
- [x] Unified plugin system (`internal/pluginhost/`) merging 3 legacy packages
- [x] Plugin lifecycle: `enable`, `disable`, `info`, `execute` subcommands
- [x] `pkg/plugin` and `internal/pluginsdk` deprecated with redirect wrappers
- [x] NEIR model extended with `Project`, `Environment`, `Type` infrastructure fields
- [x] MCP server: fixed version (0.3.0 → 0.5.0), compile_spec returns context bundle
- [x] API server: JWT auth wired into middleware, handlers use real pipeline
- [x] Dashboard: dynamic `GetStats()`, version updated to 0.5.0
- [x] Tests added for: shared/log, dashboard, docgen, testrunner, testgen, mcp (6 new test files)

### v0.5.1 — Quality & DevOps
- [x] API handlers fully wired (handleSpecs, handleArtifacts, handleMCPMessage, handlePipelineStatus)
- [x] Integration tests: full pipeline spec → parse → normalize → resolve → build → validate → compile
- [x] Cloud adapter content-based HCL tests (18 subtests: AWS/GCP/Azure × 6 resource types)
- [x] Context bundle enricher: dependency graph, security context, cloud resource mapping
- [x] Dashboard stats persistence (JSON file-based)
- [x] CI/CD pipeline (.github/workflows/ci.yml)
- [x] golangci-lint config (.golangci.yaml)
- [x] OpenAPI 3.0 spec (docs/openapi.yaml, 10 endpoints)

### v0.6.0 — Integration & Quality
- [x] Centralized version management (VERSION file, internal/version, ldflags Makefile)
- [x] All hardcoded version strings replaced with version.String()
- [x] Persistent search engine (JSON file-based, survives CLI invocations)
- [x] Plugin system wired into pipeline (PluginManager + plugin hooks)
- [x] PipelineObserver interface for dashboard/WebSocket lifecycle events
- [x] MCP validate_spec and compile_spec tool calls in API server
- [x] API artifacts endpoint uses internal/artifacts.Store with disk persistence
- [x] Cloud Destroy implemented for AWS/GCP/Azure (resource listing + plan)
- [x] Plugin host LoadAll returns combined errors instead of silent swallowing
- [x] Removed deprecated pluginsdk CLI command
- [x] Dead code cleanup (unused strconv import in db_cmd.go)

### v0.7.0 — Distributed, AI & DevOps Expansion
- [x] 10 new CLI commands (benchmark, config, deploy, distributed, events, export_compose, health, history, import, migration)
- [x] AI/LLM integration (OpenAI + Anthropic providers, enrich/spec/suggest/explain)
- [x] NATS real broker adapter (connect, publish, subscribe, ping, close)
- [x] Config hot-reload via fsnotify with debounce and config diff
- [x] PostgreSQL real adapter using pgx with transactions and versioned migrations
- [x] NEIR structural diff (project + service level, colorized output)
- [x] Distributed task execution (Coordinator, LoadBalancer, ResultAggregator, SimpleWorker)
- [x] Event sourcing (InMemory + FileStore, Aggregate, PipelineRunSnapshot)
- [x] Container artifact generation (Dockerfiles for 5 languages, docker-compose, K8s manifests)
- [x] HCL parser for project/service/infra blocks
- [x] End-to-end integration tests (full pipeline lifecycle)
- [x] Remote plugin marketplace (list, search, install, uninstall via HTTP registry)
- [x] Pipeline result cache (SHA-256 hashing, LRU eviction, disk persistence)
- [x] Pipeline middleware chain (log, metrics, auth, cache middleware)
- [x] Plugin sandbox (JSON-over-stdin/stdout, WASM execution path)
- [x] Profile detection (language/framework from marker files, confidence scoring)
- [x] Telemetry tracing (spans, batched export, HTTP exporter)
- [x] Config schema validation (schema definition, YAML/JSON validation)
- [x] WebSocket observer (PipelineObserver → EventBroadcaster bridge)
- [x] Pipeline adapter (middleware chain, event sourcing, telemetry integration)

### v0.8.0 — Quality, Security & Ecosystem
- [x] Typed error system with 12 error codes and sentinel errors
- [x] Terraform CLI integration (Init, Plan, Apply, Destroy via exec)
- [x] Cloud state management (JSON persistence, thread-safe StateManager)
- [x] Cloud cost estimation (hardcoded pricing for 11 types × 3 providers)
- [x] 5 new cloud resource types (serverless, monitoring, secrets, DNS, VPC)
- [x] WASM plugin runtime (wazero, JSON-over-WASI protocol)
- [x] Plugin marketplace SHA-256 signature verification
- [x] Plugin hot-reload via fsnotify file watcher
- [x] Plugin event bus (5 pipeline lifecycle events, PipelineObserver bridge)
- [x] API key rate limiting (X-API-Key header, per-key limiters)
- [x] Cloud API endpoints (plan, deploy, destroy, status)
- [x] Plugin API endpoints (list, execute, uninstall)
- [x] Async pipeline execution (202 Accepted + job_id)
- [x] MCP tools (list_artifacts, get_pipeline_status, export_terraform, list_plugins)
- [x] CLI commands (cloud plan/status, ai enrich, plugin test)
- [x] Consolidated OpenAPI 3.0 spec (v0.8.0, all endpoints)
- [x] ADR documents (ADR-001 Go, ADR-002 NEIR, ADR-003 MCP)
- [x] NES-041 Troubleshooting Guide (15 scenarios)
- [x] NES-028 and NES-030 stabilized with examples
- [x] golangci-lint added to CI
- [x] CI Go version fixed (1.25), Dockerfile updated to golang:1.25-alpine
- [x] fmt.Errorf %w audit and fix
- [x] Tests for generation/renderers, generation/engine, hcl, cloud, marketplace, api, pluginhost, mcp, errors
- [x] Makefile targets: docker, benchmark, security, e2e

### v0.9.0 — Production Readiness & Standards
- [x] Structured logging with log/slog (JSON, levels, request-scoped fields)
- [x] Request body size limits (10MB default, HTTP 413)
- [x] X-Request-ID propagation (UUID v4, context, response headers)
- [x] Configurable CORS (CORSConfig struct, origin whitelist, preflight)
- [x] Prometheus metrics endpoints (/metrics, /healthz, /readyz)
- [x] Real OAuth2 token exchange (Google + GitHub HTTP endpoints)
- [x] RBAC enforcement in API middleware (JWT → role → permission check)
- [x] Audit logging (FileAuditor + MemoryAuditor, wired to POST/DELETE)
- [x] OIDC discovery endpoint (/.well-known/openid-configuration, JWKS)
- [x] GoReleaser release workflow (.goreleaser.yaml + GitHub Actions)
- [x] Interactive CLI mode (naeos tui — guided wizard)
- [x] Global --output-format flag (json/yaml/table for all list commands)
- [x] Pipeline cache TTL + LRU eviction improvements
- [x] Parallel spec parsing with errgroup (GOMAXPROCS bounded)
- [x] Cloud adapter connection pooling (RunnerPool, skip re-init)
- [x] Docker multi-arch buildx (linux/amd64,linux/arm64)
- [x] CI coverage reporting (Codecov)
- [x] Expanded golangci-lint (16 linters including gosec, errorlint)
- [x] Graceful WebSocket connection draining on shutdown
- [x] gorilla/websocket replacement (replaced custom framing)
- [x] Lazy plugin loading (load on first Execute)
- [x] Shell completion install targets (bash/zsh/fish)
- [x] Docker HEALTHCHECK + .dockerignore
- [x] API ↔ OpenAPI spec alignment (fixed path mismatches)
- [x] Cleanup: removed empty api/handlers/ and api/middleware/

### v1.0.0 — Stable Release
- [x] 65.4% test coverage (105 test packages)
- [x] Comprehensive tests: pluginsdk, database, migration, configschema, telemetry, testrunner, watch, marketplace, websocket, API
- [x] OpenAPI 3.0 specification generated
- [x] Security: removed JWKS HMAC leak, fixed WebSocket CheckOrigin, random_password for DB exports
- [x] Architecture: unified MCP implementations, merged golangci-lint configs (17 linters)
- [x] Performance: WASM module caching, WebSocket busy-wait fix, pagination for list endpoints
- [x] Fuzz tests for MCP server
- [x] Documentation: SECURITY.md, CONTRIBUTING.md updated

### v1.1.0 — Database & Ecosystem
- [x] Database layer: context.Context support, configurable connection pool, retry logic
- [x] Real MySQL adapter (go-sql-driver/mysql)
- [x] Real SQLite adapter (modernc.org/sqlite)
- [x] Query logging decorator, HealthCheck on all adapters
- [x] File-based migration loader, API server database integration
- [x] WebSocket race condition fixes (3 fixes)
- [x] OpenAPI spec rewrite (27 endpoints aligned)
- [x] interface{} → any migration (247 replacements)
- [x] t.Parallel() on 109 test functions
- [x] Godoc comments (~122 symbols), 70+ new tests

### v1.3.0 — Quality, Correctness & Production Readiness
- [x] Code generation fixes: Rust/Axum 0.7, Java JUnit 5, meaningful test adapters
- [x] Persistent connection store (~/.naeos/db/connections.json)
- [x] Security hardening: security.ScanDir(), real Auditor integration
- [x] 39 new CLI integration tests, 8 database store tests, 3 security tests
- [x] Structured output (--output json/yaml) for 14 commands
- [x] 103 packages pass, 47+ new tests, 7 adapter bugs fixed

### v1.3.1 — Code Quality & Lint Compliance
- [x] golangci-lint: resolved all 999 issues (999 → 0)
- [x] Removed 22 unused functions, types, vars, and struct fields
- [x] Replaced WriteString(fmt.Sprintf(...)) with fmt.Fprintf(...) across 21 files
- [x] Added context propagation across all HTTP and exec calls
- [x] Fixed errcheck, gosec, govet, staticcheck, unparam, unconvert, ineffassign issues
- [x] Applied gofmt and goimports for consistent formatting

### v1.4.0 — Prompt Library & Platform Improvements
- [x] Prompt Library (NES-054): centralized YAML-based prompt templates for LLM and compiler adapters
- [x] Builtin prompts: 3 LLM (enrich-spec, generate-suggestions, explain-architecture) + 6 compiler adapters
- [x] Custom template functions (join, bt, code, json, yaml) with backtick support
- [x] Template CLI: `naeos template list/show` with `--kind` filter
- [x] Backward compatible: nil library falls back to hardcoded prompts
- [x] AIService wired to LLMService with rule-based fallback
- [x] Observability dashboard: /traces, /logs, /metrics endpoints now return real data
- [x] Workflow Manager: file-based persistence (~/.naeos/workflows/workflows.json)
- [x] Distributed workers: stage-aware processing with context cancellation
- [x] Bug fixes: version test mismatch, rollback Import '.' entry rejection
- [x] 19 new promptlib tests, 5 new AI-LLM integration tests
