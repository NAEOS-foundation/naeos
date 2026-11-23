# NAEOS Partner Brief — Technical Evaluation Invitation (Draft)

**Publication:** NAEOS Partner Brief  
**Document ID:** NAEOS-PUB-2026-PB-001  
**Date:** 2026-11-23  
**Version:** 0.1  
**Status:** Draft — partner-facing review required  
**Audience:** Engineering leaders, platform teams, AI engineering teams, potential design partners  
**Confidentiality:** Public draft; remove or add confidentiality labels as appropriate

## Purpose

Invite a technical team to evaluate NAEOS against one bounded AI-assisted engineering workflow using the repository's reproducible proof surfaces.

## What NAEOS is

NAEOS is an open-source engineering control plane for AI coding agents. The repository describes a workflow that connects declared engineering intent, validation, policy, authorized generation/execution, observation, evidence, and independent verification.

NAEOS is complementary to AI coding agents, source control, CI/CD, identity systems, cloud platforms, and observability tools; it is not positioned as a replacement for those systems.

## Evaluation route

1. Select one narrow agent-assisted engineering workflow.
2. Record the repository commit, environment, and scenario.
3. Run the canonical Golden Path.
4. Inspect the specification, derived NEIR, validation and policy outputs, context, run metadata, and generated artifacts.
5. Use the documented independent verifier for canonical EvidenceBundle inputs when the evaluation scope includes that evidence contract.
6. Record observations, deviations, and next actions.

## Repository-backed evaluation assets

| Asset | Evaluation use |
|---|---|
| Public Control Plane | Inspect authorization decisions in the public demo; side-effect-free boundary |
| Golden Path | Reproduce the local specification-to-artifact flow |
| Reference Demo | Reviewer checklist and evidence narrative |
| Verified Golden Path | Connect the public and local paths with verification and experiments |
| Governance Lifecycle Experiment | Test lifecycle distinctions and evidence behavior in deterministic scenarios |
| Handoff Governance Experiment | Explore documented handoff authority, capability, and replay checks |
| External Validation Runbook | Commit-pinned independent evaluation record |
| Pilot Readiness | Structure the scope, evidence, observations, and next steps for a narrow pilot |

## Proposed scope

**Initial scope:** one workflow, one repository, one agreed evaluation commit, and a defined set of expected outcomes.

**Out of scope unless separately agreed:** production deployment, broad organization-wide rollout, formal certification/compliance attestation, performance benchmarking, penetration testing, and assurance of every external agent or tool integration.

## Evaluation questions

- Can the team reproduce the documented workflow without maintainer intervention?
- Are authorization and policy outcomes understandable?
- Is run traceability sufficient for the intended workflow?
- Which evidence is useful, and what remains missing?
- What integration boundary should be evaluated next?
- What should be changed before a second evaluation?

## Success record

Use the repository's external validation or pilot record template. A successful reproduction is scoped to the commit and scenario evaluated; it is not a general production-readiness claim.

## Proposed next step

Agree on the workflow, evaluation commit, evaluator, environment, success criteria, and feedback format before starting. No partnership, deployment, or customer outcome is implied by this brief.

## Sources

- `docs/VERIFIED-GOLDEN-PATH.md`
- `docs/GOLDEN-PATH.md`
- `docs/REFERENCE-DEMO.md`
- `docs/EXTERNAL-VALIDATION.md`
- `docs/PILOT-READINESS.md`
- `docs/NAEOS-ENTERPRISE-OVERVIEW.md`
