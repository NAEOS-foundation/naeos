# Evidence Verification Boundary Matrix — Experiment v1

This experiment converts the recent NAEOS engineering discussion into a deterministic executable matrix.

## Thesis

An execution record is evidence about execution. It is not automatically proof of an external side effect. Verification must evaluate evidence against a minimum floor derived from effect authority and reversibility.

## Scenarios

| Scenario | Expected |
|---|---|
| Internal + reversible + runtime record | CONFIRMED |
| External + reversible + runtime record only | UNKNOWN |
| External + reversible + provider receipt | CONFIRMED |
| Internal + irreversible + estimated evidence | UNKNOWN |
| External + irreversible + authoritative evidence | CONFIRMED |
| Stale evidence | REVERIFY |
| Contradictory evidence | UNKNOWN |
| Missing verification policy | ESCALATE |

## Invariants exercised

1. Runtime execution is not external proof.
2. Irreversible effects require authoritative evidence.
3. Estimated evidence never masquerades as observed evidence.
4. Stale evidence triggers re-verification.
5. Contradictory evidence does not silently become success or failure.
6. Missing verification policy fails closed.
7. Every result reports which evidence signals were used.

## Scope

This is an experiment, not yet a production API. It intentionally avoids changing the core evidence/verification types until the behavior and terminology are reviewed against the existing architecture.
