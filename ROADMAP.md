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
