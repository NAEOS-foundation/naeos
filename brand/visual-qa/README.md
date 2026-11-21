# NAEOS Brand Visual QA

Automated SVG rendering and preview evidence for brand assets.

## Run locally
Prerequisites: Node.js 22+ and Playwright Chromium.

```sh
npm install --no-save playwright
npx playwright install chromium
node scripts/brand-visual-qa.mjs
```

Outputs are written to `brand-visual-qa-output/`:
- `png/`: rasterized SVG previews, preserving source canvas dimensions (1x)
- `report.json`: inventory, dimensions, warnings, render paths
- `gallery.html`: simple visual review gallery

## CI
GitHub Actions runs this on pull requests touching SVG assets, the script, or this workflow, and supports manual dispatch. The PNGs and report are uploaded as a 14-day artifact.

## QA boundaries
This pipeline validates that SVG files have measurable canvas dimensions and renders them using Chromium. It warns about external linked resources and embedded script elements. Rendering is **not** a visual pass: text wrapping, clipping, logo geometry, contrast, hierarchy, and brand consistency require human review. Any warnings or failed renders must be investigated before release.
