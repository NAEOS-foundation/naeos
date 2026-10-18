# P1.7 Policy Change Mid-Run: Explicit Capability Revocation

This example proves two distinct controls when an agent's governing policy changes mid-task:

1. An authorization issued under policy v1 is rejected at the execution boundary after v2 becomes active.
2. A fresh authorization attempt under v2 is denied because v2 explicitly denies the previously allowed `repository.write` capability—even after the demo grant is rebound to v2.

Flow:

```text
policy v1 → ALLOW
policy v2 (repository.write = DENY)
old v1 authorization → EXECUTION_BLOCKED (stale_policy)
new v2 authorization → DENY (denied_by_policy)
execution callbacks invoked: 0
evidence → independent session verification
```

The example supplies a real side-effect callback that writes `side-effect.json`. The callback must never run in either denial path. The output `result.json` records the decisions, block evidence, side-effect observation, and verification result.

## Run

```bash
./examples/control-plane-policy-change/run-demo.sh
```

Or run the focused test and example directly:

```bash
go test ./internal/controlplane -run TestPolicyChangeRevokingCapabilityBlocksStaleAndReauthorizedExecution -count=1
go run ./examples/control-plane-policy-change
```

No network, credentials, LLM API key, or external service is required.

## Acceptance criteria

- T0 authorization returns ALLOW under policy v1.
- Policy v2 becomes active before the old decision is executed.
- Reusing the v1 authorization is denied with `stale_policy`.
- A new authorization evaluated against v2 is denied with `denied_by_policy`.
- Both denial paths emit `EXECUTION_BLOCKED` evidence.
- The side-effect callback is not invoked and no `side-effect.json` is created.
- Evidence contains the original authorization, stale block, v2 policy denial, and blocked execution.
- Independent session verification returns PASS.
- The command exits non-zero if any assertion fails.

## Scope

This is a deterministic in-process control-plane experiment. It does not claim that a deployed SandBase Harness session or a production end-to-end pilot has passed. External integration evidence must be collected separately.
