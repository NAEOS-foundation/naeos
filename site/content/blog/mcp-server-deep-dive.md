---
title: "NAEOS MCP Server Deep Dive: Wire Your Specification into Every AI Tool"
description: "How the NAEOS MCP server exposes parse, validate, compile, and context tools over the Model Context Protocol — and how to use it from editors and agents."
date: 2026-08-19
author: "NAEOS Foundation"
categories: ["technical"]
---

The Model Context Protocol (MCP) is how AI tools discover and call external capabilities. NAEOS ships an MCP server that exposes its specification engine — parsing, validation, compilation, and artifact inspection — as nine tools any MCP client can call.

This is the deep dive: what the server does, how to run it, and how to call it directly with curl.

## Why MCP?

We documented the decision in [ADR-003](https://github.com/NAEOS-foundation/naeos/blob/main/docs/adr/003-why-mcp-for-ai-integration.md): rather than building a bespoke integration per vendor (Copilot, Claude, Cursor, Gemini, Codex, OpenCode), MCP gives us one vendor-neutral interface that every agent speaks. The protocol is self-describing — clients call `tools/list` and learn what's available — and it's transport-agnostic.

The other half of the strategy is the AI compiler: `naeos ai compile` writes per-agent instruction files (`CLAUDE.md`, `.cursorrules`, `.github/copilot-instructions.md`, ...). MCP is the *live* interface; instruction sets are the *persistent* context. Together they cover both modes of AI-assisted development.

## The Nine Tools

The server exposes these tools over JSON-RPC 2.0:

| Tool | What it does | Arguments |
|------|--------------|-----------|
| `parse_spec` | Parse a spec, return project/module/service summary | `spec` (required) |
| `validate_spec` | Validate the spec; PASS/FAIL + issues | `spec` (required) |
| `generate_context` | Generate an AI context bundle | `spec`, `format` (markdown/plain) |
| `compile_spec` | Compile to an instruction set for a target agent | `spec`, `target` (copilot, claude, cursor, gemini, codex, opencode) |
| `explain_concept` | Explain a NAEOS concept (pipeline, neir, kernel, policy, ...) | `concept` (required) |
| `list_artifacts` | List artifacts from the artifact store | — |
| `get_pipeline_status` | Status of a pipeline job by ID | `job_id` (required) |
| `export_terraform` | Generate Terraform HCL from a spec | `spec` (required) |
| `list_plugins` | List installed plugins and status | — |

## Running the Server

```bash
naeos mcp                     # serves on :3000
naeos mcp --port 8080         # or your own port
```

You'll see the startup banner with the health-check URL. The server implements the `initialize`, `tools/list`, and `tools/call` methods and reports protocol version `2024-11-05`.

The same MCP endpoint is also mounted inside the REST API server at `POST /api/v1/mcp/message`:

```bash
naeos api                     # REST API on :8080
```

## Calling It with curl

Health check:

```bash
curl http://localhost:3000/health
# {"status":"ok"}
```

List the tools:

```bash
curl -s -X POST http://localhost:3000/mcp \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","method":"tools/list","id":1}'
```

Validate a specification:

```bash
curl -s -X POST http://localhost:3000/mcp \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","method":"tools/call","id":2,
       "params":{"name":"validate_spec",
                 "arguments":{"spec":"project: shop\nmodules:\n  - name: core\n    path: ./core\n"}}}'
```

Parse and summarize:

```bash
curl -s -X POST http://localhost:3000/mcp \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","method":"tools/call","id":3,
       "params":{"name":"parse_spec",
                 "arguments":{"spec":"project: shop\nmodules:\n  - name: core\n    path: ./core\n"}}}'
# → "Project: shop · Modules: 1 · Services: 0"
```

Generate Terraform:

```bash
curl -s -X POST http://localhost:3000/mcp \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","method":"tools/call","id":4,
       "params":{"name":"export_terraform",
                 "arguments":{"spec":"project: shop\nservices:\n  - name: api\n    kind: http\n    port: 8080\n"}}}'
```

## Using It from the API Server

The REST API variant exposes the identical protocol at `/api/v1/mcp/message`:

```bash
curl -s -X POST http://localhost:8080/api/v1/mcp/message \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","method":"initialize","id":1}'
```

Inside the API server the MCP layer also gets access to the artifact store and plugin manager, so `list_artifacts` and `list_plugins` are fully functional there.

## What This Enables

The MCP server turns NAEOS into a *live* part of your AI workflow instead of a one-shot generator:

- **An agent can validate a spec before generating.** No more asking Claude to "write a service" and discovering the spec was wrong.
- **An agent can ask what a concept means.** `explain_concept` gives grounded, model-level answers about pipeline stages, NEIR, policy, and the kernel.
- **CI tooling can check artifact status** without a NAEOS binary — just HTTP calls.
- **Infra-as-code flows straight from the model.** `export_terraform` derives HCL from the same NEIR model as the code.

## Honest Trade-offs

The server currently implements the tools-only revision of MCP (`2024-11-05`) over HTTP — no resources, prompts, or streaming yet. If you need deeper MCP capabilities, the roadmap tracks richer protocol support alongside editor client configs.

The authoritative CLI reference lives at [`docs/cli/naeos_mcp.md`](https://github.com/NAEOS-foundation/naeos/blob/main/docs/cli/naeos_mcp.md), and the OpenAPI spec documents the API-server route. Try the [end-to-end tutorial](/blog/ecommerce-end-to-end-tutorial/) first, then wire your spec into your favorite agent with `naeos mcp`.