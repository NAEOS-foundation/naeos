---
title: Instalasi
description: Petunjuk instalasi terperinci untuk NAEOS di semua platform.
---

## Go Install (Direkomendasikan)

Membutuhkan Go 1.25+:

```bash
go install github.com/NAEOS-foundation/naeos/cmd/naeos@latest
```

## Docker

```bash
docker pull ghcr.io/naeos-foundation/naeos:latest
```

Verifikasi instalasi:

```bash
docker run --rm ghcr.io/naeos-foundation/naeos:latest naeos version
```

## Bangun dari Sumber

```bash
git clone https://github.com/NAEOS-foundation/naeos.git
cd naeos
make build
```

Biner akan tersedia di `./naeos`.

## Rilis Biner

Unduh biner terbaru untuk platform Anda dari [halaman Rilis GitHub](https://github.com/NAEOS-foundation/naeos/releases).

Platform yang didukung:
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

## Verifikasi Instalasi

```bash
naeos version
```

## Konfigurasi

NAEOS menggunakan file konfigurasi YAML. Buat dengan:

```bash
naeos init
```

Ini akan membuat `naeos.yaml` di direktori Anda saat ini dengan pengaturan default.