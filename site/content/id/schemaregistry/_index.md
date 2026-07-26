---
title: Schema Registry
description: Registry skema JSON untuk format spesifikasi NEIR — skema berversi untuk memvalidasi dokumen NAEOS Engineering Intelligence Representation.
---

## Ringkasan

NEIR Schema Registry menghosting definisi [JSON Schema](https://json-schema.org/) berversi untuk format spesifikasi NAEOS Engineering Intelligence Representation (NEIR). Skema ini memungkinkan autocompletion IDE, validasi, dan dukungan alat untuk file `.naeos.yaml` dan `.naeos.json`.

## Skema Terbaru

Skema stabil saat ini adalah **[v1](v1/neir.json)**.

| Versi | URL Skema | Status |
|-------|-----------|--------|
| v1 | [`/schemaregistry/v1/neir.json`](v1/neir.json) | Stabil |
| latest | [`/schemaregistry/latest.json`](../latest.json) | Stabil terbaru |

## Penggunaan

### Integrasi Editor

Tambahkan field `$schema` ke file spesifikasi NEIR Anda:

```yaml
# naeos.yaml
$schema: https://naeos.dev/schemaregistry/latest.json
project: my-project
modules:
  - name: core
    path: ./internal/core
```

Ini mengaktifkan IntelliSense dan validasi inline di editor seperti VS Code, JetBrains, dan lainnya yang mendukung JSON Schema.

### Validasi CLI

Gunakan CLI NAEOS untuk memvalidasi spesifikasi terhadap skema registry:

```bash
naeos schema validate spec.yaml
naeos schema validate spec.json --output json
```

### Penggunaan Programatis

Ambil skema secara programatis:

```bash
curl -s https://naeos.dev/schemaregistry/latest.json
```

## Versioning

Versi skema mengikuti versi spesifikasi NEIR. Perubahan yang tidak kompatibel ke belakang pada model NEIR menghasilkan versi skema baru. Penambahan kecil (field opsional baru) bersifat aditif dalam satu versi.

| NEIR Version | Schema Version | Catatan |
|-------------|---------------|---------|
| 1.x | v1 | Skema stabil awal |

## Konten Skema

Skema JSON NEIR mendefinisikan bagian-bagian tingkat atas ini:

- `project` — Metadata proyek (nama, versi, deskripsi, lisensi, penulis)
- `architecture` — Pola arsitektur dan lapisan
- `domain` — Model domain dengan bounded context, aggregate, entity
- `modules` — Definisi modul dengan path dan paket
- `components` — Katalog komponen dengan jenis dan dependensi
- `services` — Definisi layanan dengan endpoint dan port
- `apis` — Spesifikasi API dengan protokol dan skema
- `storage` — Backend penyimpanan (SQL, NoSQL, cache, queue, blob)
- `infrastructure` — Penyedia dan sumber daya infrastruktur
- `security` — Autentikasi, otorisasi, enkripsi, rahasia
- `ai` — Integrasi model AI, prompt, bundel konteks
- `documentation` — Panduan dokumentasi, referensi, ADR
- `deployment` — Strategi deployment, lingkungan, scaling
- `testing` — Strategi pengujian, framework, cakupan, fixture
- `metadata` — Metadata spesifikasi dengan pelacakan versi dan timestamp
- `generation` — Konfigurasi generasi kode
