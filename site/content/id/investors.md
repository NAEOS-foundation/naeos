---
title: Investor
description: NAEOS adalah engineering control plane untuk era software AI-native.
---

# Investor NAEOS

> AI coding agents menghasilkan kode. NAEOS menyediakan engineering control plane di sekitar mereka.

## Thesis utama

NAEOS sedang membangun lapisan sistem untuk engineering software AI-native.

Repositori ini sudah menunjukkan fondasi teknis yang koheren: bahasa spesifikasi deklaratif, pipeline yang mem-parsing dan menormalisasi spesifikasi, membangun model engineering terstruktur (NEIR), memvalidasi dependency dan policy, menjalankan pekerjaan, menghasilkan artefak, mengompilasi konteks AI, dan menghasilkan evidence.

Ini penting karena AI coding tools berkembang lebih cepat daripada banyak sistem engineering yang beradaptasi. Ketika kemampuan agent naik, kebutuhan akan governance, validasi, repeatability, dan auditability menjadi lebih penting, bukan lebih sedikit.

## Apa itu NAEOS hari ini

### Current

NAEOS adalah platform engineering open source yang mengubah spesifikasi perangkat lunak menjadi workflow yang tervalidasi, teratur, dan siap untuk AI.

Fondasi yang sudah ada di repositori mencakup:

- parsing spesifikasi deklaratif dengan YAML/JSON
- stage pipeline untuk parse, normalize, resolve, NEIR, dan validasi
- evaluasi policy dan governance hooks
- pembuatan AI context bundle
- generation multi-bahasa dan artifact review
- workflow CLI untuk developer
- komponen evidence, audit, observability, dan demo control-plane

