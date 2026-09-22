# Evidence Runtime Provenance v5.1

## Claim

The completion boundary must reject evidence that is structurally complete but not demonstrably tied to the expected runtime stage and event.

## Scenarios

1. Complete provenance passes.
2. Missing provenance is blocked.
3. Mismatched stage/event provenance is blocked.

## Run

    go run ./experiments/evidence-runtime-provenance-v5-1

This is a deterministic repository-level enforcement experiment. It does not claim production event-bus durability, distributed tracing completeness, or cryptographic key-management guarantees.
