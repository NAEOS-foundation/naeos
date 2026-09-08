---
title: Project Structure
description: Repository layout and directory organization for the NAEOS codebase.
weight: 24
---

This document describes the NAEOS repository structure and how the codebase is organized.

## Root Level

```text
naeos/
├── cmd/naeos/           # CLI commands (72 files)
├── internal/            # Internal packages (74 packages)
├── pkg/                 # Public packages (3 packages)
├── docs/                # NES specifications (57 files)
├── site/                # Next.js website (Cloudflare Pages)
├── governance/          # Governance documents (8 files)
├── constitution/        # Engineering constitution (8 files)
├── specification/       # Specification documents (10 files)
├── kernel/              # Kernel specifications (4 files)
├── policy/              # Policy documents (7 files)
├── profile/             # Profile documents (7 files)
├── prompts/             # AI prompt templates
├── templates/           # ADR/RFC templates and starter projects
├── examples/            # Example documents and plugins
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
├── build_cmd.go         # Build artifacts from a specification
├── deploy_cmd.go        # Deploy pipeline output to a target environment
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
├── marketplace_cmd.go   # Plugin/template marketplace
├── status_cmd.go        # Pipeline status
├── docsgen_cmd.go       # CLI documentation generation
├── docgen_cmd.go        # Documentation generation from specs
├── ai_cmd.go            # AI assistance
├── init_cmd.go          # Config initialization
├── version_cmd.go       # Version info
├── completion_cmd.go    # Shell completion
├── auth_cmd.go          # Authentication and RBAC
├── security_cmd.go      # Security and secrets management
├── supabase_cmd.go      # Supabase backend management
├── workflow_cmd.go      # Workflow and approval management
├── config_cmd.go        # Configuration management
├── db_cmd.go            # Database connections and migrations
├── gateway_cmd.go       # API gateway management
├── observability_cmd.go # Observability and telemetry
├── mcp_cmd.go           # MCP server
├── lsp_cmd.go           # Language Server Protocol server
└── helpers.go           # Shared output helpers
```

> The full per-command reference is auto-generated in [`docs/cli/`](/docs/cli-reference/).

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
│   └── adapters/        # 7 output adapters (Claude, Codex, Copilot, Cursor, dll)
├── context/             # Context bundles
│   └── bundle/          # Bundle generators (markdown, plain text, JSON)
├── generation/          # Code generation
│   ├── engine/          # Generation engine
│   ├── adapters/        # Language adapters (Go, TypeScript, Python, Java, Rust, FastAPI, Actix Web)
│   └── renderers/       # Template renderers
├── governance/          # Governance
│   ├── policy/          # Policy evaluator
│   └── review/          # Artifact review
├── artifacts/           # Artifact store with content-hash dedup
├── profiles/            # Industry profiles (6 built-in)
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
├── scheduler/           # Task scheduler
├── planner/             # DAG-based task scheduling
├── runtime/             # Runtime engine
├── profiling/           # Performance profiling
├── registry/            # Service registry
├── lint/                # Lint rules
├── create/              # Project creation
├── auth/                # Authentication, RBAC, and SSO
├── broker/              # Message broker (Redis, RabbitMQ, Kafka)
├── gateway/             # API gateway (load balancing, circuit breaker, rate limiting)
├── graphql/             # GraphQL API server
├── lsp/                 # Language Server Protocol server
├── events/              # Internal event bus (pub/sub)
├── messagequeue/        # Message queue
├── monitor/             # Monitoring
├── telemetry/           # Telemetry and metrics
├── workflow/            # Workflow and approval engine
├── schemaregistry/      # Schema registry
├── search/              # Full-text search engine
├── supabase/            # Supabase client
├── promptlib/           # Prompt library
├── pluginhost/          # Plugin host runtime
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
└── config/              # Configuration management
```

## File Count

| Directory | Files | Description |
|-----------|-------|-------------|
| `cmd/naeos/` | 72 | CLI commands (non-test `.go` files) |
| `internal/` | 74 | Internal packages |
| `pkg/` | 3 | Public API packages |
| `docs/` | 57 | NES specification documents |
| `site/` | 100+ | Next.js website content and layouts |
| `governance/` | 8 | Governance documents |
| `specification/` | 10 | Specification documents |
| **Total** | **320+** | |

See also: [Architecture](/docs/architecture/), [Pipeline Engine](/docs/pipeline-engine/)
