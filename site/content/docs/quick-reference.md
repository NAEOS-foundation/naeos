---
title: Quick Reference
description: Common commands, patterns, and configurations at a glance.
weight: 2
---

A quick reference card for NAEOS — ideal for experienced users who need a fast lookup.

## Essential Commands

```bash
# Install
go install github.com/NAEOS-foundation/naeos/cmd/naeos@latest

# Create a new project
naeos init my-project

# Run the full pipeline
naeos run --input-file spec.yaml

# Validate a spec
naeos validate --input-file spec.yaml

# Compile for AI
naeos ai compile --input-file spec.yaml --target opencode

# Start the REST API server
naeos api --port 8080

# Start dashboard
naeos dashboard
```

## Spec Minimal Example

```yaml
project: my-service
modules:
  - name: api
    path: ./api
    dependencies: [database]
  - name: database
    path: ./db
services:
  - name: rest-api
    kind: http
    port: 8080
architecture:
  pattern: microservices
generation:
  languages: [go, typescript]
```

## Module Options

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Module identifier (required) |
| `path` | string | Filesystem path (required) |
| `description` | string | Human-readable description |
| `dependencies` | list | Other module names |
| `kind` | string | Module type (default: service) |

## Service Kinds

| Kind | Protocol | Use Case |
|------|----------|----------|
| `http` | HTTP/JSON | REST APIs, gateways, WebSocket endpoints |
| `grpc` | gRPC/Protobuf | Internal service communication |
| `worker` | — | Background job processing / serverless tasks |
| `cli` | — | Command-line tools |
| `job` | — | One-off scheduled jobs |

## Architecture Patterns

| Pattern | Description | Best For |
|---------|-------------|----------|
| `microservices` | Independent, loosely coupled services | Large teams, complex domains |
| `monolithic` | Single deployable unit | Small teams, simple domains |
| `serverless` | Function-as-a-service | Event-driven, variable load |
| `event-driven` | Async message passing | High throughput, decoupling |

## Generation Languages

| Language | Adapter | Output |
|----------|---------|--------|
| Go | `go` | `.go` files with modules, packages |
| TypeScript | `typescript` | `.ts` files with interfaces |
| Python | `python` | `.py` files with classes |
| Java | `java` | `.java` files with packages |
| Rust | `rust` | `.rs` files with crates |

## CLI Quick Reference

| Command | Description |
|---------|-------------|
| `naeos init` | Create a new project |
| `naeos run` | Execute the full pipeline |
| `naeos validate` | Validate a specification |
| `naeos ai compile` | Compile spec for an AI agent |
| `naeos build` | Build artifacts from a specification |
| `naeos api` | Start the REST API server |
| `naeos dashboard` | Start the web dashboard |
| `naeos cloud plan` | Generate cloud deployment plan |
| `naeos cloud deploy` | Deploy to cloud provider |
| `naeos cloud status` | Show deployed resource status |
| `naeos plugin install` | Install a plugin |
| `naeos plugin list` | List installed plugins |
| `naeos db migrate` | Run database migrations |
| `naeos db connect` | Connect to a database |
| `naeos version` | Show version info |

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/health` | Health check |
| GET | `/api/v1/version` | Version info |
| POST | `/api/v1/specs/validate` | Validate spec |
| POST | `/api/v1/specs/compile` | Compile spec |
| POST | `/api/v1/pipeline/run` | Run pipeline |
| GET | `/api/v1/pipeline/status` | Pipeline status |
| GET | `/api/v1/artifacts` | List artifacts |
| POST | `/api/v1/context/generate` | Generate context |
| POST | `/api/v1/ai/enrich/stream` | AI enrichment (SSE) |
| POST | `/api/v1/ai/compile/stream` | AI compile (SSE) |
| GET | `/api/v1/plugins` | List plugins |
| WS | `/ws` | WebSocket real-time events |

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `NAEOS_LLM_API_KEY` | API key for LLM providers (`naeos ai compile`) | — |
| `NAEOS_LLM_PROVIDER` | LLM provider (openai, anthropic, gemini, ...) | — |
| `NAEOS_ENCRYPTION_KEY` | Passphrase for the API server's auth user store (`naeos api`) | — |
| `NAEOS_PIPELINES_FILE` | Path to the pipelines file (default: `~/.naeos/pipelines.json`) | — |

## Output Directory Structure

```
output/
├── go/                    # Generated Go code
│   ├── cmd/
│   ├── internal/
│   └── go.mod
├── typescript/            # Generated TypeScript
│   ├── src/
│   ├── package.json
│   └── tsconfig.json
├── ai/                    # AI instruction sets
│   ├── copilot-instructions.md
│   ├── CLAUDE.md
│   ├── .cursorrules
│   └── .gemini/CONFIG.md
├── context/               # Context bundles
│   └── summary.md
└── terraform/             # Cloud deployment (if configured)
    ├── main.tf
    ├── variables.tf
    └── outputs.tf
```

See also: [CLI Reference](/docs/cli-reference/), [Spec Language](/docs/spec-language/), [Architecture](/docs/architecture/)
