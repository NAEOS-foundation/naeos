---
title: "Mengapa Rekayasa Deklaratif? Kasus untuk Pengembangan Berbasis Spesifikasi"
description: "Generator kode tradisional memiliki batas. Rekayasa deklaratif dengan NAEOS mengambil pendekatan yang fundamentally berbeda."
date: 2026-07-10
categories: ["concept"]
---

Kebanyakan tim rekayasa memiliki generator kode di suatu tempat. Skrip yang membuat kerangka layanan baru, template yang mencetak endpoint REST, langkah CI yang menghubungkan boilerplate. Alat-alat ini berfungsi — sampai tidak.

Masalahnya bukan kode yang di-generate sendiri. Masalahnya adalah kebanyakan generator beroperasi di **level file**. Mereka menghasilkan file, bukan sistem. Dan saat Anda memerlukan perhatian lintas batas — tipe bersama antar layanan, penanganan error yang konsisten, konfigurasi deployment yang sejalan — generator mulai bergerak.

## Jebakan Level File

Pertimbangkan pengaturan microservices tipikal. Anda mungkin memiliki:

- Generator layanan Go untuk API backend
- Generator TypeScript untuk klien frontend
- Generator Terraform untuk infrastruktur
- Generator Docker Compose untuk pengembangan lokal

Masing-masing menghasilkan file secara independen. Masing-masing memiliki format config sendiri, konvensi sendiri, edge case sendiri. Seiring waktu, generator ini bergerak. Layanan Go menggunakan format error yang satu, klien TypeScript mengharapkan format yang lain. Konfigurasi Terraform merujuk resource dengan konvensi penamaan yang hanya cocok jika Anda menjalankan generator dengan urutan yang benar.

Ini adalah **jebakan level file**: generator yang menghasilkan artifact individual tanpa kesadaran tentang sistem yang menjadi bagiannya.

## Masuk NEIR

NAEOS mengambil pendekatan berbeda. Alih-alih menghasilkan file, ia membangun representasi perantara yang disebut **NEIR** (NAEOS Engineering Intermediate Representation) — model lengkap sistem Anda yang menangkap:

- **Modul** dan grafik dependensi mereka
- **Layanan** dan antarmuka mereka
- **Pola arsitektur** dan batasan
- **Target generasi** dan adapter per-bahasa

Pipeline tidak pergi dari YAML → file. Ia pergi dari YAML → NEIR → model tervalidasi → adapter per-bahasa → file. Setiap output diturunkan dari sumber kebenaran yang sama, dengan validasi yang sama, resolusi dependensi yang sama, jaminan struktural yang sama.

## Mengapa Ini Penting

Saat Anda generate dari NEIR alih-alih dari template, beberapa hal berubah:

**Konsistensi menjadi struktural.** Tipe error, kontrak API, dan konvensi penamaan didefinisikan sekali dalam spesifikasi dan diterapkan di semua output yang di-generate. Anda tidak bisa tanpa sengaja menghasilkan layanan Go dengan format error yang berbeda dari klien TypeScript — keduanya diturunkan dari model NEIR yang sama.

**Generasi lintas bahasa menjadi natural.** Satu spesifikasi bisa menghasilkan kode Go, TypeScript, Python, Java, dan Rust. Bukan lima generator terpisah — satu model, lima adapter. Setiap adapter tahu bagaimana mengekspresikan konsep NEIR dalam bahasa targetnya.

**Asisten AI mendapat konteks nyata.** Model NEIR bisa dikompilasi menjadi set instruksi untuk GitHub Copilot, Claude Code, Cursor, dan lainnya. Alat AI Anda tidak hanya melihat file individual — mereka melihat seluruh arsitektur, grafik dependensi, dan niat desain.

**Governance terintegrasi.** Aturan kebijakan mengevaluasi model NEIR sebelum kode apa pun di-generate. Pelanggaran ditangkap di level spesifikasi, bukan dalam code review.

## Pergeseran Mental Model

Rekayasa deklaratif bukan tentang menulis lebih sedikit kode. Ini tentang menulis kode yang diturunkan dari model presisi dan tervalidasi tentang apa yang Anda bangun.

Pikirkan seperti ini:

| Tradisional | Deklaratif |
|------------|-----------|
| Tulis kode → harap konsisten | Tulis spesifikasi → generate kode konsisten |
| Template menghasilkan file | NEIR menghasilkan sistem |
| Drift terjadi diam-diam | Drift secara struktural mustahil |
| AI membaca file | AI membaca arsitektur |

## Memulai

Jika Anda penasaran dengan pendekatan berbasis spesifikasi, cara tercepat untuk memahami adalah mencobanya:

```bash
# Instal NAEOS
go install github.com/NAEOS-foundation/naeos/cmd/naeos@latest

# Buat spesifikasi
cat > spec.yaml << 'EOF'
project: my-first-service
modules:
  - name: api
    path: ./api
services:
  - name: rest-api
    kind: rest
    port: 8080
architecture:
  pattern: microservices
generation:
  languages: [go, typescript]
EOF

# Jalankan pipeline
naeos run --input-file spec.yaml
```

Outputnya bukan hanya file — itu adalah struktur proyek lengkap dengan modul yang sadar dependensi, antarmuka bertipe, dan set instruksi AI, semuanya diturunkan dari spesifikasi 15 baris.

## Apa Selanjutnya

Kami sedang mengerjakan integrasi yang lebih dalam: dukungan LSP untuk file spesifikasi, ekstensi VS Code dengan validasi real-time, dan pipeline deployment cloud yang lebih terintegrasi. Spesifikasi menjadi antarmuka untuk seluruh siklus hidup pengembangan perangkat lunak.

Rekayasa deklaratif bukan solusi universal. Tapi untuk tim yang membangun sistem multi-bahasa, multi-layanan, ini adalah cara yang fundamentally lebih baik untuk memikirkan kode yang di-generate. Tentukan sekali. Bangun di mana saja.
