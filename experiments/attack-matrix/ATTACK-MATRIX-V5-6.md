# Attack Matrix v5.6 — Durable Runtime Receipt & Recovery

| ID | Attack / failure mode | Expected | Covered |
|---|---|---|---|
| AM-46 | Completion claims durability without a receipt | BLOCK | Partial |
| AM-47 | Receipt references mutated ledger | BLOCK | Yes |
| AM-48 | Receipt references foreign run | BLOCK | Yes |
| AM-49 | Receipt identity is tampered | BLOCK | Yes |
| AM-50 | Ledger reload after process restart | BLOCK on corruption / PASS on valid state | Yes |
| AM-51 | Receipt version mismatch | BLOCK | Yes |
| AM-52 | Receipt written non-atomically | BLOCK on incomplete receipt | Partial |

## Claim discipline

V5.6 provides a durable receipt/recovery contract at repository level. It does not provide immutable storage, process isolation, remote attestation, distributed consensus, or externally trusted telemetry.

AM-46 remains partial until the production pipeline makes durable receipt verification a mandatory completion prerequisite.
