# Brand Asset Validation Tooling v1

This validator is read-only with respect to brand source files. It does not modify the canonical logo master, raster assets, production CSS, or brand decisions.

## Scope

- Scans supported files under `brand/` and `site/public/`.
- Parses SVG XML, checks root, viewBox and positive viewBox dimensions.
- Rejects script and foreignObject elements, event-handler attributes, external href/src references and external CSS URL resources.
- Detects common placeholder markers.
- Produces a deterministic inventory containing file path, format, byte size, SHA-256 and SVG metadata.

## Run

```bash
python3 scripts/validate-brand-assets.py
python3 scripts/validate-brand-assets.py --write
python3 scripts/validate-brand-assets.py --check-generated
```

Generated outputs: `brand/exports/asset-inventory.json` and `brand/exports/asset-checksums.sha256`. The CI job publishes these as a build artifact; inventory includes only assets actually present in the repository, while planned or local-only assets are excluded.

## Validation boundaries

Version 1 does not render/visually compare assets, prove geometry equivalence to the master, decode raster images or check raster dimensions, test color contrast, or validate assets in other repositories. Internal fragment links (such as `#gradient`) are allowed.

Passing checks is evidence, not brand approval. Assets remain draft unless separately ratified.
