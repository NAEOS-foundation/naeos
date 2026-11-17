# NAEOS Brand Asset Manifest

**Manifest version:** 1.1  
**Logo master:** 2.0.0  
**Last verified:** 2026-11-18  
**Canonical repository:** `NAEOS-foundation/naeos`

This manifest separates canonical source artwork, repository-verified production derivatives, locally generated delivery exports, and planned deliverables. A file is marked **repository verified** only when its presence was checked in GitHub.

## 1. Canonical source

| Asset | Repository path | Status |
|---|---|---|
| Primary Symbol Master v2.0.0 | `brand/master/naeos-primary-symbol-master-v2.svg` | Repository verified canonical source |

## 2. Brand-system SVG files

| Asset | Repository path | Status |
|---|---|---|
| Primary logo | `brand/logo.svg` | Repository verified |
| Logo mark | `brand/logo-mark.svg` | Repository verified |
| Horizontal logo | `brand/logo-horizontal.svg` | Repository verified |
| Simplified logo | `brand/logo-simplified.svg` | Repository verified |
| Monochrome logo | `brand/logo-mono.svg` | Repository verified |

## 3. Website and CRM production assets

| Surface | Asset | Repository path | Status |
|---|---|---|---|
| NAEOS website | Favicon | `site/public/favicon.svg` | Repository verified |
| NAEOS website | Logo mark | `site/public/images/logo-mark.svg` | Repository verified |
| NAEOS CRM | Favicon | `apps/web/public/favicon.svg` in `NAEOS-foundation/naeos-crm` | Repository verified |
| NAEOS CRM | Brand mark | `apps/web/public/naeos-brand-mark.svg` in `NAEOS-foundation/naeos-crm` | Repository verified |

## 4. Generated delivery exports

These assets have been generated as local delivery packages, but have **not yet been uploaded to this repository**. Their presence in GitHub must not be assumed until a follow-up PR adds and verifies them.

| Asset | Local package | Status |
|---|---|---|
| Transparent logo export set | `NAEOS-Logo-Exports-v1.0.zip` | Generated locally; repository upload pending |
| Social avatar PNG set | `NAEOS-Social-Avatar-v1.0.zip` | Generated locally; repository upload pending |
| LinkedIn company cover | — | Not generated |
| Social post base template | — | Planned |
| Presentation theme | — | Planned |
| Business document set | — | Planned |
| Compact favicon export set | — | Planned |

## 5. Target repository layout

When the generated packages are uploaded, use this layout:

- `brand/exports/logo/` — transparent logo PNG exports and export README.
- `brand/exports/social/avatar/` — square avatar PNGs and export README.
- `brand/exports/manifest.json` — checksums and source/version provenance.

Do not commit ZIP archives as the only source of truth. Extract their contents into the repository, preserve a concise README, and record SHA-256 checksums in the manifest. Keep original delivery ZIPs as convenience downloads only if repository policy permits.

## 6. Production rules

1. Generate all derivatives from the canonical master or approved brand-system variants.
2. Never edit the canonical master in place.
3. Preserve vector source files; do not treat raster exports as the source of truth.
4. Record output dimensions, format, intended placement, background variant, source master version, and export date.
5. Verify transparent backgrounds, aspect ratio, edge quality, and legibility at actual display size.
6. Keep internal business documents and private company records in approved private storage. This public manifest must contain no secrets or private company data.
7. Update this manifest whenever an asset is added, replaced, or retired.
8. Do not mark locally generated assets as repository verified until the exact files are present on the target branch.

## 7. Source references

- `brand/brand.json`
- `brand/brand.css`
- `brand/master/naeos-primary-symbol-master-v2.svg`
- `brand/guidelines/BRAND-GUIDELINES-v1.md`
