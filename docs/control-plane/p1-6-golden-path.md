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
