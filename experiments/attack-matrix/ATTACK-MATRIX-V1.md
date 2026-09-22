# NAEOS Attack Matrix v1

Status: Active baseline  
Version: 1.0  
Scope: deterministic repository experiments

## Matrix

| ID | Attack / failure mode | Boundary | Existing experiment | Expected invariant | Status |
|---|---|---|---|---|---|
| AM-01 | Mutable instruction artifact removes governance guidance | Instruction integrity | Policy Bypass v1 | Instructions are not the policy authority; tampering cannot silently change authorization | Covered |
| AM-02 | Empty effective policy set in governed mode | Governance configuration | Policy Bypass v1 | Governed execution cannot silently proceed with zero effective policies | Covered |
| AM-03 | NaN / +Inf / -Inf satisfies finite numeric condition | Evaluator semantics | Policy Bypass v1 | Non-finite operands are rejected or explicitly defined before comparison | Covered |
| AM-04 | Policy claims a field the evaluator context does not contain | Policy-context integrity | Policy Bypass H4 | Supported governed fields are present in a versioned context; unsupported/missing context is distinguishable from DENY | Covered |
| AM-05 | DENY path accidentally reaches a real side effect | Runtime gateway | AI Agent Policy Boundary; Level-3 Evidence | DENY must not execute through the authorized gateway | Covered |
| AM-06 | REQUIRE_APPROVAL executes without approval | Runtime gateway | AI Agent Policy Boundary | Approval-required execution remains blocked until the required approval path exists | Covered |
| AM-07 | Direct side effect occurs outside the gateway | Runtime/evidence boundary | AI Agent Policy Boundary; Level-3 Evidence | Observation must distinguish an out-of-band side effect from an authorized execution | Covered |
| AM-08 | Executor reports success while durable effect is absent | Execution/observation separation | Level-3 Evidence | Execution completion is not proof of observed side effect | Covered |
| AM-09 | Artifact changes after evidence capture | Evidence integrity | Level-3 Evidence | Independent verification detects digest mismatch after observation | Covered |
| AM-10 | Replay of an already accepted handoff | Handoff authority | Agent Handoff Governance | A nonce cannot be accepted twice | Covered |
| AM-11 | Capability widening during handoff | Handoff authority | Agent Handoff Governance | Requested capability must remain within authorized capability | Covered |
| AM-12 | Downstream escalation beyond parent authority | Handoff authority | Agent Handoff Governance | A downstream handoff cannot gain authority absent from its parent | Covered |
| AM-13 | Handoff payload tampering | Handoff integrity | Agent Handoff Governance | Declared payload digest must match the received payload | Covered |
| AM-14 | Provenance mismatch | Handoff provenance | Agent Handoff Governance | Claimed source must match the authenticated initiator/provenance | Covered |
| AM-15 | Unsupported contract/canonicalization version | Protocol integrity | Agent Handoff Governance | Version mismatch fails closed | Covered |
| AM-16 | Expired authorization | Handoff lifecycle | Agent Handoff Governance | Expired authorization is rejected | Covered |
| AM-17 | Signature tampering | Handoff integrity | Agent Handoff Governance | Modified signature invalidates the handoff | Covered |
| AM-18 | Governance decision exists but evidence is absent/incomplete | Evidence boundary | Governance Lifecycle + Level-3 Evidence | Consequential execution must have auditable evidence linking intent, decision, execution, observation, and verification | Gap — promote to v2 |

## Coverage interpretation

**Covered** means the repository contains a deterministic scenario with an explicit expected invariant. It does not mean that all implementations or deployment environments are protected.

**Gap** means the matrix identifies a useful boundary that is not yet represented by a dedicated acceptance scenario.

## Promotion rule

A v1 row may be marked covered only when:

1. the referenced experiment is runnable;
2. the expected invariant is explicit;
3. the scenario result is deterministic;
4. the scenario remains in the regression suite;
5. expected failure detection is treated as a passing assertion where appropriate.

## v2 candidates

The first promotion candidate is AM-18: end-to-end evidence completeness. The v2 experiment should verify that a consequential run cannot be considered complete when one required evidence link is missing, reordered, or detached from the run identity.

Additional v2 work should be promoted only after a concrete invariant and deterministic oracle are defined.
