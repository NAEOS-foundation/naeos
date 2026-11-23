# NAEOS Brand System v2 — Stabilization Plan

**Status:** Proposed / implementation baseline  
**Owner:** NAEOS  
**Scope:** Brand governance, canonical specification, token reconciliation, asset lifecycle, social templates, product design-system integration, validation automation.  
**Non-goal:** Redesign or mutate the approved Logo Master v2.0.0.

## 1. Current repository baseline

Evidence reviewed:
- `brand/brand.json` — metadata status currently `FROZEN`, logo master `2.0.0`, colors and typography values.
- `brand/brand.css` — currently hand-maintained token declarations.
- `brand/docs/README.md` and `brand/source-of-truth/README.md` — approved/frozen Logo Master v2.0.0 and preservation rules.
- `brand/guidelines/BRAND-GUIDELINES-v1.1.md` — Draft and explicit open ratification decisions.
- `brand/ASSET-MANIFEST-v1.md` — mixed verified and pending assets.
- `brand/marketing/social-media/README.md` and `ASSET-REGISTER.md` — social templates exist, Draft; raster exports/previews are pending.
- `docs/SOCIAL-IDENTITY.md` — positioning/channel governance and distribution loop.

## 2. Findings and risks

### P0 — Status and authority are ambiguous
The global `brand.json` status `FROZEN` can be interpreted as the whole brand system being frozen, while guidelines, social templates, and export assets remain Draft or incomplete.

**Decision:** freeze only the approved logo geometry/source. Track maturity independently for foundations, tokens, guidelines, templates, product implementation, and asset releases.

### P0 — Multiple sources can drift
Color and typography values are repeated across JSON, CSS, README, and templates. The guideline explicitly says some items remain unratified.

**Decision:** designate the forthcoming Brand System Spec as policy authority; treat generated token files as outputs. Reconcile existing sources after owner ratification, not by silently changing visual values.

### P1 — Binary assets and release evidence incomplete
PNG exports, avatar sizes, template previews/contact sheet, and checksum records are not fully repository-published/verified.

**Decision:** release asset packages only when files are present, validated, and registered with dimensions, source master, intended use, version, and SHA-256.

### P1 — Social template coverage is mostly generic
Existing layouts primarily use the same navy/grid/glow treatment. More content-specific patterns are needed for architecture, engineering implementation, evidence, releases, case studies, founder notes, community, and partner calls.

### P1 — Product design system lacks evidence of integration
Brand principles and tokens are not yet demonstrated as a single validated system shared by website, CRM/dashboard, and control-plane product surfaces.

### P2 — No brand CI governance
No verified automated checks currently enforce token consistency, asset dimensions, SVG validity, placeholders, contrast, provenance, or deprecation status.

## 3. Proposed maturity/status model

| Subsystem | Current status | Next gate |
|---|---|---|
| Logo geometry/master v2.0.0 | Approved / frozen | Preserve integrity |
| Positioning/social identity | Baseline documented | Public profile synchronization |
| Brand color tokens | Draft / reconcile | Ratification |
| Typography | Documented in metadata; guideline pending ratification | Ratification + font/licensing validation |
| Brand guidelines v1.1 | Draft | Review and approval |
| Social Media Kit v1.0 | Draft | Visual/technical validation |
| Binary asset release | Partial/pending | Generate, verify, publish |
| Product design system | Not evidenced as unified | Inventory + integration plan |
| Brand CI | Not evidenced | Validation pipeline |

## 4. Development workstreams

### Phase 1 — Stabilize and reconcile (now)
1. Add per-subsystem status rather than a misleading global freeze.
2. Establish spec authority and decision log.
3. Reconcile brand metadata, guideline, CSS, social kit, and identity source.
4. Record all open ratification questions with evidence.
5. Keep logo master immutable.

### Phase 2 — Foundations
1. Confirm semantic color roles and contrast pairings from current source and actual usage.
2. Ratify typography roles and validate font availability, licensing, and fallbacks.
3. Specify clear space, minimum size, grid, spacing, responsive breakpoints, light/dark use.
4. Publish accessibility/contrast rules.

### Phase 3 — Asset production and release
1. Export derivatives from canonical approved source using reproducible tooling.
2. Produce SVG, PNG, favicon, app icon, avatars, previews, and contact sheets.
3. Validate dimensions, transparency, aspect ratio, SVG parsing, and rendering.
4. Generate manifest and SHA-256.
5. Publish a versioned release package; no ZIP-only source of truth.

### Phase 4 — Social template family v2
Add function-specific layouts for Architecture, Engineering/Build, Evidence/Verification, Release, Case Study, Founder Journal, Community, and Partner/Recruitment. Keep templates editable, safe-area tested, and claims tied to evidence.

### Phase 5 — Product design system
Inventory actual website and CRM/dashboard/control-plane implementations. Build shared foundations and component patterns for navigation, cards, tables, status, approvals, authorization, evidence, audit trail, and lifecycle states. Keep marketing effects separate from product tokens.

### Phase 6 — CI validation
Introduce checks for token consistency, SVG validity, required asset manifest fields, dimensions, placeholders, contrast, source/master provenance, and obsolete assets. Fail closed for release validation when provenance or required metadata is missing.

## 5. Acceptance criteria

- Logo Master v2.0.0 and original source remain unchanged.
- Every subsystem has explicit maturity status and owner.
- Canonical spec and decision log identify authority and unresolved approvals.
- Tokens are reconciled; generated files are distinguishable from source.
- No production-ready claim for unverified assets.
- Released assets have source/version, intended use, dimensions, and checksum.
- Social templates include proof-oriented content families and platform safe areas.
- Product design-system integration is based on implementation inventory.
- CI checks validate release assets and fail on missing provenance or placeholders.
- Public brand claims remain evidence-backed.

## 6. Risks and controls

| Risk | Control |
|---|---|
| Silent palette change | Explicit decision record and visual comparison before token ratification |
| Logo geometry drift | Immutable master, source checksum, derived-asset validation |
| Documentation status conflict | Per-layer status registry |
| Incorrect public claims | Content evidence requirement and editorial review |
| Broken template rendering | SVG validation and raster preview checks |
| Product/marketing token coupling | Separate semantic token namespaces |
| Unverified binary publication | Manifest + checksum + presence verification |

## 7. Immediate implementation scope

This PR establishes governance artifacts only. It does not ratify new colors/fonts, modify logo geometry, publish missing binary exports, or claim CI validation exists. Those require separate implementation and verification steps.
