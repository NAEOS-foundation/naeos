# ADR Example

- ADR ID: ADR-001
- Title: Adopt NAEOS Specification-First Workflow
- Status: Proposed
- Date: 2026-07-09
- Authors: NAEOS Foundation
- Related Documents: README.md, specification/NAEOS-SPEC-001.md

## Context
Proyek NAEOS membutuhkan alur kerja yang konsisten untuk mengubah spesifikasi menjadi artefak implementasi. Tanpa workflow yang jelas, dokumen, generator, dan validator dapat berkembang secara tidak selaras.

## Decision
Mengadopsi workflow specification-first di mana setiap implementasi dimulai dari spesifikasi yang terdokumentasi dan dapat ditelusuri.

## Rationale
Workflow ini memastikan konsistensi, traceability, dan kemampuan validasi lintas komponen. Selain itu, pendekatan ini memudahkan integrasi dengan AI agent dan tooling eksternal.

## Alternatives Considered
- Code-first approach: cepat untuk prototyping, tetapi sulit diaudit dan dipelihara.
- Documentation-only approach: mudah dibaca, tetapi tidak cukup untuk automation.

## Consequences
Dampak positif:
- kualitas artefak lebih baik,
- proses review lebih terstruktur,
- traceability lebih kuat.

Risiko:
- kebutuhan belajar lebih tinggi bagi tim baru,
- proses awal bisa terasa lebih lambat.

## Notes
Keputusan ini harus didukung oleh template RFC, validator, dan panduan kontribusi.
