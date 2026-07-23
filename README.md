# NAEOS Foundation — NAEOS

[![CI](https://github.com/NAEOS-foundation/naeos/actions/workflows/ci.yml/badge.svg)](https://github.com/NAEOS-foundation/naeos/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/go-1.25+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/NAEOS-foundation/naeos)](https://goreportcard.com/report/github.com/NAEOS-foundation/naeos)
[![Release](https://img.shields.io/github/v/release/NAEOS-foundation/naeos)](https://github.com/NAEOS-foundation/naeos/releases)

> Specify Once. Build Anywhere.

NAEOS (Nusantara Engineering & Architecture Operating System) is a declarative engineering platform that transforms specifications into high-quality software systems through a consistent, validated, and extensible pipeline.

NAEOS is not just a project generator. NAEOS is an engineering runtime that understands specifications, builds an internal model, orchestrates execution plans, generates artifacts, validates results, and keeps projects aligned with specifications throughout their lifecycle.

## Vision

Build an open-source engineering platform that enables developers and organizations to describe their system once, then build, validate, and evolve software across multiple languages, frameworks, and platforms.

## Quick Start

```bash
# Clone dan build
git clone https://github.com/NAEOS-foundation/naeos.git
cd naeos
go build ./cmd/naeos/

# Buat spesifikasi
cat > spec.yaml << 'EOF'
project: my-app
modules:
  - name: auth
    path: ./auth
  - name: api
    path: ./api
    dependencies: [auth]
services:
  - name: gateway
    kind: http
    port: 8080
architecture:
  pattern: hexagonal
generation:
  languages: [go, typescript]
EOF

# Jalankan pipeline
naeos run --config config.yaml --input-file spec.yaml

# Generate AI context
naeos context --input-file spec.yaml

# Compile ke AI tools
naeos compile --all --input-file spec.yaml
```

## Features

### Core Pipeline
- **Parser** — YAML/JSON specification parsing with variable interpolation
- **Normalizer** — data normalization
- **Resolver** — cross-reference resolution
- **NEIR Builder** — unified project model
- **Validator** — comprehensive validation (circular deps, port conflicts, module boundaries)
- **Scheduler** — DAG-based task scheduling
- **Generator** — multi-language code generation (Go, TypeScript, Python, Java, Rust)

### Spec Language v2
- `${var}` — variable interpolation
- `$env{VAR}` — environment variable resolution
- `$ref{path}` — cross-reference resolution
- `$include{file}` — multi-file spec composition
- `$fn{name(args)}` — custom functions (upper, lower, slug, default, len, coalesce)
- `$if{condition}` / `$endif` — conditional sections
- Schema versioning with auto-check (minimum v0.1.0)

### AI Integration
- **Compiler** — transform NEIR ke AI instruction sets
- **6 Output Adapters**:
  - GitHub Copilot — `.github/copilot-instructions.md`
  - Claude Code — `CLAUDE.md`
  - Cursor — `.cursorrules`
  - Gemini CLI — `.gemini/CONFIG.md`
  - Codex — `AGENTS.md`
  - OpenCode — `AGENTS.md`
- **MCP Server** — Model Context Protocol untuk AI agent integration
- **Context Bundles** — LLM-optimized project summaries

### Marketplace
- **Profile Marketplace** — publish, search, download industry profiles
- **Plugin Marketplace** — install, uninstall, search plugins
- **5 Built-in Profiles**: SaaS, AI Agent, FinTech, Healthcare, Government

### Governance
- **Policy Evaluator** — 7 operators, 5 default rules
- **Artifact Review** — governance rules
- **Audit Trail** — traceability

### Developer Tools
- **35+ CLI Commands** — run, validate, compile, context, test, docgen, mcp, marketplace, etc.
- **Watch Mode** — hot-reload pipeline on spec changes
- **Diff Engine** — compare specs with colorized output
- **Migration Engine** — schema version transforms (v0.1→v0.2→v0.3)
- **Testing Framework** — multi-language test runner
- **Documentation Generator** — auto-generate API/module docs
- **Benchmarks & Fuzz Testing** — performance and robustness
- **Docker** — multi-stage Dockerfile

## Architecture

```text
┌─────────────────────────────────────────────────────────┐
│                    NAEOS Architecture                     │
├─────────────┬──────────────┬──────────────┬─────────────┤
│    Input    │  Core Layer  │  Generation  │   Output    │
├─────────────┼──────────────┼──────────────┼─────────────┤
│  Spec YAML  │   Parser     │   Generator  │  Code Files │
│  CLI cmds   │   Normalizer │   Adapters   │  Configs    │
│  Profiles   │   Resolver   │   Renderers  │  Docs       │
│  Context    │   Validator  │   Compiler   │  AI Context │
│             │   Scheduler  │   Profiles   │  Artifacts  │
│             │   Kernel     │              │             │
│             │   Policy     │              │             │
│             │   Review     │              │             │
└─────────────┴──────────────┴──────────────┴─────────────┘
```

## Core Components

### Kernel
The kernel provides the runtime foundation:
- Service Registry
- Event Bus (pub/sub)
- Telemetry Collection
- Lifecycle Management

### Specification
Specifications use NAEOS Specification Language v2 as the single source of truth.

### NEIR
NAEOS Engineering Intermediate Representation is the central engineering model representing the entire system. NEIR encompasses project, architecture, domain, module, component, service, API, storage, infrastructure, security, AI, documentation, deployment, testing, and metadata.

### Compiler
The compiler transforms NEIR into AI instruction sets for 6 target tools.

### Marketplace
A marketplace for profiles, plugins, and templates that can be published, searched, and installed.

## CLI Commands

| Command | Description |
|---------|-------------|
| `naeos run` | Execute full pipeline |
| `naeos validate` | Validate specification |
| `naeos compile` | Compile to AI instruction sets |
| `naeos context` | Generate AI context bundle |
| `naeos test` | Run tests for generated code |
| `naeos docgen` | Generate documentation |
| `naeos mcp` | Start MCP server |
| `naeos marketplace` | Browse marketplace |
| `naeos profile` | Manage industry profiles |
| `naeos artifacts` | Manage artifact store |
| `naeos migrate` | Schema migration |
| `naeos doctor` | System health check |
| `naeos diff` | Compare specifications |
| `naeos watch` | Watch for changes |
| `naeos init` | Initialize config |
| `naeos create` | Create project |
| `naeos scaffold` | Generate scaffold |
| `naeos export` | Export artifacts |
| `naeos audit` | Audit specification |
| `naeos kernel` | Inspect kernel |
| `naeos plugin` | Manage plugins |
| `naeos template` | Manage templates |
| `naeos workspace` | Manage workspace |
| `naeos rollback` | Rollback changes |
| `naeos repair` | Repair specification |
| `naeos status` | Pipeline status |
| `naeos ai` | AI assistance |
| `naeos docs` | Documentation |
| `naeos lock` | Lock dependencies |
| `naeos version` | Version info |
| `naeos completion` | Shell completion |

## Repository Structure

```text
cmd/naeos/           # CLI commands (35+ files)
internal/
  specification/     # Parser, normalizer, resolver
  neir/             # NEIR model and builder
  compiler/         # AI instruction compiler
  context/          # Context bundle generator
  generation/       # Code generation
  governance/       # Policy and review
  artifacts/        # Artifact store
  profiles/         # Industry profiles
  marketplace/      # Profile & plugin marketplace
  migration/        # Schema migration
  mcp/              # MCP server
  testrunner/       # Test framework
  docgen/           # Documentation generator
  diff/             # Diff engine
  watch/            # File watcher
  security/         # Security rules
  knowledge/        # Knowledge graph
  database/         # Database layer (PostgreSQL, MySQL, SQLite)
  websocket/        # WebSocket real-time communication
  eventsourcing/    # Event sourcing and aggregate snapshots
  distributed/      # Distributed task execution
  configreload/     # Configuration hot-reload
  pipelinecache/    # Pipeline result caching
  pipelinemiddleware/ # Composable pipeline middleware
  audit/            # Audit logging layer
  hcl/              # HCL configuration parser
  profiledetect/    # Automatic language/framework detection
  ai/               # AI service and LLM integration
  pluginsdk/        # Plugin SDK with WASM runtime
pkg/
  pipeline/         # Main pipeline
  kernel/           # System kernel
  config/           # Configuration
  plugin/           # Plugin system
docs/               # Documentation (56 NES specs)
```

## Documentation

- [DOCUMENTATION-INDEX.md](DOCUMENTATION-INDEX.md) — document index
- [GETTING-STARTED.md](GETTING-STARTED.md) — onboarding guide
- [CONTRIBUTING.md](CONTRIBUTING.md) — contribution guidelines
- [CHANGELOG.md](CHANGELOG.md) — version history
- [docs/](docs/) — 56 NES specifications (NES-000 to NES-053)

## Roadmap

### Completed
- [x] v0.1.0 — Foundation (parser, NEIR, pipeline, CLI)
- [x] v0.2.0 — Compiler Foundation (6 adapters, artifact store, profiles)
- [x] v0.3.0 — Core Specification (Spec v2, validation, context bundles)
- [x] v0.4.0 — MCP Server, migration engine, marketplace, benchmarks
- [x] v1.0.0 — Stable release (test coverage, security hardening, 35+ commands)
- [x] v1.1.0 — Critical fixes (WebSocket races, interface{}→any, godoc, OpenAPI)
- [x] v1.2.0 — Database layer (PostgreSQL/MySQL/SQLite, retry, logging, health checks)
- [x] v1.3.0 — Quality, Correctness & Production Readiness (code gen fixes, security audit, CLI --output json/yaml)
- [x] v1.3.1 — Code Quality & Lint Compliance (999 issues resolved, 22 unused symbols removed)
- [x] v1.4.0 — Prompt Library & Platform Improvements (YAML templates, observability dashboard, workflow manager)
- [x] v1.5.0 — Production Hardening (HTTP timeouts, context propagation, error logging, SSE fixes, test fixes)

### In Progress
- [ ] v1.6.0 — Ecosystem & Documentation
- [ ] v2.0.0 — Dashboard UI, distributed builds

## License

Apache License 2.0

## Status

🟢 **Active Development** — NAEOS is under active development with full features for specification-driven engineering. Latest version: v1.5.0 (Dashboard, Distributed Builds, Prompt Library, Production Hardening).
