---
title: Project Structure
description: Repository layout and directory organization for the NAEOS codebase.
weight: 24
---

This document describes the NAEOS repository structure and how the codebase is organized.

## Root Level

```text
naeos/
├── cmd/naeos/           # CLI commands (35+ files)
├── internal/            # Internal packages (60+ files)
├── pkg/                 # Public packages (4 packages)
├── docs/                # NES specifications (56 files)
├── site/                # Hugo website
├── governance/          # Governance documents (8 files)
├── constitution/        # Engineering constitution
├── specification/       # Specification documents (10 files)
├── kernel/              # Kernel specifications (4 files)
├── policy/              # Policy documents (7 files)
├── prompts/             # AI prompt templates
├── templates/           # ADR/RFC templates
├── examples/            # Example documents
├── wiki/                # Project wiki (17 pages)
├── Reference Architecture/ # Reference architecture docs
├── go.mod               # Go module definition
├── go.sum               # Go module checksum
├── Makefile             # Build automation
├── README.md            # Project readme
├── CHANGELOG.md         # Version history
├── CONTRIBUTING.md      # Contribution guidelines
├── LICENSE              # Apache 2.0 license
├── .github/             # GitHub workflows and config
├── .gitignore           # Git ignore rules
├── .golangci.yml        # Linter configuration
├── .goreleaser.yaml     # Release configuration
└── Dockerfile           # Multi-stage Docker build
```

## cmd/naeos/

CLI commands organized by feature area:

```text
cmd/naeos/
├── main.go              # Root command and entry point
├── run_cmd.go           # Pipeline execution
├── validate_cmd.go      # Spec validation
├── lint_cmd.go          # Spec linting
├── create_cmd.go        # Project creation
├── scaffold_cmd.go      # Code scaffolding
├── export_cmd.go        # Artifact export
├── compile_cmd.go       # AI compilation
├── context_cmd.go       # Context bundles
├── profile_cmd.go       # Industry profiles
├── artifacts_cmd.go     # Artifact management
├── migrate_cmd.go       # Schema migration
├── doctor_cmd.go        # System health check
├── diff_cmd.go          # Spec comparison
├── watch_cmd.go         # File watching
├── workspace_cmd.go     # Workspace management
├── kernel_cmd.go        # Kernel inspection
├── plugin_cmd.go        # Plugin management
├── template_cmd.go      # Template management
├── lock_cmd.go          # Dependency locking
├── rollback_cmd.go      # Rollback changes
├── audit_cmd.go         # Spec auditing
├── marketplace_cmd.go   # Profile marketplace
├── status_cmd.go        # Pipeline status
├── docs_cmd.go          # Documentation generation
├── ai_cmd.go            # AI assistance
├── init_cmd.go          # Config initialization
├── version_cmd.go       # Version info
├── completion_cmd.go    # Shell completion
├── repair_cmd.go        # Spec repair
├── preview_cmd.go       # Preview mode
├── test_cmd.go          # Test runner
├── mcp_cmd.go           # MCP server
├── helpers.go           # Shared helpers
├── e2e_test.go          # End-to-end tests
└── main_test.go         # Unit tests
```

## internal/

```text
internal/
├── specification/       # Spec processing
│   ├── parser/          # YAML/JSON parser with variable interpolation
│   ├── normalizer/      # Data normalization
│   └── resolver/        # Cross-reference resolution
├── neir/                # NEIR model
│   ├── model/           # Model definitions (Project, Module, Service, etc.)
│   ├── builder/         # NEIR builder from parsed specs
│   └── validator/       # NEIR model validator
├── compiler/            # AI instruction compiler
│   └── adapters/        # 6 output adapters (Copilot, Claude, Cursor, etc.)
├── context/             # Context bundles
│   └── bundle/          # Bundle generator (markdown, plain text, JSON)
├── generation/          # Code generation
│   ├── engine/          # Generation engine
│   ├── adapters/        # Language adapters (Go, TS, Python, Java, Rust)
│   └── renderers/       # Template renderers
├── governance/          # Governance
│   ├── policy/          # Policy evaluator
│   └── review/          # Artifact review
├── artifacts/           # Artifact store with content-hash dedup
├── profiles/            # Industry profiles (5 built-in)
├── migration/           # Schema migration engine
├── security/            # Security rules and scanning
├── marketplace/         # Profile & plugin marketplace
├── pluginsdk/           # Plugin SDK with WASM runtime
├── ai/                  # AI service and LLM integration
├── mcp/                 # MCP server implementation
├── knowledge/           # Knowledge graph
├── database/            # Database layer (PostgreSQL, MySQL, SQLite)
├── websocket/           # WebSocket real-time communication
├── eventsourcing/       # Event sourcing and aggregate snapshots
├── distributed/         # Distributed task execution
├── configreload/        # Configuration hot-reload
├── pipelinecache/       # Pipeline result caching
├── pipelinemiddleware/  # Composable pipeline middleware
├── audit/               # Audit logging layer
├── hcl/                 # HCL configuration parser
├── profiledetect/       # Automatic language/framework detection
├── testrunner/          # Multi-language test runner
├── docgen/              # Documentation generator
├── diff/                # Diff engine with colorized output
├── watch/               # File watcher for hot-reload
├── lock/                # Dependency locking
├── rollback/            # Rollback management
├── workspace/           # Workspace management
├── templates/           # Template engine
├── planner/             # DAG-based task scheduling
├── runtime/             # Runtime engine
├── profiling/           # Performance profiling
├── registry/            # Service registry
├── lint/                # Lint rules
├── create/              # Project creation
└── shared/              # Shared utilities
    ├── log/             # Structured logging (slog)
    ├── strutil/         # String utilities
    └── contracts/       # Shared contracts
```

## pkg/

Public packages that external consumers can import:

```text
pkg/
├── pipeline/            # Main pipeline orchestration
├── kernel/              # System kernel (registry, event bus, telemetry)
├── config/              # Configuration management
└── plugin/              # Plugin system interface
```

## File Count

| Directory | Files | Description |
|-----------|-------|-------------|
| `cmd/naeos/` | 35 | CLI commands and entry point |
| `internal/` | 60+ | Internal packages |
| `pkg/` | 4 | Public API packages |
| `docs/` | 56 | NES specifications |
| `site/` | 100+ | Hugo website content and layouts |
| `governance/` | 8 | Governance documents |
| `specification/` | 10 | Specification documents |
| **Total** | **270+** | |

See also: [Architecture](/docs/architecture/), [Pipeline Engine](/docs/pipeline-engine/)
