# NAEOS Adversarial Governance v1 — Post-Hardening Evidence

Status: H1/H2/H3 implementation evidence recorded; CI harness verification is the authoritative runtime check.

## Purpose

This document records the before/after verification contract for the three architectural findings selected for Adversarial Governance v1.

The evidence is intentionally scoped. It does not claim that NAEOS is secure or that unrelated findings in the broader 17-scenario landscape have been resolved.

## Baseline

The broader deterministic policy-bypass harness contains 17 scenarios. The original baseline identified 12 reproducible enforcement weaknesses under the defined scenarios.

For v1, three architectural findings were promoted for hardening:

| ID | Finding | Baseline scenarios | Hardened target |
|---|---|---|---|
| AGV1-01 | Instruction Layer Is Not an Enforcement Boundary | prompt override dir neutralizes policy | 0 bypass |
| AGV1-02 | Empty Governance Configuration Silently Disables Enforcement | no configured policies => no checks | 0 bypass |
| AGV1-03 | Policy Evaluator Accepts Non-Finite Numeric Values | NaN bypasses gt threshold; Inf bypasses lt bound | 0 bypass |

## Hardening implemented

### AGV1-01 — Instruction integrity

Governance-sensitive built-in compiler templates are protected from repository-writable overrides. A matching override is rejected rather than silently replacing the built-in template.

Regression coverage:
- `TestLoadOverrides_ProtectedCompilerTemplate`
- adversarial scenario: `prompt override dir neutralizes policy`

Scope note: this closes the tested mutable compiler-template override path. It does not establish tamper-evidence for every instruction artifact.

### AGV1-02 — Governance configuration

The pipeline now distinguishes explicit governed execution from intentionally policy-free execution.

In governed mode:
- zero effective policies fail closed;
- the result is not returned as a successful governed execution;
- `governance.unconfigured` records the blocked state.

In intentionally policy-free mode:
- execution is marked `ungoverned`;
- status is `intentionally-disabled`;
- effective policy count is recorded.

Regression coverage:
- `TestPipelineGovernedModeFailsClosedWithoutEffectivePolicies`
- `TestPipelineGovernanceEvidenceDistinguishesPolicyFreeExecution`
- adversarial scenario: `no configured policies => no checks`

### AGV1-03 — Numeric semantics

The policy evaluator rejects non-finite numeric operands before `gt`, `lt`, `gte`, or `lte` comparisons.

Rejected classes:
- `NaN`
- `+Inf`
- `-Inf`

Finite numeric comparisons retain their existing semantics.

Regression coverage:
- `TestEvaluateNumericRulesRejectNonFiniteOperands`
- `TestEvaluateNumericRulesKeepFiniteComparisons`
- adversarial scenarios: `NaN bypasses gt threshold`, `Inf bypasses lt bound`

## Verification protocol

The same deterministic harness is rerun without deleting or weakening baseline scenarios.

Authoritative CI command:

```text
go run ./experiments/policy-bypass
```

The GitHub Actions workflow publishes:

```text
experiments/policy-bypass/reports/EXPERIMENT-REPORT.md
```

as the `policy-bypass-report` artifact.

The complete suite must remain deterministic, and the selected AGV1 scenarios must report `Bypassed=false`.

## Before / after claim boundary

The supported claim is:

> The same deterministic scenarios that reproduced the three selected v1 enforcement weaknesses are now expected to be blocked by explicit regression-tested controls.

The unsupported claim is:

> NAEOS is secure.

The broader 17-scenario landscape remains a separate measurement surface. Remaining findings should be addressed only after the current evidence is independently verified and used for external technical validation.

## Evidence checklist

- [x] Baseline scenarios retained.
- [x] H1 regression test retained.
- [x] H2 governed fail-closed regression test retained.
- [x] H2 explicit ungoverned-state regression test retained.
- [x] H3 non-finite operand regression coverage retained.
- [x] Finite numeric comparison regression coverage retained.
- [ ] Fresh CI run confirms selected AGV1 scenarios are all `Bypassed=false`.
- [ ] Fresh CI artifact is reviewed and linked from external validation material.

## Next action

Do not promote another finding to hardening yet.

First obtain a fresh green CI run and inspect the generated report. Then use the resulting evidence as the technical proof artifact for outreach to engineers evaluating NAEOS.
