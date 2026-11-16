# Control Plane Persistence Contracts — Implementation Design

Status: proposed implementation contract for #796
Parent: #786
Related foundation: #793

## Runtime boundary

The production Control Plane must not depend on concrete in-memory stores. Persistence calls are context-aware and return explicit errors. An error is never converted into an empty query result, a synthetic allow/deny, or an unpersisted success. The investor demo remains composed separately with memory adapters.

## Ledger repository

The production ledger contract must provide append event returning canonical committed event and sequence or error; load/verify the complete ordered event stream; retrieve canonical authorization decision by decision ID; detect whether execution intent/outcome already exists for a decision; and append execution lifecycle and reconciliation evidence under repository transaction semantics.

Append must serialize sequence/head updates, verify predecessor hash, hash canonical event payload, commit before returning, and preserve failure as an error. Reconciliation needs its own authorized transactional workflow and cannot use generic append.

## Policy repository

Policy persistence retains immutable versions and an active-version pointer. Required operations: put new version with monotonic version validation; fetch active policy and specified immutable version; atomically validate active version while holding a database lock/transaction across the execution boundary.

An in-process mutex is insufficient across multiple Railway replicas. Do not hold a DB transaction across an arbitrary external side effect without an explicit execution protocol; use durable execution intent, idempotency, and reconciliation.

## Grant repository

Grant operations include create/read, revoke, active-state validation, and binding to agent, policy ID/version, capability scope, and expiry. Revocation is durable and checked at execution boundary.

## Approval repository

Approval operations include create, approve, reject, get, expire, and consume. Consume is a compare-and-set/transactional one-time transition from approved to consumed. Approval is bound to decision ID, agent, capability, policy ID/version, artifact hash, and expiry.

## Context and error semantics

- Every repository call takes context.Context for cancellation and bounded request deadlines.
- Missing record is distinguishable from database unavailable, integrity failure, and conflict.
- A conflict/revocation/stale policy returns policy denial only when authoritative state was successfully read.
- Infrastructure/integrity failures at authorization boundary return INDETERMINATE and prevent side effects.
- Gateway must not expose ALLOW until canonical decision event commits.
- Execution is idempotent by decision/execution ID and must not rely on local process state.

## Verification plan

Use ephemeral PostgreSQL CI only: memory and PostgreSQL repository contract suites; restart recovery for policy versions, grants, approvals, ledger; concurrent ledger append and chain continuity; concurrent approval consumption (exactly one winner); concurrent execution idempotency (at most one durable execution intent); policy activation/revocation race against execution; tampering, sequence gaps, DB unavailable, transaction rollback; runtime-role privilege assertions for private schema.

## Rollout order

1. Define contracts and error taxonomy; add memory adapters and contract tests.
2. Complete PostgreSQL adapters and isolated integration tests.
3. Refactor gateway to propagate context and fail closed on repository errors.
4. Add production composition root with required DATABASE_URL, TLS enforcement, bounded pool, secret-safe logs.
5. Validate deployment configuration in CI; only then separately plan production cutover. No production DB migrations, credentials, or Railway settings are changed by these implementation PRs.
