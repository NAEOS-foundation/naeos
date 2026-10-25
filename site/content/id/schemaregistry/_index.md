---
title: Schema Registry
description: JSON Schema berversi untuk memvalidasi spesifikasi sumber NAEOS dan dokumen model NEIR terserialisasi.
---

## Ringkasan

Registry menerbitkan [JSON Schema](https://json-schema.org/) terpisah untuk file spesifikasi sumber NAEOS dan dokumen model NEIR terserialisasi. Gunakan skema input-spec untuk file sumber `.naeos.yaml` dan `.naeos.json`. Gunakan skema model NEIR untuk JSON yang dihasilkan NAEOS.

## Skema Terbaru

Skema stabil saat ini:

- **Spesifikasi input:** [`/schemaregistry/spec/v1/spec.json`](/schemaregistry/spec/v1/spec.json)
- **Model NEIR terserialisasi:** [`/schemaregistry/v1/neir.json`](/schemaregistry/v1/neir.json)

| Kontrak | URL berversi | URL terbaru |
|---------|-------------|------------|
| Spesifikasi sumber | [`/schemaregistry/spec/v1/spec.json`](/schemaregistry/spec/v1/spec.json) | [`/schemaregistry/spec/latest.json`](/schemaregistry/spec/latest.json) |
| Model NEIR terserialisasi | [`/schemaregistry/v1/neir.json`](/schemaregistry/v1/neir.json) | [`/schemaregistry/latest.json`](/schemaregistry/latest.json) |

## Penggunaan

### Integrasi Editor

Tambahkan URI `$schema` input-spec ke file sumber Anda:

```yaml
# naeos.yaml
$schema: https://registry.naeos.dev/schemaregistry/spec/latest.json
project: my-project
modules:
  - name: core
    path: ./internal/core
```

Ini mengaktifkan IntelliSense dan validasi inline di editor seperti VS Code, JetBrains, dan lainnya yang mendukung JSON Schema.

### Validasi CLI

Gunakan CLI NAEOS untuk memvalidasi spesifikasi sumber terhadap skema input:

```bash
naeos schema validate spec.yaml
naeos schema validate spec.json --output json
```

### Penggunaan Programatis

Ambil skema input-spec secara programatis:

```bash
curl -s https://registry.naeos.dev/schemaregistry/spec/latest.json
```

## Versioning

Versi skema input-spec dan model NEIR adalah kontrak terpisah. Perubahan pada salah satunya tidak otomatis mengubah versi yang lain. Perubahan yang tidak kompatibel memerlukan path berversi baru untuk kontrak terkait; alias `latest` menunjuk ke versi stabil terbaru.

| Kontrak | Versi skema | Catatan |
|---------|-------------|---------|
| Input spec | v1 | YAML/JSON sumber; `project` berupa string (atau alias top-level `name`) |
| Model NEIR | v1 | Model terserialisasi; `project` berupa object |

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
