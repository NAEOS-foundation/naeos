# P1.6 — Public Control-Plane Golden Path

## Milestone

`P0 DONE → P1.1 DONE → P1.2 DONE → P1.3 DONE → P1.4 DONE → P1.5 AUTOMATED VALIDATION GATE DONE → P1.6`

## Protocol

### ALLOW

```
request → policy evaluation → ALLOW → execute → side effect → evidence → independent verification → PASS
```

### DENY

```
request → policy evaluation → DENY → execution blocked → no side effect → evidence → independent verification → PASS
```

Both paths use the same agent, policy, decision gateway, ledger, and verifier. Only the requested capability changes.

## Evidence contract

Every path is correlated by `request_id`. The ledger records:

- `AUTHORIZATION_DECISION`
- `EXECUTION_ALLOWED` or `EXECUTION_BLOCKED`
- `SIDE_EFFECT_OBSERVED`

The final result is written to `result.json`.

## Safety boundary

The side effect is a disposable local file in an isolated temporary directory. No production resource or external credential is touched.

This proves the control-plane protocol and evidence boundary; it does not claim production deployment or external-system durability.

## CI verification record

The repository now executes this golden path as a dedicated required CI job.

- Pull request: [#702 — Execute and verify P1.6 control-plane golden path](https://github.com/NAEOS-foundation/naeos/pull/702)
- Merged commit: `183192240501ea76cd6b597a6ba8a50b9c50823b`
- CI job: `P1.6 Control Plane Golden Path`
- Result on the merged change: `success`
- Required CI gate on the PR head: `success`

The CI job runs the demo in an isolated temporary directory and verifies that:

1. The ALLOW path creates its expected local side-effect file.
2. The DENY path creates no side-effect file.
3. Both decisions are represented in `result.json` with evidence and a successful verification result.

This is repeatable repository-level acceptance evidence. It supplements the protocol above; it does not expand the claim to production deployment, external-system durability, or arbitrary real-world actions.

