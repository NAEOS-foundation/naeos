---
title: "Build an E-Commerce Platform with NAEOS: End-to-End Tutorial"
description: "From a single YAML specification to a validated multi-language e-commerce platform with AI context — a complete NAEOS walkthrough."
date: 2026-08-18
author: "NAEOS Foundation"
categories: ["tutorial"]
---

In this tutorial you'll build a complete e-commerce platform from a single YAML specification. No hand-written boilerplate, no separate generators per language — one spec, and NAEOS derives the rest: module structure, service endpoints, deployment strategy, and even AI instruction sets for your coding assistants.

We'll use `examples/spec-full.yaml` from the repository — it's the reference spec that exercises most of the NAEOS pipeline.

## Prerequisites

```bash
curl -fsSL https://naeos.dev/install.sh | sh
naeos version   # should print 3.4.0
```

## Step 1: Create the Project

The interactive wizard scaffolds the project layout:

```bash
naeos create
```

Enter `e-commerce-platform` as the project name. The wizard sets up the workspace structure and a starter configuration.

## Step 2: Write the Specification

Save this as `spec.yaml`:

```yaml
project: e-commerce-platform

modules:
  - name: auth
    path: ./internal/auth
    description: Authentication and authorization module
    dependencies: [core, user]

  - name: core
    path: ./internal/core
    description: Core business logic and shared utilities

  - name: user
    path: ./internal/user
    description: User management and profile module
    dependencies: [core]

  - name: order
    path: ./internal/order
    description: Order processing module
    dependencies: [core, user, payment]

  - name: payment
    path: ./internal/payment
    description: Payment processing module
    dependencies: [core]

services:
  - name: api-gateway
    kind: http
    port: 8080
    description: Main REST API gateway
    endpoints:
      - { method: POST, path: /auth/login, action: login }
      - { method: POST, path: /auth/register, action: register }
      - { method: GET,  path: /users, action: listUsers }
      - { method: GET,  path: /users/:id, action: getUser }
      - { method: POST, path: /orders, action: createOrder }
      - { method: GET,  path: /orders/:id, action: getOrder }
      - { method: POST, path: /payments, action: processPayment }

  - name: worker
    kind: worker
    port: 9090
    description: Background job processor

architecture:
  pattern: hexagonal
  description: Hexagonal architecture with clear separation of core logic from infrastructure

deployment:
  strategy: blue-green
  environments: [development, staging, production]

testing:
  strategy: unit
  coverage: "85%"

generation:
  languages: [go, typescript]
  output_dir: ./out
  module_dir: ./internal
```

Notice what's *not* in the spec: no HTTP framework details, no ORM choices, no file layout. The architecture pattern and the language adapters handle those.

## Step 3: Validate Before Generating

The pipeline can validate the spec without writing anything:

```bash
naeos validate --input-file spec.yaml
```

This runs the parse → normalize → resolve → NEIR build stages and reports issues, warnings, and cross-reference errors. Fix any reported problems before generating.

## Step 4: Run the Pipeline

```bash
naeos run --input-file spec.yaml
```

The 11-stage pipeline executes: build the NEIR model, validate, evaluate policy, schedule the DAG, generate artifacts through the Go and TypeScript adapters, review, and write to `./out/`. You'll see per-stage progress and a summary of artifacts written.

Check the output:

```bash
find out -type f | head -20
```

You should see the Go module skeletons (`internal/auth`, `internal/order`, ...), TypeScript artifacts, Dockerfiles, and deployment manifests — all consistent, all derived from the same model.

## Step 5: Iterate with Watch Mode

While you refine the spec, let NAEOS regenerate on every change:

```bash
naeos watch --input-file spec.yaml
```

Save the spec — the pipeline re-runs and only rewrites what changed. This is where the v3.1.0 pipeline cache shines: unchanged stages are reused via `--cache-dir`:

```bash
naeos run --input-file spec.yaml --cache-dir .naeos-cache
```

## Step 6: Give Your AI Assistants Real Context

Now the fun part. Compile the NEIR model into instruction sets your AI tools can actually read:

```bash
# GitHub Copilot
naeos ai compile --input-file spec.yaml --target copilot

# Claude Code
naeos ai compile --input-file spec.yaml --target claude

# Cursor
naeos ai compile --input-file spec.yaml --target cursor
```

Each produces the right file for that agent (`CLAUDE.md`, `.cursorrules`, `.github/copilot-instructions.md`, ...) containing the architecture, dependency graph, and conventions — so your AI assistant stops guessing and starts building against the actual system.

You can also generate a full context bundle:

```bash
naeos context --input-file spec.yaml
```

## Step 7: Test the Generated Code

The multi-language test runner executes the generated test suites:

```bash
naeos test --input-file spec.yaml
```

## What You Built

In about ten minutes, from one spec:

- A Go backend with five modules and a dependency graph (`auth`, `user`, `order`, `payment`, `core`)
- An HTTP API gateway with typed endpoints
- A background worker service
- TypeScript artifacts for tooling
- Docker + deployment manifests (blue-green strategy, three environments)
- AI instruction sets for six coding assistants
- Unit test scaffolding with an 85% coverage target

Hand-written, this is weeks of boilerplate — and it would start drifting the moment you merged it. With NAEOS, the spec is the system. Everything else is derived, validated, and reproducible.

## Next Steps

- Explore more examples in [`examples/`](https://github.com/NAEOS-foundation/naeos/tree/main/examples)
- Read the [specification language reference](/docs/specification-language/)
- Try the [MCP server deep dive](/blog/mcp-server-deep-dive/) to wire this spec into your editor