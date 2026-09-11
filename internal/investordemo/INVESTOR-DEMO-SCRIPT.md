# NAEOS Investor Demo — Strict English Script

This is the deterministic and reproducible investor demo intended for technical
and investment audiences. The sequence runs against the actual control plane:
policy engine, capability authority, handoff validator, execution gate,
independent verifier, and audit ledger. No simulated decisions are used.

Trigger the demo from the dashboard, or execute it directly:

```bash
curl -X POST http://localhost:9090/api/investor-demo
```

## Presenter Script

### Step 1 — Task Intake
"The agent receives a task to update the payment service and run the relevant verification steps."

Result: Success.

### Step 2 — Approved Read Access
"The agent requests read access to the repository. This is within scope and is granted."

Result: ALLOW.

### Step 3 — Approved Write Access
"The agent requests write access to the repository within the approved work scope. This is also permitted."

Result: ALLOW.

### Step 4 — Approved Test Execution
"The agent requests execution of the test suite. This is permitted under the active grant and policy."

Result: ALLOW.

### Step 5 — Protected Capability Attempt
"The agent then attempts to rotate credentials. This action is not in the grant, and it is protected by policy. The request is rejected."

Result: BLOCK.

### Step 6 — Policy Tampering Attempt
"The agent attempts to modify the governance policy that governs its own permissions. This is a direct trust-boundary violation and is denied."

Result: BLOCK.

### Step 7 — Cross-Agent Escalation Attempt
"Agent A attempts to hand off to Agent B with a request for credential rotation. The handoff contract is evaluated and rejected because the requested capability exceeds the authorized scope."

Result: HANDOFF REJECTED.

### Step 8 — Replay Attempt
"The agent reuses a previously valid handoff contract. The nonce is recognized as previously seen, and the replay is rejected."

Result: REPLAY REJECTED.

### Step 9 — Stale Authorization Check
"The system updates policy version POLICY-017 to POLICY-018. The agent attempts to reuse a stale grant generated under the older policy. The authorization is revalidated and denied."

Result: STALE AUTHORIZATION → BLOCK.

### Step 10 — Independent Verification
"NAEOS performs an independent verification of the full session. The verifier checks the audit ledger, validates the policy history, and confirms that the unauthorized attempts were not accepted. The session fails verification."

Result: VERIFICATION: FAIL.

## The Core Message

"This system demonstrates a critical principle: the agent may generate intent, but it does not control authorization. The control plane decides what is allowed, what is denied, and what must be independently verified."

## Scenario Summary Table

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
| 9 | System | Updates POLICY-017 to POLICY-018 and reuses old grant | STALE AUTHORIZATION → BLOCK |
| 10 | NAEOS | Performs independent verification of the session | VERIFICATION: FAIL |

## Final Security Posture

The response includes a posture summary computed from the audit ledger:

- Authorized actions
- Blocked actions
- Handoff violations
- Replay attempts
- Policy changes
- Verification status
- Audit event count

## Verification Expectation

Because unauthorized actions were attempted, the independent verifier must report
**FAIL**. This is the essential point of the demo: reasoning is delegated,
authorization is not.