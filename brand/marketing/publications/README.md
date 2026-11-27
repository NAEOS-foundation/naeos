# NAEOS Publication Family — Brand & Editorial System

**Status:** Draft for review  
**Version:** 0.1  
**Parent brand:** NAEOS — The Intelligent Core  
**Canonical symbol:** Primary Symbol Master v2.0.0  
**Applies to:** NAEOS Engineering Notes, NAEOS Founder Journey, NAEOS Partner Brief

## 1. Family rule

These three publications are a **single NAEOS publication family**, not three sub-brands. They share the NAEOS master logo, core palette references, typography hierarchy, cover/header/footer structure, metadata format, and editorial quality bar. Their distinction is their audience and editorial purpose—not separate logos, mascots, or custom palettes.

## 2. Publication identities

| Name | Purpose | Primary audience | Editorial promise |
|---|---|---|---|
| **NAEOS Engineering Notes** | Explain engineering systems, technical decisions, mechanisms, experiments, and evidence | Engineers, technical evaluators, OSS community | Precise, technically useful, evidence-aware |
| **NAEOS Founder Journey** | Share the real process of building NAEOS: decisions, trade-offs, lessons, setbacks, and progress | Builders, founders, developers, early community | Honest, reflective, transparent; no manufactured success story |
| **NAEOS Partner Brief** | Explain collaboration context, partner value, capabilities, boundaries, and a practical next step | Potential partners, customers, ecosystem collaborators | Clear, credible, business-relevant, and appropriately scoped |

## 3. Brand references

- Canonical SVG: `brand/master/naeos-primary-symbol-master-v2.svg`
- Dark base / Obsidian: `#0B0F19`
- Signal Blue: `#2563FF`
- Ion Blue (source documented): `#60A5FA`
- Core Cyan reference: `#00DFFF` (use only as an accent pending palette ratification)
- Mist: `#E5E7EB`
- Cloud White: `#FFFFFF`

**Status:** color references are documented in source/master context but are not yet a ratified semantic token palette. Use a restrained combination; do not add new hex values or gradients. Validate text contrast before release.

**Typography:** no official typeface is ratified. Use a consistent system sans-serif fallback for working drafts; Inter/Arial/sans-serif may be used as a practical fallback only, not advertised as the official brand font. Keep typography consistent across the publication family.

## 4. Shared cover and document structure

Recommended cover or opening page:
1. NAEOS master symbol (preserve aspect ratio; no redraw)
2. Publication label (Engineering Notes / Founder Journey / Partner Brief)
3. Primary headline
4. Optional one-line subtitle / dek
5. Metadata: issue or document ID, date, author, status, and audience/confidentiality where appropriate

Internal page/document structure:
- One primary title; clear section headings
- Short paragraphs and scannable lists
- Diagram/table only when it improves understanding
- Consistent running header/footer where applicable
- Footer metadata: publication name, version/date, page number when paginated

## 5. Editorial design variations

Use the same visual system with a restrained editorial accent. The accent may be a label, small tag, or rule; it must not become a separate logo or require an unratified color.

| Publication | Suggested label | Content layout |
|---|---|---|
| Engineering Notes | TECHNICAL / ENGINEERING NOTES | Structured sections, architecture or process diagram, evidence/caveat block, takeaway |
| Founder Journey | BUILD IN PUBLIC / FOUNDER JOURNEY | Narrative sections, timeline or decision block, reflection, open question |
| Partner Brief | PARTNERSHIP / PARTNER BRIEF | Executive summary, value/fit, verified capability, scope/boundaries, proposed next step |

## 6. Claim and evidence rules

- Facts: link to a source, repository, issue, PR, test output, or dated evidence where feasible.
- Assumptions: explicitly label as **Assumption** or **Hypothesis**.
- Roadmap or planned work: label **Planned**, **Proposed**, or **Not yet available**.
- Opinions and lessons: identify as reflection, opinion, or lesson learned.
- Do not claim customer adoption, certifications, compliance, performance metrics, production readiness, or general availability unless verifiable and correctly scoped.
- Partner briefs must not disclose confidential information without permission; use a clear confidentiality label when appropriate.

