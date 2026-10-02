# NAEOS Funding Readiness

**Status:** Preparation baseline
**Date:** 2026-10-02

## Purpose

This document maps NAEOS evidence to open-source funding and startup-support applications. It is an evidence index, not a statement that any program has approved NAEOS.

## Project thesis

NAEOS is an open-source engineering control plane for AI coding agents. It establishes explicit boundaries between engineering intent, policy decisions, authorized execution, observations, evidence, and independent verification.

## Existing evidence

- Architecture: Reference Architecture, Master Technical Specification, and normative specifications.
- Golden Path: `docs/GOLDEN-PATH.md` for a reproducible specification-to-evidence flow.
- External validation: `docs/EXTERNAL-VALIDATION.md` for independent evaluation.
- Handoff security: `docs/PHASE-4.3-AGENT-HANDOFF-CONTRACT.md` for version, capability, provenance, and replay boundaries.
- Evidence: Evidence V5.6 experiments for durable receipts and verification behavior.
- Supply chain: SBOM, artifact signing, Dependabot, CodeQL, Gitleaks, and SHA-pinned Actions.
- Governance: Constitution, governance material, CODEOWNERS, and DCO.
- CI: automated tests, race detection, fuzzing, vulnerability checks, and governance gates.

## Funding tracks

### GitHub Secure Open Source Fund

Fit: security-focused open-source infrastructure.

Emphasize security boundaries, fail-closed authorization, reproducible security experiments, supply-chain controls, external evaluation, and maintainer governance.

Current gap: stronger external adoption/traction and completion of an independent evaluator run.

### GitHub Fund

Fit: early-stage open-source developer infrastructure, AI, security, and developer tooling.

Emphasize the AI coding-agent control-plane thesis, differentiated architecture, technical proof, open-source distribution, early adoption, and a credible commercial path.

Current gap: adoption evidence and a concise investment narrative.

### GitHub Sponsors

The repository now exposes the organization through `.github/FUNDING.yml`. Sponsor tiers must be configured on the GitHub organization profile before public promotion.

### Microsoft for Startups

Relevant to a separate eligible for-profit startup entity. Do not represent an OSS foundation as a for-profit startup unless the legal structure and eligibility are established.

## Use of funds

1. Security hardening and independent review.
2. Reproducible verification infrastructure.
3. AI-agent integrations and handoff interoperability.
4. Documentation and onboarding.
5. External evaluator and pilot support.
6. Community and maintainer operations.

## Readiness gaps

- [ ] Complete independent external evaluator run.
- [ ] Capture evaluator commit SHA, environment, results, and deviations.
- [ ] Establish repeatable adoption metrics: users, contributors, downloads, pilots, and integrations.
- [ ] Remove credential-like hardcoded demo material from security-sensitive examples.
- [ ] Prepare one-page funding brief and application narratives.
- [ ] Confirm legal entity/funding recipient for equity or startup-credit applications.

## Evidence discipline

Funding materials must distinguish repository experiments from production guarantees and cite the exact implementation, test, experiment, or external evaluation supporting each security or reliability claim.
