# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

### Added
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
