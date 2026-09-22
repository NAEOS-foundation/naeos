# Independent Runtime Observer v5.4

V5.4 introduces an explicit observer boundary between runtime execution and evidence construction.

## Trust boundary

`Runtime → Independent Runtime Observer → Runtime Event Ledger → Evidence Builder → Verification → Completion`

The observer owns event publication. Evidence construction receives a read-only observation interface and cannot publish or manufacture runtime events.

## Scenarios

- **VALID** — all five lifecycle observations exist before evidence is built; completion passes.
- **OBSERVER_BYPASS** — forged event references are rejected because they do not exist in the observer.
- **LATE_OBSERVATION** — observations after the sealed boundary are rejected.
- **FOREIGN_RUN** — an observation from another run cannot be bound to target evidence.

Run:

```bash
go run ./experiments/evidence-runtime-observer-v5-4
```

This is a deterministic repository experiment. V5.4 establishes a logical observer boundary; the observer and ledger are still in-process. It does not claim an externally trusted telemetry service or process-isolated observer.
