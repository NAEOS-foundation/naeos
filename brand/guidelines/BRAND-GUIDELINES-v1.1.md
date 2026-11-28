# NAEOS Brand Guidelines v1.1

**Status:** Draft for review  
**Version:** 1.1  
**Applies to:** NAEOS corporate, product, engineering, and social communication  
**Canonical logo source:** `brand/master/naeos-primary-symbol-master-v2.svg`

---

## 1. Brand foundation

NAEOS communicates AI Engineering Governance & Verification. Brand expression should convey engineering discipline, trust, clarity, technical depth, and operational accountability.

Use these principles in every brand touchpoint:

- **Precise:** use specific language, explicit evidence, and clear hierarchy.
- **Trustworthy:** avoid exaggerated claims, unsupported metrics, and decorative complexity that weakens legibility.
- **Technical:** use structured layouts, restrained visual systems, and purposeful diagrams.
- **Accountable:** distinguish verified product capabilities from roadmap items and proposals.

The positioning above is the working communication direction for this guideline. It is not a claim that every capability is already generally available.

## 2. Logo governance

### 2.1 Canonical source

Use the NAEOS Primary Symbol Master v2.0.0 as the canonical logo source:

`brand/master/naeos-primary-symbol-master-v2.svg`

The distributed Logo Exports v1.0 package identifies its SVG and PNG files as derivatives of this master. The original 260:165 aspect ratio is preserved. Do not edit the canonical master, alter mark geometry, distort proportions, or change approved gradients.

### 2.2 Approved derivative exports

The inspected Logo Exports v1.0 package contains:

- SVG source copy: `naeos-primary-symbol-master-v2.svg`
- Transparent PNG: 512 px wide
- Transparent PNG: 1024 px wide
- Transparent PNG: 2048 px wide

Use SVG for scalable digital placements where supported. Use PNG exports where raster artwork is required. Preserve aspect ratio in all placements.

### 2.3 Social avatar

The Social Avatar v1.0 package is derived from a Primary Symbol Master v2.0 PNG export. It provides square PNG sizes of 64, 180, 320, 512, and 1024 px. Its documented treatment is Deep Navy `#0B0F19`, a restrained blue glow, and a subtle cyan-blue circular keyline.

Use these files for social profile avatars and compact brand placements. Do not crop or rebuild the mark from screenshots.

### 2.4 Historical variants

The separate Logo System v1.0 package documents variants derived from Candidate v1.0 geometry and Brand Identity v3.1. Treat that package as historical/reference material until each variant is checked against the v2.0 canonical master. Do not silently substitute its favicon or app-icon files for current production assets.

### 2.5 Incorrect logo handling

Do not:

- stretch, skew, rotate, or non-uniformly scale the mark;
- redraw, trace, or modify its vector paths;
- apply unapproved colors, gradients, outlines, shadows, or effects;
- place the mark on backgrounds that make its silhouette unclear;
- use a low-resolution export when a vector or larger raster export is available.

## 3. Color system

The inspected source packages explicitly document these colors in the older Logo System v1.0:

| Token | Hex | Source status |
|---|---|---|
| Obsidian | `#0B0F19` | Documented in Logo System v1.0 |
| Signal Blue | `#2563FF` | Documented in Logo System v1.0 |
| Ion Blue | `#60A5FA` | Documented in Logo System v1.0 |
| Mist | `#E5E7EB` | Documented in Logo System v1.0 |
| Cloud White | `#FFFFFF` | Documented in Logo System v1.0 |
| Deep Navy | `#0B0F19` | Documented in Social Avatar v1.0 |

**Governance note:** these values are source-documented references, not yet a fully ratified semantic token system for the current brand. The exact current master SVG gradient stops and contrast pairings must be checked before declaring a final palette. Do not invent additional brand hex values in production assets.

## 4. Typography

**Pending approval.** The inspected logo/avatar packages do not specify a typeface, font weights, type scale, or fallback stack. Until an approved typography specification is added:

- use the typography already approved in the relevant product or document system;
- do not claim a particular font as the official NAEOS typeface;
- keep headings, body copy, labels, and code visually distinct and consistently applied;
- prioritize legibility across desktop and mobile sizes.

## 5. Layout and composition

- Preserve clear separation between the logo and nearby text or UI controls.
- Use consistent alignment, spacing, and a clear content hierarchy.
- Keep high-density technical information structured into short sections, tables, diagrams, or evidence blocks.
- Avoid visual noise, excessive glow, ornamental gradients, and unnecessary decoration.
- Ensure layouts remain legible at mobile widths.

**Pending specification:** numerical clear-space, minimum-size, grid, spacing, and breakpoint values have not been established by the inspected source packages. Do not invent mandatory measurements; define and validate them in a subsequent design-system specification.

## 6. Social and marketing expression

For social posts, company announcements, and educational content:

