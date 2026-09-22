# Durable Runtime Event Ledger v5.5

V5.5 hardens V5.4 with a durable append-only JSONL runtime event ledger.

## Trust boundary

`Runtime → Independent Observer → Durable Append-Only Ledger → Evidence Builder → Verification → Completion`

The ledger:
- persists events across observer restarts;
- verifies deterministic event identity on load;
- rejects mutation, truncation/reordering, invalid sequences, and writes after sealing;
- flushes each accepted record before returning.

This implementation is intentionally repository-local and deterministic. It does **not** claim distributed consensus, remote attestation, or an externally operated telemetry service.

Run:

```bash
go test ./internal/evidence -run DurableRuntimeEventLedger
```

Next hardening: connect the pipeline observer to the durable ledger and define an explicit recovery/receipt contract.
