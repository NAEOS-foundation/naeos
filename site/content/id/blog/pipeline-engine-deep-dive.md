---
title: "Pipeline Engine: Bagaimana NAEOS Mengubah Spesifikasi Menjadi Sistem"
description: "Panduan teknis 11-stage DAG pipeline — dari parsing YAML hingga pembuatan kode multi-bahasa."
date: 2026-07-20
author: "NAEOS Foundation"
categories: ["tutorial"]
---

Setiap kali Anda menjalankan `naeos run --input-file spec.yaml`, sebuah pipeline DAG 11-stage bekerja. Setiap stage dapat diamati secara independen, dapat diperluas via plugin, dan dirancang untuk menangani spesifikasi dalam skala berapa pun.

## Pipeline Sekilas

```text
┌────────┐ ┌──────────┐ ┌────────┐ ┌──────────┐ ┌─────────┐
│ Parse  │→│Normalize │→│Resolve │→│Build NEIR│→│Validate │
└────────┘ └──────────┘ └────────┘ └──────────┘ └────┬────┘
                                                     │
┌──────────┐ ┌────────┐ ┌───────┐ ┌──────────┐      │
│ Write    │←│ Review │←│Generate│←│Schedule  │←─────┘
│ Artifacts│ └────────┘ └───────┘ └──────────┘
└──────────┘
```

Dua stage koordinasi berjalan antara validasi dan penjadwalan: **Build Graph** (membangun DAG eksekusi) dan **Policy Evaluation** (menegakkan aturan governance).

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

Validator memeriksa struktur model NEIR: proyek harus terdefinisi, minimal satu modul dengan `name` dan `path` yang unik, service dengan nama valid dan port dalam rentang 0–65535 (port duplikat menjadi peringatan), serta pola arsitektur harus salah satu dari `layered`, `clean`, `hexagonal`, `microkernel`, `event-driven`, `cqrs`, `monolith`, `monolithic`, `microservices`, atau `serverless`. Secara opsional, spesifikasi dapat dicocokkan dengan NEIR JSON Schema bila sumber skema dikonfigurasi.

## Stage 6: Build Graph

Pipeline membangun DAG eksekusi dari model NEIR — setiap modul dan service menjadi node, dependensi menjadi edge. Graf ini menentukan apa yang dijalankan selanjutnya.

## Stage 7: Policy Evaluation

Aturan governance dievaluasi terhadap model. Pelanggaran menghentikan pipeline sebelum kode apa pun dihasilkan — governance ditegakkan pada waktu build, bukan setelahnya.

## Stage 8: Schedule

Penjadwal DAG mengidentifikasi grup eksekusi paralel, melakukan topological sort, dan mendukung build inkremental. Modul tanpa interdependensi menghasilkan kode secara paralel.

## Stage 9: Generate

Generator berbasis template untuk setiap bahasa target membuat file proyek, scaffold modul, stub service, test, Dockerfile, dan konfigurasi CI.

```bash
naeos run --input-file spec.yaml --language go --language typescript
```

## Stage 10: Review

Artifact yang dihasilkan melewati mesin review — proses linting ringan yang menghasilkan temuan yang melekat pada output pipeline sebelum apa pun ditulis.

## Stage 11: Write Artifacts

Semua hasil ditempatkan di direktori output: kode sumber, dokumentasi, file konteks AI, manifest deployment, dan laporan build. Manifest machine-readable juga ditulis agar sistem CI dapat memeriksa apa yang dihasilkan.

## Menjalankan Pipeline

```bash
# Pipeline penuh
naeos run --input-file spec.yaml

# Kompilasi instruksi AI (perintah terpisah)
naeos ai compile --input-file spec.yaml --target claude

# Mode watch dengan hot-reload
naeos watch --input-file spec.yaml
```

Untuk referensi lengkap, lihat [dokumentasi Pipeline Engine](/docs/pipeline-engine/).
