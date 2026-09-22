# Attack Matrix v5.1 — Runtime Evidence Provenance

## Purpose

v5.1 extends the completion boundary from structural evidence integrity to runtime provenance integrity.

## Enforcement invariant

RUN COMPLETE MUST be rejected unless each required evidence record:

- carries the expected runtime stage and event;
- carries a non-empty payload digest;
- carries a provenance digest derived from stage, event, and payload digest;
- remains covered by the v5 run binding and hash-chain checks.

## Scenarios

| ID | Attack / failure mode | v5.1 result |
|---|---|---|
| AM-19A | Missing runtime provenance | Covered |
| AM-19B | Mismatched runtime stage/event | Covered |
| AM-19C | Payload provenance digest mismatch | Covered |

## Claim discipline

This is a deterministic repository-level lifecycle enforcement experiment. It is not a production distributed tracing proof, durability guarantee, or penetration test.
