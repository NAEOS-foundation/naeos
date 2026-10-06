# NAEOS External Validation Report #001

**Status:** PASS  
**Scope:** External runtime validation of authorization, execution binding, policy freshness, denial, and crash recovery  
**Repository:** NAEOS-foundation/naeos  
**Validation issue:** #502  
**Validation baseline:** main after PR #506 merge, commit `567aad2054bc639d416fed4db5f036f14fb0b4f0`

## 1. Executive summary

External Validation #001 closes the first NAEOS runtime evidence loop.

The validation combines three independent runtime validations:

- **CrewAI + Attenu:** consequential execution authorization, denial, revocation, durable ledger integrity, and independent verification.
- **Temporal:** crash after a consequential side effect and recovery on another worker, demonstrating that a logical invocation can be retried after a worker failure.
- **CrewAI policy-boundary validation:** mutation after authorization and policy-version freshness enforcement at the public `before_tool_call` boundary.

The result is **PASS for the documented scenarios**.

This report does not claim production readiness, exactly-once semantics for arbitrary external side effects, enterprise compliance, or blanket security guarantees.

## 2. Acceptance matrix

| Scenario | Runtime | Result | Evidence |
|---|---|---:|---|
| ALLOW → exact invocation → observed outcome | CrewAI + Attenu | PASS | PR #503 |
| BLOCK/DENY → no execution outcome | CrewAI + Attenu / CrewAI | PASS | PR #503, #506 |
| Mutation after ALLOW | CrewAI | PASS | PR #506 |
| Policy v1 → v2 → stale authorization blocked | CrewAI | PASS | PR #506 |
| Worker/process crash → recovery | Temporal | PASS | PR #505 |
| Same logical invocation preserved across recovery | Temporal | PASS | PR #505 |
| Independent evidence verification | Attenu | PASS | PR #503 |

## 3. Evidence records

### 3.1 CrewAI + Attenu

PR #503 executed the upstream CrewAI + Attenu integration against a real runtime.

Observed:

- authorized `crm_query` executed and produced an outcome;
- unauthorized `crm_export` was denied;
- revoked `crm_query` was denied;
- the ledger contained 10 events;
- the ledger hash chain verified;
- offline verification reported integrity, monotonicity, containment, and a verified anchor;
- the baseline without the enforcement bridge demonstrated the corresponding exfiltration path.

The retained artifact was independently hashed:

`sha256:2fe4c6df6225bae14d397f5e634c42b484f758ca10db1181effff5960ade1799`

### 3.2 Temporal crash/recovery

PR #505 used a real Temporal worker boundary.

The test intentionally performed the consequential side effect, persisted the side-effect record, and killed worker 1 before the Activity completed. Worker 2 then resumed the same Workflow/Activity.

The evidence demonstrates:

`authorize → execute → side effect → worker crash → retry/recovery → same logical invocation`

The result is deliberately **not** an exactly-once claim. It demonstrates why authorization, execution attempt, observation, and evidence must be bound to a stable logical invocation identity across worker recovery.

### 3.3 CrewAI mutation and freshness

PR #506 used CrewAI 1.15.16 public `before_tool_call` hook dispatch and an actual `BaseTool` body.

Mutation scenario:

`authorize → mutate canonical input → verify binding → BLOCK`

The mutation changed the canonical invocation digest and the real tool body did not execute.

Freshness scenario:

`ALLOW(v1) → policy changes to v2 → consequential call → BLOCK`

The stale authorization did not reach the real tool body.

Runtime artifact:

`external-validation-crewai-boundaries-d0648e80101779cbe8c655ef4a28834ae58ef23e`

Artifact SHA-256:

`5776dfef2d0952f35c5eeff6426775b91b884a4cb60b25a2c70e8e6674869994`

## 4. Evidence contract validated

Across the external validations, the core contract is:

`invocation_id + tool/target + canonical args digest + policy/context version/digest + decision`

Execution outcome is bound separately to the same logical invocation.

The independent verification question is therefore:

1. Was this exact invocation authorized?
2. Does the recorded outcome belong to that authorized invocation?

This separation is important because a policy decision, an execution attempt, an observed side effect, and a persisted evidence record can fail at different points in time.

## 5. What this establishes

The evidence establishes that the documented NAEOS control-plane properties can be exercised against real external runtimes for the tested boundaries:

`Intent → Authorization → Execution Boundary → Observation → Evidence → Verification`

It also establishes three important failure-boundary behaviors:

- **mutation invalidates the previously authorized invocation;**
- **policy changes invalidate stale authorization;**
- **worker failure can cause execution retry, so logical invocation identity must survive process boundaries.**

## 6. What this does not establish

This validation does **not** establish:

- production readiness;
- exactly-once execution for arbitrary external side effects;
- security of every AI agent or provider;
- enterprise compliance or certification;
- scalability or performance targets;
- correctness of arbitrary policy definitions;
- safety of every consequential action.

Those require separate evidence appropriate to each claim.

## 7. Reproduction references

The canonical repository paths remain:

- [Golden Path](GOLDEN-PATH.md)
- [Reference Demo](REFERENCE-DEMO.md)
- [External Validation Runbook](EXTERNAL-VALIDATION.md)

The external runtime validations are retained in their corresponding pull-request workflows and artifacts.

## 8. Conclusion

**External Validation #001: PASS.**

The first NAEOS external evidence loop is complete. The next engineering phase should prioritize making this evidence easy for an independent engineer or design partner to reproduce and understand, rather than expanding the core feature surface.

> Evidence should be reproducible before it is persuasive.
