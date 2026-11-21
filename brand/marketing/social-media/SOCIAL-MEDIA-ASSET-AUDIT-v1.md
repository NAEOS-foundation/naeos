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

### Documentation discrepancy remediated

The README described `assets/naeos-primary-symbol-v2-2048.png` as the logo reference, but the repository currently has only the SVG logo reference copy in the kit. README and Asset Register now clarify that the PNG export is missing and that templates currently use the SVG. This is a documentation correction; no asset geometry or production brand tokens were changed.

### Raster exports and previews

Not committed:
- PNG symbol export `naeos-primary-symbol-v2-2048.png`
- Social avatar PNG
- Seven template previews
- Contact sheet

These are release-candidate gaps, not path failures in the editable SVG templates.

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
- [x] Document missing raster previews/contact sheet.
- [x] Classify generic template placeholders as intentional.
- [ ] Download and inspect generated inventory/checksum artifact line-by-line.
- [ ] Decode and validate raster dimensions/format and transparency.
- [ ] Render all seven SVG templates to PNG for visual QA.
- [ ] Produce contact sheet and compare to approved master v2.0 geometry/appearance.
- [ ] Verify LinkedIn safe areas and crops on platform-specific previews.
- [ ] Fill campaign placeholders in RC copies and re-run placeholder check on release assets.
- [ ] Review all SHA-256 values, provenance and release manifest.
- [ ] Independent review and explicit ratification/approval.

## 6. Governance

This audit and the corrective documentation are draft evidence only. No Logo Master v2.0.0, production CSS, semantic brand tokens, or brand decisions were changed. No release was published and no PR was merged.
