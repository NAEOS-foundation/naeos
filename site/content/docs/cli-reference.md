---
title: CLI Reference
description: Complete command reference for all NAEOS CLI commands.
---

NAEOS provides a complete CLI with commands for every stage of the engineering pipeline.

## Core Commands

| Command | Description |
|---------|-------------|
| `naeos run` | Execute the full pipeline |
| `naeos validate` | Validate a specification using the NAEOS pipeline |
| `naeos ai compile` | Compile a spec into AI instruction sets |
| `naeos context` | Generate AI context bundles from specifications |
| `naeos test` | Run tests for generated code |
| `naeos docgen` | Generate documentation from specification |

## Project & Spec Commands

| Command | Description |
|---------|-------------|
| `naeos init` | Initialize a new NAEOS project or generate config |
| `naeos create` | Interactive project creation wizard |
| `naeos scaffold` | Generate a starter project scaffold |
| `naeos import` | Import specifications from HCL format to NAEOS YAML/JSON |
| `naeos export` | Export generated artifacts to a directory |
| `naeos audit` | Security audit of generated or source files |
| `naeos diff` | Compare generated artifacts with existing output directory |
| `naeos repair` | Repair the NAEOS output directory |
| `naeos lint` | Lint a specification file |
| `naeos inspect` | Inspect the NAEOS pipeline result |
| `naeos preview` | Preview generated artifacts without writing them |
| `naeos distributed` | Run pipeline tasks in distributed mode across multiple workers |

## Build, Deploy & CI/CD

| Command | Description |
|---------|-------------|
| `naeos build` | Build artifacts from a specification |
| `naeos cicd` | Generate CI/CD pipeline configuration |
| `naeos cloud` | Cloud deployment commands |
| `naeos deploy` | Deploy the pipeline output to a target environment |
| `naeos gateway` | API gateway management |
| `naeos graphql` | Start GraphQL API server |

## Management Commands

| Command | Description |
|---------|-------------|
| `naeos mcp` | Start MCP (Model Context Protocol) server |
| `naeos marketplace` | Browse and install templates, profiles, and plugins |
| `naeos profile` | Manage industry-specific project profiles |
| `naeos plugin` | Manage NAEOS plugins |
| `naeos template` | Manage generation templates, prompt library, and template marketplace |
| `naeos workspace` | Manage multi-module workspaces |
| `naeos artifacts` | Manage generated project artifacts |
| `naeos migrate` | Manage spec schema migrations |
| `naeos migration` | Database migration management |
| `naeos rollback` | Rollback to a previous snapshot of generated artifacts |
| `naeos lock` | Manage lock files for reproducible builds |

## Monitoring & Operations

| Command | Description |
|---------|-------------|
| `naeos status` | Show current pipeline, system and project status |
| `naeos doctor` | Run diagnostics on the NAEOS environment and configuration |
| `naeos watch` | Watch for specification changes and re-run the pipeline |
| `naeos kernel` | Inspect the NAEOS kernel and service registry |
| `naeos dashboard` | Start NAEOS web dashboard |
| `naeos api` | Start NAEOS REST API server |
| `naeos health` | Run system health checks and diagnostics |
| `naeos monitor` | Start monitoring server with Prometheus metrics |
| `naeos observability` | Observability and telemetry management |
| `naeos events` | Event sourcing commands for pipeline audit trail and replay |
| `naeos history` | Show pipeline run history from persisted events |

## Security & Compliance

| Command | Description |
|---------|-------------|
| `naeos auth` | Authentication and authorization management |
| `naeos security` | Security and secrets management |
| `naeos compliance` | Compliance reporting and audit log export |
| `naeos config` | Configuration management commands |

## Data & Integration

| Command | Description |
|---------|-------------|
| `naeos db` | Database connection and migration management |
| `naeos broker` | Message broker management |
| `naeos ws` | Start WebSocket server for real-time updates |
| `naeos lsp` | Start NEIR Language Server Protocol server |
| `naeos supabase` | Supabase backend management |

## Developer Experience & Utility

| Command | Description |
|---------|-------------|
| `naeos ai` | AI-powered assistance commands |
| `naeos dx` | Developer experience tools |
| `naeos tui` | Terminal user interface tools |
| `naeos schema` | NEIR schema registry operations |
| `naeos search` | Full-text search engine management |
| `naeos perf` | Performance optimization tools |
| `naeos benchmark` | Run pipeline benchmarks |
| `naeos docs` | Generate project documentation |
| `naeos completion` | Generate shell completion scripts |
| `naeos version` | Show NAEOS version |

For detailed usage of each command, run `naeos <command> --help`.

## Download

- [CLI Reference PDF](/downloads/naeos-cli-reference.pdf)
