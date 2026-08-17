---
title: "Kenapa Kami Membangun NAEOS: Rekayasa Terfragmentasi, dan Itu Masalah Struktural"
description: "Kisah di balik NAEOS — masalah fragmentasi dalam rekayasa perangkat lunak modern, jebakan spec-drift, dan mengapa kami membangun platform rekayasa deklaratif."
date: 2026-08-17
author: "NAEOS Foundation"
categories: ["announcement"]
---

Setiap tim rekayasa mengenal perasaan ini. Dokumentasi bilang satu hal, kode bilang hal lain, dan engineer terbaru di tim menghabiskan minggu-minggu pertama mereka untuk reverse-engineering sistem alih-alih membangun di atasnya.

Ini bukan kegagalan proses. Ini kegagalan struktural. Dan inilah alasan kami membangun NAEOS.

## Masalah Fragmentasi

Sistem perangkat lunak modern dibangun dari bagian-bagian yang tidak pernah benar-benar sepakat satu sama lain:

- **Banyak bahasa dan framework.** Produk tipikal mengirim layanan Go, frontend TypeScript, tooling Python, dan mungkin infrastruktur Java atau Rust — masing-masing dengan konvensi sendiri, boilerplate sendiri, drift sendiri.
- **Specification–implementation drift.** Dokumentasi dan kode menyimpang seiring waktu, dan tidak ada mekanisme otomatis yang menjamin keduanya tetap selaras. Spec bilang hexagonal; kodenya adalah tumpukan layer yang saling kusut.
- **Konteks engineering yang hilang.** Keputusan arsitektur hidup di ADR dan di kepala orang — tidak terdokumentasi, tidak dapat ditelusuri, dan hilang saat orangnya pergi.
- **Ledakan alat AI.** Enam coding assistant, enam format konteks berbeda. Tim memelihara file instruksi duplikat untuk Copilot, Claude, Cursor, dan lainnya — dan file-file itu tetap usang.
- **Governance yang tidak pernah dieksekusi.** Kebijakan ada sebagai dokumen statis. Tidak ada apa pun di pipeline yang menegakkannya.

Biayanya adalah rework, audit mahal, migrasi menyakitkan, kehilangan pengetahuan, dan kegagalan kepatuhan — semuanya menumpuk secara diam-diam.

## Tesis

NAEOS dibangun di atas satu klaim:

> **Spesifikasi adalah satu-satunya sumber kebenaran. Segala sesuatu — kode, dokumentasi, konfigurasi, konteks AI, artifact deployment — harus diturunkan dari spesifikasi melalui pipeline yang deterministik, tervalidasi, dan dapat diaudit.**

Kalimat itu adalah seluruh produknya. NAEOS (Nusantara Engineering & Architecture Operating System) adalah runtime rekayasa deklaratif yang mengambil spesifikasi YAML/JSON dan menjalankannya melalui pipeline 11 stage: parse, normalize, resolve, build NEIR, validate, build graph, evaluate policy, schedule, generate, review, dan write artifacts.

Semua yang di hilir diturunkan dari satu model — **NEIR** (NAEOS Engineering Intermediate Representation) — representasi kanonik dan versioned dari sistem Anda yang mencakup modul, layanan, pola arsitektur, API, storage, keamanan, target AI, dan deployment. Tidak ada generator independen yang saling menyimpang. Satu model, banyak adapter.

## Mengapa Kami Yakin Ini Berhasil

Tiga hal terjadi setelah kami berkomitmen pada tesis ini:

**1. Drift menjadi mustahil secara struktural.** Ketika layanan Go dan klien TypeScript di-generate dari model NEIR yang sama, keduanya tidak mungkin diam-diam tidak sepakat soal format error, kontrak API, atau konvensi penamaan. Konsistensi bukan lagi checklist review — ini invariant build.

**2. Asisten AI akhirnya mendapat konteks nyata.** Model NEIR yang sama dikompilasi menjadi set instruksi untuk GitHub Copilot, Claude Code, Cursor, Gemini CLI, Codex, dan OpenCode. Alat AI Anda berhenti membaca file terisolasi dan mulai membaca arsitektur — dependensi, pola, batasan, dan niat.

**3. Governance menjadi executable.** Aturan kebijakan mengevaluasi model NEIR sebelum satu baris kode pun di-generate. Pelanggaran tertangkap di level spesifikasi, bukan di code review. Lengkap dengan RBAC, audit trail, dan template kepatuhan SOC 2 / HIPAA / GDPR yang terintegrasi di pipeline.

## Di Mana Kami Sekarang

NAEOS tumbuh dari fondasi v0.1.0 pada Juli 2026 menjadi **v3.1.0** hari ini:

- **5 bahasa** — Go, TypeScript, Python, Java, Rust — dari satu spesifikasi
- **6 platform AI** — set instruksi yang dikompilasi dari model NEIR
- **67 command CLI** — run, validate, test, watch, diff, deploy, cloud, dan lainnya
- **56 dokumen NES** — proyek yang digerakkan spesifikasi, terdokumentasi seperti yang dibangunnya
- **Marketplace plugin WASM** — ekstensi pihak ketiga yang di-sandbox dan diverifikasi tanda tangannya
- **Schema registry, policy engine, server MCP, server LSP, dashboard** — operating system untuk engineering, bukan sekadar code generator

## Langkah Berikutnya

Kami sedang menuju target rilis ekosistem: marketplace plugin yang lebih kaya, integrasi kepatuhan yang lebih dalam, distributed builds, dan integrasi editor yang lebih ketat. Roadmap-nya publik — [ROADMAP.md](https://github.com/NAEOS-foundation/naeos/blob/main/ROADMAP.md) — dan kami senang mendapat bantuan Anda membentuknya.

## Coba

Cara tercepat memahami NAEOS adalah dengan membangun sesuatu:

```bash
curl -fsSL https://naeos.dev/install.sh | sh
naeos create
naeos run --input-file spec.yaml
```

Satu spesifikasi. Lima bahasa. Nol drift. Baca [whitepaper](/id/whitepaper/) untuk tesis lengkapnya, dan bergabunglah dengan kami di [GitHub](https://github.com/NAEOS-foundation/naeos).

Kami membangun NAEOS karena kami percaya rekayasa bisa deklaratif, tervalidasi, dan dapat diaudit — untuk manusia dan untuk AI. Tentukan sekali. Bangun di mana saja.