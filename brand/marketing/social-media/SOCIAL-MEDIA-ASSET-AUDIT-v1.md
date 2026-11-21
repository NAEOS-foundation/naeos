# Social Media Kit Asset Audit v1 — Findings and Remediation

**Audit date:** 2026-11-21  
**Audited source:** NAEOS repository at validator CI head `7dd19a71b295848a6aed4b989a877bd7875495a1`  
**Status:** Draft / not ratified

## 1. Validator gate

- CI run for Brand Asset Validation completed successfully on the audited commit.
- Unit tests passed.
- Repository asset scope includes `brand/**` and `site/public/**`.
- Automated validation is limited to inventory/hash, SVG parsing and structural/security checks, and marker detection; it does not perform raster decoding, visual rendering, geometry equivalence, or platform crop testing.

## 2. Inventory and checksum evidence

The workflow produced the `brand-asset-inventory` artifact (8,899 bytes; artifact digest SHA-256 `a67376ec81b2722b7d8b7183ccb4c36cdfacb06c30ea6d42717b1565d9aa4bbb`). The artifact is retained by CI and is not committed in this audit.

**Limitation:** the available GitHub connector cannot download the artifact ZIP. Individual inventory rows and per-asset checksums therefore remain pending direct artifact inspection. The artifact digest verifies the artifact object, not each asset's checksum.

## 3. Social Media Kit references

The seven template files are present:
- `linkedin-post-1x1.svg`
- `linkedin-post-4x5.svg`
- `carousel-cover-4x5.svg`
- `carousel-content-4x5.svg`
- `quote-insight-1x1.svg`
- `announcement-4x5.svg`
- `linkedin-company-banner.svg`

All seven templates reference `../assets/naeos-primary-symbol-v2.svg`, which is present. Relative path is correct from the template directory.

### Repository-level export reconciliation

The user-uploaded export packages are confirmed present under `brand/exports/` on `main`, separate from this Social Media Kit folder:
- `NAEOS-Logo-Exports-v1.0/`: canonical Master v2 SVG and transparent PNG exports (512, 1024, 2048); manifest includes file SHA-256 and byte sizes.
- `NAEOS-Social-Avatar-v1.0/`: avatar PNG exports (64, 180, 320, 512, 1024); manifest includes SHA-256 values.
- `NAEOS-Logo-System-v1.0/`: historical derivatives, app icons, favicons, monochrome variants, and contact sheet. Package README labels it historical; its v1 master must not be treated as current canonical master.

The Social Media Kit README points to these repository-level packages and clarifies that templates use a kit-local SVG copy. This resolves the previous false claim that symbol PNG and avatar exports were globally absent from the repository.

### Render automation added

- Added `scripts/render-social-media-kit.py` to render the seven SVG templates, validate PNG decodability, create a contact sheet, and write a dimensions/SHA-256 manifest.
- Added `.github/workflows/social-media-render.yml` to run on manual dispatch or relevant PR path changes and upload a 14-day QA artifact.
- Documented the reproducible local command in the kit README.

**Important:** automation source is committed, but its CI execution and resulting render artifact have not yet been verified. Do not claim rendered previews or visual QA are complete until workflow succeeds and artifact is inspected.

### Remaining validation

- [ ] Run render workflow and inspect generated previews/contact sheet.
- [x] **SVG source equivalence verified:** fetched kit-local `assets/naeos-primary-symbol-v2.svg` and canonical `brand/master/naeos-primary-symbol-master-v2.svg` contain identical SVG text (including viewBox, gradients, paths, title/desc). This establishes content equivalence for the fetched text files.
- [ ] Raster decode, dimensions, transparency, visual rendering, crop/safe-area validation pending render artifact inspection.

These are release-candidate validation gaps, not missing repository-level logo/avatar exports.

## 4. Placeholder classification

Bracketed content is intentionally retained in reusable template files:
- `[SERIES / DATE]` across post templates
- `[Add one insight or explanation]` and `[Support it with a workflow, example, or source]` in carousel content
- `[Insert a precise engineering insight.]` and `[Author / role / source]` in quote template
- Announcement headline, body and CTA placeholders

**Release control:** these placeholders must be filled with campaign-specific content and reviewed before export/publication. Do not remove them from generic templates merely to make the template validator report zero placeholders. Template placeholder detection is expected to flag editable template content; release-candidate copies must be placeholder-free.

## 5. Asset Release Candidate v1 preparation checklist

- [x] Automated validator gate green at the audited commit.
- [x] Confirm all seven social template files exist.
- [x] Confirm template logo SVG reference path exists.
- [x] Correct README mismatch about missing PNG export.
- [x] Document repository-level logo/avatar export packages.
- [x] Classify generic template placeholders as intentional.
- [x] Add reproducible render script and CI workflow source.
- [ ] Execute render workflow and inspect the QA artifact.
- [ ] Download and inspect generated inventory/checksum artifact line-by-line.
- [ ] Verify LinkedIn safe areas and crops on platform-specific previews.
- [ ] Fill campaign placeholders in RC copies and re-run placeholder check on release assets.
- [ ] Review all SHA-256 values, provenance and release manifest.
- [ ] Independent review and explicit ratification/approval.

## 6. Governance

This audit and corrective tooling are draft evidence only. No Logo Master v2.0.0, production CSS, semantic brand tokens, or brand decisions were changed. No release was published and no PR was merged.
