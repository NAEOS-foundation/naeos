# Attack Matrix v5.4 — Independent Runtime Observer

| ID | Attack / failure mode | Expected | Covered |
|---|---|---|---|
| AM-32 | Evidence builder attempts to publish through observer | BLOCK by read-only interface | Yes |
| AM-33 | Evidence references forged observer event | BLOCK | Yes |
| AM-34 | Runtime observation arrives after seal | BLOCK | Yes |
| AM-35 | Foreign-run observation is bound to target evidence | BLOCK | Yes |
| AM-36 | Consumer mutates observer snapshot | BLOCK by defensive copy | Yes |
| AM-37 | Observer ledger identity tampering | BLOCK by ledger verification | Yes |

## Claim discipline

V5.4 establishes a repository-level logical observer boundary. The observer and ledger remain in-process, so this is not yet a process-isolated or externally trusted telemetry system.

The next hardening step is durable append-only persistence and an independently operated/process-isolated observer.
