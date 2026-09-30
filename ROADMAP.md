# NAEOS Roadmap

> **Operational roadmap — September 2026**
>
> This file is the public execution navigator. Normative roadmap authority remains
> NAEOS-GOV-007, while the executable development plan is NAEOS-RDP-001.

## Current baseline

- **Current documented release:** NAEOS 3.6.0 (VERSION)
- **Core direction:** Engineering Control Plane for AI Coding Agents
- **Canonical model:** Specification → NEIR → Validation + Policy → Agent Intent → Authorized Execution → Observation → Evidence → Independent Verification
- **Repository proof path:** Golden Path → Reference Demo → External Validation
- **Current engineering milestone:** **P1.9 — Evidence & Verification Plane**
- **Primary public surface:** /control-plane on the NAEOS website

The project has already implemented substantial policy, control-plane, runtime,
evidence, verification, observability, signing, SBOM, plugin, and production
server capabilities. The next work should therefore prioritize **proof,
reproducibility, adoption, and real-world evaluation** over adding unrelated
surface area.

## Current execution track

### P1.7 — Policy Change Mid-Run / Stale Authorization

**Status: DONE**

Goal: prove that an authorization issued under one active policy version cannot
be reused after the active policy advances.

Evidence:
- `docs/control-plane/p1-7-policy-change-mid-run.md`
- `docs/experiments/EXP-001-policy-change-mid-run.md`
- `examples/control-plane-policy-change/`

Acceptance:
- [x] Authorization issued under policy v1
- [x] Policy advances to v2
- [x] Execution-boundary freshness check rejects stale authorization
- [x] Execution is blocked
- [x] No side effect occurs
- [x] Evidence is retained
- [x] Independent verification passes

### P1.8 — Atomic Execution Commit Boundary

**Status: DONE**

Goal: close the in-process check-to-side-effect race by making policy freshness
validation and the authorized side-effect callback share one execution boundary.

Evidence:
- `docs/control-plane/p1-8-atomic-execution.md`
- `docs/experiments/EXP-002-atomic-execution-commit-boundary.md`
- `examples/control-plane-atomic-execution/`

Acceptance:
- [x] Authorization validated against active policy
- [x] Atomic execution boundary acquired
- [x] Authorized side effect executes inside the boundary
- [x] Concurrent policy update cannot commit mid-execution
- [x] Execution evidence is recorded
- [x] Independent verification passes

### P1.9 — Evidence & Verification Plane

**Status: IMPLEMENTED**

Goal: make one agent action independently reconstructable through a canonical
verifier-facing evidence contract.

Evidence:
- `docs/control-plane/p1-9-evidence-verification.md`
- `internal/controlplane/evidence.go`
- `internal/controlplane/evidence_test.go`

Acceptance:
- [x] Reconstruct one action lifecycle
- [x] Bind policy/version
- [x] Bind grant
- [x] Bind artifact
- [x] Bind execution
- [x] Independent verification
- [x] Tamper detection

## Track 1 — Five-minute developer onboarding

**Status: ACTIVE**

The repository already has a canonical five-minute CLI Golden Path. The next
step is to make the public control-plane proof and the local Golden Path feel
like one coherent onboarding journey.

- [x] Canonical CLI demo
- [x] Golden Path acceptance contract
- [x] Reference Demo evidence story
- [x] External Validation runbook
- [ ] Link public Control Plane → Golden Path → Reference Demo
- [ ] Verify fresh-checkout onboarding on the current release
- [ ] Remove or fix broken onboarding links
- [ ] Add one concise "what you just proved" explanation

References:
- docs/GOLDEN-PATH.md
- docs/REFERENCE-DEMO.md
- docs/EXTERNAL-VALIDATION.md
- START-HERE.md

## Track 2 — Documentation Truth Sync

**Status: ACTIVE**

Public documentation must agree with the current repository state. The
normative hierarchy remains authoritative; public navigators should point to it
rather than reproduce competing roadmaps.

- [x] README states current release and control-plane positioning
- [x] START-HERE defines the supported first-run path
- [x] Golden Path defines reproducible acceptance criteria
- [x] Reference Demo defines the evidence narrative
- [x] External Validation defines independent evaluation
- [x] Whitepapers identify repository version 3.6.0
- [x] Top-level roadmap aligned with the current execution milestone
- [ ] Audit public website copy against the canonical positioning
- [ ] Audit roadmap/version references for stale phase language
- [ ] Ensure release notes, website, README, and whitepaper distinguish
      software releases from experiment milestones

## Track 3 — Adoption Engineering

**Status: NEXT**

Goal: turn technical proof into repeatable developer adoption.

~~~text
Public proof
    ↓
Developer runs NAEOS
    ↓
Developer understands the control boundary
    ↓
Developer challenges / contributes
    ↓
Developer becomes evaluator
~~~

Targets:

- [ ] Clear Developer / Contributor / Organization entry points
- [ ] One copy-paste public demo path
- [ ] Issue/discussion template for external validation reports
- [ ] First cohort of technical evaluators
- [ ] Capture reproducible deviations and failure modes
- [ ] Convert repeated evaluator needs into engineering work

Success should be measured by **reproducible usage and technical feedback**, not
only traffic, followers, or impressions.

## Track 4 — Design Partner Pilot

**Status: AFTER PUBLIC PROOF**

Start with a narrow boundary rather than a full enterprise deployment:

~~~text
1 team
  ↓
1 repository
  ↓
1 AI-agent workflow
  ↓
1 consequential control boundary
  ↓
policy → authorization → execution → evidence → verification
~~~

Pilot questions:

1. Can the team define a meaningful policy boundary?
2. Can the agent request an action without becoming the authority?
3. Can NAEOS deterministically allow or deny the action?
4. Can the runtime enforce the decision?
5. Can the team inspect evidence afterward?
6. Can an independent verifier validate the resulting claim?

Do not expand the pilot scope until these questions produce concrete evidence.

## Track 5 — Ecosystem / P2

**Status: GATED BY ADOPTION EVIDENCE**

Potential areas:

- agent adapters and protocol-neutral handoffs
- plugin/extension ecosystem
- durable evidence and verification integrations
- organization-level governance
- operational dashboards
- compliance workflows
- enterprise deployment patterns

The order should be driven by observed evaluator and pilot requirements, not by
feature volume.

## Strategic sequence

~~~text
P1.7 Policy Freshness
        ↓
P1.8 Atomic Execution
        ↓
P1.9 Evidence & Verification
        ↓
P1.6 Public Control Plane
        ↓
Five-minute onboarding
        ↓
Documentation truth sync
        ↓
Technical evaluators
        ↓
Design partner pilot
        ↓
Evidence-backed P2 priorities
~~~

## What we deliberately do not optimize for yet

- A large dashboard before real users require it
- Provider-specific AI features that weaken vendor neutrality
- Generic AI marketing claims without repository evidence
- Enterprise packaging before a repeatable technical pilot
- Feature expansion that makes the core control boundary harder to explain

## Source-of-truth rules

When documents disagree:

1. Follow the normative hierarchy in DOCUMENTATION-AUTHORITY.md.
2. Treat NAEOS-GOV-007 as the stable strategic roadmap.
3. Treat NAEOS-RDP-001 as the executable development-plan reference.
4. Treat this file as the current public execution navigator.
5. Treat CHANGELOG.md and VERSION as the release-history sources.

**Roadmap principle:** prove the control boundary, make it reproducible, then
earn adoption before scaling the ecosystem.
