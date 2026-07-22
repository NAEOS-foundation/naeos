---
title: API Reference
description: Interactive API documentation for the NAEOS REST API.
weight: 17
---

The NAEOS REST API allows you to interact with the NAEOS runtime programmatically. This page provides an interactive reference generated from the OpenAPI 3.0 specification.

## Overview

| Info | Value |
|------|-------|
| **Base URL** | `http://localhost:8080` |
| **API Version** | 1.5.0 |
| **Format** | JSON |
| **Authentication** | None (local development) |

## Quick Reference

### System

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/health` | Health check |
| GET | `/api/v1/version` | Version info |
| GET | `/api/v1/config/schema` | Config schema |
| GET | `/healthz` | Liveness probe |
| GET | `/readyz` | Readiness probe |
| GET | `/metrics` | Prometheus metrics |

### Specifications

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/specs` | List all specifications |
| POST | `/api/v1/specs` | Create a new specification |
| POST | `/api/v1/specs/validate` | Validate a specification |
| POST | `/api/v1/specs/compile` | Compile a specification |

### Pipeline

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/pipeline/run` | Run the pipeline |
| GET | `/api/v1/pipeline/status` | Get pipeline status |
| GET | `/api/v1/pipelines` | List all pipelines |

### AI (SSE Streaming)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/ai/enrich/stream` | AI enrichment (streaming) |
| POST | `/api/v1/ai/explain/stream` | AI explanation (streaming) |
| POST | `/api/v1/ai/compile/stream` | AI compilation (streaming) |

### Other

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/artifacts` | List artifacts |
| POST | `/api/v1/context/generate` | Generate context bundle |
| GET/POST | `/api/v1/plugins` | Plugin management |
| POST | `/api/v1/mcp/message` | MCP message |
| POST | `/api/v1/cloud/plan` | Cloud deployment plan |
| POST | `/api/v1/cloud/deploy` | Deploy to cloud |
| POST | `/api/v1/cloud/destroy` | Destroy cloud resources |
| GET | `/api/v1/cloud/status` | Cloud status |

## Interactive Documentation

The full interactive API documentation is available below. You can test API calls directly from the browser.

{{< swagger-ui >}}

## CLI Usage

You can also interact with the API via the CLI:

```bash
# Start the API server
naeos serve --port 8080

# Validate via API
curl -X POST http://localhost:8080/api/v1/specs/validate \
  -H "Content-Type: application/json" \
  -d '{"spec": "project: my-app"}'

# Run pipeline via API
curl -X POST http://localhost:8080/api/v1/pipeline/run \
  -H "Content-Type: application/json" \
  -d '{"spec_path": "spec.yaml"}'
```

## OpenAPI Specification

Download the full OpenAPI 3.0 specification:

- [openapi.yaml](/openapi.yaml)

See also: [CLI Reference](/docs/cli-reference/), [Pipeline Engine](/docs/pipeline-engine/)
