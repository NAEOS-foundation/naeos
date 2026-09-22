# NAEOS Attack Matrix v2

Status: Active
Version: 2.0
Scope: deterministic repository experiments

## AM-18 promotion

AM-18 from Attack Matrix v1 is promoted from **Gap** to **Covered** by the Evidence Completeness v2 experiment.

| ID | Attack / failure mode | Boundary | Experiment | Expected invariant | Status |
|---|---|---|---|---|---|
| AM-18 | Governance decision exists but evidence is absent, reordered, or detached | Evidence boundary | Attack Matrix v2 — Evidence Completeness | A consequential run is VERIFIED only when required evidence is complete, ordered, linked, and bound to one run identity | Covered |

## Deterministic scenarios

1. **Complete chain** — all required evidence kinds are present and correctly linked.
2. **Missing evidence** — a required observation is absent; verification must fail.
3. **Reordered evidence** — the logical sequence is invalid even though the underlying hash chain remains intact; verification must fail.
4. **Detached evidence** — one evidence record belongs to another run identity; verification must fail.

The experiment treats expected verification failures as passing assertions.

## Boundary distinction

The evidence store's cryptographic/hash-chain integrity and the run's logical completeness are separate controls:

- hash-chain verification answers whether evidence records were altered;
- completeness verification answers whether the required evidence exists and belongs to the same run in the required order.

A hash-valid evidence store can therefore still describe an incomplete or invalid consequential run.

## Claim discipline

Covered means the repository contains a deterministic scenario and explicit oracle for this boundary. It does not establish distributed-system guarantees, production audit durability, or security against every external tampering mechanism.
