---
title: "Pipeline Engine: Bagaimana NAEOS Mengubah Spesifikasi Menjadi Sistem"
description: "Panduan teknis 9-stage DAG pipeline — dari parsing YAML hingga pembuatan kode multi-bahasa."
date: 2026-07-20
author: "NAEOS Foundation"
categories: ["tutorial"]
---

Setiap kali Anda menjalankan `naeos run --input-file spec.yaml`, sebuah pipeline DAG 9-stage bekerja. Setiap stage dapat diamati secara independen, dapat diperluas via plugin, dan dirancang untuk menangani spesifikasi dalam skala berapa pun.

## Pipeline Sekilas

```text
┌────────┐ ┌──────────┐ ┌────────┐ ┌───────┐ ┌─────────┐
│ Parse  │→│Normalize │→│Resolve │→│ Build │→│Validate │
└────────┘ └──────────┘ └────────┘ └───────┘ └─────┬───┘
                                                    │
┌────────┐ ┌──────────┐ ┌─────────┐ ┌────────┐    │
│ Export │←│ Compile  │←│Generate │←│Schedule│←───┘
└────────┘ └──────────┘ └─────────┘ └────────┘
```

## Stage 1: Parse

Pipeline membaca spesifikasi YAML/JSON dan mengubahnya menjadi AST. Mendukung interpolasi variabel (`${var}`), resolusi environment (`$env{VAR}`), komposisi multi-file via `$include`, dan validasi versi skema.

```yaml
project: blog-platform
$include:
  - ./shared/data-model.yaml
variables:
  region: us-east-1
```

## Stage 2: Normalize

Data mentah jarang konsisten. Normalizer mengubah notasi singkat ke bentuk kanonik, menerapkan nilai default, memvalidasi batasan tipe, dan menggabungkan file yang disertakan.

## Stage 3: Resolve

Referensi silang diselesaikan di sini. Resolver berjalan di seluruh pohon spesifikasi dan menyelesaikan referensi `$ref{path}`, referensi eksternal, dan graf dependensi.

## Stage 4: Build NEIR

NEIR (NAEOS Engineering Intermediate Representation) adalah model kanonik — representasi lengkap dan bertipe dari sistem Anda. Model ini mencakup graf modul dan service, pola arsitektur, kebutuhan infrastruktur, dan metadata tata kelola.

## Stage 5: Validate

Tujuh lapis validasi berjalan secara berurutan: kesesuaian skema, dependensi sirkuler, referensi cross-modul, aturan kebijakan, aturan bisnis via plugin, konvensi penamaan, dan konflik port.

## Stage 6: Schedule

Penjadwal DAG mengidentifikasi grup eksekusi paralel, melakukan topological sort, dan mendukung build inkremental. Modul tanpa interdependensi menghasilkan kode secara paralel.

## Stage 7: Generate

Generator berbasis template untuk setiap bahasa target membuat file proyek, scaffold modul, stub service, test, Dockerfile, dan konfigurasi CI.

```bash
naeos run --input-file spec.yaml --language go,typescript
```

## Stage 8: Compile

Kompiler AI mengubah NEIR menjadi set instruksi untuk enam platform AI: GitHub Copilot, Claude Code, Cursor, Gemini CLI, OpenAI Codex, dan OpenCode.

## Stage 9: Export

Semua hasil ditempatkan di direktori output: kode sumber, dokumentasi, file konteks AI, manifest deployment, dan laporan build.

## Menjalankan Pipeline

```bash
# Pipeline penuh
naeos run --input-file spec.yaml

# Lewati kompilasi AI
naeos run --input-file spec.yaml --skip compile

# Mode watch dengan hot-reload
naeos watch --input-file spec.yaml
```

Untuk referensi lengkap, lihat [dokumentasi Pipeline Engine](/docs/pipeline-engine/).
