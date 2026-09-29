# P1.7 — Policy Change Mid-Run / Stale Authorization

## Purpose

An authorization decision issued at T0 must not be executable at T1 when the policy version that produced it is no longer the active policy.

## Contract

request -> policy v1 -> ALLOW / authorization -> policy changes -> v2 active -> execution-boundary re-check -> STALE / BLOCK -> no side effect -> evidence -> verification

The agent is not trusted to remember that policy changed. The control plane establishes freshness from its policy store and the append-only evidence ledger.

## Evidence

The execution attempt remains correlated to the original request_id and decision_id. The ledger records AUTHORIZATION_DECISION, EXECUTION_BLOCKED, stale_policy, policy identity, and the authorization policy version.

## Acceptance criteria

1. T0 returns ALLOW under policy v1.
2. Policy v2 becomes active before execution.
3. The v1 authorization is rejected at execution.
4. The rejection reason is stale_policy.
5. No side effect occurs.
6. Evidence remains queryable by the original request ID.
7. Independent verification returns PASS.

## Scope

P1.7 does not attempt production authorization, distributed consensus, external receipts, or multi-agent handoff.