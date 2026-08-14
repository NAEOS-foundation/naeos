---
title: Memulai dengan NAEOS — Dari Spesifikasi ke Kode dalam 5 Menit
description: Tutorial praktis menunjukkan cara mendeklarasikan proyek microservices dengan NAEOS dan menghasilkan kode Go + TypeScript dalam waktu kurang dari lima menit.
date: 2026-07-26
author: "NAEOS Foundation"
categories: ["tutorial"]
---

NAEOS memungkinkan Anda mendeskripsikan seluruh proyek — modul, layanan, API, dependensi — dalam satu spesifikasi YAML, kemudian menghasilkan kode siap-produksi untuk bahasa dan platform pilihan Anda.

Dalam tutorial ini Anda akan membangun proyek microservices sederhana dengan gateway API Go dan layanan auth TypeScript.

## Prasyarat

- [Go 1.25+](https://go.dev/dl/)
- Terminal

Instal NAEOS:

```bash
go install github.com/NAEOS-foundation/naeos/cmd/naeos@latest
```

## Langkah 1 — Buat Spesifikasi

Mulai dengan menginisialisasi proyek:

```bash
naeos init --template microservices my-app
cd my-app
```

Ini membuat file `naeos.yaml`. Buka dan ganti isinya dengan:

```yaml
project: my-app
version: "1.0"
description: Proyek demo microservices

modules:
  - name: gateway
    path: ./gateway
  - name: auth
    path: ./auth
    dependencies: []
  - name: api
    path: ./api
    dependencies: [auth]

services:
  - name: gateway
    kind: http
    port: 8080
    module: gateway
  - name: auth-service
    kind: grpc
    port: 50051
    module: auth

generation:
  languages: [go, typescript]
```

## Langkah 2 — Validasi

Periksa bahwa spesifikasi Anda valid:

```bash
naeos validate
```

Anda akan melihat `✓ Specification is valid`.

## Langkah 3 — Hasilkan Kode

Jalankan pipeline penuh untuk menghasilkan kode Go dan TypeScript:

```bash
naeos run --input naeos.yaml
```

Dalam hitungan detik, NAEOS menghasilkan:

```
out/
├── go/
│   ├── gateway/
│   │   ├── main.go
│   │   ├── handler.go
│   │   ├── go.mod
│   │   └── Dockerfile
│   └── auth/
│       ├── server.go
│       ├── proto/
│       ├── go.mod
│       └── Dockerfile
└── typescript/
    ├── api/
    │   ├── src/
    │   ├── package.json
    │   └── tsconfig.json
    └── auth/
        ├── src/
        ├── package.json
        └── tsconfig.json
```

## Langkah 4 — Pratinjau NEIR

Lihat representasi antara yang dibangun NAEOS dari spesifikasi Anda:

```bash
naeos run --input naeos.yaml --verbose
```

Model NEIR mencakup dependensi yang telah di-resolve, endpoint layanan, dan metadata arsitektur yang digunakan generator untuk menghasilkan kode yang benar dan konsisten di semua bahasa target.

## Selanjutnya

- Jelajahi [dokumentasi](/id/docs/getting-started/) untuk fitur spesifikasi lanjutan
- Coba template berbeda: `naeos init --template fullstack my-app`
- Tambahkan kompilasi konteks AI dengan `naeos ai compile`
- Bergabung dengan [komunitas di GitHub](https://github.com/NAEOS-foundation/naeos)
