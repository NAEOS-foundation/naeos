# NAEOS Evidence Tampering Benchmark v1

Status: reproducible proof benchmark

## Purpose

This benchmark tests one bounded security property:

> If an evidence record is modified after append, the modified record must not verify against the original integrity hash.

It uses the existing `internal/evidence.EvidenceStore` and `internal/verification.EvidenceChainVerifier`. No new integrity primitive is introduced.

## Scenarios

| Scenario | Mutation | Expected result |
|---|---|---|
| execution-output | change recorded execution output | FAILED |
| authorization-decision | change recorded policy decision | FAILED |
| stored-hash | replace stored hash | FAILED |

The benchmark first verifies an untampered baseline, then applies each mutation to an independent copy of the record. The authoritative store must remain intact throughout.

## Run

```bash
go run ./experiments/evidence-tampering-v1
```

A successful run emits machine-readable JSON and exits zero.

## What this proves

- The current evidence hash covers semantic evidence fields.
- A changed evidence payload cannot reuse the original stored hash.
- The existing evidence-chain verifier detects the mutation.
- The authoritative evidence store remains intact during the negative tests.

## What this does not prove

- Cryptographic signatures or key management.
- Protection against compromise of the process holding the evidence store.
- Artifact tampering detection.
- Authorization replay prevention.
- Policy mutation or capability escalation resistance.
- Production durability or distributed-store guarantees.

Those are separate evidence tiers and must not be inferred from this benchmark.

## Promotion rule

A future change that causes any tampered record to verify is a benchmark regression.
