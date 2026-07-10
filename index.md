---
layout: default
title: NAEOS
---

# NAEOS — Declarative Platform Engineering System

NAEOS transforms YAML/JSON specifications into validated, multi-language project structures with full traceability from intent to implementation.

## Quick Start

```bash
# Install
go install github.com/NAEOS-foundation/naeos/cmd/naeos@latest

# Initialize
naeos init

# Run pipeline
naeos run --config config.yaml --input spec.yaml

# Validate
naeos validate --config config.yaml --input spec.yaml

# Export artifacts
naeos export --config config.yaml --input spec.yaml --output-dir ./out
```

## Architecture

{% include_relative ARCHITECTURE-OVERVIEW.md %}

## Documentation

### Core
- [Foundation](docs/NES-000-Foundation.md) — Architectural principles
- [Repository](docs/NES-001-Repository.md) — Repository structure
- [Kernel](docs/NES-002-Kernel.md) — Runtime kernel layer
- [Workspace](docs/NES-003-Workspace.md) — Project workspace model
- [Bootstrap](docs/NES-004-Bootstrap.md) — Initialization modes

### Pipeline
- [Pipeline](docs/NES-026-Pipeline.md) — Orchestration layer
- [Compiler](docs/NES-013-Compiler.md) — 9-stage compilation pipeline
- [Generator](docs/NES-007-Generator.md) — Code generation engine
- [Validator](docs/NES-014-Validator.md) — Validation pipeline
- [Graph](docs/NES-011-Graph.md) — Dependency graph model

### NEIR Model
- [NEIR](docs/NES-023-NEIR.md) — Canonical NEIR model
- [NEIR Model](docs/NES-023-NEIR-Model.md) — Go domain model reference
- [Internal Structure](docs/NES-024-Internal-Structure.md) — Package structure

### SDK & Adapters
- [SDK](docs/NES-019-SDK.md) — Programmatic integration
- [SDK Multi-Language](docs/NES-039-SDK-MultiLanguage.md) — Multi-language support
- [Output Adapter Architecture](docs/NES-040-Output-Adapter-Architecture.md) — Adapter pattern

### Specifications
- [SPEC-001 Overview](specification/NAEOS-SPEC-001.md)
- [SPEC-008 Compiler Model](specification/NAEOS-SPEC-008.md)
- [SPEC-006 Dependency Graph](specification/NAEOS-SPEC-006.md)
- [SPEC-005 Rule Model](specification/NAEOS-SPEC-005.md)
- [SPEC-009 Reasoning Graph](specification/NAEOS-SPEC-009.md)
- [SPEC-010 Intent Model](specification/NAEOS-SPEC-010.md)

### Governance
- [Governance](docs/NES-027-Governance.md) — Policy and review system
- [Policy](docs/NES-012-Policy.md) — Policy model
- [Release](docs/NES-022-Release.md) — Release process
- [Security](docs/NES-020-Security.md) — Security model

### Operations
- [CLI Reference](docs/NES-028-CLI-Reference.md) — Command reference
- [Configuration](docs/NES-029-Configuration.md) — Config format
- [Cloud](docs/NES-018-Cloud.md) — Deployment targets
- [Telemetry](docs/NES-032-Telemetry.md) — Observability
- [Event Bus](docs/NES-034-Event-Bus.md) — Internal pub/sub

### Reference
- [Knowledge](docs/NES-010-Knowledge.md) — Knowledge model
- [AI Integration](docs/NES-016-AI.md) — AI assistant support
- [Plugin](docs/NES-009-Plugin.md) — Extension mechanism
- [Testing](docs/NES-021-Testing.md) — Test patterns
- [Errors](docs/NES-031-Errors.md) — Error catalog
- [Version Management](docs/NES-035-Version-Management.md) — SemVer
- [Shared Types](docs/NES-038-Shared-Types-Contracts.md) — Common types

## Supported Languages

| Language | Adapter | Status |
|----------|---------|--------|
| Go | `internal/generation/adapters/go.go` | Active |
| TypeScript | `internal/generation/adapters/typescript.go` | Active |
| Python | `internal/generation/adapters/python.go` | Active |
| Java | `internal/generation/adapters/java.go` | Active |
| Rust | `internal/generation/adapters/rust.go` | Active |

## License

Apache License 2.0
