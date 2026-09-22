# NAEOS Attack Matrix v2 — Evidence Completeness

Status: Active baseline
Version: 2.0
Scope: deterministic repository experiment

## Purpose

Attack Matrix v1 identified AM-18 as the next evidence boundary:

> A consequential run must not be considered complete when required evidence is missing, reordered, or detached from the run identity.

This experiment promotes AM-18 into a deterministic acceptance scenario using the existing NAEOS evidence store and verification chain.

## Boundary

`INTENT → POLICY DECISION → EXECUTION → OBSERVATION → VERIFICATION`

Every consequential run in this experiment must produce an ordered evidence sequence with a stable `run_id`.

Required evidence kinds:

1. `intent`
2. `decision`
3. `execution`
4. `observation`
5. `verification`

Each evidence record carries:

- `run_id`;
- `kind`;
- monotonic `sequence`;
- `previous_evidence_id` where applicable.

## Scenarios

| Scenario | Expected result |
|---|---|
| Complete evidence chain | VERIFIED |
| Missing required evidence | FAILED |
| Reordered evidence sequence | FAILED |
| Detached evidence from run identity | FAILED |

Expected verification failures are assertions: the experiment passes when the verifier detects the invalid condition.

## Run

From the repository root:

```bash
go run ./experiments/attack-matrix-v2
```

Exit code `0` means all expected assertions passed.

## Claim discipline

This experiment demonstrates deterministic evidence-completeness behavior in the repository. It is not a production audit-log guarantee, distributed consistency proof, or penetration test.
