# NAEOS Social Media Kit v1.0

Public-facing social media source assets for NAEOS.

## Source of truth

- Brand metadata: `brand/brand.json`
- Brand CSS tokens: `brand/brand.css`
- Approved logo geometry: `brand/master/naeos-primary-symbol-master-v2.svg`
- Logo version: Logo Master v2.0.0

Do not alter the approved logo path geometry. Use the canonical brand tokens and typography roles from the files above.

## Asset 001 — AI governance

- Source: `001-ai-governance-square.svg`
- Canvas: 1080 × 1080
- Format: editable SVG
- Theme: AI systems need more than intelligence; they need governance.
- Pillars: Control, Verify, Account

### Suggested LinkedIn caption

AI systems need more than intelligence. They need governance.

As AI systems become increasingly capable, the challenge is no longer just what they can do.

It is whether their behavior can be controlled, verified, and accounted for.

Three foundations matter:

- **Control** — establish boundaries and permissions.
- **Verification** — validate system behavior and outputs.
- **Accountability** — maintain evidence, traceability, and responsibility.

At NAEOS, we believe engineering governance must be part of the architecture, not an afterthought.

**Architecture Drives Engineering.**

#NAEOS #AIGovernance #AIEngineering #EngineeringGovernance #ResponsibleAI

## Production notes

- The SVG is the editable source asset.
- Typography uses Sora, Inter, and JetBrains Mono with system fallbacks.
- Confirm font availability and inspect line wrapping before raster export.
- Export PNG at 1080 × 1080 for LinkedIn.
- Check contrast, safe margins, logo fidelity, and mobile legibility before publishing.

## Scope

This directory contains public brand/marketing assets only. Do not add private customer information, credentials, internal operational records, or confidential business documents.


## Asset 002 — Trust must be verified, not assumed

- Directory: `carousel-002/`
- Canvas: 1080 × 1350 per slide
- Format: five editable SVG slides
- Narrative: Control → Verify → Account → Governance in the architecture
- Intended channel: LinkedIn carousel

### Slide sequence

1. **Trust must be verified, not assumed** — capability is not proof of reliability.
2. **Control** — define what AI is authorized to do; intent is not authorization.
3. **Verify** — test behavior under real conditions and record observed outcomes.
4. **Account** — preserve traceable evidence across request, decision, execution, and observation.
5. **Governance belongs in the architecture** — close with the NAEOS engineering principle.

### Suggested LinkedIn caption

Trust in AI should never be based on capability alone.

A system may produce impressive outputs and still behave unpredictably in production.

Trust requires an engineering discipline:

1. **Control** what the system is allowed to do.
2. **Verify** what it actually does under relevant conditions.
3. **Account** for decisions, execution, and outcomes with traceable evidence.

These are not separate compliance tasks. They are architectural requirements.

At NAEOS, we believe governance must be built into the engineering lifecycle—from authorization to execution and evidence.

**Trust must be verified, not assumed.**

Architecture Drives Engineering.

#NAEOS #AIGovernance #AIEngineering #AgenticAI #EngineeringGovernance #AIAccountability

### Production checks

- Preserve the five-slide order (01–05).
- Export each SVG to PNG at 1080 × 1350 for LinkedIn carousel upload.
- Inspect line wrapping, font availability, contrast, safe margins, and logo fidelity at mobile size.
- Confirm the exported PNGs visually before publication; the repository SVGs are editable source files, not proof of raster-export validation.


## Asset 003 — LinkedIn Portrait Master

- Source: `003-linkedin-portrait-master.svg`
- Canvas: 1080 × 1350 (4:5 portrait)
- Format: editable SVG
- Purpose: reusable portrait composition for engineering insights, product education, and NAEOS announcements
- Demonstration message: “Governance is an engineering requirement.”

### Layout specification

- **Canvas:** 1080 × 1350.
- **Safe area:** keep essential copy and logo within the 92 px horizontal inset; avoid placing critical content against the canvas edge.
- **Header:** category / series label in JetBrains Mono.
- **Brand mark:** canonical Logo Master v2.0.0 geometry; do not redraw or alter the paths.
- **Headline:** Sora, high contrast, short stacked lines; use NAEOS Blue / Electric Cyan only for emphasis.
- **Supporting copy:** Inter, concise, ideally no more than two lines.
- **Insight panel:** up to three concise pillars or proof points.
- **Footer:** “Architecture Drives Engineering.” plus a short series identifier.

### Reuse rules

