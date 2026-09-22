# NAEOS Level-3 Evidence Experiment v1

## Thesis

> **A policy decision is not proof that an unauthorized side effect was prevented.**

This experiment closes the remaining evidence gap between deterministic governance experiments and deployment-level observation by exercising the real NAEOS execution gateway against a real local filesystem side effect.

The lifecycle is:

`AGENT INTENT → POLICY DECISION → GATEWAY → REAL SIDE EFFECT → INDEPENDENT OBSERVATION → EVIDENCE → INDEPENDENT VERIFICATION`

No LLM, network service, or external dependency is required.

## What is real

The experiment reuses existing NAEOS components:

- `internal/governance/control` — policy authorization;
- `internal/runtime/gateway` — the execution enforcement boundary;
- `internal/evidence` — append-only, hash-chained evidence;
- `internal/verification` — independent verification contracts and verifier chains.

The experiment adds only a small local filesystem sandbox and an independent filesystem observer as experiment harnesses. It does **not** introduce a second policy engine, runtime gateway, or evidence store.

## Scenarios

| Scenario | Expected result |
|---|---|
| ALLOW | Gateway executes a real file write; an independent observer reads the file; evidence records the observed artifact; verification passes. |
| DENY | Gateway denies the request; sandbox is never invoked; observer confirms no side effect. |
| Direct bypass | A file is written outside the gateway; the observer sees the side effect, while gateway history contains no authorized execution. The experiment records this as a governance failure, not as a successful authorization. |
| Execution/observation separation | A sandbox can report completion without persisting the effect; the independent observer detects the mismatch and verification fails. |
| Tamper after observation | The observed artifact is changed after evidence capture; independent verification detects the digest mismatch. |
| Reproducibility | The same deterministic run produces the same scenario outcomes and artifact contents. |

## Run

From the repository root:

```bash
go run ./experiments/level3-evidence
```

Exit code `0` means all expected assertions, including expected failure detections, passed.

## Evidence boundary

The experiment demonstrates a concrete local side effect and independent observation under a controlled test harness. It is stronger evidence than a simulated execution result, but it is **not** production security validation, OS-level isolation, or proof that an external agent cannot bypass a deployment boundary.

The direct-bypass scenario is intentionally important: NAEOS can detect that a side effect occurred outside the authorized execution path; preventing every possible out-of-band write requires an actual deployment enforcement boundary.
