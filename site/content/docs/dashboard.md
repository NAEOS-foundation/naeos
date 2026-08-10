---
title: Dashboard
description: Monitor pipeline activity, stats, and component health in real time.
weight: 14
---

The NAEOS web dashboard gives you a live view of pipeline activity, system
stats, and component health, with real-time updates over WebSocket.

## Starting the Dashboard

```bash
naeos dashboard
```

Serving on `http://localhost:3000` by default. Use `--port` to change it:

```bash
naeos dashboard --port 8080
```

## Features

- **Live activity log** — pipeline events and log messages stream in real
  time over WebSocket (`/ws`)
- **Stats** — pipeline statistics broadcast every 5 seconds
- **Component health** — API server, parser, compiler, and MCP server status
- **API endpoints** — `GET /api/stats`, `GET /api/activity`,
  `GET /api/health`

## Dashboard Components

| Component | Status Source |
|-----------|---------------|
| API Server | `api.NewServer` with auth disabled |
| Parser | Ready on startup |
| Compiler | Ready on startup |
| MCP Server | Stopped (degraded) unless enabled |

## Using the Dashboard with the API Server

The dashboard reuses the NAEOS API server, so both the dashboard UI and the
REST API run on the same port. WebSocket upgrades happen at `/ws`, and the
dashboard itself is served at `/`.

## Related

- [CLI Reference](/docs/cli-reference/) — all CLI commands including `dashboard`
- [Distributed Builds](/docs/distributed-builds/) — parallel pipeline execution
