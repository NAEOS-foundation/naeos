# Roadmap

Roadmap ini memberikan arah pengembangan dokumentasi dan ekosistem NAEOS.

## Fase 1 — Fondasi
- menyempurnakan dokumen inti,
- memastikan konsistensi terminologi,
- menambahkan panduan kontribusi dan onboarding.

## Fase 2 — Tooling dan Validasi
- menyiapkan template untuk ADR dan RFC,
- memperjelas mekanisme review,
- mengembangkan aturan validasi dokumen.

## Fase 3 — Referensi Implementasi
- menyediakan contoh implementasi referensi,
- memperjelas alur kerja dari requirement ke deployment,
- menyiapkan profil untuk skenario industri tertentu.

## Fase 4 — Ekosistem
- memperluas interoperabilitas dengan AI agent dan toolchain,
- memperkuat dokumentasi publik,
- mendukung adopsi lintas organisasi.

## Prinsip roadmap
Prioritas utama adalah menjaga kualitas, konsistensi, dan keterpahaman dokumen bagi komunitas serta implementer.

---

## Implementasi Teknis (Completed)

### Core Improvements
- [x] Fix `FindByContentSubstring` bug (was hardcoded false)
- [x] Resolver cross-reference: dependency filtering, endpoint normalization, defaults
- [x] Wire `--verbose` CLI flag to pipeline
- [x] Integrate `renderers.Renderer` into pipeline kernel service
- [x] Implement `GenerateForLanguage` with per-language code generation
- [x] Add `ParallelGroups()` to scheduler for priority-based execution
- [x] Add `extractDeployment()` and `extractTesting()` to NEIR builder
- [x] Add `SetOutputDir()` and file write to RuntimeEngine
- [x] 180+ tests passing with race detector
- [x] Clean up duplicate governance files
