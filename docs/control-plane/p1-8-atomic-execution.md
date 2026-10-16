# P1.8 — Atomic Execution Commit Boundary

## Purpose

P1.7 proves authorization freshness at the execution boundary. P1.8 closes the in-process check-to-side-effect race by making the policy freshness check and the side-effect callback share one execution lock.

## Contract

    T0: request -> policy v1 -> ALLOW
    T1: execution boundary -> active policy v1 -> commit lock
    T2: durable EXECUTION_STARTED intent is committed
    T3: side effect executes while policy update is excluded
    T4: final receipt is committed or the durable start remains unresolved -> callback returns -> policy update may commit

A concurrent policy update cannot become active between the freshness check and the side-effect callback when both operations use the same PolicyStore execution boundary.

## Evidence

The ledger records AUTHORIZATION_DECISION, then records EXECUTION_STARTED before invoking the side effect. When ledger persistence is configured, that start event is committed to the durable snapshot before the callback runs; a persistence failure prevents invocation. A successful execution adds EXECUTION_ALLOWED with the same execution ID. If the final receipt cannot be committed, the unmatched start remains durable and verification reports UNKNOWN until reconciliation. Failed or stale attempts that never enter execution remain fail-closed and record EXECUTION_BLOCKED.

## Acceptance criteria

1. T0 returns ALLOW under policy v1.
2. Atomic execution validates that v1 is still active.
3. The side-effect callback runs while policy updates are excluded.
4. A concurrent v2 update cannot commit before the side effect completes.
5. EXECUTION_STARTED is durably committed before the callback runs.
6. Successful execution records EXECUTION_ALLOWED with the same execution ID.
7. A missing final receipt leaves the session UNKNOWN, including after reload and later successful appends.
8. A stale authorization is rejected before the callback.
9. Regression tests cover the serialization and durable-evidence boundaries.

## Scope

P1.8 is an in-process control-plane primitive. It does not claim distributed transaction semantics, external-system atomicity, or consensus across multiple policy stores.
