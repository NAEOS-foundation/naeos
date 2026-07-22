---
title: Pipeline Engine
description: Pipeline DAG 9-tahap — dari parsing hingga ekspor.
---

## Ikhtisar

Pipeline engine NAEOS adalah DAG (directed acyclic graph) 9-tahap yang mengubah spesifikasi YAML/JSON mentah menjadi output multi-bahasa yang tervalidasi. Setiap tahap dapat diamati secara independen dan dapat diperluas melalui plugin.

## Tahapan Pipeline

```text
┌────────┐ ┌──────────┐ ┌────────┐ ┌───────┐ ┌─────────┐
│ Parse  │→│Normalisasi│→│Resolusi│→│ Bangun│→│Validasi │
└────────┘ └──────────┘ └────────┘ └───────┘ └─────┬───┘
                                                   │
┌────────┐ ┌──────────┐ ┌─────────┐ ┌────────┐    │
│ Ekspor │←│ Kompilasi│←│Hasilkan │←│Jadwalkan│←───┘
└────────┘ └──────────┘ └─────────┘ └────────┘
```

### 1. Parse

Membaca dan parse file spesifikasi YAML/JSON. Mendukung:
- Interpolasi variabel (`${var}`)
- Resolusi variabel lingkungan (`$env{VAR}`)
- Komposisi multi-file via `$include`
- Validasi versi skema

### 2. Normalisasi

Normalisasi struktur data untuk pemrosesan hilir yang konsisten.

### 3. Resolusi

Menyelesaikan referensi silang dan dependensi.

### 4. Bangun

Membangun NEIR (NAEOS Engineering Intermediate Representation).

### 5. Validasi

Validasi komprehensif termasuk deteksi dependensi sirkuler.

### 6. Jadwalkan

Penjadwalan tugas berbasis DAG dengan grup eksekusi paralel.

### 7. Hasilkan

Generasi kode multi-bahasa dengan adapter per-bahasa.

### 8. Kompilasi

Kompilasi NEIR ke set instruksi AI untuk 6 platform.

### 9. Ekspor

Ekspor artefak, dokumentasi, dan manifes deployment.

## Konfigurasi Pipeline

Pipeline dapat dikonfigurasi melalui `naeos.yaml`:

```yaml
pipeline:
  stages:
    - parse
    - normalize
    - resolve
    - build
    - validate
    - schedule
    - generate
    - compile
    - export
  parallel: true
  cache: true
  output_dir: ./generated
```

## Menjalankan Pipeline

```bash
# Pipeline lengkap
naeos run --input-file spec.yaml

# Lewati tahap tertentu
naeos run --input-file spec.yaml --skip compile,export

# Jalankan hanya tahap tertentu
naeos run --input-file spec.yaml --only validate,generate

# Mode watch dengan hot-reload
naeos watch --input-file spec.yaml
```
