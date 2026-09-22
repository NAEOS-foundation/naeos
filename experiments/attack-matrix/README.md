# NAEOS Attack Matrix v1

## Purpose

The Attack Matrix v1 maps adversarial scenarios to the NAEOS enforcement boundary they exercise.

It is a coverage artifact, not a vulnerability count and not a production penetration-test result. Each row must point to an existing deterministic experiment or an explicitly tracked gap.

## Boundary model

`INTENT → POLICY → GATEWAY → SIDE EFFECT → OBSERVATION → EVIDENCE → VERIFICATION`

Cross-cutting controls:

- instruction integrity;
- governance configuration;
- policy-context integrity;
- evaluator semantics;
- handoff authority;
- replay/provenance/integrity;
- evidence integrity.

## Coverage rule

A matrix row is **covered** only when a deterministic scenario exists and has a documented expected invariant. A row marked **gap** requires a future experiment or implementation work; it must not be treated as a demonstrated weakness.

## Current v1 scope

The matrix consolidates the existing NAEOS experiments:

- AI Agent Policy Boundary;
- Governance Lifecycle;
- Level-3 Evidence;
- Policy Bypass / Adversarial Governance;
- Agent Handoff Governance.

The matrix deliberately reuses those experiments rather than creating a parallel security harness.

## Execution

Run the deterministic experiment suites from the repository root:

```bash
go test ./...
go run ./experiments/ai-agent-policy-boundary
go run ./experiments/governance-lifecycle
go run ./experiments/level3-evidence
go run ./experiments/policy-bypass
go run ./experiments/handoff-governance
```

The exact CI workflow remains the authoritative regression environment.

## Claim discipline

Results should be reported as behavior reproduced by the named deterministic harness. A covered row does not establish that every production deployment is protected against the corresponding class of attack.
