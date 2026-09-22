# Evidence Runtime Event Ledger v5.3

V5.3 separates runtime observation from evidence construction.

## Trust boundary

`Event Producer → Runtime Event Ledger → Evidence Builder → Verification → Completion`

The evidence builder has no capability to publish runtime events. It can only bind evidence to an event that already exists in the independent ledger.

## Scenarios

- **VALID** — all five lifecycle events exist before evidence is built; completion passes.
- **MISSING_EVENT** — one lifecycle event is absent; completion is blocked.
- **FOREIGN_EVENT** — an event from another run cannot be bound to target evidence.
- **LATE_EVENT** — the sealed ledger rejects events after the completion boundary.

Run:

```bash
go run ./experiments/evidence-runtime-event-ledger-v5-3
```

This is a deterministic repository experiment. It demonstrates separation of observation and evidence construction; it is not a claim of an externally trusted production telemetry system.
