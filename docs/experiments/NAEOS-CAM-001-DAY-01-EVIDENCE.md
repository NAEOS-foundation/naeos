# Day 01 — Runtime Evidence Record

Use this file only after executing `go run ./experiments/governance-lifecycle` in the publication environment.

## Run metadata
- Campaign: NAEOS-CAM-001
- Experiment: Governance Lifecycle v1
- Repository commit SHA: `<commit-sha>`
- Environment: `<OS / Go version / runner>`
- Command: `go run ./experiments/governance-lifecycle`
- Timestamp: `<UTC timestamp>`
- Exit code: `<actual exit code>`

## Observed output

```text
<paste relevant machine-readable output here>
```

## Scenario results
| Scenario | Expected | Observed | Status |
|---|---|---|---|
| Complete lifecycle | lifecycle verifies | `<record actual>` | `<PASS/FAIL>` |
| Agent claim vs observation | claim alone is insufficient | `<record actual>` | `<PASS/FAIL>` |
| Artifact mutation after approval | mutated artifact rejected | `<record actual>` | `<PASS/FAIL>` |
| Stale authorization replay | stale authorization rejected | `<record actual>` | `<PASS/FAIL>` |

## Evidence integrity
- [ ] Output copied from actual runtime execution.
- [ ] Commit SHA recorded.
- [ ] Exit code recorded.
- [ ] No result was inferred from source code alone.
- [ ] Limitations are preserved from the experiment README.

## Publication statement

This record reports one reproducible NAEOS experiment run. It demonstrates only the tested governance invariants and is not, by itself, a production security, compliance, or universal AI-agent safety claim.

## Independent reproduction
Re-run the command from the repository root and compare the scenario results with this record.
