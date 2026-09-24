# NAEOS Control Plane Live Proof

The Control Plane page can evaluate a request against the real NAEOS control-plane engine through the existing investor-demo API.

## Run locally

Start the NAEOS demo server:

```bash
go run ./cmd/naeos demo --addr :9091
```

The API exposes:

```
POST http://localhost:9091/api/control-plane/decision
```

The endpoint performs an authorization evaluation through the reusable `internal/controlplane` gateway. It records the authorization decision in the control-plane ledger and does not execute the requested side effect.

The hardened boundary applies:
- explicit CORS origin allowlisting
- optional Bearer authentication
- request-size and field-length limits
- strict JSON decoding for the decision request
- per-client rate limiting (30 requests/minute by default)
- policy selection from the agent's active grant rather than caller-supplied policy identity

## Connect the website

For local development, set:

```bash
NAEOS_CONTROLPLANE_ALLOWED_ORIGINS=http://localhost:3000
NEXT_PUBLIC_CONTROL_PLANE_API_URL=http://localhost:9091
```

For a deployed website, set `NAEOS_CONTROLPLANE_ALLOWED_ORIGINS` to the exact browser origins that may call the demo API. The server no longer uses wildcard CORS. Optional `NAEOS_CONTROLPLANE_API_TOKEN` enables Bearer authentication for server-to-server use; do not put that token in browser code.

Then start the website normally.

The `/control-plane` page sends:

```json
{
  "agent_id": "agent-demo",
  "capability": "database.delete",
  "artifact_hash": "sha256:demo-artifact"
}
```

and displays the returned decision ID, policy identity/version, reason, execution status, and an evidence endpoint for the recorded decision.

## Trust boundary

This proof deliberately stops at authorization. It does **not** claim that a browser click executed a production side effect.

The evidence chain demonstrated here is:

```
Request
  ↓
NAEOS Control Plane
  ↓
Policy / Grant Evaluation
  ↓
ALLOW / DENY / REQUIRE_APPROVAL
  ↓
Control-Plane Ledger
```

Execution and external observation remain separate boundaries.

## Production deployment

A public demo deployment must also provide:
- HTTPS/TLS at the edge
- authentication when the endpoint is not intentionally public
- strict CORS allowlisting
- rate limiting and request-size limits
- durable ledger persistence
- health/readiness monitoring
- logging and alerting
- deployment-level network controls

The demo decision endpoint is intentionally side-effect-free. It is suitable for a public proof only after those controls are configured and the ledger persistence is operational.

Do not expose the broader demo API surface (`/api/execute`, `/api/reset`, scenario runners, approvals) as an unauthenticated public service.
