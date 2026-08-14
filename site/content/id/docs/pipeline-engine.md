---
title: Pipeline Engine
description: Pipeline DAG 11-tahap — dari parsing hingga output artifact.
---

## Ikhtisar

Pipeline engine NAEOS adalah DAG (directed acyclic graph) 11-tahap yang mengubah spesifikasi YAML/JSON mentah menjadi output multi-bahasa yang tervalidasi. Setiap tahap dapat diamati secara independen dan dapat diperluas melalui plugin.

## Tahapan Pipeline

```text
┌────────┐ ┌──────────┐ ┌────────┐ ┌──────────┐ ┌─────────┐
│ Parse  │→│Normalisasi│→│Resolusi│→│Build NEIR│→│Validasi │
└────────┘ └──────────┘ └────────┘ └──────────┘ └────┬────┘
                                                     │
┌──────────┐ ┌────────┐ ┌───────┐ ┌──────────┐      │
│ Write    │←│ Review │←│Hasilkan│←│Jadwalkan │←─────┘
│ Artifacts│ └────────┘ └───────┘ └──────────┘
└──────────┘
```

Alur eksekusi juga menyertakan dua tahap koordinasi antara validasi dan penjadwalan:

```text
Validasi → Build Graph → Policy Evaluation → Jadwalkan → Hasilkan → Review → Write Artifacts
```

### 1. Parse

Membaca dan parse file spesifikasi YAML/JSON. Mendukung:
- Interpolasi variabel (`${var}`)
- Resolusi variabel lingkungan (`$env{VAR}`)
- Komposisi multi-file via `$include`
- Validasi versi skema

### 2. Normalisasi

Normalisasi struktur data untuk pemrosesan hilir yang konsisten:
- Mengubah notasi singkat ke bentuk kanonik
- Menerapkan nilai default
- Memvalidasi batasan tipe
- Menggabungkan file yang disertakan

### 3. Resolusi

Menyelesaikan referensi silang dan dependensi:
- Resolusi `$ref{path}` di seluruh pohon spesifikasi
- Resolusi referensi eksternal
- Deteksi referensi sirkuler
- Konstruksi graf dependensi

### 4. Build NEIR

Membangun NEIR (NAEOS Engineering Intermediate Representation):
- Menyusun model kanonik dari data ternormalisasi
- Menerapkan pola dan template arsitektur
- Mengonstruksi graf modul dan service
- Menghasilkan pengenal internal dan metadata

### 5. Validasi

Validasi komprehensif meliputi:
- Validasi model semantik
- Pemeriksaan kesesuaian skema opsional (via schema registry)
- Validasi referensi lintas-modul
- Validasi aturan bisnis via plugin

### 6. Build Graph

Membangun DAG eksekusi dari model NEIR:
- Konstruksi graf modul dan service
- Urutan dependensi untuk penjadwalan hilir

### 7. Policy Evaluation

Mengevaluasi aturan kebijakan governance terhadap model yang dihasilkan:
- Evaluasi aturan multi-kebijakan
- Menghentikan pipeline jika terjadi pelanggaran

### 8. Jadwalkan

Penjadwalan tugas berbasis DAG:
- Identifikasi grup eksekusi paralel
- Topological sort tugas dependen
- Penjadwalan sadar sumber daya
- Dukungan build inkremental

### 9. Hasilkan

Generasi kode multi-bahasa:
- Output berbasis template per bahasa
- Adapter per-bahasa (Go, TypeScript, Python, Java, Rust, FastAPI, Actix Web)
- Generasi konkuren lintas modul
- Pembuatan manifest artifact
- Caching hasil opsional

### 10. Review

Meninjau artifact yang dihasilkan:
- Review dan linting artifact otomatis
- Hasil review dilampirkan pada hasil pipeline

### 11. Write Artifacts

Menulis semua artifact ke disk:
- Kode sumber yang dihasilkan
- Bundel dokumentasi
- File konteks AI
- Manifes deployment
- Laporan dan ringkasan build

## Konfigurasi Pipeline

Pipeline dikonfigurasi melalui `naeos.yaml` (atau `naeos.json`):

```yaml
pipeline:
  name: demo            # nama pipeline
  output_dir: ./output  # tempat artifact ditulis
  verbose: false        # logging stage verbose
  language: [go]        # target generasi
```

Caching diaktifkan per-jalankan dengan `--cache-dir` (lihat di bawah); bukan
kunci pada file konfigurasi.

## Menjalankan Pipeline

```bash
# Pipeline lengkap
naeos run --input-file spec.yaml

# Pipeline terbatas dengan preview (tidak menulis ke disk)
naeos run --input-file spec.yaml --dry-run

# Target bahasa tertentu saja
naeos run --input-file spec.yaml --language go --language typescript

# Mengaktifkan caching antar eksekusi
naeos run --input-file spec.yaml --cache-dir .naeos-cache

# Profiling pipeline
naeos run --input-file spec.yaml --profile --profile-out profile.json

# Mode watch dengan hot-reload
naeos watch --input-file spec.yaml
```

## Terkait

- [Referensi CLI](/id/docs/cli-reference/) — referensi flag lengkap untuk `naeos run` dan `naeos watch`
- [Arsitektur](/id/docs/architecture/) — pipeline dalam platform yang lebih luas
- [Plugin SDK](/id/docs/plugin-sdk/) — memperluas pipeline dengan hook dan plugin