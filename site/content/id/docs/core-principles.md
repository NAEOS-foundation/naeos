---
title: Prinsip Inti
description: 8 prinsip engineering yang memandu semua keputusan desain NAEOS.
weight: 2
---

NAEOS dibangun di atas 8 prinsip inti yang memandu semua keputusan teknis dan arsitektural. Prinsip ini memastikan konsistensi, keandalan, dan ekstensibilitas di seluruh platform.

## 1. Spesifikasi adalah Single Source of Truth

Spesifikasi mendefinisikan apa yang ada, apa yang valid, dan apa yang dibutuhkan. Kode, dokumentasi, dan keputusan engineering harus tetap tersinkronisasi dengannya.

**Implikasi:**
- Semua perubahan dimulai dari spesifikasi
- Kode adalah produk dari spesifikasi, bukan sebaliknya
- Dokumentasi harus konsisten dengan spesifikasi
- Spesifikasi di-version dan tidak berubah setelah dipublikasikan

## 2. Arsitektur Mendahului Implementasi

Keputusan arsitektur harus dibuat dan didokumentasikan sebelum kode ditulis.

**Implikasi:**
- Buat Architecture Decision Record (ADR) untuk setiap keputusan signifikan
- Dokumentasikan trade-offs secara eksplisit
- Review arsitektur sebelum implementasi dimulai
- Gunakan NEIR sebagai model arsitektur kanonikal

## 3. Semua Bersifat Deklaratif

Semua konfigurasi dan perilaku didefinisikan secara deklaratif.

**Implikasi:**
- Spesifikasi YAML/JSON adalah satu-satunya cara mendefinisikan sistem
- Tidak ada skrip konfigurasi imperatif
- Semua state dapat direkonstruksi dari spesifikasi
- Output deterministik dari input yang sama

## 4. Semua adalah Artifact

Semua output pipeline adalah artifact yang dapat di-version, di-review, dan di-audit.

**Implikasi:**
- Kode, konfigurasi, dan dokumentasi adalah artifact
- Semua artifact didokumentasikan dan dapat ditelusuri
- Artifact dapat di-cache dan di-share
- Deduplikasi content-hash di artifact store

## 5. Semua Di-version

Semua komponen menggunakan semantic versioning.

**Implikasi:**
- Spesifikasi punya versi (minimum: 0.1.0)
- Artifact menyertakan metadata versi
- Dependensi di-lock
- Versi skema memungkinkan jalur migrasi

## 6. Semua Dapat Diperluas

Sistem dirancang untuk diperluas tanpa memodifikasi inti.

**Implikasi:**
- Sistem plugin untuk ekstensi kustom
- Adapter kustom untuk target output baru
- Profil kustom untuk kebutuhan spesifik industri
- Runtime WASM untuk eksekusi plugin yang ter-sandbox

## 7. Kernel Kecil, Ekstensi Kuat

Kernel minimalis; semua fitur ada di ekstensi.

**Implikasi:**
- Kernel hanya menangani layanan inti (registry, event bus, telemetry)
- Semua fitur ada di package terpisah
- Ekstensi dapat di-install dan di-uninstall secara independen
- Plugin adalah warga negara kelas satu

## 8. Review Manusia Sebelum Otomatisasi

Otomatisasi harus didahului oleh review manusia.

**Implikasi:**
- Review artifact sebelum deployment
- Evaluasi kebijakan untuk governance
- Audit trail untuk semua perubahan
- Guardrails mencegah aksi otomatis yang tidak diinginkan

## Menerapkan Prinsip Ini

Prinsip ini diterapkan melalui:

- **Validasi pipeline** — Validator memeriksa kepatuhan spesifikasi di setiap tahap
- **Kebijakan governance** — Aturan kebijakan memblokir atau memperingatkan pelanggaran
- **Audit trail** — Semua perubahan di-log dan dapat ditelusuri
- **Review kode** — Review manusia diperlukan sebelum merge

Lihat juga: [Arsitektur](/id/docs/architecture/), [Governance](/id/docs/governance/), [Pipeline Engine](/id/docs/pipeline-engine/)
