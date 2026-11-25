# NAEOS Outreach Server — Phase 1B

**Status:** Foundation skeleton with authorization boundary and configuration hardening.  
**Safety:** No email sending, provider integration, prospect ingestion, or production configuration.

NAEOS Outreach is a separate service boundary from NAEOS core and CRM business logic. This initial skeleton is placed in the canonical NAEOS repository only because automated repository creation is not available through the connected GitHub tooling. Move it to a standalone NAEOS Outreach repository once repository creation/administration is available and the boundary is ratified.

## Run

```bash
cd services/naeos-outreach
go test ./...
go run ./cmd/naeos-outreach-server
```

Server listens on `127.0.0.1:8088` by default. Override with `NAEOS_OUTREACH_ADDR`. Binding to loopback is enforced by configuration to reduce accidental exposure.

## Build (Docker)

```bash
cd services/naeos-outreach
docker build -t naeos-outreach:local .
docker run --rm -p 8088:8088 naeos-outreach:local
```

## Endpoints

- `GET /health`: process health (public, no authentication required).
- `GET /ready`: readiness (public, no authentication required).

All other routes require a trusted principal supplied by a NAEOS identity adapter (`401 Unauthorized` otherwise). Only these endpoints are implemented in Phase 1B. API contract files under `api/outreach/v1` are design artifacts and not yet fully implemented.

## Phase 1B changes

- Fail-closed principal boundary (`internal/authboundary`): identity must be supplied by a trusted upstream NAEOS identity adapter; this package intentionally does not parse credentials.
- Strict configuration (`internal/config`): validates `NAEOS_OUTREACH_ADDR`, enforces loopback-only host binding, rejects malformed addresses.
- Graceful shutdown on `SIGINT`/`SIGTERM` with 10-second shutdown timeout.
- Security headers (`X-Content-Type-Options: nosniff`, `Cache-Control: no-store`).

## Explicitly disabled

- Live sending or sending provider adapters.
- Autonomous outreach.
- LinkedIn automation/scraping.
- Production credentials or deployment.
- Persistent data stores.
- JWT identity parsing.

Next: authentication/RBAC, migration and repository layer, prospect CRUD with provenance, qualification, campaign draft/approval and dry-run.
