---
title: AI Compiler
description: Transform NEIR into AI instruction sets for 7 coding assistants.
---

## Overview

The NAEOS AI Compiler transforms a NAEOS specification into platform-specific instruction files for AI coding assistants. This ensures your AI tools understand your project architecture, conventions, and dependencies — reducing manual prompt engineering and improving code quality.

## Supported Platforms

| Platform | Target ID | File |
|----------|-----------|------|
| GitHub Copilot | `copilot` | `.github/copilot-instructions.md` |
| Claude Code | `claude` | `CLAUDE.md` |
| Cursor | `cursor` | `.cursorrules` |
| Gemini CLI | `gemini` | `.gemini/CONFIG.md` |
| Codex | `codex` | `AGENTS.md` |
| OpenCode | `opencode` | `AGENTS.md` |
| Windsurf | `windsurf` | `.windsurfrules` |

## How It Works

The compiler sends your specification to an LLM (OpenAI, Anthropic, or Ollama) together with a target-specific prompt, then streams the generated instruction file back:

```text
┌──────────┐     ┌────────────┐     ┌──────────────────┐
│ Spec     │────→│   AI       │────→│  CLAUDE.md       │
│ (YAML/   │     │  Compiler  │     │  .cursorrules    │
│  JSON)   │     │ (LLM-based)│     │  AGENTS.md       │
│          │     └────────────┘     └──────────────────┘
└──────────┘
```

## Usage

```bash
# Compile for a specific AI agent
naeos ai compile --input-file spec.yaml --target claude

# Use a specific LLM provider (default: openai)
naeos ai compile --input-file spec.yaml --target opencode --provider anthropic
```

Remember to set `NAEOS_LLM_API_KEY` (and optionally `NAEOS_LLM_PROVIDER`) before running:

```bash
export NAEOS_LLM_API_KEY="your-api-key"
export NAEOS_LLM_PROVIDER="anthropic"
```

The compiled instruction file is printed to stdout — pipe it to your target file:

```bash
naeos ai compile --input-file spec.yaml --target claude > CLAUDE.md
```

For offline, deterministic context generation without an LLM, use:

```bash
# Generate an AI context bundle in the current directory
naeos context --input-file spec.yaml
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

- Run `naeos ai compile` as part of your project setup script
- Commit generated instruction files to your repository
- Recompile whenever your architecture changes
- Use `naeos context` for a bundled, offline approach
- Pipe output to platform-specific files (e.g. `CLAUDE.md`, `AGENTS.md`)
