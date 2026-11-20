# NAEOS Brand Guidelines v1.0

**Status:** Initial operational guideline  
**Brand system:** NAEOS  
**Logo master:** 2.0.0  
**Brand metadata:** 3.1.0  
**Source of truth:** `brand/brand.json`, `brand/brand.css`, and `brand/master/naeos-primary-symbol-master-v2.svg`

## 1. Brand foundation

- **Brand name:** NAEOS
- **Description:** Nusantara AI Engineering Operating System
- **Corporate tagline:** Architecture Drives Engineering.
- **Product-facing descriptor:** The Intelligent Core

Use the corporate tagline in company-level and institutional communication. Use “The Intelligent Core” as the product-facing descriptor where a compact product expression is more appropriate. Do not combine, rewrite, or replace these phrases without an approved brand revision.

## 2. Logo governance

The founder-approved Primary Symbol is registered as **Logo Master v2.0.0**. The canonical vector is:

`brand/master/naeos-primary-symbol-master-v2.svg`

Production lockups and derivatives must remain traceable to this master. The master geometry is frozen. Any change to the symbol geometry requires explicit approval and a new master version.

### Safe-use rules

- Use the supplied logo variants; do not redraw the symbol manually.
- Preserve the original aspect ratio and proportions.
- Do not stretch, skew, rotate, crop, outline, or add effects to the symbol.
- Do not alter the approved gradient or substitute arbitrary colors.
- Keep the mark legible against its background and provide clear surrounding space.
- Use a supplied monochrome variant when color reproduction is unavailable or inappropriate.
- Do not place the logo over visually noisy imagery where its silhouette or contrast is compromised.
- Do not create new lockups by rearranging logo components without approval.

### Variant selection

| Variant | Intended use |
|---|---|
| Primary logo / lockup | Corporate documents, presentations, and brand introductions |
| Symbol / mark | Avatar, favicon, compact UI, and small-format placements |
| Horizontal lockup | Website headers, document headers, and landscape layouts |
| Simplified mark | Constrained small-size contexts, only where the supplied file is intended |
| Monochrome | Single-color printing and contexts where color is unavailable |

The canonical master is the geometry authority. A derivative file is not a new master.

## 3. Color system

Use the values below exactly as recorded in `brand/brand.json` and `brand/brand.css`.

| Token | Hex | Role |
|---|---|---|
| NAEOS Blue | `#2563FF` | Primary brand accent |
| Electric Cyan | `#60A5FA` | Secondary accent and highlights |
| Deep Navy | `#0B0F19` | Primary dark foundation |
| Obsidian | `#0B0F19` | Dark surface |
| Slate | `#64748B` | Secondary text and neutral UI |
| Silver | `#E5E7EB` | Light neutral surface / text |
| White | `#FFFFFF` | High-contrast light surface |

### Color application

- Use Deep Navy / Obsidian as the primary dark foundation for high-level brand compositions.
- Use NAEOS Blue as the principal action and emphasis color.
- Use Electric Cyan as a supporting accent, not as a replacement for the primary blue.
- Use Silver and White for readable contrast on dark surfaces.
- Use Slate for secondary information; verify contrast before using it for small text.
- Avoid introducing additional brand colors into official assets without a documented revision.

## 4. Typography

| Role | Typeface | Typical application |
|---|---|---|
| Display | Sora | Headlines, section titles, brand statements |
| Body | Inter | Paragraphs, descriptions, labels, and UI text |
| Technical | JetBrains Mono | Code, identifiers, technical labels, and data snippets |

Maintain a clear hierarchy: display typography for short statements, body typography for sustained reading, and technical typography only where technical content is present. If a typeface is unavailable, use a metrically and visually compatible sans-serif fallback without changing the hierarchy.

## 5. Layout and visual language

- Prefer structured, grid-aligned compositions with clear hierarchy.
- Use generous negative space around the symbol and major headlines.
- Keep information density controlled; technical detail should be grouped and labeled.
- Use blue/cyan accents to direct attention rather than decorate every element.
- Favor restrained gradients and subtle glow effects; never let effects obscure the mark or reduce text contrast.
- Use consistent alignment, spacing, and corner treatments within a single asset family.

## 6. Social and marketing assets

- Profile avatars should use the symbol/mark variant, centered and tested at small sizes.
- Wide cover artwork should keep essential text and marks inside platform-safe areas.
- Social posts should use one principal message, one clear visual focal point, and a restrained supporting line.
- Use the corporate tagline for company positioning; use “The Intelligent Core” for product-oriented creative.
- Keep claims factual and avoid implying certifications, customer outcomes, or capabilities that have not been verified.

## 7. Business documents and presentations

- Use the primary or horizontal lockup in title areas; use the mark sparingly in page furniture.
- Keep body copy in Inter and headings in Sora.
- Use NAEOS Blue for emphasis and Deep Navy for major headings or dark cover surfaces.
- Maintain accessible contrast and consistent page margins.
- Keep legal identity, registration, address, and tax fields as approved company data or explicit placeholders; never invent them.

## 8. Digital product usage

- Use the mark variant for favicon and compact navigation contexts.
- Preserve consistent logo geometry across naeos.dev and crm.naeos.dev.
- Product UI should prioritize legibility and functional state communication over decorative brand effects.
- Do not recolor status indicators in ways that conflict with their semantic meaning.

## 9. Asset production and version control

Every derivative asset should record:
- source master version;
- output format and dimensions;
- intended placement;
- background variant;
- creation or revision date.

Recommended asset groups:
1. Logo exports: SVG and transparent PNG in approved color and monochrome variants.
2. Social: avatar and platform-specific cover artwork.
3. Marketing: reusable social post layouts.
4. Business: presentation, proposal, invoice, and letterhead templates.
5. Governance: this guideline and an asset manifest.

Do not overwrite the canonical master when exporting derivatives. Keep generated exports in a clearly named derivative directory and update the asset manifest when files are added.

## 10. Change control

Changes to wording, color tokens, typography, logo geometry, or approved usage rules must be reviewed as a brand-system change. Logo geometry changes require a new explicit master version. This guideline does not authorize changes to the frozen master.

---

**Canonical references**
- `brand/brand.json`
- `brand/brand.css`
- `brand/master/naeos-primary-symbol-master-v2.svg`
