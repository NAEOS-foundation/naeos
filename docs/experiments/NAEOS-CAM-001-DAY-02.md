# NAEOS-CAM-001 — Day 02: Atomic Execution Commit Boundary

**Campaign:** 30 Days of AI Agent Governance  
**Day:** 02  
**Experiment:** EXP-002 — Atomic Execution Commit Boundary  
**Status:** Public experiment brief  
**Evidence policy:** Runtime results must be captured from an actual execution; no fabricated output.

## Question

What happens when a policy update races with an already-authorized AI-agent execution at the check-to-side-effect boundary?

## Hypothesis

If policy freshness validation and the authorized side effect share one in-process execution boundary, a concurrent policy update cannot commit between the authorization check and completion of that side effect.

## Reproduction

```bash
go run ./examples/control-plane-atomic-execution
```

No network, credentials, LLM API key, or external service is required.

## Expected acceptance criteria

1. Policy v1 is active before authorization.
2. `repository.write` receives `ALLOW`.
3. The atomic execution boundary is entered.
4. The authorized local side effect occurs while the boundary is held.
5. A concurrent policy update to v2 is serialized behind the execution.
6. Execution evidence records `EXECUTION_ALLOWED`.
7. The authorization remains bound to policy v1 for that execution.
8. Policy v2 becomes active after the atomic execution completes.
9. Independent session verification returns `PASS`.
10. The regression test confirms the policy update cannot complete while the side-effect callback is blocked.

## Engineering invariant

> Policy freshness validation and the authorized side effect must share the same in-process execution boundary when atomic execution is claimed.

## Scope boundary

This experiment demonstrates an in-process control-plane property. It does **not** establish distributed transaction semantics, external-system atomicity, or consensus across multiple policy stores.

## Publication evidence

Any public result should record:

- exact command
- exit code
- commit SHA
- execution environment
- UTC timestamp
- observed output
- scenario/regression result

Do not convert this experiment into a blanket production-security or compliance claim.

## Discussion

If an AI coding agent has already received authorization and policy changes milliseconds later, what execution boundary should determine whether the side effect remains authorized?

**CTA:** Run the experiment and compare the observed result.
