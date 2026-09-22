# Evidence Integrity & Run Binding v5

Deterministic experiment for the completion-boundary invariants introduced in Attack Matrix v5.

## Scenarios

- complete run: accepted;
- interleaved evidence from another run: rejected;
- duplicate required evidence kinds: rejected;
- run-bound evidence contract: accepted only when the deterministic binding is valid.

The reusable enforcement lives in internal/evidence.ValidateCompletion, while the pipeline lifecycle supplies the run binding and exact predecessor identity.

## Run

    go run ./experiments/evidence-integrity-run-binding-v5

## Claim discipline

This is a repository-level deterministic experiment. It demonstrates lifecycle enforcement and tamper-evident chain verification; it is not a production distributed audit-durability guarantee, cryptographic key-management system, or penetration test.
