# NAEOS-CAM-001 — Day 05: Atomic Execution Boundary

**Campaign:** 30 Days of AI Agent Governance  
**Day:** 05  
**Experiment:** Atomic Execution Commit Boundary (P1.8)  
**Status:** Public experiment brief

## Question

What happens when a policy update races with an already-authorized execution at the check-to-side-effect boundary?

## Reproduction

Run from the repository root:

```bash
./examples/control-plane-atomic-execution/run-demo.sh
```

or:

```bash
go run ./examples/control-plane-atomic-execution
```

The experiment is local and deterministic. It requires no network, credentials, LLM API key, or external service.

## Scenario

The executable example models this sequence:

1. Policy v1 is active.
2. `repository.write` receives an `ALLOW` decision.
3. `ExecuteAtomic` enters the policy-store execution boundary.
4. The authorized side-effect callback creates a disposable local artifact while the execution lock is held.
5. A concurrent policy update to v2 is attempted while the callback is still active.
6. The policy update cannot commit until the atomic side effect returns.
7. Successful execution records `EXECUTION_ALLOWED`.
8. Policy v2 becomes active only after the atomic execution completes.
9. Independent session verification returns `PASS`.

## Acceptance criteria

A valid run should demonstrate all of the following:

- Initial authorization is `ALLOW`.
- Execution remains authorized at the execution boundary.
- The side effect is actually observed.
- Execution evidence records `EXECUTION_ALLOWED`.
- The authorization is bound to policy v1.
- Policy v2 becomes current only after the atomic execution completes.
- Session verification returns `PASS`.
- The regression condition shows that the concurrent policy update cannot complete while the side-effect callback is blocked.

The experiment should exit non-zero if an assertion fails.

## Engineering invariant

> Policy freshness validation and the authorized side effect must share the same in-process execution boundary when atomic execution is claimed.

This is the concrete P1.8 control-plane property exercised by the experiment.

## Evidence to publish

When publishing a run, record:

- exact command
- commit SHA
- environment
- UTC timestamp
- process exit code
- relevant machine-readable runtime output
- scenario-level results
- verification result
- evidence/artifact identifiers, if produced

Do not fabricate runtime output. The publication record should correspond to the exact commit and environment that produced it.

## Why it matters

A governance decision is not sufficient if policy can change between the final authorization check and the authorized side effect. This experiment makes that boundary explicit and tests whether the implementation serializes the policy update with the side effect.

The goal is not to claim that every external operation is atomic. The goal is to make one narrow, testable control-plane guarantee reproducible.

## Scope limitations

This experiment demonstrates an **in-process control-plane** property. It does not prove:

- distributed transaction semantics
- atomicity across external systems
- consensus across multiple policy stores
- production security or compliance
- universal safety of AI coding agents

## Discussion

If an AI coding agent is authorized to perform a side effect and governance changes at the same moment, **where should the system draw the atomic execution boundary?**

## CTA

**Run the experiment and compare your observed result.**
