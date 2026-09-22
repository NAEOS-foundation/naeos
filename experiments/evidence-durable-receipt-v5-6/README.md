# Durable Runtime Receipt & Recovery v5.6

V5.6 introduces an explicit durable receipt contract for a sealed runtime event ledger.

## Trust boundary

`Runtime → Independent Observer → Durable Append-Only Ledger → Seal → Durable Receipt → Restart/Recovery Verification → Completion`

The receipt binds:
- run identity;
- event count;
- first and terminal event identity;
- canonical ledger digest;
- receipt version;
- seal timestamp;
- deterministic receipt identity.

## Covered scenarios

1. valid receipt survives reload;
2. ledger mutation after sealing is rejected;
3. foreign run receipt verification is rejected;
4. receipt tampering is rejected.

Run:

```bash
go run ./experiments/evidence-durable-receipt-v5-6
```

## Claim discipline

V5.6 establishes a repository-level durable receipt and recovery contract. It does not claim process isolation, remote attestation, distributed consensus, immutable storage, or externally trusted telemetry.

Pipeline integration remains a separate boundary: production callers must persist and verify the receipt before treating durable completion as established.
