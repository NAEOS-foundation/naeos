# Attack Matrix v5.2 — Runtime Event Evidence Binding

| ID | Attack / failure mode | Control | Expected result |
|---|---|---|---|
| AM-19 | Evidence references no runtime event | Runtime event binding | BLOCK |
| AM-20 | Referenced event does not exist | Event lookup | BLOCK |
| AM-21 | Event payload digest differs from evidence | Payload binding | BLOCK |
| AM-22 | Event belongs to another run | Run binding | BLOCK |
| AM-23 | Event type differs from lifecycle kind | Event-type binding | BLOCK |
| AM-24 | Runtime event occurs after completion boundary | Exact event count / sequence | BLOCK |
| AM-25 | Valid observed event and matching evidence | Full runtime-event binding | PASS |

## Scope

v5.2 strengthens the repository completion boundary by requiring lifecycle evidence to reference runtime observations recorded in a separate RuntimeEventStore.

## Claim discipline

The experiment demonstrates deterministic repository-level enforcement. It is not a production distributed event-sourcing implementation, tamper-proof external audit system, or penetration test.
