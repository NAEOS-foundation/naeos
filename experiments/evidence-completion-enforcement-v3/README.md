# Evidence Completion Enforcement v3

This experiment promotes the AM-18 evidence-completeness invariant from an experiment-only verifier into a reusable lifecycle gate.

## Invariant

A consequential run MUST NOT transition to RUN COMPLETE unless the completion gate confirms:

1. every required evidence kind is present;
2. evidence is contiguous and ordered;
3. non-root evidence records identify their predecessor; and
4. evidence is bound to the requested run identity.

The gate returns a structured CompletionResult. The execution lifecycle is expected to treat Complete == false as a hard stop.

## Deterministic scenarios

| Scenario | Expected gate result |
|---|---|
| Complete evidence chain | Complete=true |
| Missing required evidence | Complete=false |
| Reordered sequence | Complete=false |
| Evidence detached to another run | Complete=false |

## Why v3 exists

Attack Matrix v2 demonstrated that an independent verifier can reject an incomplete run. That alone does not establish an enforcement boundary.

v3 moves the invariant into internal/evidence.ValidateCompletion, creating a reusable control point that an executor can call before declaring a consequential run complete.

This experiment still does not claim production-grade distributed durability, concurrency control across external stores, or a security audit.

## Run

    go run ./experiments/evidence-completion-enforcement-v3
