# NAEOS Investor Demo — Scripted Walkthrough

This is the deterministic, reproducible investor demo intended for technical
and investor audiences. The complete sequence runs against the real control
plane (policy engine, capability authority, handoff validator, execution gate,
independent verifier, audit ledger) — no mocked BLOCK/ALLOW values.

Trigger it from the dashboard ("Run Investor Demo") or directly:

```bash
curl -X POST http://localhost:9090/api/investor-demo
```

## Steps

| # | Actor | Action | Result |
|---|-------|--------|--------|
| 1 | Agent | Receives task: "Update the payment service and run the tests." | OK |
| 2 | Agent | Requests `repository.read` | ALLOW |
| 3 | Agent | Requests `repository.write` | ALLOW |
| 4 | Agent | Requests `test.execute` | ALLOW |
| 5 | Agent | Attempts `credential.rotate` | BLOCK |
| 6 | Agent | Attempts `policy.modify` | BLOCK |
| 7 | Agent A → Agent B | Handoff requesting `credential.rotate` | HANDOFF REJECTED |
| 8 | Agent | Replays a previously valid handoff contract | REPLAY REJECTED |
| 9 | System | POLICY-017 updated v17 → v18; agent reuses old grant | STALE AUTHORIZATION → BLOCK |
| 10 | NAEOS | Independent verification of the whole session | VERIFICATION: FAIL |

## Final Security Posture

The response includes a posture summary computed from the audit ledger:

- Authorized actions
- Blocked actions
- Handoff violations
- Replay attempts
- Policy changes
- Verification status
- Audit event count

## Verification expectation

Because unauthorized actions were attempted, the independent verifier must
report **FAIL** — this is the point of the demo: reasoning is delegated,
authorization is not.