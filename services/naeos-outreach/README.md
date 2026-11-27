# NAEOS Outreach Server — Phase 1A

**Status:** Foundation skeleton; repository placement is provisional.  
**Safety:** No email sending, provider integration, prospect ingestion, or production configuration.

NAEOS Outreach is a separate service boundary from NAEOS core and CRM business logic. This initial skeleton is placed in the canonical NAEOS repository only because automated repository creation is not available through the connected GitHub tooling. Move it to a standalone NAEOS Outreach repository once repository creation/administration is available and the boundary is ratified.

## Run

```bash
cd services/naeos-outreach
go test ./...
go run ./cmd/naeos-outreach-server
```

Server listens on `127.0.0.1:8088` by default. Override with `NAEOS_OUTREACH_ADDR`. Binding to loopback is the default to reduce accidental exposure.

## Endpoints

- `GET /health`: process health.
- `GET /ready`: readiness (currently ready if process started; future dependencies must be checked).

Only these endpoints are implemented in Phase 1A. API contract files under `api/outreach/v1` are design artifacts and not yet fully implemented.

## Explicitly disabled

- Live sending or sending provider adapters.
- Autonomous outreach.
- LinkedIn automation/scraping.
- Production credentials or deployment.
- Persistent data stores.

Next: configuration hardening, authentication/RBAC, migration and repository layer, prospect CRUD with provenance, qualification, campaign draft/approval and dry-run.
