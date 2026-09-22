# Attack Matrix v5.5 — Durable Runtime Event Ledger

| ID | Attack / failure mode | Expected | Covered |
|---|---|---|---|
| AM-38 | Event deletion/truncation | BLOCK on reload/integrity | Yes |
| AM-39 | Durable record mutation | BLOCK | Yes |
| AM-40 | Ledger reordering | BLOCK | Yes |
| AM-41 | Replay/duplicate sequence | BLOCK by sequence contract | Partial |
| AM-42 | Observer write after seal | BLOCK | Yes |
| AM-43 | Corrupt journal record | BLOCK | Yes |
| AM-44 | Restart/recovery | Reload and verify | Yes |
| AM-45 | Completion without durable receipt | Not yet wired | No |

## Claim discipline

V5.5 establishes durable local persistence and integrity verification. It is not yet a process-isolated observer or externally trusted telemetry system.
