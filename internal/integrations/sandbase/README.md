# SandBase Authorization Adapter

This package provides the NAEOS side of the sandbase.authz/v1 external authorization hook merged in SandBase Harness PR #947.

SandBase remains responsible for session lifecycle, local permissions/approvals, sandbox execution, MCP/tool invocation, and its audit/replay trail. NAEOS evaluates the external authorization request using its configured policy and grant. The HTTP endpoint is veto-only: only an explicit NAEOS ALLOW becomes allow; every other decision refuses execution.

## Production daemon wiring

The `naeos serve` daemon registers this endpoint at
`POST /api/v1/integrations/sandbase/authorize` when
`NAEOS_SANDBASE_AUTHZ_CONFIG` is set. It requires both
`NAEOS_SANDBASE_AUTHZ_TOKEN` (the shared secret; configure the same value as
SandBase's `MANAGED_AGENTS_EXTERNAL_AUTHZ_TOKEN`) and a valid config file.
Startup fails closed if the config is supplied without a token or cannot be
validated. Keep the route on a private HTTPS listener/network boundary.

The config file is JSON and contains the active policy and the grant assigned
to the trusted NAEOS agent:

```json
{
  "policy": {
    "id": "sandbase-policy",
    "version": 1,
    "status": "active",
    "allowed_capabilities": ["tool.execute"],
    "denied_capabilities": [],
    "protected_capabilities": [],
    "approval_required": [],
    "requires_explicit_auth": true
  },
  "grant": {
    "grant_id": "grant-sandbase",
    "agent_id": "trusted-agent",
    "policy_id": "sandbase-policy",
    "policy_version": 1,
    "capabilities": ["tool.execute"],
    "status": "active",
    "revoked": false,
    "expires_at": "2099-01-01T00:00:00Z"
  }
}
```

Use real timestamps and a least-privilege capability set in deployment. Do not
commit the bearer token. The optional `NAEOS_SANDBASE_AUTHZ_LEDGER` selects the
append-only authorization ledger snapshot path; by default it is derived from
the config path. Protect both files and their containing directory.

The handler rereads the config for every authorization request. Publish policy
changes as a **new monotonically increasing version**; changing policy content
without incrementing its version or rolling a version backward is rejected
fail-closed. A new policy version with a grant still bound to the old version
returns `reauthorize` until the grant is updated to match. This makes a
policy rollover visible to a still-running SandBase session rather than
silently retaining the previous ALLOW.

## HTTP contract

The daemon wiring above registers the configured handler. Embedders may instead register NewHTTPHandler(adapter, bearerToken) directly on a private HTTPS route. Configure the same secret in SandBase's MANAGED_AGENTS_EXTERNAL_AUTHZ_TOKEN environment variable. The NAEOS handler rejects an empty configured token and requires the matching Authorization: Bearer <token> header. Do not expose the endpoint publicly without TLS and network-level restrictions.

Request (POST, JSON):

```json
{
  "schema": "sandbase.authz/v1",
  "session_id": "session-42",
  "invocation_id": "call-17",
  "capability": "tool.execute",
  "target": "repository.write",
  "arguments_digest": "64 lowercase or uppercase hexadecimal SHA-256 characters",
  "policy_context_digest": "64 lowercase or uppercase hexadecimal SHA-256 characters",
  "digest_schema": "sandbase.digest/v1"
}
```

The handler validates both schema versions and digests, limits the request body, rejects unknown fields and extra JSON values, and obtains the NAEOS agent identity from the configured grant rather than trusting the request. It binds invocation_id as the request ID, binds both digests into the decision artifact hash, and includes both digests in the evaluated action context.

Response:

```json
{
  "decision": "allow",
  "reason": "allowed_by_policy",
  "policy_version": "1",
  "decision_id": "DEC-...",
  "context_digest": "..."
}
```

The decision is lowercase to match SandBase's hook. NAEOS ALLOW maps to allow; policy denial maps to deny; stale policy/grant version maps to reauthorize. Malformed requests and unavailable authorization return non-2xx responses, which SandBase treats as refusal. The configured NAEOS policy and grant must explicitly permit the incoming capability (SandBase currently sends tool.execute).

## Freshness boundary

The HTTP handler answers the synchronous pre-tool authorization check. It does not itself claim that a prior decision authorizes a later execution. NAEOS DecisionGateway.ExecuteDecision / ExecuteAtomic revalidates the canonical decision at the execution boundary and blocks stale policy versions. The SandBase hook remains veto-only and does not widen a local denial.

## Verification

- adapter_test.go covers the protocol-neutral adapter and stale-policy execution block.
- http_handler_test.go covers the SandBase wire envelope, lowercase decision mapping, fail-closed denial, schema/digest validation, bearer authentication, and method restrictions.
- configured_handler_test.go covers active policy reload, stale-grant reauthorization, policy-version immutability, and persisted ledger snapshots.

This is an implementation of the contract against the merged SandBase hook, not evidence that a deployed endpoint or end-to-end production pilot has run.