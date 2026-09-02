---
title: "NAEOS MCP Server Grows: Resources, Prompts, Completions, and Pagination"
description: "The NAEOS MCP server now speaks the full 2025 MCP model layer — resources, prompts, completions, liveness ping, and cursor pagination — so agents can discover specs, docs, and artifacts on demand."
date: 2026-08-30
author: "NAEOS Foundation"
categories: ["technical", "release"]
---

When we first shipped the [MCP server deep dive](/blog/mcp-server-deep-dive/), the server implemented the tools-only revision of the Model Context Protocol (`2024-11-05`) — nine tools and nothing else. No resources, no prompts, no streaming. That last line, "no resources, prompts, or streaming yet," was written as an honest caveat. Today we're closing that gap.

The upcoming NAEOS release turns the MCP server into the full model-layer experience: the existing tools are joined by **resources**, **prompts**, **argument completions**, a **liveness ping**, and **cursor-based pagination** across every list operation.

## Verified against the checklist

In the previous post we wrote that the server "currently implements the tools-only revision... no resources, prompts, or streaming yet." Here's what changed:

- **Resources** — read NAEOS concept docs, artifact-store contents, and pipeline job status as MCP resources.
- **Prompts** — four builtin templates agents can fetch and instantiate.
- **Completions** — serve argument autocompletions to a client mid-conversation.
- **Ping** — a lightweight liveness method for connection health checks.
- **Pagination** — cursor-based listing across every `*_list` method.

## What are MCP resources?

The Model Context Protocol models more than tool calls. A *resource* is a piece of content a client can read on demand — think of it as a live file system the agent can open rather than a function it invokes. NAEOS exposes three namespaces:

| Resource namespace | What it exposes |
|--------------------|-----------------|
| `naeos://docs/{concept}` | NAEOS concept documentation (pipeline, NEIR, kernel, policy, ...) |
| `naeos://artifacts/{path}` | Contents of the artifact store by path |
| `naeos://jobs/{id}` | Status of a pipeline job by ID |

So when your agent wants to know what "kernel" means in the NAEOS model, or check whether a build artifact landed, it can *read the resource* instead of guessing.

```bash
# resources/list with pagination
curl -s -X POST http://localhost:3000/mcp \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","method":"resources/list","id":1,
       "params":{"cursor":"<opaque-cursor>"}}'

# read one resource
curl -s -X POST http://localhost:3000/mcp \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","method":"resources/read","id":2,
       "params":{"uri":"naeos://docs/neir"}}'
```

## Builtin prompt templates

Resources answer "what do you know?"; prompts answer "what should I ask you to do?". NAEOS ships four templates the client can list and instantiate, substituting arguments verbatim into the prompt message:

- `review-spec` — hand a spec to the agent for review.
- `enrich-spec` — suggest improvements to a spec.
- `explain-architecture` — explain a chosen architecture pattern.
- `generate-spec` — scaffold a spec from a loose description.

```bash
curl -s -X POST http://localhost:3000/mcp \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","method":"prompts/list","id":1}'

curl -s -X POST http://localhost:3000/mcp \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","method":"prompts/get","id":2,
       "params":{"name":"explain-architecture",
                 "arguments":{"architecture":"microservices"}}}'
```

## Completions: autocomplete inside the conversation

Many agents let a tool fill in arguments as you type. MCP models this with `completion/complete`. NAEOS wires two completion sources:

- **`ref/prompt`** — completing the `architecture` argument of `explain-architecture` suggests from all ten NEIR architecture patterns.
- **`ref/resource`** — completing a resource URI filters against the live resource list (concept docs, artifact paths, pipeline job IDs), with case-insensitive prefix matching and the spec's 100-value cap.

The `initialize` handshake now advertises the `completions` capability, so clients know to offer autocomplete.

## Healthy by design: ping

Connection health checks need a cheap, side-effect-free round trip. `ping` provides it:

```bash
curl -s -X POST http://localhost:3000/mcp \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","method":"ping","id":3}'
```

## Cursor pagination everywhere

Every list method — `tools/list`, `resources/list`, `prompts/list` — now paginates with a cursor instead of drowning a client in one giant response. Responses are 50 items per page; when more remain, the server returns an opaque base64 cursor:

```json
{ "resources": [ ...50 items... ], "nextCursor": "bmV4dDoxMDA=" }
```

Pass that value back as the `cursor` parameter to fetch the next page. Malformed cursors return a `-32602` (Invalid Params) error.

## What this unlocks for your agent workflow

The NAEOS MCP server is no longer just a tool on a shelf. It's a **conversational model layer**:

- An agent can *read* the artifact store to check a build result, not just call a function that summarizes it.
- A richer `initialize` announces resources, prompts, and completions — so the client configures the right UX up front.
- Large registries (plugins, templates, schemas) can be walked page by page without blowing up memory or token context.

The one thing we still haven't shipped is streaming/tool-call progress events (SSE). It's tracked on the roadmap; everything else in the model layer is now live.

## Try it

Grab the latest build, run the server, and poke at it with curl:

```bash
naeos mcp --port 3000
```

Then walk the resources:

```bash
curl -s -X POST http://localhost:3000/mcp \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","method":"resources/list","id":1}'
```

Pair it with the [AI-driven development guide](/blog/ai-driven-development/) and the [end-to-end tutorial](/blog/ecommerce-end-to-end-tutorial/) to see the whole picture. The authoritative CLI reference lives at [`docs/cli/naeos_mcp.md`](https://github.com/NAEOS-foundation/naeos/blob/main/docs/cli/naeos_mcp.md).
