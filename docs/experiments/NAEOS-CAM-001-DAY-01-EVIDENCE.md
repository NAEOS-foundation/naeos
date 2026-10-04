# Day 01 — Runtime Evidence Record

This record was populated from an actual GitHub Actions execution of the experiment.

## Run metadata
- Campaign: NAEOS-CAM-001
- Experiment: Governance Lifecycle v1
- Repository commit SHA: `9f09a8a11725b421d7db396b0fd7c26b200a54a4`
- Environment: GitHub Actions, Ubuntu 24.04.5 LTS, Go 1.26.6
- Command: `go run ./experiments/governance-lifecycle`
- Timestamp: `2026-10-03T10:40:18Z`
- Workflow run: `37117144805`
- Exit code: `0`

## Observed output

```json
{
  "experiment": "NAEOS Governance Lifecycle v1",
  "thesis": "The agent's claim is not the evidence.",
  "results": [
    {
      "name": "01-complete-lifecycle",
      "expected": "ALLOW + executed + observed + VERIFIED",
      "observed": "ALLOW + executed + observed + VERIFIED",
      "passed": true,
      "verification_state": "VERIFIED"
    },
    {
      "name": "02-agent-claim-vs-observation",
      "expected": "execution claim must not equal observed side effect",
      "observed": "agent claimed success; observation=false",
      "passed": true
    },
    {
      "name": "03-artifact-mutated-after-approval",
      "expected": "VERIFICATION FAILED",
      "observed": "FAILED",
      "passed": true,
      "verification_state": "FAILED"
    },
    {
      "name": "04-stale-authorization-replay",
      "expected": "OLD AUTHORIZATION REJECTED AFTER POLICY CHANGE",
      "observed": "stored=1.0.0/current=2.0.0/verification=FAILED",
      "passed": true,
      "verification_state": "FAILED"
    }
  ],
  "summary": "4/4 lifecycle scenarios passed; evidence chain intact; the agent's claim is not treated as authoritative evidence"
}
```

## Scenario results
| Scenario | Expected | Observed | Status |
|---|---|---|---|
| Complete lifecycle | lifecycle verifies | ALLOW + executed + observed + VERIFIED | PASS |
| Agent claim vs observation | claim alone is insufficient | agent claimed success; observation=false | PASS |
| Artifact mutation after approval | mutated artifact rejected | verification FAILED | PASS |
| Stale authorization replay | stale authorization rejected | stored=1.0.0/current=2.0.0/verification=FAILED | PASS |

## Evidence integrity
- [x] Output copied from actual runtime execution.
- [x] Commit SHA recorded.
- [x] Exit code recorded.
- [x] No result was inferred from source code alone.
- [x] Limitations are preserved from the experiment README.

## Publication statement

This record reports one reproducible NAEOS experiment run. It demonstrates only the tested governance invariants and is not, by itself, a production security, compliance, or universal AI-agent safety claim.

## Independent reproduction

Re-run the command from the repository root and compare the scenario results with this record.
