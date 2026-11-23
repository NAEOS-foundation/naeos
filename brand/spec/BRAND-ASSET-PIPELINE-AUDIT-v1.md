# NAEOS Brand Asset Pipeline Audit v1

**Audit date:** 2026-11-21
**Status:** Audit findings — not a release approval
**Related:** Brand System Stabilization PR #851
**Scope:** Repository manifests, SVG assets, extraction/export records, social templates, and binary-publication constraints.

## Executive summary

The repository has a usable canonical SVG master and a meaningful collection of SVG derivatives/templates, but the release pipeline is incomplete. The main risk is not lack of design files; it is **provenance/status ambiguity and unverified binary delivery**.

The audit confirms:
- Canonical Logo Master v2.0.0 SVG is present in the repo.
- Five core logo SVG variants are documented as repository-verified; website favicon and mark are documented.
- The social media kit has 7 SVG template canvases documented as repository-verifiable, but templates remain draft.
- The extraction register records 37 files from three local packages, with PNG and other binaries pending repository publication.
- The upload register notes that its companion SHA-256 data file exists locally, not in repository.
- Existing manifest and social register carry branch names/verification language that may be stale and should not be treated as current verification on main.
- No automated asset validation/release pipeline was confirmed in this audit.

## 1. Canonical source and provenance

| Item | Path / reference | Status | Audit note |
|---|---|---|---|
| Primary Symbol Master v2.0.0 | `brand/master/naeos-primary-symbol-master-v2.svg` | Verified source file | Keep immutable |
| Primary logo | `brand/logo.svg` | Listed as verified | Revalidate exact geometry and provenance |
| Logo mark | `brand/logo-mark.svg` | Listed as verified | Consumer derivative |
| Horizontal logo | `brand/logo-horizontal.svg` | Listed as verified | Revalidate against current master |
| Simplified logo | `brand/logo-simplified.svg` | Listed as verified | Confirm size-use criteria |
| Monochrome logo | `brand/logo-mono.svg` | Listed as verified | Verify both light/dark use and semantics |
| Website mark | `site/public/images/logo-mark.svg` | Listed as verified | Canonical consumer, should compare bytes/geometry |
| Original visual reference | `brand/source-of-truth/naeos-logo-source-of-truth.png` | Immutable/historical | Preserve unchanged |

**Control:** derivative validation must compare geometry/provenance to master, not just visual similarity. Master edits are out of scope.

## 2. Binary assets publication gap

Extraction records describe three local packages with **37 files total**:
- Logo Exports package: 6 files (README/manifest/master SVG and transparent PNG at 512, 1024, 2048)
- Logo System package: 24 files (includes historical v1, variants, app icons/favicons, PNGs, contact sheet)
- Social Avatar package: 7 files (README/manifest and PNG sizes 64, 180, 320, 512, 1024)

Repository manifests explicitly state PNG/JPEG/ICO binaries are not yet individually uploaded and verified. Local checksum companion file `asset-register.json` is also not present in the public repository according to the extraction register.

**Risk:** users or maintainers can see documentation naming planned exports and mistakenly assume files are downloadable/published.

**Recommended disposition:** keep these as `PENDING_PUBLICATION`; only promote to `VERIFIED` after exact path, binary content, size, format, dimensions, SHA-256, and provenance are checked in the repository.

## 3. Social Media Kit inventory

Documented editable SVG canvases:
| Template | Canvas | Status |
|---|---:|---|
| LinkedIn post square | 1080×1080 | Draft |
| LinkedIn post portrait | 1080×1350 | Draft |
| Carousel cover | 1080×1350 | Draft |
| Carousel content | 1080×1350 | Draft |
| Quote / insight | 1080×1080 | Draft |
| Announcement | 1080×1350 | Draft |
| LinkedIn company banner | 1584×396 | Draft; platform crop check needed |

Known gaps: PNG preview/contact sheet and PNG logo/avatar references are described as local/not uploaded. The kit references a PNG asset path that is not repository-verified; current templates must be checked for whether they use SVG logo copy or PNG references, and corrected to use a verified SVG source before release.

## 4. Manifest integrity and freshness

