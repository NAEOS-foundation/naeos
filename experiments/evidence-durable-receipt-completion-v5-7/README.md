# V5.7 — Durable Receipt Completion Boundary

V5.7 makes the durable runtime receipt a mandatory prerequisite for pipeline completion.

## Architecture

`Runtime → Independent Observer → Durable Append-Only Ledger → Seal → Durable Receipt → Reload/Verify → Completion`

The pipeline now:
1. persists runtime observations through the independent observer;
2. seals the observer and durable ledger;
3. creates and atomically persists a durable runtime receipt;
4. reloads and verifies that receipt;
5. only then enters the completion validator.

## Security property

Completion is blocked when the durable ledger is missing or invalid, or when the receipt cannot be verified against the run identity and ledger digest.

## Claim discipline

This is a repository/pipeline enforcement boundary. It does not claim immutable storage, process isolation, distributed consensus, remote attestation, or externally trusted telemetry.

## Regression coverage

- completion rejects a missing durable receipt;
- completion accepts a verified durable receipt;
- runtime observations are persisted before evidence references them.
