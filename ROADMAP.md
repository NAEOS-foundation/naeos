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
- **Current engineering milestone:** **P1.6 — Public Control-Plane Golden Path**
- **Primary public surface:** /control-plane on the NAEOS website

The project has already implemented substantial policy, control-plane, runtime,
evidence, verification, observability, signing, SBOM, plugin, and production
server capabilities. The next work should therefore prioritize **proof,
reproducibility, adoption, and real-world evaluation** over adding unrelated
surface area.

## Current execution track

### P1.6 — Public Control Plane

**Status: IN PROGRESS**

Goal: make the control-plane contract observable from a public browser surface
without implying that a browser request executes a production side effect.

Target proof:

~~~text
Request
  ↓
Policy / Grant Evaluation
  ↓
ALLOW / DENY / REQUIRE_APPROVAL
  ↓
Decision Record / Ledger
  ↓
Inspectable Evidence
~~~

Acceptance:

- [x] Real control-plane decision endpoint
- [x] Public website Control Plane page
- [x] Explicit CORS allowlisting
- [x] Rate and request-size limits
- [x] Optional server-to-server authentication
- [x] Side-effect-free public decision path
- [x] Production deployment
- [x] Website CSP permits the public control-plane endpoint
- [ ] Browser-level ALLOW proof verified against production
- [ ] Browser-level DENY proof verified against production
- [ ] Public evidence walkthrough captured
- [ ] P1.6 documentation and screenshots synchronized

Reference: docs/control-plane/live-proof.md

### Track 1 — Five-minute developer onboarding

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

### Track 2 — Documentation Truth Sync

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

### Track 3 — Adoption Engineering

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

### Track 4 — Design Partner Pilot

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

### Track 5 — Ecosystem / P2

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
P1.6 Public Control Plane
        ↓
Production browser proof
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
