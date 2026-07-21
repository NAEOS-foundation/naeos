---
title: Panduan Awal
description: Pasang NAEOS dan jalankan pipeline pertama Anda dalam hitungan menit.
---

## Prasyarat

- Go 1.25+ (untuk metode `go install`)
- Terminal dengan pengetahuan dasar command-line

## Instalasi

Pilih salah satu metode berikut:

### Go Install

```bash
go install github.com/NAEOS-foundation/naeos/cmd/naeos@latest
```

### Docker

```bash
docker pull ghcr.io/naeos-foundation/naeos:latest
docker run --rm -v $(pwd):/workspace ghcr.io/naeos-foundation/naeos:latest naeos version
```

### Bangun dari Sumber

```bash
git clone https://github.com/NAEOS-foundation/naeos.git
cd naeos
go build ./cmd/naeos/
```

## Pipeline Pertama Anda

### 1. Buat file spesifikasi

Buat `spec.yaml`:

```yaml
project: my-app
modules:
  - name: auth
    path: ./auth
  - name: api
    path: ./api
    dependencies: [auth]
services:
  - name: gateway
    kind: http
    port: 8080
generation:
  languages: [go, typescript]
```

### 2. Inisialisasi konfigurasi

```bash
naeos init
```

### 3. Jalankan pipeline

```bash
naeos run --input-file spec.yaml
```

### 4. Generate konteks AI

```bash
naeos context --input-file spec.yaml
```

### 5. Kompilasi untuk asisten AI

```bash
naeos compile --all --input-file spec.yaml
```

## Langkah Selanjutnya

- Jelajahi [Referensi CLI](/id/docs/cli-reference/) untuk semua perintah yang tersedia
- Baca tentang [Arsitektur](/id/docs/architecture/) untuk memahami cara kerja NAEOS
- Lihat halaman [Fitur](/id/features/) untuk gambaran lengkap