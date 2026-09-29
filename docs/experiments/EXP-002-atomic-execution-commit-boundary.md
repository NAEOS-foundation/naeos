# Evidence Experiment 002 — Atomic Execution Commit Boundary

**Status:** Reproducible repository evidence  
**Scope:** P1.8 control-plane atomic execution  
**Purpose:** Demonstrate that policy freshness validation and the side-effect callback share one in-process execution lock.

## Question

What happens when a policy update races with an already-authorized execution at the check-to-side-effect boundary?

## Existing implementation

The repository contains a focused executable example:

`examples/control-plane-atomic-execution/`

Run:

```bash
./examples/control-plane-atomic-execution/run-demo.sh
```

or:

```bash
go run ./examples/control-plane-atomic-execution
```

The example requires no network, credentials, LLM API key, or external service.

## Experiment

1. Policy v1 is active.
2. An authorization request for `repository.write` receives `ALLOW`.
3. `ExecuteAtomic` enters the policy-store execution boundary.
4. The side-effect callback creates a disposable local artifact while the execution lock is held.
5. A concurrent policy update to v2 is attempted during the callback.
6. The update cannot commit until the atomic side effect returns.
7. Successful execution records `EXECUTION_ALLOWED`.
8. Policy v2 becomes active only after the atomic execution completes.
9. Session verification independently returns `PASS`.

## Expected evidence

The executable example asserts:

- initial decision: `ALLOW`
- execution decision: `ALLOW`
- side effect observed
- execution evidence: `EXECUTION_ALLOWED`
- authorization policy version: v1
- current policy version after execution: v2
- session verification: `PASS`

The regression test additionally asserts that a concurrent policy update cannot complete while the side-effect callback is still blocked.

## Engineering invariant

> Policy freshness validation and the authorized side effect must share the same in-process execution boundary when atomic execution is claimed.

This closes the in-process check-to-side-effect race represented by P1.8.

## Source evidence

The behavior is covered by:

- `docs/control-plane/p1-8-atomic-execution.md`
- `examples/control-plane-atomic-execution/main.go`
- `internal/controlplane/gateway.go`
- `internal/controlplane/gateway_test.go`

The regression test specifically covers serialization of the policy update with the atomic side-effect callback.

## Reproduction note

This document records repository implementation evidence. Publication should report runtime output only after the command is actually executed in the publication environment.

## Scope boundary

P1.8 is an in-process control-plane primitive. This experiment does **not** prove distributed transaction semantics, external-system atomicity, or consensus across multiple policy stores.

It demonstrates one concrete property: an active policy update cannot commit between the authorization freshness check and the completion of the authorized side-effect callback when both use the same execution lock.
