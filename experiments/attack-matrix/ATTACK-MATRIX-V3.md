# Attack Matrix v3 — Evidence Completion Enforcement

## Purpose

AM-18 in Attack Matrix v1 identified evidence completeness as a gap. Attack Matrix v2 demonstrated deterministic rejection of missing, reordered, and detached evidence. v3 turns that invariant into a reusable lifecycle gate.

## Enforcement invariant

RUN COMPLETE MUST be rejected unless:

- required evidence exists for the requested run identity;
- evidence count matches the required contract;
- sequence is contiguous;
- predecessor links are present for non-root evidence.

The controllable boundary is evidence.ValidateCompletion. A caller must gate the final run-completion transition on its result.

## Scenarios

| ID | Attack / failure mode | v3 result |
|---|---|---|
| AM-18A | Complete evidence chain | Covered |
| AM-18B | Missing required evidence | Covered |
| AM-18C | Reordered evidence sequence | Covered |
| AM-18D | Evidence detached to another run | Covered |

## Architectural significance

hash-chain integrity proves that stored evidence was not altered; completion enforcement determines whether a consequential run may be declared complete.

A valid hash chain is therefore necessary but not sufficient for completion.

## Claim discipline

This is a deterministic repository-level enforcement experiment. It is not a production security audit, distributed durability proof, or penetration test.
