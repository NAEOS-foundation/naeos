# Control Plane E2E Proof

## Purpose

This workflow proves the browser-facing control-plane contract against the real NAEOS engine without executing an external side effect.

The test starts the existing `naeos demo` server and verifies the complete decision/evidence path:

```text
HTTP request
   ↓
NAEOS demo API
   ↓
Control-plane gateway
   ↓
Policy / grant evaluation
   ↓
ALLOW or DENY
   ↓
Control-plane ledger
   ↓
Evidence endpoint
```

## Assertions

The E2E workflow proves three properties:

1. **ALLOW** — `repository.read` for `agent-payment-01` returns `ALLOW`.
2. **DENY** — `credential.rotate` for `agent-payment-01` returns `DENY`.
3. **Evidence** — both decisions are present through `/api/control-plane/evidence`.

Each decision must also return a decision ID and evidence endpoint.

## Important boundary

This is an engine-backed E2E proof, not a public production deployment.

The public `naeos.dev/control-plane` page currently has no production control-plane API configured. A public deployment still requires an independently operated backend with authentication, authorization, rate limiting, CORS restrictions, persistence, health monitoring, and operational ownership.

The workflow deliberately does not use a temporary tunnel or expose the CI runner publicly. This keeps the proof deterministic and avoids presenting an ephemeral test server as a production service.

## Next boundary

The remaining step for public E2E is infrastructure, not policy logic:

```text
naeos.dev/control-plane
        ↓
authenticated public API
        ↓
NAEOS Control Plane
        ↓
decision
        ↓
verifiable evidence receipt
```

The public endpoint should be configured only after the backend has a stable deployment identity and the operational controls above are in place.
