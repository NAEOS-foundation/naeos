---
title: Arsitektur
description: Penjelasan mendalam tentang arsitektur NAEOS dan prinsip desain.
---

## Arsitektur Sistem

NAEOS dibangun di atas arsitektur berlapis dengan lima lapisan utama:

```text
┌─────────────────────────────────────────────────────────┐
│                      Lapisan Input                        │
│            Spec YAML/JSON · Perintah CLI                  │
├─────────────────────────────────────────────────────────┤
│                      Core Runtime                         │
│   Parser · Normalizer · Resolver · Validator · Scheduler │
├─────────────────────────────────────────────────────────┤
│                    Lapisan Penalaran                       │
│            Reasoning Graph · Knowledge Graph               │
├─────────────────────────────────────────────────────────┤
│                    Lapisan Generasi                        │
│   Generator · Adapters · Template Engine · Compiler       │
├─────────────────────────────────────────────────────────┤
│                      Lapisan Output                        │
│   NEIR Model · Kode · Dokumen · Konteks AI · Manifes      │
└─────────────────────────────────────────────────────────┘
```

### 1. Lapisan Input
Titik masuk di mana spesifikasi (YAML/JSON) dan perintah CLI masuk ke sistem.

### 2. Core Runtime
Menangani parsing, normalisasi, resolusi referensi silang, validasi, dan penjadwalan berbasis DAG.

### 3. Lapisan Penalaran
Lapisan pengambilan keputusan dengan reasoning graph untuk ketelusuran dan knowledge graph untuk pemahaman domain.

### 4. Lapisan Generasi
Generasi kode multi-bahasa dengan adapter per-bahasa dan kompilasi instruksi AI.

### 5. Lapisan Output
Menghasilkan model NEIR, kode yang dihasilkan, dokumentasi, bundel konteks AI, dan manifes deployment.

## Prinsip Desain

- **Spesifikasi yang dapat dibaca manusia** — YAML/JSON sebagai sumber kebenaran tunggal
- **NEIR yang dapat dibaca mesin** — Representasi antara kanonikal untuk semua pemrosesan hilir
- **Netral vendor** — Multi-bahasa, multi-cloud, multi-platform-AI
- **Ekstensibel** — Adapter, plugin, dan profil untuk kustomisasi
- **Deterministik** — Input yang sama selalu menghasilkan output yang sama

## Komponen Utama

### Model NEIR
NAEOS Engineering Intermediate Representation mendeskripsikan seluruh sistem: proyek, arsitektur, modul, layanan, API, penyimpanan, infrastruktur, keamanan, AI, dokumentasi, deployment, pengujian, dan metadata.

### Pipeline Engine
Pipeline berbasis DAG 9-tahap: Parse → Normalisasi → Resolusi → Bangun → Validasi → Jadwalkan → Hasilkan → Kompilasi → Ekspor.

### Kompiler AI
Mengubah NEIR menjadi set instruksi AI untuk 6 platform target: GitHub Copilot, Claude Code, Cursor, Gemini CLI, Codex, dan OpenCode.

### Tata Kelola
Evaluasi kebijakan, RBAC, jejak audit, dan alur kerja tinjauan artefak.