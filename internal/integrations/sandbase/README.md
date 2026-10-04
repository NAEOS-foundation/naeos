# SandBase Authorization Adapter

This package defines the smallest protocol-neutral boundary between SandBase Harness and the NAEOS control plane.

SandBase Harness remains responsible for session lifecycle, sandbox execution, MCP/tool invocation, approvals, and audit/replay. NAEOS evaluates the external authorization request and returns a compact decision envelope. SandBase integration can consume this envelope without importing NAEOS implementation types.

## Contract

Request:

```json
{
  "session_id": "session-1",
  "request_id": "req-1",
  "agent_id": "agent-1",
  "capability": "repository.write",
  "target": "repo.write",
  "artifact_hash": "sha256:...",
  "context": {},
  "timestamp": "2026-10-05T09:00:00Z"
}
```

Decision:

```json
{
  "decision": "ALLOW",
  "policy_version": 1,
  "capability": "repository.write",
  "target": "repo.write",
  "expires_at": "2026-10-05T10:00:00Z",
  "context_digest": "sha256:...",
  "request_id": "req-1",
  "decision_id": "DEC-...",
  "reason": "allowed_by_policy"
}
```

## Freshness boundary

The adapter issues authorization; it does not make an authorization durable across policy changes. NAEOS `DecisionGateway.ExecuteDecision` / `ExecuteAtomic` revalidates the decision at the execution boundary and blocks stale policy versions.

This preserves the intended separation:

`SandBase execution → NAEOS policy/authorization → SandBase execution`

The adapter is a prototype contract, not a claim that SandBase Harness currently has a native external-authorization hook.

## Verification

`adapter_test.go` verifies:

1. a fresh v1 request produces an ALLOW decision with policy version and context digest;
2. after policy v2 becomes active, reuse of the v1 authorization is blocked as `stale_policy`.
