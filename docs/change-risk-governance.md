# Change Risk Governance

**Document ID:** NAEOS-GOV-CHG-001  
**Version:** 1.0.0  
**Status:** Proposed  
**Owner:** NAEOS Foundation

## Purpose

Dependency risk governance answers whether a dependency update is safe to accept. Change Risk Governance extends the same boundary to repository changes.

The governing model is:

**change surface × architectural criticality × verification evidence**

The classifier is deterministic and policy-versioned. It does not authorize execution; it records the risk classification and the verification gates required before merge.

## Decision boundary

```
Change proposal
      ↓
Change surface classification
      ↓
Architectural domain classification
      ↓
Risk classification
      ↓
Policy decision
      ├── allow
      ├── require_review
      └── deny
      ↓
Verification
      ↓
Human review when required
      ↓
Merge
      ↓
Post-merge verification
```

The model or agent may propose a change. The policy determines the governance requirements. CI executes verification. Durable evidence records the classification and decision.

## Change surfaces

| Surface | Candidate criticality |
|---|---|
| docs | low |
| tests | low |
| configuration | medium |
| code | medium |
| mixed | medium |
| unknown | deny |

A surface is only a starting point. Architectural domain and change size can raise criticality.

## Critical domains

Changes touching these domains have a minimum criticality of **high**:

- `policy`
- `security`
- `audit_evidence`
- `runtime`
- `deployment_ci`
- `architecture`

## Change size

Total additions plus deletions raise the minimum criticality:

- **500+ changed lines:** high
- **1000+ changed lines:** critical

## Risk matrix

| Criticality | Risk | Required gates | Decision |
|---|---|---|---|
| low | low | CI | allow |
| medium | medium | CI + governance | allow |
| high | high | CI + security + governance + human review | require_review |
| critical | critical | CI + security + benchmark + governance + human review | require_review |
| unknown | unknown | CI + security + governance + human review | deny |

A `require_review` result does not perform the review. Repository branch protection and review controls remain the human authorization boundary.

## Fail-closed rules

The evaluator denies the change when:

1. the base or head revision cannot be verified;
2. the change surface cannot be classified;
3. an affected domain is unknown;
4. required evidence cannot be established;
5. the policy or schema version is unsupported.

## Durable evidence

Each evaluation emits machine-readable evidence containing:

- policy ID and policy version;
- schema version;
- base and head revisions;
- change surface;
- architectural domains;
- changed-file count;
- additions and deletions;
- criticality;
- risk;
- required gates;
- final decision.

This record describes the governance decision. It is not proof that code executed successfully. Execution and verification evidence remain separate concerns.

## Relationship to dependency risk governance

Dependency Risk Governance and Change Risk Governance use the same NAEOS principle but govern different decision inputs:

- dependency governance: dependency change × architectural criticality × verification evidence;
- change governance: change surface × architectural criticality × verification evidence.

They intentionally remain separate evaluators.

## Implementation boundary

The current implementation:

1. loads and validates a machine-readable policy;
2. computes the changed-file set from the Git base/head range;
3. maps paths to architectural domains;
4. classifies surface, criticality, risk, and decision deterministically;
5. emits durable JSON evidence;
6. fails closed for unknown or unsupported conditions;
7. runs in CI on pull requests and main.

The next boundary is to bind high-risk change decisions to explicit Policy Decision Records and the existing tamper-evident evidence infrastructure.
