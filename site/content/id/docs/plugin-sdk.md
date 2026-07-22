---
title: Plugin SDK
description: Perluas NAEOS dengan plugin kustom, generator, dan validator.
---

## Ikhtisar

NAEOS menyediakan Plugin SDK untuk memperluas platform dengan fungsionalitas kustom. Plugin dapat menambahkan generator kode baru, validator, deployer, analis, dan banyak lagi. Plugin dapat ditulis dalam Go (native) atau bahasa apa pun yang dikompilasi ke WASM.

## Tipe Plugin

| Tipe | Deskripsi |
|------|-----------|
| **Generator** | Hasilkan kode dalam bahasa kustom |
| **Validator** | Aturan validasi kustom |
| **Deployer** | Deploy ke platform kustom |
| **Analis** | Analisis dan pelaporan kustom |
| **Hook** | Hook siklus hidup untuk tahap pipeline |

## Memulai

### Prasyarat

- Go 1.25+ (untuk plugin native)
- TinyGo (untuk plugin WASM)

### Memasang Plugin

```bash
# Dari marketplace
naeos plugin install my-generator

# Dari file lokal
naeos plugin install ./path/to/plugin.wasm
```

### Mengelola Plugin

```bash
# Daftar plugin terpasang
naeos plugin list

# Perbarui plugin
naeos plugin update my-generator

# Hapus plugin
naeos plugin remove my-generator
```

## Konfigurasi Plugin

Plugin dapat menerima konfigurasi melalui file spesifikasi:

```yaml
plugins:
  - name: my-generator
    config:
      template_dir: ./templates
      output_style: compact
      features: [typescript, openapi]
```

## Publikasi ke Marketplace

```bash
# Paket plugin Anda
naeos plugin package ./my-generator --output my-generator.tar.gz

# Publikasikan
naeos marketplace publish my-generator.tar.gz
```

## Referensi SDK

Plugin SDK menyediakan:

- `sdk.Context` — Konteks pipeline dengan konfigurasi, logging, dan akses file
- `sdk.Register()` — Daftarkan implementasi plugin Anda
- `sdk.ReadNEIR()` — Deserialisasi model NEIR dari memori WASM
- `sdk.WriteResult()` — Serialisasi hasil kembali ke memori WASM
- `sdk.Artifact` — Output file yang dihasilkan
- `sdk.Issue` — Isu validasi dengan tingkat keparahan, lokasi, dan pesan

## Praktik Terbaik

- Uji plugin dengan `naeos test --plugin my-plugin`
- Gunakan semantic versioning untuk rilis plugin
- Sertakan manifes `plugin.yaml` dengan metadata
- Manfaatkan logging bawaan via `sdk.Context.Logger`
- Tangani error dengan baik dan kembalikan pesan yang bermakna
