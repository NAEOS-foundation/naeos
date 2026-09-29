# P1.8 — Atomic Execution Commit Boundary

## Purpose

P1.7 proves authorization freshness at the execution boundary. P1.8 closes the in-process check-to-side-effect race by making the policy freshness check and the side-effect callback share one execution lock.

## Contract

    T0: request -> policy v1 -> ALLOW
    T1: execution boundary -> active policy v1 -> commit lock
    T2: side effect executes while policy update is excluded
    T3: execution evidence -> policy update may commit

A concurrent policy update cannot become active between the freshness check and the side-effect callback when both operations use the same PolicyStore execution boundary.

## Evidence

The ledger records AUTHORIZATION_DECISION followed by EXECUTION_ALLOWED for a successful atomic execution. Failed or stale attempts remain fail-closed and record EXECUTION_BLOCKED.

## Acceptance criteria

1. T0 returns ALLOW under policy v1.
2. Atomic execution validates that v1 is still active.
3. The side-effect callback runs while policy updates are excluded.
4. A concurrent v2 update cannot commit before the side effect completes.
5. Successful execution records EXECUTION_ALLOWED.
6. A stale authorization is rejected before the callback.
7. Regression tests cover the serialization boundary.

## Scope

P1.8 is an in-process control-plane primitive. It does not claim distributed transaction semantics, external-system atomicity, or consensus across multiple policy stores.