## 7. Reusable metadata

Use this block at the beginning of a draft:

```yaml
publication: NAEOS Engineering Notes | NAEOS Founder Journey | NAEOS Partner Brief
document_id: NAEOS-PUB-YYYY-NNN
title: ""
subtitle: ""
author: Bayu / NAEOS
date: YYYY-MM-DD
version: 0.1
status: Draft
audience: ""
confidentiality: Public | Internal | Confidential
source_links: []
```

## 8. Template and quality gate

Use `DOCUMENT-TEMPLATE.md` as the reusable source template. Before publication, check:

- [ ] Correct publication label and audience
- [ ] Canonical v2.0 logo source; no stretched or rebuilt logo
- [ ] Shared visual hierarchy and palette references; no invented palette/font claims
- [ ] Title and subtitle fit the intended format
- [ ] Metadata is complete and correct
- [ ] Facts have evidence; assumptions and roadmap clearly labeled
- [ ] No unsupported claims or confidential material
- [ ] Layout checked at intended output size, including mobile for social derivatives
- [ ] Draft status is removed only after owner review and approval

## 9. Cover template assets

Three editable SVG cover templates are provided at 1080 × 1350 px (4:5 portrait). All share the same NAEOS symbol placement, Obsidian base, restrained grid, typography hierarchy, blue accent rule, metadata area, and footer. Publication differentiation is expressed by publication name and eyebrow copy only.

- Engineering Notes: `templates/engineering-notes-cover.svg`
- Founder Journey: `templates/founder-journey-cover.svg`
- Partner Brief: `templates/partner-brief-cover.svg`

The templates use a fallback typography stack (`Inter, Arial, sans-serif`; Inter is not ratified as the official font), no new color values, and no logo redraw. The logo asset is referenced using the correct relative path `../../social-media/assets/naeos-primary-symbol-v2.svg` to `brand/marketing/social-media/assets/naeos-primary-symbol-v2.svg`.

### Editing rules

- Replace bracketed title, issue, date, author, and status placeholders.
- Keep the master symbol aspect ratio and relative placement.
- Maintain title/subtitle hierarchy and the shared footer.
- Do not introduce publication-specific logos, palette, gradients, or ornamental effects.
- Review text fit at actual 4:5 output size before release.
- These SVGs are editable source templates; visual rendering/export QA remains required.

## 10. Governance

This specification is a draft extension of Brand Guidelines v1.1. It does not ratify typography, semantic color tokens, clear-space/minimum-size, contrast pairings, or platform export dimensions. When those decisions are approved, update this publication spec and the master brand guideline together.

## 11. Repository-grounded editorial integration

Publication content must reflect the repository's current documented and implemented state, not merely the brand narrative. The repository snapshot and starter drafts below are editorial starting points, not approval to publish.

### Snapshot

- [Repository snapshot — 2026-11-23](REPOSITORY-SNAPSHOT-2026-11-23.md) records the reviewed main commit, proof surfaces, release reference, claim boundaries, and source links.
- Update or replace this snapshot before publication if the source repository has materially changed.

### Draft starter publications

- [Engineering Notes #002 — Run Traceability Is Not Independent Evidence Verification](engineering-notes/NAEOS-PUB-2026-002-run-traceability-vs-independent-verification.md)
- [Founder Journey #001 — Building a Proof Before Making a Claim](founder-journey/NAEOS-PUB-2026-FJ-001-proof-before-claim.md)
- [Partner Brief #001 — Technical Evaluation Invitation](partner-brief/NAEOS-PUB-2026-PB-001-technical-evaluation-invitation.md)

All starter documents are **Draft**. Engineering Notes and Partner Brief require source/technical review before publishing. Founder Journey includes explicit placeholders for real founder experience and must not be published without Bayu's personal review and completion.

## 12. Related documents

- `brand/guidelines/BRAND-GUIDELINES-v1.1.md`
- `brand/marketing/social-media/README.md`
- `brand/marketing/social-media/ASSET-REGISTER.md`
- `DOCUMENTATION-AUTHORITY.md` — source authority and claim discipline
- `docs/READINESS-CONTRACT.md` — repository readiness gates

