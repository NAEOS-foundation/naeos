---
title: Features
description: Explore the complete feature set of NAEOS — from specification parsing to AI-assisted code generation.
---

NAEOS provides a complete engineering platform with the following capabilities:

## Pipeline Engine

The 9-stage DAG pipeline is the heart of NAEOS:

1. **Parse** — Read and parse YAML/JSON specifications with variable interpolation
2. **Normalize** — Normalize data structures for consistent processing
3. **Resolve** — Resolve cross-references, dependencies, and external includes
4. **Build** — Build the NEIR (NAEOS Engineering Intermediate Representation)
5. **Validate** — Comprehensive validation including circular dependency detection
6. **Schedule** — DAG-based task scheduling with parallel execution groups
7. **Generate** — Multi-language code generation (Go, TypeScript, Python, Java, Rust)
8. **Compile** — Compile NEIR to AI instruction sets for 6 platforms
9. **Export** — Export artifacts, documentation, and deployment manifests

## Spec Language v2

The NAEOS Specification Language v2 provides a rich set of features:

- `${var}` — Variable interpolation
- `$env{VAR}` — Environment variable resolution
- `$ref{path}` — Cross-reference resolution
- `$include{file}` — Multi-file spec composition
- `$fn{name(args)}` — Custom functions (upper, lower, slug, default, len, coalesce)
- `$if{condition}` / `$endif` — Conditional sections
- Schema versioning with auto-check (minimum v0.1.0)

## NEIR Model

The NAEOS Engineering Intermediate Representation is the canonical model representing the entire system. It encompasses:

- Project metadata and configuration
- Architecture patterns (microservices, serverless, monolithic, hexagonal)
- Domain model and bounded contexts
- Module structure and dependencies
- Service definitions with endpoints and ports
- API contracts (REST, GraphQL, WebSocket)
- Storage and database configuration
- Infrastructure and cloud resources
- Security policies and rules
- AI integration settings
- Documentation requirements
- Deployment targets (Kubernetes, Docker, serverless)
- Testing strategies
- CI/CD pipeline configuration

## Code Generator

Generate production-ready code with per-language adapters:

- **Go** — Standard library patterns, net/http, testing
- **TypeScript** — Express/NestJS patterns, type safety
- **Python** — FastAPI patterns, type hints
- **Java** — Spring Boot patterns, JUnit 5
- **Rust** — Axum patterns, async/await

## AI Compiler

Transform NEIR into AI instruction sets for 6 coding assistants:

- **GitHub Copilot** — `.github/copilot-instructions.md`
- **Claude Code** — `CLAUDE.md`
- **Cursor** — `.cursorrules`
- **Gemini CLI** — `.gemini/CONFIG.md`
- **Codex** — `AGENTS.md`
- **OpenCode** — `AGENTS.md`

## Governance & Policy

Built-in policy engine and governance framework:

- 7 policy operators for rule definition
- 5 default policy rules
- RBAC permission system
- Audit trail with full traceability
- Artifact review and approval workflows
- Policy evaluator with structured output

## Marketplace

Publish, discover, and install extensions:

- **Profile Marketplace** — Industry profiles (SaaS, AI Agent, FinTech, Healthcare, Government)
- **Plugin Marketplace** — WASM and native plugins
- **Template Marketplace** — Project templates and scaffolds
- SHA-256 signature verification for security

## Developer Tools

Additional tools for a complete development workflow:

- 35+ CLI commands
- Watch mode for hot-reload
- Diff engine for spec comparison
- Migration engine for schema versioning
- Documentation generator
- Multi-language test runner
- MCP server for AI agent integration
- WebSocket dashboard for real-time monitoring
- Distributed task execution
- Event sourcing and audit logging