# Dependency Risk Governance

**Document ID:** NAEOS-GOV-DEP-001  
**Version:** 1.0.0  
**Status:** Proposed  
**Owner:** NAEOS Foundation

## Purpose

Dependabot and semantic versioning identify dependency updates, but they do not by themselves establish operational risk. NAEOS therefore classifies dependency changes using:

**dependency change × architectural criticality × verification evidence**

This policy is intentionally small. It defines the governance boundary before automation is expanded.

## Decision boundary

The model may propose a dependency classification. The policy determines the required verification and merge decision. CI executes only the checks required by that decision.

```
Dependabot proposal
      ↓
Risk classification
      ↓
Policy decision
      ├── allow
      ├── require_review
      └── deny
      ↓
Verification gates
      ↓
Human review when required
      ↓
Merge
      ↓
Post-merge verification
```

## Critical domains

The following dependency domains are architecturally sensitive:

- `policy`
- `security`
- `audit_evidence`
- `runtime`
- `deployment_ci`

A dependency touching one of these domains has a minimum criticality of `high`.

## Risk classes

| Criticality | Risk | Minimum verification | Decision |
|---|---|---|---|
| low | low | CI | allow |
| medium | medium | CI + governance | allow |
| high | high | CI + security + governance + human review | require_review |
| critical | critical | CI + security + benchmark + governance + human review | require_review |
| unknown | unknown | CI + security + governance + human review | deny |

Semantic-version labels are inputs, not authorization.

## Fail-closed rules

The policy denies the change when:

1. required verification evidence is unavailable;
2. dependency impact cannot be classified;
3. the dependency's architectural domain cannot be determined where classification is required.

## Version semantics

- patch → candidate low risk
- minor → candidate medium risk
- major → candidate high risk
- unknown → unknown

Architectural criticality and verification evidence may raise the resulting risk above the version candidate.

## Durable evidence

A classification record should preserve:

- dependency ecosystem and name;
- version-change class;
- criticality;
- risk;
- affected domains;
- policy ID and policy version;
- schema version;
- required verification gates;
- verification results;
- final decision.

The classification record proves the policy decision. It does not prove that a dependency was successfully executed or deployed.

## Relationship to PR #259

PR #259 only changes Dependabot grouping behavior. This policy is deliberately separate so maintenance automation does not silently become governance authorization.

## Next implementation boundary

The next implementation should:

1. load and validate the machine-readable policy;
2. map repository dependency paths to architectural domains;
3. emit a deterministic classification record;
4. connect risk classes to CI verification gates;
5. test low, high-criticality, and unknown cases;
6. preserve policy/schema versions in durable evidence.
