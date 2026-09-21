# NAEOS Adversarial Governance Experiment v1 — Test Matrix

Purpose: machine-readable acceptance matrix for the three v1 architectural findings.

| ID | Finding | Baseline scenario(s) | Required invariant | Hardened outcome |
|---|---|---|---|---|
| AGV1-01 | Instruction Layer Is Not an Enforcement Boundary | prompt override dir neutralizes policy | Mutable instruction artifacts cannot silently become the policy authority | Override detected/rejected or execution remains governed by an external deterministic policy boundary |
| AGV1-02 | Empty Governance Configuration Silently Disables Enforcement | no configured policies => no checks | A governed run with zero effective policies cannot execute as if governed | Governed mode blocks; `governance.unconfigured` telemetry records the blocked state; policy-free mode is explicitly `ungoverned` |
| AGV1-03 | Policy Evaluator Accepts Non-Finite Numeric Values | NaN bypasses gt threshold; Inf bypasses lt bound | Non-finite operands cannot silently satisfy finite numeric constraints | `NaN`, `+Inf`, and `-Inf` are rejected before numeric comparison; finite comparisons remain unchanged |
| AGV1-04 | Policy Context Integrity | policy context integrity: security claim not evaluated | A policy must receive the data it claims to govern, or the run must be classified as not evaluated | H4 is explicitly `NOT_EVALUATED` until a versioned policy-context contract and regression tests are implemented |

## Verification protocol

1. Run the unchanged baseline harness.
2. Record the exact scenario result and evidence.
3. Change only the implementation needed for the selected invariant.
4. Re-run the same scenario.
5. Run the complete policy-bypass suite.
6. Add/update regression tests.
7. Record the final evidence and governance state.

## Promotion rule

A finding is considered hardened only when its adversarial scenario no longer reproduces AND the complete suite remains deterministic AND a regression test preserves the invariant. `NOT_EVALUATED` findings are tracked separately from ordinary DENY/ALLOW outcomes and are not counted as successful enforcement.

Do not delete, rename, or weaken a scenario merely to make the tally improve.