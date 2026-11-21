# NAEOS Social Media Kit v1.0

**Status:** Draft for review  
**Purpose:** Editable starter templates for NAEOS LinkedIn and social communication.

## Included templates

| Template | Canvas | Intended use |
|---|---:|---|
| LinkedIn post 1:1 | 1080 × 1080 | Single-image post |
| LinkedIn post 4:5 | 1080 × 1350 | Portrait post |
| Carousel cover 4:5 | 1080 × 1350 | Opening slide |
| Carousel content 4:5 | 1080 × 1350 | Repeatable insight/content slide |
| Quote / insight 1:1 | 1080 × 1080 | Quote, principle, engineering insight |
| Announcement 4:5 | 1080 × 1350 | Product/company announcement |
| LinkedIn company banner | 1584 × 396 | Company page cover; verify crop on destination before publishing |

## Brand source and governance

- Templates reference the canonical NAEOS Primary Symbol Master v2.0.0-derived SVG at `assets/naeos-primary-symbol-v2.svg`.
- The canonical master and raster PNG exports are already published under `brand/exports/NAEOS-Logo-Exports-v1.0/` (master SVG plus 512/1024/2048 PNGs). They are repository-level exports, not copies inside this Social Media Kit package; templates currently reference the kit-local SVG copy. Use the export package as the source when verifying or producing derivatives.
- Colors used are source-documented references only: Obsidian / Deep Navy `#0B0F19`, Signal Blue `#2563FF`, Ion Blue `#60A5FA`, Mist `#E5E7EB`, Cloud White `#FFFFFF`.
- These colors are not yet a ratified semantic palette.
- Typography uses `Inter, Arial, sans-serif` as a practical template fallback only; this is not an official NAEOS typeface selection.
- SVG files are editable. Replace all bracketed placeholder text before exporting a campaign-specific final asset.
- Templates intentionally avoid claims of customer adoption, certification, compliance, or measured outcomes.

## Important limitations

This is a first layout system, not a final approved campaign kit. Palette, typography, logo clear-space/minimum-size, and LinkedIn crop/safe-area validation remain open for review.

## Current repository structure

- `templates/` — seven editable SVG layouts.
- `assets/naeos-primary-symbol-v2.svg` — canonical symbol SVG copy used by template references.
- `previews/` — planned; template PNG render previews are not committed yet.
- `SOCIAL-MEDIA-KIT-PREVIEW.png` — planned template contact sheet; not committed yet.
- Repository-level exports are available under `brand/exports/NAEOS-Logo-Exports-v1.0/`, `brand/exports/NAEOS-Social-Avatar-v1.0/`, and `brand/exports/NAEOS-Logo-System-v1.0/`. These packages are separate from the kit folder and require provenance/hash reconciliation before RC packaging.

Do not mark this package production-approved until open items are reviewed. Record SHA-256 for released binary files.
