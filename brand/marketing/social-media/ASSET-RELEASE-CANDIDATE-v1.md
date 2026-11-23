# NAEOS Social Media Asset Release Candidate v1

**Version:** RC v1 (preliminary / incomplete)  
**Source:** Social Media Kit v1.0 draft  
**Logo source:** NAEOS Primary Symbol Master v2.0.0  
**Status:** BLOCKED — not ratified and not publishable

## Candidate assets

| Asset | Source SVG | Candidate status |
|---|---|---|
| LinkedIn 1:1 | `templates/linkedin-post-1x1.svg` | Render QA passed; campaign content pending |
| LinkedIn 4:5 | `templates/linkedin-post-4x5.svg` | Render QA passed; campaign content pending |
| Carousel cover 4:5 | `templates/carousel-cover-4x5.svg` | Render QA passed; campaign content pending |
| Carousel content 4:5 | `templates/carousel-content-4x5.svg` | Render QA passed; generic placeholder retained, RC copy must be filled |
| Quote/insight 1:1 | `templates/quote-insight-1x1.svg` | Render QA passed; verified quote/source and campaign copy pending |
| Announcement 4:5 | `templates/announcement-4x5.svg` | Render QA passed; approved announcement copy and CTA pending |
| LinkedIn company banner | `templates/linkedin-company-banner.svg` | Render QA passed and text clipping fixed; platform crop/safe-area validation pending |

## Repository-level export evidence

- Logo Master v2.0.0-derived SVG and PNG exports are available in `brand/exports/NAEOS-Logo-Exports-v1.0/` (512, 1024, 2048 px; manifest present).
- Social Avatar v1.0 PNG exports are available in `brand/exports/NAEOS-Social-Avatar-v1.0/` (64, 180, 320, 512, 1024 px; manifest present).
- Historical Logo System package and contact sheet are available in `brand/exports/NAEOS-Logo-System-v1.0/`; do not promote its legacy v1 master as canonical.

## Required gates before RC can be complete

- [ ] Generated inventory and per-file checksums inspected.
- [x] SVG templates rendered to PNG and visually inspected (workflow #4; artifact inspected).
- [x] Rendered symbol visually matches approved Master v2.0 direction; kit-local SVG source text equivalence verified against canonical source.
- [x] PNG decoding and dimensions checked for seven outputs; RGB/opaque output verified. Color profile/platform export check remains pending.
- [ ] LinkedIn crop/safe-area checks on actual platform preview remain pending.
- [ ] All campaign-specific placeholder text filled in candidate exports.
- [ ] Campaign factual claims and source/attribution verified.
- [ ] Final asset set and SHA-256 manifest recorded.
- [ ] Reviewer sign-off recorded.

## Governance

This document is a release-candidate preparation record only. It is not a release approval. Do not tag, publish, merge, or ratify this kit based on this document. Preserve the generic editable templates and create separate campaign-specific export copies for placeholder replacement.
