---
title: "Bangun Platform E-Commerce dengan NAEOS: Tutorial End-to-End"
description: "Dari satu spesifikasi YAML menjadi platform e-commerce multi-bahasa yang tervalidasi lengkap dengan konteks AI — walkthrough NAEOS lengkap."
date: 2026-08-18
author: "NAEOS Foundation"
categories: ["tutorial"]
---

Dalam tutorial ini Anda akan membangun platform e-commerce lengkap dari satu spesifikasi YAML. Tanpa boilerplate yang ditulis manual, tanpa generator terpisah per bahasa — satu spec, dan NAEOS menurunkan sisanya: struktur modul, endpoint layanan, strategi deployment, dan bahkan set instruksi AI untuk coding assistant Anda.

Kita akan menggunakan `examples/spec-full.yaml` dari repository — spec referensi yang melatih sebagian besar pipeline NAEOS.

## Prasyarat

```bash
curl -fsSL https://naeos.dev/install.sh | sh
naeos version   # seharusnya mencetak 3.4.0
```

## Langkah 1: Buat Proyek

Wizard interaktif membuatkan struktur proyek:

```bash
naeos create
```

Masukkan `e-commerce-platform` sebagai nama proyek. Wizard menyiapkan struktur workspace dan konfigurasi awal.

## Langkah 2: Tulis Spesifikasi

Simpan ini sebagai `spec.yaml`:

```yaml
project: e-commerce-platform

modules:
  - name: auth
    path: ./internal/auth
    description: Authentication and authorization module
    dependencies: [core, user]

  - name: core
    path: ./internal/core
    description: Core business logic and shared utilities

  - name: user
    path: ./internal/user
    description: User management and profile module
    dependencies: [core]

  - name: order
    path: ./internal/order
    description: Order processing module
    dependencies: [core, user, payment]

  - name: payment
    path: ./internal/payment
    description: Payment processing module
    dependencies: [core]

services:
  - name: api-gateway
    kind: http
    port: 8080
    description: Main REST API gateway
    endpoints:
      - { method: POST, path: /auth/login, action: login }
      - { method: POST, path: /auth/register, action: register }
      - { method: GET,  path: /users, action: listUsers }
      - { method: GET,  path: /users/:id, action: getUser }
      - { method: POST, path: /orders, action: createOrder }
      - { method: GET,  path: /orders/:id, action: getOrder }
      - { method: POST, path: /payments, action: processPayment }

  - name: worker
    kind: worker
    port: 9090
    description: Background job processor

architecture:
  pattern: hexagonal
  description: Hexagonal architecture with clear separation of core logic from infrastructure

deployment:
  strategy: blue-green
  environments: [development, staging, production]

testing:
  strategy: unit
  coverage: "85%"

generation:
  languages: [go, typescript]
  output_dir: ./out
  module_dir: ./internal
```

Perhatikan apa yang *tidak* ada di spec: tidak ada detail framework HTTP, tidak ada pilihan ORM, tidak ada tata letak file. Pola arsitektur dan adapter bahasa yang menangani itu.

## Langkah 3: Validasi Sebelum Generate

Pipeline bisa memvalidasi spec tanpa menulis apa pun:

```bash
naeos validate --input-file spec.yaml
```

Ini menjalankan stage parse → normalize → resolve → build NEIR dan melaporkan isu, warning, serta error referensi silang. Perbaiki masalah apa pun sebelum generate.

## Langkah 4: Jalankan Pipeline

```bash
naeos run --input-file spec.yaml
```

Pipeline 11 stage berjalan: bangun model NEIR, validasi, evaluasi kebijakan, jadwalkan DAG, generate artifact melalui adapter Go dan TypeScript, review, dan tulis ke `./out/`. Anda akan melihat progres per stage dan ringkasan artifact yang ditulis.

Cek hasilnya:

```bash
find out -type f | head -20
```

Anda akan melihat skeleton modul Go (`internal/auth`, `internal/order`, ...), artifact TypeScript, Dockerfile, dan manifest deployment — semuanya konsisten, semuanya diturunkan dari model yang sama.

## Langkah 5: Iterasi dengan Watch Mode

Sambil menyempurnakan spec, biarkan NAEOS regenerate setiap kali berubah:

```bash
naeos watch --input-file spec.yaml
```

Simpan spec — pipeline berjalan ulang dan hanya menulis ulang yang berubah. Di sinilah pipeline cache v3.1.0 berperan: stage yang tidak berubah dipakai ulang lewat `--cache-dir`:

```bash
naeos run --input-file spec.yaml --cache-dir .naeos-cache
```

## Langkah 6: Beri Asisten AI Anda Konteks Nyata

Sekarang bagian serunya. Kompilasi model NEIR menjadi set instruksi yang benar-benar bisa dibaca alat AI Anda:

```bash
# GitHub Copilot
naeos ai compile --input-file spec.yaml --target copilot

# Claude Code
naeos ai compile --input-file spec.yaml --target claude

# Cursor
naeos ai compile --input-file spec.yaml --target cursor
```

Masing-masing menghasilkan file yang tepat untuk agen itu (`CLAUDE.md`, `.cursorrules`, `.github/copilot-instructions.md`, ...) berisi arsitektur, grafik dependensi, dan konvensi — sehingga asisten AI Anda berhenti menebak dan mulai membangun sesuai sistem yang sebenarnya.

Anda juga bisa membuat bundle konteks lengkap:

```bash
naeos context --input-file spec.yaml
```

## Langkah 7: Tes Kode yang Di-generate

Test runner multi-bahasa menjalankan suite tes hasil generate:

```bash
naeos test --input-file spec.yaml
```

## Yang Anda Bangun

Dalam sekitar sepuluh menit, dari satu spec:

- Backend Go dengan lima modul dan grafik dependensi (`auth`, `user`, `order`, `payment`, `core`)
- API gateway HTTP dengan endpoint bertipe
- Layanan worker background
- Artifact TypeScript untuk tooling
- Manifest Docker + deployment (strategi blue-green, tiga environment)
- Set instruksi AI untuk enam coding assistant
- Scaffolding tes unit dengan target coverage 85%

Ditulis manual, ini berminggu-minggu boilerplate — dan akan mulai menyimpang begitu di-merge. Dengan NAEOS, spec adalah sistemnya. Segala sesuatu yang lain diturunkan, tervalidasi, dan dapat direproduksi.

## Langkah Berikutnya

- Jelajahi contoh lain di [`examples/`](https://github.com/NAEOS-foundation/naeos/tree/main/examples)
- Baca [referensi bahasa spesifikasi](/id/docs/specification-language/)
- Coba [deep dive server MCP](/id/blog/mcp-server-deep-dive/) untuk menghubungkan spec ini ke editor Anda