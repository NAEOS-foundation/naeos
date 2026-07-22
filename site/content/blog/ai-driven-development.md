---
title: "AI-Driven Development with NAEOS: Teaching Your Assistants to Think in Architecture"
description: "How NAEOS compiles NEIR into AI instruction sets that give Copilot, Claude, and Cursor real architectural context."
date: 2026-07-15
categories: ["tutorial"]
---

AI coding assistants are powerful — but they're working blind. When you ask GitHub Copilot to generate a function, it sees the file you're in. When you ask Claude Code to refactor a module, it sees the directory you pointed it at. None of them see the architecture.

NAEOS changes this with its **AI Compiler** — a system that transforms your NEIR model into instruction sets that coding assistants can consume. The result: your AI tools don't just see code, they see the system.

## The Context Problem

Modern AI assistants excel at local tasks. They can autocomplete a function, suggest a type, or refactor a class with impressive accuracy. But they fail at **architectural tasks** — the kind that require understanding how modules connect, what patterns your team uses, and what constraints apply.

Consider asking an AI: "Add a new endpoint to the order service that calls the payment service."

Without context, the AI might:
- Generate an HTTP client that doesn't match your existing RPC pattern
- Miss the error handling contract between services
- Ignore the rate limiting policy defined in your governance rules
- Use a different logging format than the rest of the codebase

With NAEOS context, the AI knows:
- The order service uses gRPC, not REST
- Payment errors must be wrapped in `PaymentError` with a specific error code
- All inter-service calls go through the service mesh with retry policies
- Structured logging uses the `slog` package with a request ID field

## How the AI Compiler Works

The AI Compiler takes your NEIR model and generates **context bundles** — structured instruction sets tailored for each AI platform:

```bash
# Generate context for all supported AI platforms
naeos compile --all --input-file spec.yaml

# Generate for a specific platform
naeos compile --target copilot --input-file spec.yaml
```

The compiler supports six platforms:

| Platform | Output Format | How It's Used |
|----------|--------------|---------------|
| GitHub Copilot | `copilot-instructions.md` | Loaded automatically by Copilot |
| Claude Code | `CLAUDE.md` | Read by Claude on session start |
| Cursor | `.cursorrules` | Applied to all Cursor sessions |
| Gemini CLI | `GEMINI.md` | Context for Gemini CLI |
| Codex | `AGENTS.md` | Instructions for OpenAI Codex |
| OpenCode | `AGENTS.md` | Instructions for OpenCode |

## What's Inside a Context Bundle

A context bundle isn't just a README dump. It's a structured document that captures:

### Architecture Overview
```markdown
## Architecture Pattern: Microservices

### Service Mesh
- Protocol: gRPC with HTTP/2
- Load balancing: Round-robin
- Circuit breaking: 5 consecutive failures → open for 30s

### Services
- order-service (port 9001): Handles order CRUD and lifecycle
- payment-service (port 9002): Processes payments via Stripe API
- inventory-service (port 9003): Manages stock levels
```

### Dependency Graph
```markdown
## Module Dependencies

order-service
  ├── payment-service (gRPC)
  └── inventory-service (gRPC)

payment-service
  └── (external: Stripe API)

inventory-service
  └── postgres-db (connection pool)
```

### Coding Conventions
```markdown
## Conventions

### Error Handling
- All service errors return `ServiceError` struct
- Error codes follow: `{module}.{action}.{reason}`
- Example: `order.payment.insufficient_funds`

### Logging
- Use `slog` with structured attributes
- Every log entry must include `request_id` and `service_name`
```

### Governance Rules
```markdown
## Policies

- No direct database access between services (use APIs)
- All API endpoints must have rate limiting
- Authentication via JWT with RS256 signing
```

## Real-World Example

Let's walk through a concrete example. Say you have a spec for a chat application:

```yaml
project: chat-app
modules:
  - name: gateway
    path: ./gateway
    dependencies: [message-service, user-service]
  - name: message-service
    path: ./services/messages
    dependencies: [redis-cache, postgres-db]
  - name: user-service
    path: ./services/users
    dependencies: [postgres-db]
services:
  - name: ws-gateway
    kind: websocket
    port: 8080
  - name: message-api
    kind: rest
    port: 9001
  - name: user-api
    kind: rest
    port: 9002
architecture:
  pattern: microservices
generation:
  languages: [go, typescript]
```

Running `naeos compile --all` produces six different instruction files, each optimized for its target platform. When you open your project in Cursor, the `.cursorrules` file tells Claude about the WebSocket gateway, the message service's Redis caching strategy, and the user service's JWT validation flow — before you type a single prompt.

## The Workflow

The full workflow looks like this:

1. **Write the spec** — Define your system in YAML
2. **Run the pipeline** — Generate code + context bundles
3. **Open your IDE** — AI assistants load the context automatically
4. **Code with context** — AI suggestions respect your architecture

```bash
# One command generates everything
naeos run --input-file spec.yaml --compile-all

# Your project now has:
# - Generated Go and TypeScript code
# - AI instruction sets for all 6 platforms
# - Validated dependency graph
# - Governance-compliant structure
```

## Tips for Better AI Context

The quality of AI output depends on the quality of your spec. Here are some tips:

**Be explicit about patterns.** Instead of just listing services, describe their interaction patterns:

```yaml
architecture:
  pattern: microservices
  description: |
    Event-driven architecture with:
    - Synchronous gRPC for request/response
    - Redis Streams for async events
    - Circuit breakers on all external calls
```

**Document conventions in the spec.** The AI compiler includes your spec's `description` fields in the context bundle:

```yaml
modules:
  - name: order-service
    description: |
      Order lifecycle management.
      Uses CQRS pattern: commands go to the write model,
      queries hit the read model (Elasticsearch).
      All state changes emit domain events to Redis Stream.
```

**Keep the dependency graph accurate.** The AI uses this to understand what can call what. Missing dependencies lead to AI-generated code that bypasses your service boundaries.

## What's Next

We're working on making the context bundles even richer:

- **Incremental compilation** — Only regenerate context for changed modules
- **Custom instruction templates** — Override the default format for your team's conventions
- **Live context sync** — Context bundles update as your spec evolves, so AI always has the latest architecture

The goal is simple: make your AI coding assistants as knowledgeable about your system as your senior engineers are. Declarative specs are the perfect source for that knowledge — they're precise, complete, and machine-readable.
