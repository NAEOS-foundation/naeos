# Evidence Experiment 001 — Policy Change Mid-Run

**Status:** Reproducible repository evidence  
**Scope:** P1.7 control-plane policy freshness  
**Purpose:** Demonstrate that an authorization issued under policy v1 is not reusable after policy v2 becomes active.

## Question

What happens when an already-authorized engineering action reaches the execution boundary after its bound policy version is no longer active?

## Existing implementation

The repository already contains a focused executable example:

`examples/control-plane-policy-change/`

Run:

```bash
./examples/control-plane-policy-change/run-demo.sh
```

or:

```bash
go run ./examples/control-plane-policy-change
```

The example requires no network, credentials, LLM API key, or external service.

## Experiment

1. Policy v1 is active.
2. An authorization request for `repository.write` receives `ALLOW`.
3. Policy v2 becomes active before execution.
4. The v1 authorization is presented at the execution boundary.
5. The gateway re-checks authorization freshness.
6. The stale authorization is denied with `stale_policy`.
7. Execution evidence records `EXECUTION_BLOCKED`.
8. The example verifies that the side effect did not occur.
9. The session evidence is independently verified.

## Expected evidence

The executable example asserts:

- initial decision: `ALLOW`
- execution decision: `DENY`
- denial reason: `stale_policy`
- execution event: `EXECUTION_BLOCKED`
- no side-effect file
- original authorization evidence retained
- blocked execution evidence retained
- session verification: `PASS`

The command exits non-zero if an assertion fails.

## Engineering invariant

> An authorization decision is valid only while its bound policy version remains the active policy at the execution boundary.

This is an authorization-freshness property, not a claim about a specific AI model or agent.

## Source evidence

The behavior is covered by:

- `examples/control-plane-policy-change/README.md`
- `examples/control-plane-policy-change/main.go`
- `internal/controlplane/gateway_test.go`

The control-plane tests additionally cover rejection of stale authorization before a side-effect callback is invoked and serialization of policy changes with atomic execution.

## Reproduction note

This document records repository implementation evidence. A publication should report runtime output only after the command is actually executed in the publication environment.

## What this does not prove

This experiment does not by itself prove production readiness, security of every integration, customer adoption, compliance, or correctness of arbitrary AI-agent workflows.

It demonstrates one concrete control-plane property: policy-version freshness at the execution boundary.
