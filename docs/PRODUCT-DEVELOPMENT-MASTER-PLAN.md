# NAEOS Product Development Master Plan

Status: Active  
Scope: P1.12 → P4  
Product thesis: NAEOS is the engineering control plane for AI coding agents.

## 1. Product North Star

Every product decision should strengthen one observable property:

> AI-assisted engineering can be controlled against explicit intent, policy, authorization, execution, evidence, and independent verification.

The primary product proof is not feature count. It is the time and reliability with which an independent engineer can reproduce a controlled workflow and verify its evidence.

## 2. Current Product State

P1 control-plane work has established:

- a canonical CLI Golden Path;
- specification → NEIR → validation → policy → AI context → generation → evidence;
- a Reference Demo and External Validation Runbook;
- an independent, read-only EvidenceBundle verifier;
- a public Control Plane entry point;
- CI coverage for the canonical demo.

The next product constraint is adoption engineering: make the proof easy for an external engineer to reproduce, challenge, report, and extend.

## 3. Phase Roadmap

### P1 — Product Proof

**Objective:** Make the existing control-plane proof stable, legible, and regression-protected.

Required outcomes:

- Golden Path remains the single first-run path.
- Public Control Plane links into the same proof story.
- Reference Demo and verifier remain aligned.
- Evidence boundaries are explicit.
- CI failures do not silently block unrelated external evaluation.
- Security-sensitive dependency changes remain fail-closed until reviewed.

Acceptance criteria:

1. Fresh checkout can execute the documented Golden Path.
2. Required evidence files are produced.
3. EvidenceBundle can be independently verified.
4. Documentation contains no contradictory first-run paths.
5. CI health is sufficient to distinguish product failures from infrastructure/governance failures.

### P2 — External Developer Validation

**Objective:** Test NAEOS with engineers who did not implement the current proof.

Initial cohort:

- 5–10 technical evaluators;
- 5–10 repositories or controlled sample repositories;
- at least one real AI-agent workflow;
- one consequential control boundary per pilot.

Each evaluation records:

- commit SHA;
- environment/toolchain;
- workflow;
- control boundary;
- run ID;
- specification and NEIR hashes;
- policy decision;
- evidence references;
- deviations;
- follow-up questions.

Success is based on observed reproducibility and actionable feedback, not testimonials.

### P3 — Integration & Ecosystem

**Objective:** Turn the validated core into an engineering integration surface.

Priority order:

1. Git/GitHub integration.
2. CI/CD integration.
3. AI coding-agent adapters.
4. CLI and stable API contracts.
5. IDE/runtime integrations.
6. Plugin SDK.
7. Plugin registry.

The registry should follow demonstrated extension demand rather than precede it.

### P4 — Commercialization

**Objective:** Package proven control-plane capabilities for organizations with team and governance requirements.

Potential product surfaces:

- team policy administration;
- centralized authorization;
- audit and evidence retention;
- deployment/runtime governance;
- organization-level analytics;
- enterprise integration;
- support and operational controls.

Commercial features must preserve the open core's observable proof model.

## 4. Product Metrics

### Activation

- time to first successful Golden Path run;
- time to first verified EvidenceBundle;
- first-run completion rate.

### Reproducibility

- successful external evaluations / total evaluations;
- deviation rate;
- time to diagnose a deviation;
- percentage of evaluation reports containing complete evidence anchors.

### Engineering reliability

- canonical-demo CI pass rate;
- policy decision test coverage;
- verifier success/failure classification;
- dependency/change-risk gate health.

### Adoption

- active external repositories;
- active evaluators;
- repeated evaluation runs;
- external contributors;
- integrations in active use.

Do not use GitHub stars as a primary product KPI.

## 5. Product Gates

### Gate A — P1 complete

Do not expand feature scope until:

- Golden Path is reproducible;
- independent verification is documented;
- onboarding is coherent;
- CI health is understood.

### Gate B — P2 validated

Do not scale ecosystem investment until external evaluators demonstrate:

- reproducible execution;
- understandable evidence;
- actionable deviation reporting;
- at least one real workflow/control-boundary use case.

### Gate C — P3 validated

Do not broaden commercialization until integrations demonstrate repeated use and stable contracts.

## 6. Flagship Demonstration

The flagship product story is:

```
Agent receives engineering task
        ↓
Intent is represented
        ↓
Validation runs
        ↓
Policy authorizes or rejects
        ↓
Authorized execution occurs
        ↓
Execution produces evidence
        ↓
Independent verifier checks the evidence
```

A particularly important evaluation scenario is a policy/authorization change during an active workflow. The expected product behavior must be explicit and evidence-backed rather than inferred from an agent claim.

## 7. Non-Goals for the Current Cycle

Do not prioritize:

- broad feature expansion without an external use case;
- a large plugin marketplace before extension demand exists;
- complex dashboards before evidence workflows require them;
- enterprise packaging before external validation;
- vanity metrics;
- multiple competing first-run demos.

## 8. Execution Backlog

### Now

- [ ] Close P1.12 onboarding gaps.
- [ ] Resolve CI dependency-risk handling for non-Go manifests.
- [ ] Keep Dependabot/secret-boundary behavior explicit.
- [ ] Validate fresh-checkout onboarding.
- [ ] Run the first external evaluator cohort.
- [ ] Capture deviations as GitHub issues.

### Next

- [ ] Select one pilot repository/workflow/control boundary.
- [ ] Define stable evaluator intake.
- [ ] Establish a repeatable evidence-report format.
- [ ] Stabilize CLI/API surfaces used by external evaluators.

### Later

- [ ] Build first agent adapter.
- [ ] Build CI integration.
- [ ] Define plugin SDK contract.
- [ ] Validate extension demand.
- [ ] Introduce registry only after extension usage is demonstrated.

## 9. Decision Rule

When prioritizing a new feature, ask:

1. Does it strengthen control, evidence, or verification?
2. Does an external engineer need it to reproduce the proof?
3. Is there observed user demand?
4. Can its acceptance criteria be tested?
5. Does it introduce a new competing workflow?

If the feature cannot answer these questions clearly, defer it.

**Product principle:** prove the control plane, validate it externally, then scale the ecosystem.

