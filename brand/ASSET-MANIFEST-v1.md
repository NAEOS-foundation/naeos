# NAEOS Brand Asset Manifest

**Manifest version:** 1.0  
**Logo master:** 2.0.0  
**Last verified:** 2026-11-17  
**Canonical repository:** `NAEOS-foundation/naeos`

This manifest separates canonical source artwork, production derivatives, and planned delivery exports. A file is marked **verified** only when its presence was checked in the repository.

## 1. Canonical source

| Asset | Repository path | Status |
|---|---|---|
| Primary Symbol Master v2.0.0 | `brand/master/naeos-primary-symbol-master-v2.svg` | Verified canonical source |

## 2. Brand-system SVG files

| Asset | Repository path | Status |
|---|---|---|
| Primary logo | `brand/logo.svg` | Verified |
| Logo mark | `brand/logo-mark.svg` | Verified |
| Horizontal logo | `brand/logo-horizontal.svg` | Verified |
| Simplified logo | `brand/logo-simplified.svg` | Verified |
| Monochrome logo | `brand/logo-mono.svg` | Verified |

## 3. Website and CRM production assets

| Surface | Asset | Repository path | Status |
|---|---|---|---|
| NAEOS website | Favicon | `site/public/favicon.svg` | Verified |
| NAEOS website | Logo mark | `site/public/images/logo-mark.svg` | Verified |
| NAEOS CRM | Favicon | `apps/web/public/favicon.svg` in `NAEOS-foundation/naeos-crm` | Verified |
| NAEOS CRM | Brand mark | `apps/web/public/naeos-brand-mark.svg` in `NAEOS-foundation/naeos-crm` | Verified |

## 4. Delivery exports

The following are planned deliverables. They are **not yet recorded as generated or verified files**.

| Asset | Target formats / dimensions | Status |
|---|---|---|
| Transparent logo export set | PNG at practical presentation and web resolutions | Planned |
| Social avatar | Square 1:1, platform-ready PNG | Planned |
| LinkedIn company cover | Platform-specific landscape export | Planned |
| Social post base template | Editable source plus PNG preview | Planned |
| Presentation theme | PPTX theme/template | Planned |
| Business document set | DOCX/PDF templates for proposal, invoice, and letterhead | Planned |
| Compact favicon export set | SVG and PNG/ICO where required by deployment | Planned |

## 5. Production rules

1. Generate all derivatives from the canonical master or approved brand-system variants.
2. Never edit the canonical master in place.
3. Preserve vector source files; do not treat raster exports as the source of truth.
4. Record output dimensions, format, intended placement, background variant, source master version, and export date.
5. Verify transparent backgrounds, aspect ratio, edge quality, and legibility at actual display size.
6. Keep internal business documents and private company records in approved private storage. This public manifest must contain no secrets or private company data.
7. Update this manifest whenever an asset is added, replaced, or retired.

## 6. Source references

- `brand/brand.json`
- `brand/brand.css`
- `brand/master/naeos-primary-symbol-master-v2.svg`
- `brand/guidelines/BRAND-GUIDELINES-v1.md`