- lead with one clear engineering or governance problem;
- use a concise headline and a single primary message per frame;
- support claims with a concrete explanation, workflow, example, or evidence;
- use the current canonical logo/export and approved avatar assets;
- maintain a restrained technical aesthetic consistent with the documented navy/blue direction;
- avoid implying customer adoption, certification, compliance, or measured outcomes unless verified.

Recommended formats to produce as separate templates: 1:1 post, 4:5 post, carousel, quote/insight, announcement, and LinkedIn company banner. Exact platform dimensions should be checked at production time.

## 6.1 NAEOS Publication Family — Naros

The NAEOS publication family consists of three editorial formats. They are **one NAEOS publication family, not separate sub-brands**. Each publication uses the canonical Primary Symbol Master v2.0.0, the shared brand palette references, and one consistent cover/header/footer system. Distinction comes from editorial purpose and labels, not separate logos or invented colors.

| Publication | Editorial role | Content pattern |
|---|---|---|
| **NAEOS Engineering Notes** | Technical education and engineering evidence | Problem → architecture/mechanism → implementation/evidence → takeaway |
| **NAEOS Founder Journey** | Transparent founder/building-in-public narrative | Context → decision/experiment → outcome or current status → lesson/open question |
| **NAEOS Partner Brief** | Partner-facing opportunity, collaboration, and alignment | Context → mutual value → verified capability/offer → next step |

### Publication visual rules

- Use the canonical logo: `brand/master/naeos-primary-symbol-master-v2.svg`. Never redraw or use historical logo variants.
- Use Obsidian `#0B0F19` as the dark foundation; Signal Blue `#2563FF` and documented blue/cyan references as restrained accents. Color status remains draft/reference until the semantic palette is ratified.
- Cloud White `#FFFFFF` or Mist `#E5E7EB` may be used for text/surface contrast, subject to contrast validation.
- Use the same typography hierarchy and grid across all three publication types. Typography is not yet officially ratified; templates use a practical fallback only.
- Cover/header structure: logo → publication label → title → optional subtitle → issue/date/author metadata. Keep one dominant headline and avoid decorative clutter.
- Use an editorial accent label or thin accent rule to identify the publication type; do not create a new symbol, logo, or standalone color palette for each.
- Inside documents, maintain clear section hierarchy, readable body width, short paragraphs, diagrams/tables where they add information, and consistent footer metadata.
- Claims must be evidence-backed. Clearly label assumptions, opinions, roadmap items, and experiments. Do not imply adoption, certification, compliance, customer results, or general availability without verification.
- Keep editable source documents. Include version/date/status/author and confidentiality label when relevant.

The working specification and reusable markdown document template are maintained at `brand/marketing/publications/README.md` and `brand/marketing/publications/DOCUMENT-TEMPLATE.md`. Treat both as Draft until visual review and the open brand decisions are completed.

## 7. Product UI expression

The CRM and dashboard should apply the same principles of precision, clarity, and accountability:

- distinguish navigation, workspace content, status, and destructive actions;
- use consistent component states and accessible contrast;
- make system status and verification evidence explicit;
- avoid using color as the sole carrier of meaning;
- keep product UI tokens separate from marketing-only effects.

**Pending specification:** production design tokens, component library, icon rules, and UI state patterns must be documented from the actual product implementation before being designated official.

## 8. Business documents

Company profiles, proposals, assessment reports, pilot plans, invoices, and presentations should:

- use a consistent cover, header/footer, typography hierarchy, and logo placement;
- identify document title, version, date, owner, and confidentiality level where relevant;
- separate facts, assumptions, recommendations, and commitments;
- avoid unverified legal, commercial, security, or product claims;
- use editable source formats for controlled business documents.

## 9. Asset governance

Every production asset should have:

- a unique filename and version;
- a documented source/master;
- format and dimensions;
- intended use and restrictions;
- approval status: Draft, Review, Approved, Deprecated, or Archived;
- SHA-256 checksum for released binary files.

Only assets marked **Approved** may be treated as production-ready. Keep historical assets available for traceability, but clearly label them so they cannot be mistaken for current masters.

## 10. Open decisions before final approval

The following items require source validation or explicit brand-owner approval:

1. Final semantic color palette and exact gradient-stop values from the current master SVG.
2. Official typography family, weights, and fallback stack.
3. Logo clear-space and minimum-size measurements.
4. Approved light/dark background combinations and contrast checks.
5. Current production favicon and app-icon exports derived from master v2.0.
6. Product design tokens and UI component standards based on the live CRM/dashboard implementation.
7. Social template dimensions and export specifications for each destination platform.

---

## Source package reviewed

- `NAEOS-Logo-Exports-v1.0/README.md` and `manifest.json`
- `NAEOS-Logo-System-v1.0/README.md`
- `NAEOS-Social-Avatar-v1.0/README.md` and `manifest.json`

This guideline intentionally records gaps rather than filling them with unverified brand decisions.