Current manifest `brand/ASSET-MANIFEST-v1.md` reports version 1.2 and last verified 2026-11-18. It mixes:
- assets repository-verified,
- local-only exports,
- planned assets,
- cross-repository CRM assets,
- branch-qualified verification claims.

The social asset register identifies branch `docs/brand-guidelines-v1`, and says files were uploaded/repository-verifiable, while the current default-branch snapshot is the reference being audited. This branch metadata is not a robust verification record for a released asset.

**Recommendation:** introduce a release manifest generated from actual repository files, not manually maintained “repository verified” prose. Use a per-asset record:
- asset ID and logical name
- canonical source master + version
- output path
- format, width, height, viewBox (if SVG)
- intended placement and light/dark background
- generator/tool/version and reproducible command
- SHA-256
- status: `DRAFT`, `PENDING_PUBLICATION`, `VERIFIED`, `RETIRED`
- verification date and verifier (avoid sensitive details)
- license/IP evidence reference if applicable

## 5. Pipeline recommendation

### Stage A — inventory (read-only)
Enumerate every file under `brand/`, `site/public/`, and related social templates; capture metadata; group by master version; mark cross-repo consumer references separately.

### Stage B — validation
Automate:
- SVG parse and XML well-formedness.
- viewBox presence, positive dimensions, path/shape bounds.
- no unexpected external links, scripts, or embedded raster.
- logo-master geometry/provenance checks against canonical source.
- raster dimensions/format/alpha channel and image integrity.
- expected filename and naming convention.
- checksum calculation and manifest reconciliation.
- contrast checks for UI tokens only (do not misapply to logo artwork).
- template placeholders scan (block release if bracketed placeholders remain).
- preview render/contact-sheet generation in CI.

### Stage C — release candidate
Create a versioned release directory, e.g. `brand/releases/assets-vX.Y.Z/` with:
- manifest JSON,
- SVG sources and raster exports,
- previews/contact sheet,
- validation report,
- checksum file,
- release notes and deprecation map.

### Stage D — release
Publish through a controlled PR; verify every file in target branch; tag/release only after approval. Keep canonical master untouched; all generated derivatives should be reproducible from source.

## 6. Priority actions

### P0 — before publishing current social kit
1. Revalidate all asset paths and references; especially template logo references and PNG paths.
2. Reconcile manifest/register branch claims and last-verified dates.
3. Produce repository inventory and checksum data for actual committed files.
4. Check SVG templates for placeholders and external dependencies.
5. Correct the “downloadable/published” language where binaries are missing.

### P1 — pipeline implementation
1. Add a deterministic SVG/raster asset validation script.
2. Add CI job for manifest reconciliation, placeholders, dimensions, XML, and checksums.
3. Generate previews/contact sheet automatically.
4. Add a private upload/publication workflow for binary asset files if direct local binary upload becomes available; don't publish archives as the only source.

### P2 — later improvements
- Asset usage/crop guide per platform.
- Logo minimum size and clear-space measurement records.
- Consumer audit across site, CRM, dashboard, and control plane.
- Automated deprecation alerts for old logo versions.

## 7. Safe acceptance criteria

- Every released asset resolves to an exact committed path.
- Every derivative maps to a canonical source/version.
- Every binary has verified dimensions, format, alpha/background, and SHA-256.
- Every SVG passes XML parsing and policy checks.
- Every template has no placeholder text and has platform crop review.
- Manifest is generated or CI-verified from repository reality.
- No logo geometry changes without explicit owner approval and a versioned master.
- Draft/pending/verified/retired states are explicit and accurate.

## 8. Limitations of this audit

This was a repository documentation and source audit. It did not download and inspect every binary package, visually render every SVG template, compare all derivative geometry at pixel/vector level, or validate CRM assets in the separate repository. Those tasks remain queued as a second verification pass.

## Sources reviewed

- `brand/ASSET-MANIFEST-v1.md`
- `brand/exports/UPLOAD-REGISTER-v1.md`
- `brand/exports/EXTRACTION-REGISTER-v1.0.md`
- `brand/marketing/social-media/ASSET-REGISTER.md`
- `brand/marketing/social-media/README.md`
- `brand/source-of-truth/README.md`