Untuk tinjauan repositori saat ini, lihat [README.md](https://github.com/NAEOS-foundation/naeos/blob/main/README.md), [ARCHITECTURE-OVERVIEW.md](https://github.com/NAEOS-foundation/naeos/blob/main/ARCHITECTURE-OVERVIEW.md), [WHITEPAPER.md](https://github.com/NAEOS-foundation/naeos/blob/main/WHITEPAPER.md), dan [ROADMAP.md](https://github.com/NAEOS-foundation/naeos/blob/main/ROADMAP.md).

## Mengapa masalah ini muncul sekarang

AI coding agents dapat menghasilkan kode dengan cepat, tetapi engineering perangkat lunak bukan hanya code generation. Ia juga membutuhkan:

- intent engineering yang konsisten
- validasi arsitektur dan dependency
- enforcement policy
- manajemen konteks antar tool
- traceability dari spesifikasi ke artefak
- governance terhadap aksi agent
- eksekusi yang dapat direproduksi

Ini adalah celah yang NAEOS dirancang untuk isi.

## Layer arsitektur yang penting

### Current

Repositori saat ini sudah memperlakukan NEIR sebagai model engineering yang persisten.

```text
Specification
      ↓
Parse → Normalize → Resolve → Build NEIR
      ↓
Validate → Policy → AI Context
      ↓
Execution / Generation
      ↓
Artifacts → Evidence / Audit
```

Ini bukan hanya AI coding assistant. Ini adalah sistem engineering di sekitar eksekusi berbasis AI.

## NEIR sebagai keunggulan strategis

### Current

NEIR adalah representasi engineering sentral di repositori. Ia memberi NAEOS model terstruktur yang bisa divalidasi, diatur, diubah, didokumentasikan, dan disuplai ke tools AI downstream.

Ini merupakan salah satu diferensiasi paling penting:

- spesifikasi adalah source of truth
- model dapat diproses secara mesin
- validasi dan policy terjadi sebelum generation
- konteks dapat dikompilasi secara konsisten
- evidence dapat ditautkan ke artefak dan eksekusi

## Mengapa NAEOS berbeda

| Kategori | AI coding tools | NAEOS |
|---|---|---|
| Peran inti | Code generation | Engineering control plane |
| Source of truth | Prompt atau perubahan file | Spesifikasi deklaratif + NEIR |
| Governance | Sering eksternal atau ad hoc | Terintegrasi ke pipeline |
| Validasi | Variatif atau parsial | Deterministik dan terstruktur |
| AI context | Sering spesifik per tool | Dikompilasi secara sistematis |
| Auditability | Terbatas | Evidence dan traceability dibangun di dalam sistem |
| Model eksekusi | Aksi agent | Pipeline terkontrol dengan policy dan review |

## Fondasi teknis saat ini

### Current

Repositori ini sudah berisi sebagian besar fondasi yang dibutuhkan:

- `pkg/pipeline` — pipeline eksekusi utama
- `internal/neir` — model NEIR dan komponen validasi
- `internal/governance/policy` — evaluasi kebijakan dan rules
- `internal/context/bundle` — pembuatan AI context
- `internal/evidence` — dukungan evidence dan audit
- `internal/investordemo` — demo control-plane dan alur otorisasi deterministik
- `internal/demoobs` — observability dan dukungan export SIEM/OTLP
- `cmd/naeos` — antarmuka CLI untuk validasi, context, run, demo, policy, dan workflow verifikasi

Proof point terkuat saat ini adalah bahwa repositori ini sudah mengimplementasikan arsitektur dalam kode, bukan hanya dalam pitch deck.

## Open source dan monetisasi masa depan

### Current

NAEOS adalah open source di bawah Apache 2.0 dan dibangun secara terbuka. Ini adalah keunggulan adopsi karena memungkinkan validasi teknis, kontribusi eksternal, dan pengembangan ekosistem.

### Target

Jalur komersial yang realistis adalah menjaga open core tetap berguna secara luas sambil memonetisasi layer control plane melalui:

- governance dan policy management terpusat
- workflow audit dan evidence
- integrasi enterprise
- layanan control-plane yang dikelola
- permukaan deployment yang dikelola
- kontrol kolaborasi dan compliance untuk tim

### Vision

Arah jangka panjangnya adalah menjadi control plane default untuk tim engineering AI-native yang membutuhkan trust, struktur, dan governance di sekitar pembuatan perangkat lunak berbasis AI.

## Tahap pengembangan saat ini

### Current

Repositori sudah cukup besar dan berfungsi. Implementasi saat ini mencakup workflow CLI yang bisa dijalankan, pipeline yang bekerja, evaluasi policy, pembuatan konteks, kemampuan demo control-plane, serta hook evidence dan audit.

### Target

Prioritas berikutnya adalah mengubah fondasi teknis yang ada menjadi wedge developer yang lebih tajam dan story komersial yang lebih jelas:

- alur kontrol-plane yang lebih jelas
- contoh end-to-end yang lebih kuat
- narasi governance dan audit yang lebih baik
- positioning enterprise dan mitra yang lebih kuat
- validasi design-partner yang lebih luas

### Vision

Dalam 12 bulan ke depan, tujuan NAEOS adalah menjadi lapisan engineering default bagi tim yang ingin AI produktif tanpa mengorbankan kejelasan arsitektur, compliance policy, atau traceability operasional.

## Target 12 bulan

### Target

Target berikut adalah target, bukan hasil saat ini:

- 10–20 design partner
- 100+ tim engineering aktif yang menggunakan proyek ini
- 3–5 enterprise pilot
- 8–10 integrasi teknis di sekitar AI dan developer tooling
- 50+ contributor eksternal
- workflow control-plane yang lebih kuat dan berulang untuk tim nyata

## Mengapa investor harus peduli

NAEOS menarik bagi investor yang memahami developer infrastructure, AI systems, open-source platforms, dan kategori baru dari AI-native engineering tools.

Argumen terkuatnya bukan bahwa NAEOS adalah “AI tool lain.” Argumen terkuatnya adalah bahwa NAEOS sedang membangun infrastruktur sistem yang dibutuhkan oleh delivery perangkat lunak berbasis AI.

## Arah founder

NAEOS dibangun secara terbuka di depan publik. Proyek ini secara eksplisit diposisikan sebagai investasi infrastruktur untuk masa depan engineering, bukan sekadar wrapper produktivitas di sekitar LLM.

## Kontak

Repositori adalah sumber kebenaran utama untuk evaluasi teknis dan bukti publik.

- GitHub: [NAEOS Foundation / naeos](https://github.com/NAEOS-foundation/naeos)
- README: [README.md](https://github.com/NAEOS-foundation/naeos/blob/main/README.md)
- Architecture: [ARCHITECTURE-OVERVIEW.md](https://github.com/NAEOS-foundation/naeos/blob/main/ARCHITECTURE-OVERVIEW.md)
- Roadmap: [ROADMAP.md](https://github.com/NAEOS-foundation/naeos/blob/main/ROADMAP.md)
- Whitepaper: [WHITEPAPER.md](https://github.com/NAEOS-foundation/naeos/blob/main/WHITEPAPER.md)

## Kesimpulan

NAEOS sedang membangun engineering control plane untuk era software AI-native.

Proyek ini sudah memiliki substansi teknis yang nyata. Peluangnya adalah mengubah substansi itu menjadi platform developer yang tahan lama, control plane yang dipercaya, dan perusahaan infrastruktur jangka panjang yang kuat.
