---
title: "Dari Satu Spesifikasi Menjadi Proyek Go dan TypeScript"
description: "Demo CLI NAEOS yang dapat dijalankan ulang untuk validasi, pembuatan konteks AI, dan generasi artifact multi-bahasa."
date: 2026-09-08
author: "NAEOS Foundation"
categories: ["tutorial", "case-study"]
---

Studi kasus ini menggunakan [demo CLI NAEOS](https://github.com/NAEOS-foundation/naeos/tree/main/examples/demo-cli) yang tersimpan di repository. Ini adalah contoh yang dapat dijalankan ulang, bukan klaim tentang deployment production.

## Contohnya

Spesifikasi mendeskripsikan `demo-app` dengan:

- modul `auth` dan `api`,
- dependency `api` terhadap `auth`,
- service HTTP `gateway` pada port 8080,
- pola arsitektur hexagonal,
- target generasi Go dan TypeScript.

Intent engineering disimpan di [`spec.yaml`](https://github.com/NAEOS-foundation/naeos/blob/main/examples/demo-cli/spec.yaml), sedangkan pengaturan pipeline ada di [`naeos.yaml`](https://github.com/NAEOS-foundation/naeos/blob/main/examples/demo-cli/naeos.yaml).

## Menjalankannya

Dari root repository:

```bash
go build -o naeos ./cmd/naeos
./examples/demo-cli/run-demo.sh
```

Demo menjalankan tiga tahap:

```text
spesifikasi
    ├── validasi
    ├── context.md
    └── run
          └── generated/
```

Run saat ini melaporkan 61 artifact yang dihasilkan. Script kemudian memverifikasi context bundle, README proyek, Go module, dan package TypeScript, lalu membuat file `summary.md`.

## Arti output

Folder generated berisi folder modul, source code dan test Go, metadata package TypeScript, Dockerfile, konfigurasi CI, serta dokumentasi arsitektur. Semua artifact tersebut berasal dari spesifikasi yang sama dan dapat diperiksa sebelum digunakan dalam proyek.

Context bundle adalah ringkasan Markdown yang memuat proyek, modul, service, dan dependency graph. Developer dapat meninjaunya atau menggunakannya sebagai konteks untuk tool AI.

## Mengapa workflow ini penting

Nilai contoh ini ada pada alur yang dapat ditelusuri:

```text
spec.yaml
→ validasi
→ context bundle
→ generated artifacts
→ summary yang dapat diperiksa
```

Contoh ini sengaja lebih kecil daripada sistem nyata agar mudah direproduksi dan dijadikan titik awal eksperimen.

## Batasan bukti

- Command dan file yang disebutkan tersedia di repository NAEOS.
- Jumlah artifact dihasilkan oleh CLI dan dapat berubah ketika generator berkembang.
- Contoh ini tidak membuktikan deployment production, adopsi pelanggan, benchmark performa, atau compliance enterprise.

Mulai dari [source demo CLI](https://github.com/NAEOS-foundation/naeos/tree/main/examples/demo-cli) dan [panduan Getting Started](/id/docs/getting-started/).
