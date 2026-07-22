---
title: AI Compiler
description: Transform NEIR into AI instruction sets for 6 coding assistants.
---

## Overview

The NAEOS AI Compiler transforms the NEIR model into platform-specific instruction files for AI coding assistants. This ensures your AI tools understand your project architecture, conventions, and dependencies — reducing manual prompt engineering and improving code quality.

## Supported Platforms

| Platform | File | Status |
|----------|------|--------|
| GitHub Copilot | `.github/copilot-instructions.md` | ✅ |
| Claude Code | `CLAUDE.md` | ✅ |
| Cursor | `.cursorrules` | ✅ |
| Gemini CLI | `.gemini/CONFIG.md` | ✅ |
| Codex | `AGENTS.md` | ✅ |
| OpenCode | `AGENTS.md` | ✅ |

## How It Works

```text
┌──────────┐     ┌────────────┐     ┌──────────────────┐
│ NEIR     │────→│   AI       │────→│  copilot-         │
│ Model    │     │  Compiler  │     │  instructions.md  │
│          │     │            │     │  CLAUDE.md        │
│ Project  │     │  ┌──────┐  │     │  .cursorrules     │
│ Modules  │     │  │Adapter│  │     │  CONFIG.md        │
│ Services │     │  │Pipeline│ │     │  AGENTS.md        │
│ APIs     │     │  └──────┘  │     └──────────────────┘
│ Security │     └────────────┘
└──────────┘
```

### Compilation Pipeline

1. **Extract** — Pull relevant sections from NEIR (architecture, modules, dependencies, conventions)
2. **Translate** — Convert to platform-specific format and syntax
3. **Optimize** — Prioritize information by relevance and token budget
4. **Format** — Apply platform-specific formatting rules
5. **Emit** — Write to the target file location

## What Gets Included

Each instruction file typically includes:

- **Project overview** — Name, purpose, architecture pattern
- **Module structure** — Component tree with dependencies
- **Service definitions** — API endpoints, ports, protocols
- **Language stack** — Languages, frameworks, toolchains
- **Code conventions** — Naming, formatting, testing patterns
- **Key dependencies** — Critical libraries and their usage
- **Architecture rules** — Constraints and patterns to follow
- **Security policies** — Authentication, authorization, data handling

## Usage

```bash
# Compile for all platforms
naeos compile --all --input-file spec.yaml

# Compile for specific platform
naeos compile --platform copilot --input-file spec.yaml

# Compile with custom output directory
naeos compile --all --input-file spec.yaml --output-dir .github

# Generate AI context bundle
naeos context --input-file spec.yaml --format bundle
```

## Example Output

When you compile a microservices project, the generated `CLAUDE.md` might contain:

```markdown
# Project: ecommerce-platform
## Architecture: Microservices
## Language Stack: Go, TypeScript, PostgreSQL

### Modules
- api-gateway → depends on: user-service, product-service
- user-service → depends on: database
- product-service → depends on: database, search-engine

### Conventions
- Go: standard library + chi router, sqlx for DB
- TypeScript: Express with Zod validation
- All services expose Prometheus metrics on :9090
```

## Best Practices

- Run `naeos compile` as part of your project setup script
- Commit generated instruction files to your repository
- Recompile whenever your architecture changes
- Use `naeos context` for a bundled approach
- Combine with `naeos watch` for automatic recompilation
