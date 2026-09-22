# Attack Matrix v5.3 — Independent Runtime Event Ledger

| ID | Attack / failure mode | Expected | Covered |
|---|---|---|---|
| AM-26 | Evidence created without a pre-existing runtime event | BLOCK | Yes |
| AM-27 | Runtime event belongs to another run | BLOCK | Yes |
| AM-28 | Runtime event arrives after completion boundary | BLOCK | Yes |
| AM-29 | Missing lifecycle runtime event | BLOCK | Yes |
| AM-30 | Evidence builder attempts to manufacture event identity | BLOCK by API boundary | Yes |
| AM-31 | Runtime event identity tampering | BLOCK by ledger verification | Yes |

## Claim discipline

V5.3 demonstrates a repository-level trust-boundary separation between runtime event production and evidence construction. The in-memory ledger is not an independent external telemetry service and should not be described as one.
