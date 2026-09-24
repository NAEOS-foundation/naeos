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

## Connect the website

For local development, set:

```bash
NEXT_PUBLIC_CONTROL_PLANE_API_URL=http://localhost:9091
```

Then start the website normally.

The `/control-plane` page sends:

```json
{
  "agent_id": "agent-demo",
  "capability": "database.delete",
  "artifact_hash": "sha256:demo-artifact"
}
```

and displays the returned decision ID, policy identity/version, reason, and execution status.

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

The public website should only be configured with a deployed API endpoint after that backend has its own authentication, authorization, rate limiting, CORS policy, persistence, and operational controls.

Do not expose the demo server directly to the public internet without those controls.
