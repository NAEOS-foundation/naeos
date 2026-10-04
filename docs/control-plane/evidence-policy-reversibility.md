# Evidence Policy + Reversibility Experiment

## Thesis

An execution state must not transition from UNKNOWN merely because some evidence exists.

Evidence must be evaluated against an explicit policy considering evidence type, freshness, source independence, conflicting observations, and reversibility.

Irreversible actions require a higher verification threshold than reversible actions.

## Scenarios

1. Fresh provider receipt confirms a reversible operation.
2. A single signal is insufficient for an irreversible operation.
3. Two fresh independent signals confirm an irreversible operation.
4. Stale evidence keeps the state UNKNOWN.
5. Conflicting evidence keeps the state UNKNOWN.
6. Two signals from the same source do not count as independent proof.

## Intended transition

UNKNOWN -> CONFIRMED/REJECTED only when the evidence policy is satisfied.

Otherwise:

UNKNOWN -> UNKNOWN

The experiment deliberately does not infer certainty from the agent's narrative or a convenient single signal.