1. Duplicate the SVG before adapting it for a new post.
2. Replace the category, headline, supporting copy, insight labels, and footer identifier while preserving the layout hierarchy.
3. Keep the palette tied to `brand/brand.json` and `brand/brand.css`.
4. Keep the approved logo geometry unchanged.
5. Check text overflow and mobile legibility after every copy change.
6. Export a PNG at 1080 × 1350 only after visual inspection of the SVG rendering.

The checked-in SVG is an editable source master. Raster export and cross-device visual QA remain separate production steps.


## Asset 004 — Quote & Engineering Insight Master

- Directory: quote-engineering-insight/
- Specification: quote-engineering-insight/MASTER.md
- Editable sample: quote-engineering-insight/quote-trust-is-engineering-outcome.svg
- Canvas: 1080 × 1350 (4:5 portrait)
- Master statement: “Trust is not a feature. It is an engineering outcome.”
- Content variants: Quote, Engineering Insight, Evidence Principle

### Production checks

- Use the canonical brand palette and typography from brand/brand.json and brand/brand.css.
- The sample uses a text-only NAEOS footer placeholder; insert the approved Logo Master v2.0.0 lockup before publication.
- Inspect actual font rendering, line wrapping, contrast, and safe margins.
- Export PNG only after visual review; repository SVGs are editable source assets, not proof of rendered-output QA.


## Asset 004 — Quote & Engineering Insight Master

- Source: `004-quote-engineering-insight-master.svg`
- Canvas: 1080 × 1080
- Format: editable SVG
- Purpose: reusable square quote card for concise engineering principles and NAEOS point-of-view statements.
- Example statement: **“Trust is not a feature. It is an engineering outcome.”**
- Attribution: NAEOS / Engineering Principle
- Logo: canonical Logo Master v2.0.0 path geometry; do not redraw or modify.
- Palette: canonical tokens from `brand/brand.json` and `brand/brand.css` only.
- Typography roles: Sora for the statement, Inter for supporting copy, JetBrains Mono for the asset marker.

### Reusable quote variants

Use the same layout and replace only the quote and supporting copy:

1. **“An AI agent should never authorize its own actions.”**  
   Supporting copy: Separate agent intent from policy authorization and runtime execution.
2. **“No evidence, no defensible assurance.”**  
   Supporting copy: Preserve traceable evidence of decisions, execution, and observed outcomes.
3. **“Trust is not a feature. It is an engineering outcome.”**  
   Supporting copy: Control what is allowed. Verify what happens. Preserve evidence of every meaningful outcome.

### Reuse and production rules

- Keep the 1080 × 1080 canvas, logo clear space, footer baseline, and text safe area.
- Replace quote text without changing canonical color tokens or approved logo geometry.
- Keep statements concise; adjust line breaks rather than compressing type below a legible size.
- Export PNG at 1080 × 1080 for LinkedIn after confirming Sora, Inter, and JetBrains Mono availability.
- Inspect contrast, line wrapping, logo fidelity, and mobile legibility before publishing.
- SVG is the editable source. Raster export and rendered visual QA remain separate production steps.

## Asset 005 — Announcement Master

- Source: `005-announcement-master.svg`
- Canvas: 1080 × 1350 (4:5 portrait)
- Format: editable SVG
- Purpose: reusable announcement layout for verified product milestones, releases, and company updates.
- Demonstration headline: “A new milestone for NAEOS.”
- Replaceable fields: date, milestone/release name, and one concise verifiable detail.
- Logo: canonical Logo Master v2.0.0 path geometry; do not redraw or modify.
- Palette: canonical NAEOS colors only; typography roles use Sora, Inter, and JetBrains Mono.

### Reuse and production rules

1. Duplicate the SVG before adapting it for a specific announcement.
2. Replace all bracketed placeholders with approved, factual information; remove any field that is not applicable.
3. Keep claims measurable and verifiable. Do not imply a release, customer deployment, certification, or milestone that has not occurred.
4. Preserve the logo geometry, 4:5 canvas, safe margins, and footer hierarchy.
5. Inspect font rendering, line wrapping, contrast, and mobile legibility before exporting.
6. Export PNG at 1080 × 1350 only after visual review. The checked-in SVG is an editable source master; raster export and rendered visual QA remain separate production steps.

### Suggested LinkedIn caption

A new milestone for NAEOS.

We are continuing to build governance into the AI engineering lifecycle—connecting authorization, execution, verification, and evidence as engineering concerns.

[Insert the specific milestone and one verifiable detail.]

Architecture Drives Engineering.

#NAEOS #AIGovernance #AIEngineering #EngineeringGovernance #AgenticAI

