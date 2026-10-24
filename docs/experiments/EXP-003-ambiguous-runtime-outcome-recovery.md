# Experiment 003 — Ambiguous Runtime Outcome Recovery

**Status:** Regression coverage added; CI validation pending  
**Scope:** Runtime gateway execution outcomes and durable invocation replay protection  
**Motivation:** Microsoft Agent Framework issue [#9033](https://github.com/microsoft/agent-framework/issues/9033) highlights a crash window in which a tool may have run but its result was not durably recorded.

## Failure model

An external side effect can commit even when the caller receives a timeout or transport error. The runtime therefore cannot infer “not executed” from an execution error alone.

The gateway now uses the following contract:

- A sandbox error after execution begins produces `unknown`, not `failed`.
- An observer error, identity mismatch, or non-confirming observation produces `unknown`; absence of authoritative evidence is not proof of absence.
- Fail-closed mode still returns an error to the caller. It does not relabel the outcome as a confirmed failure.
- With replay protection enabled, the invocation ID is durably claimed before the side effect. Reuse of that ID is denied, including by a newly constructed gateway after restart.
- An unknown outcome must be resolved through independent observation/reconciliation. It must not trigger an automatic retry with a new invocation identity.

## Regression scenario

The test `TestAmbiguousExecutionIsUnknownAndCannotBeReplayedAfterRestart` simulates:

1. A sandbox performs the modeled external effect.
2. The sandbox loses its response and returns an error.
3. The first gateway reports `unknown`.
4. A new gateway instance receives the same invocation ID.
5. The durable replay store denies the replay.
6. The modeled external effect count remains exactly one.

Observer regression tests also verify that missing, mismatched, or unavailable observation cannot be reported as success.

## Reproduce

Run the focused regression tests:

```bash
go test ./internal/runtime/gateway -run 'TestGatewaySandboxFailure|TestGatewayFailOpen|TestGatewayObserverFailuresCannotProduceSuccess|TestAmbiguousExecutionIsUnknownAndCannotBeReplayedAfterRestart' -count=1
```

Run the full package tests before merging:

```bash
go test ./internal/runtime/gateway
```

## Acceptance criteria

1. Ambiguous execution and observation failures remain `unknown`.
2. No unverified outcome is promoted to `completed`.
3. A durable invocation ID cannot be claimed twice across gateway instances.
4. A crash/response-loss scenario does not cause the same invocation to reach the sandbox twice.
5. Recovery requires independent evidence or explicit reconciliation.

## Scope boundary

This experiment proves conservative outcome classification and replay suppression at the NAEOS gateway boundary. It does **not** prove distributed exactly-once execution or make arbitrary external systems transactional. A provider-side idempotency key, authoritative external observation, and/or an authenticated reconciliation workflow is still required to determine whether an ambiguous effect committed.
