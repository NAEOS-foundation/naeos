# NAEOS Logo Source of Truth

Status: **FROZEN / IMMUTABLE**

## Visual Source of Truth

`brand/source-of-truth/naeos-logo-source-of-truth.png` is the canonical **original visual reference supplied for the NAEOS logo**.

This file is preserved as supplied. It is the visual reference against which derived logo geometry, rendering, and production assets are evaluated.

> **Important:** the original reference PNG is not required to be byte-identical to any later frozen packaging artifact or derived vector. Its role is visual reference/source material, not a claim that every copy or derivative has identical file bytes.

## Reference Integrity

The repository copy must not be edited, redrawn, recompressed, recolored, or replaced in place.

The SHA-256 recorded in the repository documentation refers to the previously frozen reference-package artifact, not as an assertion that the currently uploaded original-reference PNG has identical bytes.

For byte-level verification of a specific repository copy, download that exact PNG and compute its SHA-256 locally.

## Derived Assets

- `brand/master/naeos-mark-master.svg` — current production vector derivative (Logo Master v1.0).
- `brand/logo-mark.svg` — public mark consumer.
- `brand/variants/` — derived presentation variants.
- `brand/lockups/` — derived lockups.

The derived vector is **not a pixel-identical copy of the visual reference**. Any future geometry replacement requires an explicit versioned review and must not silently overwrite the original visual reference.

## Governance

1. The original visual reference is the absolute visual source of truth for visual intent.
2. Geometry changes require a new version.
3. Rendering/material changes may be made without changing the reference.
4. Production consumers must point to approved derived assets.
5. Do not delete or mutate the original visual reference.
6. Do not treat a package checksum as the checksum of a separately uploaded repository copy unless it has been independently verified.
