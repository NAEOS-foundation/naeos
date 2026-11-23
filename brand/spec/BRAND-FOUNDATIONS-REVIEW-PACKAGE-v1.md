# Brand Foundations Review Package v1

**Status:** Review package — no ratification
**Branch:** `chore/brand-system-stabilization-v2`
**Related PR:** #851
**Purpose:** Give the brand owner an evidence-based comparison of the two current source systems before any production CSS, typography, or token changes.

## Guardrails

- This package is a review aid, not an approval.
- Keep the canonical Logo Master v2.0.0 immutable.
- Do not change production website, CRM, dashboard, or social assets as part of this review.
- Logo gradient values remain in a separate logo-art namespace and are not semantic UI tokens.

## 1. Current source snapshots

| Foundation | Brand source / guideline | Website implementation | Decision needed |
|---|---|---|---|
| Primary blue | `#2563FF` | `#2563FF` | No disagreement in inspected sources |
| Cyan / accent | `#60A5FA` | `#00D9FF` | Choose one semantic role or define distinct roles (e.g. accent vs highlight) |
| Dark base | `#0B0F19` | `#05070D` | Decide base/background semantics |
| Navy / elevated | `#0B0F19` (brand source duplicates base) | `#06132B`, plus surfaces `#0b1730`, `#09162c` | Define dark surface hierarchy rather than one generic navy |
| Neutral | `#E5E7EB` | `#E2E8F0` | Decide canonical neutral or assign semantic roles |
| Display font | Sora | Space Grotesk | Select one display family or document brand-vs-product distinction |
| Body font | Inter | Inter | Aligned |
| Technical font | JetBrains Mono | JetBrains Mono (fallbacks present) | Aligned; formalize fallback stack |

## 2. Option A — Brand-document direction (current documented values)

**Snapshot:** Signal Blue `#2563FF`, Electric Cyan/Ion Blue `#60A5FA`, Deep Navy/Obsidian `#0B0F19`, Silver `#E5E7EB`; Sora display, Inter body, JetBrains Mono technical.

### Strengths
- Consistent with currently declared brand metadata and brand CSS.
- Compact and relatively conservative palette.
- Sora has a distinct geometric display character and can provide clear heading differentiation.

### Risks / open validation
- Website implementation does not currently match these declared values.
- `#0B0F19` is duplicated for Deep Navy and Obsidian in the current brand CSS, leaving role ambiguity.
- Cyan `#60A5FA` may function more like a blue accent than a bright cyan; verify perceived color identity and contrast before naming it “Electric Cyan”.
- Sora is documented but not currently loaded in the website layout.

## 3. Option B — Website implementation direction (current code values)

**Snapshot:** Signal Blue `#2563FF`, Cyan `#00D9FF`, Navy `#06132B`, Obsidian `#05070D`, Silver `#E2E8F0`; Space Grotesk display, Inter body, JetBrains Mono technical.

### Strengths
- Matches the active website's current implementation, so fewer immediate visual changes if chosen.
- Distinguishes dark base, elevated navy, surface, hover, and border roles more explicitly.
- Space Grotesk is already loaded by the website and used in display/headline contexts.

### Risks / open validation
- Website values are not yet ratified as the global brand foundation.
- The dark palette contains multiple values; semantic naming and accessibility checks are still required.
- Current site implementation contains legacy decorative and syntax colors; they should not be wholesale promoted to brand tokens.

## 4. Recommendation for review (not approval)

Prefer a **semantic token model** over choosing one flat palette and replacing every existing value:

1. **Core brand:** blue and approved logo gradient remain distinct, immutable namespaces.
2. **Marketing accent:** define a bright cyan role only if it is intentional for campaign/hero visuals.
3. **Product accent:** use a separately named UI accent token if it has different accessibility or interaction requirements.
4. **Dark surfaces:** define base, elevated, surface, alternate surface, hover, border, and focus roles. Do not treat `#0B0F19`, `#05070D`, and `#06132B` as interchangeable.
5. **Neutral:** choose the silver value based on role/contrast, not by forcing a single gray across text, borders, and surfaces.
6. **Typography:** decide whether display font is global or context-specific. The current site is evidence that Space Grotesk is in active use; it is not automatically proof of approval. Sora remains documented but is not loaded by the inspected website layout.

## 5. Ratification checklist

Mark only after brand-owner review and evidence:

- [ ] Display family: Sora / Space Grotesk / role-based split / other
- [ ] Cyan: `#60A5FA` / `#00D9FF` / separate marketing and product roles / other
- [ ] Dark base: `#0B0F19` / `#05070D` / other
- [ ] Navy/elevated surfaces: assign distinct semantic roles
- [ ] Neutral silver: `#E5E7EB` / `#E2E8F0` / role-based split / other
- [ ] Light-theme accent and contrast behavior
- [ ] Dark-theme text, status, border, focus, and contrast behavior
- [ ] Typography weights and fallback stacks
- [ ] Spacing scale, grid, clear-space, and minimum-size measurements
- [ ] Logo gradients remain separate from semantic tokens

## 6. Proposed verification criteria before implementation

- Contrast validation for text, icons, borders, focus indicators, and status colors in both themes.
- Visual comparison across website hero, article/technical page, CRM/dashboard surfaces, social template, and document template.
- Check small screens and reduced-motion behavior where animated accents exist.
- Run visual regression screenshots after tokens are implemented, not before.
- Ensure all production consumers reference a canonical token source or generated outputs.
- Validate logo derivative files and their checksums independently of semantic token changes.

## 7. Safe next phase after ratification

1. Record decisions in `brand/spec/BRAND-DECISIONS.md` with date, scope, and rationale.
2. Add canonical machine-readable semantic tokens (separate core, marketing, product, light/dark).
3. Generate CSS outputs; avoid manual duplicate declarations.
4. Update website in a scoped PR with visual regression evidence.
5. Audit CRM/dashboard/control plane token consumers separately; do not assume they share the website stylesheet.
6. Add brand checks to CI only after asset and token requirements are unambiguous.

## Source files inspected

- `brand/brand.json`
- `brand/brand.css`
- `brand/guidelines/BRAND-GUIDELINES-v1.1.md`
- `site/styles/globals.css`
- `site/app/[lang]/layout.tsx`
- `site/public/manifest.json`
- `site/public/images/og-dark.svg`
- `brand/spec/BRAND-SOURCE-RECONCILIATION.md`

**Conclusion:** Source evidence shows the website direction is implemented and the brand-document direction is declared, but neither should be globally ratified by inference. Use owner review to approve semantic roles, then implement tokens and migration in controlled stages.
