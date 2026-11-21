# NAEOS Brand Asset Extraction Register v1.0

- Extracted files: **37**
- Total uncompressed file size: **1,188,235 bytes**
- Source archive: `NAEOS-Brand-Assets-Extracted-v1.0.zip`
- Status: extracted and checksummed locally; **not yet published as individual GitHub assets**.

## Package inventory

| Package | Files |
|---|---:|
| NAEOS Logo Exports v1.0 | 6 |
| NAEOS Logo System v1.0 | 24 |
| NAEOS Social Avatar v1.0 | 7 |

## Validation notes

- The Logo Exports package contains the v2 master SVG and transparent PNG exports at 512×325, 1024×650, and 2048×1300.
- The Logo System package contains a historical `naeos-primary-symbol-master-v1.svg`; retain it as historical source, not as the current master.
- The Logo System favicon PNGs are logo-aspect-ratio renders (for example 16×10), not square canvases. Keep them as supplied exports and verify platform use before replacing production favicons.
- Social avatar PNGs are square RGB exports at 64, 180, 320, 512, and 1024 px.
- SHA-256 checksums for all 37 extracted files are recorded in the local companion file `asset-register.json`.

## Publication boundary

The archive has been extracted and inspected. The GitHub connector available in this workflow can commit UTF-8 text and metadata, but does not provide a local-file upload operation for binary PNG/JPEG/ICO payloads. Therefore no binary asset is marked as uploaded until it is actually present in the repository and verified.
