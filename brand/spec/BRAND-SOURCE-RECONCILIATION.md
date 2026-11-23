# Brand Source Reconciliation — Evidence Matrix

**Audit stage:** Source inspection, not visual ratification  
**Base revision inspected:** `main` at PR #851 base SHA `3555c9c1c726c585c694b8fbdf9c0209b5c6f329`  
**Logo master:** v2.0.0 — immutable

## 1. Confirmed current logo-master gradient values

The canonical SVG uses two gradients, not the semantic brand palette.

| Gradient | Stop | Value |
|---|---|---|
| White gradient | 0% | `#FFFFFF` |
| White gradient | 62% | `#F7F9FF` |
| White gradient | 100% | `#DCE8FA` |
| Blue gradient | 0% | `#168BFF` |
| Blue gradient | 52% | `#0563FF` |
| Blue gradient | 100% | `#0737D9` |

These are logo-art values and should not be automatically promoted to semantic UI tokens.

## 2. Source conflict matrix

| Item | Brand metadata / guideline | Website implementation | Social kit | Finding |
|---|---|---|---|---|
| Primary blue | `#2563FF` | `#2563FF` | `#2563FF` | Consistent in sampled files |
| Cyan accent | `#60A5FA` | `#00D9FF` | `#60A5FA` | **Conflict**; site uses a more saturated cyan |
| Dark foundation | `#0B0F19` | `#05070D` and `#06132B` | `#0B0F19` | **Conflict**; site has different dark foundations |
| Silver neutral | `#E5E7EB` | `#E2E8F0` | `#E5E7EB` | **Conflict** in neutral value |
| Display type | Sora | Space Grotesk | Inter fallback (template) | **Conflict**; site uses Space Grotesk, not Sora |
| Body type | Inter | Inter | Inter fallback | Mostly aligned; template fallback differs by policy |
| Technical type | JetBrains Mono | JetBrains Mono | Not specified | Mostly aligned, but token policy differs |
| UI accent | No ratified semantic mapping | Cyan `#00D9FF`, blue `#2563FF`, light-theme accent variants | Not applicable | Semantic roles and theme values not reconciled |
| Logo colors | Master gradient values | Some OG SVG references cyan `#08D6FF` in decorative elements | Social uses palette references | Logo colors must remain separate from UI and marketing accent tokens |

## 3. Concrete source evidence

- `brand/brand.json` declares colors including NAEOS Blue `#2563FF`, Electric Cyan `#60A5FA`, Deep Navy `#0B0F19`, Silver `#E5E7EB`, and typography Sora / Inter / JetBrains Mono.
- `brand/brand.css` mirrors those declared values and font families.
- `brand/guidelines/BRAND-GUIDELINES-v1.md` describes Sora / Inter / JetBrains Mono as the role mapping and records the palette.
- `site/styles/globals.css` declares brand cyan `#00D9FF`, navy `#06132B`, obsidian `#05070D`, silver `#E2E8F0`, and display font Space Grotesk.
- `site/app/[lang]/layout.tsx` loads Inter, JetBrains Mono, and Space Grotesk from Google Fonts.
- `site/public/manifest.json` uses `#05070D` as background and `#2563FF` as theme color.
- `site/public/images/og-dark.svg` and related OG artwork use cyan `#08D6FF` in decorative elements.
- Social kit README records colors as source-documented references and explicitly says the semantic palette is not ratified; it uses Inter as a practical fallback, not a declared final typeface.

## 4. Audit conclusion

There is real implementation drift—not merely a documentation mismatch—between the current brand foundation documents and website CSS/metadata. The highest-impact divergences are cyan accent, dark foundation/surface colors, neutral silver, and display typography.

**Do not immediately overwrite website styles.** The site may have a distinct, evolved visual system or design choices with functional contrast/theme rationale. The next step is an owner-led visual review and a token decision record:

1. Decide whether the canonical display family should be Sora or Space Grotesk.
2. Decide whether canonical cyan is `#60A5FA`, `#00D9FF`, or a formally separated marketing/product accent role.
3. Decide whether `#0B0F19`, `#05070D`, and `#06132B` are distinct semantic surfaces or accidental drift.
4. Decide whether silver neutral `#E5E7EB` or `#E2E8F0` is canonical, or assign semantic roles.
5. Reconcile light theme accent and contrast after semantic roles are defined.

## 5. Recommended minimal correction path

- Preserve all current production values until review.
- Ratify semantic roles, including dark/light theme.
- Make canonical token source machine-readable.
- Generate brand CSS and product CSS token outputs (or shared package) rather than maintaining duplicate declarations.
- Apply tokens in staged, visual-regression-tested changes.
- Keep logo gradient stops immutable and in a separate namespace.
