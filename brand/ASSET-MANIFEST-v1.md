# NAEOS Brand Asset Manifest

**Manifest version:** 1.2  
**Logo master:** 2.0.0  
**Last verified:** 2026-11-18  
**Canonical repository:** `NAEOS-foundation/naeos`

A file is marked **repository verified** only after its presence has been checked on this PR branch.

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

## 4. Published export sources

| Asset | Repository path | Status |
|---|---|---|
| Social avatar vector | `brand/exports/social/avatar/naeos-social-avatar.svg` | Repository verified on this PR branch |
| Social avatar README | `brand/exports/social/avatar/README.md` | Repository verified on this PR branch |
| Extraction register | `brand/exports/UPLOAD-REGISTER-v1.md` | Repository verified on this PR branch |
| Transparent logo PNG exports | Local package `NAEOS-Logo-Exports-v1.0.zip` | Generated locally; PNG upload pending |
| Social avatar PNG sizes (64, 180, 320, 512, 1024 px) | Local package `NAEOS-Social-Avatar-v1.0.zip` | Generated locally; PNG upload pending |
| LinkedIn company cover | — | Not generated |
| Social post base template | — | Planned |
| Presentation theme | — | Planned |
| Business document set | — | Planned |

The SVG avatar uses the canonical Primary Symbol Master v2.0.0 paths without changing symbol geometry. PNG delivery exports are not claimed as repository assets until their exact files are uploaded and verified.

## 5. Extracted package inventory

The three supplied ZIP packages were extracted and inspected locally: Logo System v1.0 (24 files), Logo Exports v1.0 (6 files), and Social Avatar v1.0 (7 files), 37 files total. The extraction register is published at `brand/exports/UPLOAD-REGISTER-v1.md`. Binary PNG/JPEG/ICO files remain pending individual repository upload and verification; no binary file is marked as published by this manifest.

## 6. Target repository layout

- `brand/exports/logo/` — transparent logo PNG exports and export README.
- `brand/exports/social/avatar/` — vector avatar source, raster avatar exports, and README.
- `brand/exports/manifest.json` — checksums and source/version provenance.

Do not commit ZIP archives as the only source of truth. Prefer extracted files, retain vector sources, and record SHA-256 checksums for raster outputs.

## 7. Production rules

1. Generate all derivatives from the canonical master or approved brand-system variants.
2. Never edit the canonical master in place.
3. Preserve vector source files; do not treat raster exports as the source of truth.
4. Record output dimensions, format, intended placement, background variant, source master version, and export date.
5. Verify transparent backgrounds, aspect ratio, edge quality, and legibility at actual display size.
6. Keep internal business documents and private company records in approved private storage. This public manifest must contain no secrets or private company data.
7. Update this manifest whenever an asset is added, replaced, or retired.
8. Do not mark locally generated assets as repository verified until the exact files are present on the target branch.

## 8. Source references

- `brand/brand.json`
- `brand/brand.css`
- `brand/master/naeos-primary-symbol-master-v2.svg`
- `brand/guidelines/BRAND-GUIDELINES-v1.md`
