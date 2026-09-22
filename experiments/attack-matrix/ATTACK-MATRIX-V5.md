# Attack Matrix v5 — Evidence Integrity & Run Binding

## Purpose

v5 closes the remaining completion-boundary integrity gaps identified after v4.

## Enforcement invariant

RUN COMPLETE MUST be rejected unless:

- required evidence kinds are unique;
- the backing evidence hash chain verifies;
- evidence for the target run is not interleaved with another run;
- every selected record carries the deterministic binding for the requested run;
- sequence numbers are contiguous;
- predecessor identity matches the actual previous record.

## Scenarios

| ID | Attack / failure mode | v5 result |
|---|---|---|
| AM-18E | Duplicate required evidence kind | Covered |
| AM-18F | Mixed-run evidence interleaving | Covered |
| AM-18G | Run-binding mismatch | Covered |
| AM-18H | Tampered evidence content / broken chain | Covered |

## Architectural significance

v4 established that evidence completion is a real runtime boundary. v5 adds integrity conditions so that a structurally complete-looking evidence set cannot pass merely because records have the expected labels and counts.

The completion gate now evaluates both lifecycle completeness and evidence integrity.

## Claim discipline

This is a deterministic repository-level enforcement experiment. It is not a production security audit, distributed durability proof, or penetration test.
