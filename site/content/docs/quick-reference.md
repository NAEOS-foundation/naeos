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
naeos compile --all --input-file spec.yaml

# Start API server
naeos serve --port 8080

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
    kind: rest
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
| `rest` | HTTP/JSON | REST APIs |
| `grpc` | gRPC/Protobuf | Internal service communication |
| `graphql` | HTTP/GraphQL | Flexible query APIs |
| `websocket` | WebSocket | Real-time communication |
| `worker` | — | Background job processing |
| `lambda` | — | Serverless functions |
| `reverse-proxy` | HTTP | API gateway / load balancer |

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
| `naeos compile` | Compile spec for AI assistants |
| `naeos gen` | Generate code for a specific language |
| `naeos serve` | Start the API server |
| `naeos dashboard` | Start the web dashboard |
| `naeos cloud plan` | Generate cloud deployment plan |
| `naeos cloud deploy` | Deploy to cloud provider |
| `naeos cloud destroy` | Tear down cloud resources |
| `naeos plugin install` | Install a plugin |
| `naeos plugin list` | List installed plugins |
| `naeos db migrate` | Run database migrations |
| `naeos db reset` | Reset the database |
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
| `NAEOS_LLM_API_KEY` | API key for LLM providers | — |
| `NAEOS_DB_DRIVER` | Database driver (postgres, mysql, sqlite) | sqlite |
| `NAEOS_DB_DSN` | Database connection string | — |
| `NAEOS_PORT` | API server port | 8080 |
| `NAEOS_LOG_LEVEL` | Log level (debug, info, warn, error) | info |

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
│   └── GEMINI.md
├── context/               # Context bundles
│   └── summary.md
└── terraform/             # Cloud deployment (if configured)
    ├── main.tf
    ├── variables.tf
    └── outputs.tf
```

See also: [CLI Reference](/docs/cli-reference/), [Spec Language](/docs/spec-language/), [Architecture](/docs/architecture/)
