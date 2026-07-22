---
title: Fitur
description: Jelajahi rangkaian fitur lengkap NAEOS — dari parsing spesifikasi hingga生成 kode berbantuan AI.
---

NAEOS menyediakan platform rekayasa lengkap dengan kemampuan berikut:

## Pipeline Engine

Pipeline DAG 9-tahap adalah jantung dari NAEOS:

1. **Parse** — Baca dan parse spesifikasi YAML/JSON dengan interpolasi variabel
2. **Normalisasi** — Normalisasi struktur data untuk pemrosesan yang konsisten
3. **Resolusi** — Selesaikan referensi silang, dependensi, dan include eksternal
4. **Bangun** — Bangun NEIR (NAEOS Engineering Intermediate Representation)
5. **Validasi** — Validasi komprehensif termasuk deteksi dependensi sirkuler
6. **Jadwalkan** — Penjadwalan tugas berbasis DAG dengan grup eksekusi paralel
7. **Hasilkan** —生成 kode multi-bahasa (Go, TypeScript, Python, Java, Rust)
8. **Kompilasi** — Kompilasi NEIR ke set instruksi AI untuk 6 platform
9. **Ekspor** — Ekspor artefak, dokumentasi, dan manifes deployment

## Spec Language v2

Bahasa Spesifikasi NAEOS v2 menyediakan fitur yang kaya:

- `${var}` — Interpolasi variabel
- `$env{VAR}` — Resolusi variabel lingkungan
- `$ref{path}` — Resolusi referensi silang
- `$include{file}` — Komposisi spesifikasi multi-file
- `$fn{name(args)}` — Fungsi kustom (upper, lower, slug, default, len, coalesce)
- `$if{condition}` / `$endif` — Bagian bersyarat
- Versioning skema dengan pengecekan otomatis (minimum v0.1.0)

## Model NEIR

NAEOS Engineering Intermediate Representation adalah model kanonikal yang merepresentasikan seluruh sistem. Mencakup:

- Metadata dan konfigurasi proyek
- Pola arsitektur (microservices, serverless, monolithic, hexagonal)
- Model domain dan bounded context
- Struktur modul dan dependensi
- Definisi layanan dengan endpoint dan port
- Kontrak API (REST, GraphQL, WebSocket)
- Konfigurasi penyimpanan dan database
- Sumber daya infrastruktur dan cloud
- Kebijakan dan aturan keamanan
- Pengaturan integrasi AI
- Persyaratan dokumentasi
- Target deployment (Kubernetes, Docker, serverless)
- Strategi pengujian
- Konfigurasi pipeline CI/CD

## Generator Kode

Hasilkan kode siap-produksi dengan adapter per-bahasa:

- **Go** — Pola standard library, net/http, testing
- **TypeScript** — Pola Express/NestJS, type safety
- **Python** — Pola FastAPI, type hints
- **Java** — Pola Spring Boot, JUnit 5
- **Rust** — Pola Axum, async/await

## Kompiler AI

Ubah NEIR menjadi set instruksi AI untuk 6 asisten coding:

- **GitHub Copilot** — `.github/copilot-instructions.md`
- **Claude Code** — `CLAUDE.md`
- **Cursor** — `.cursorrules`
- **Gemini CLI** — `.gemini/CONFIG.md`
- **Codex** — `AGENTS.md`
- **OpenCode** — `AGENTS.md`

## Tata Kelola & Kebijakan

Mesin kebijakan bawaan dan kerangka tata kelola:

- 7 operator kebijakan untuk definisi aturan
- 5 aturan kebijakan default
- Sistem izin RBAC
- Jejak audit dengan ketelusuran penuh
- Alur kerja tinjauan dan persetujuan artefak
- Evaluator kebijakan dengan output terstruktur

## Marketplace

Publikasikan, temukan, dan pasang ekstensi:

- **Marketplace Profil** — Profil industri (SaaS, AI Agent, FinTech, Healthcare, Government)
- **Marketplace Plugin** — Plugin WASM dan native
- **Marketplace Template** — Template dan scaffold proyek
- Verifikasi tanda tangan SHA-256 untuk keamanan

## Alat Pengembang

Alat tambahan untuk alur kerja pengembangan yang lengkap:

- 35+ perintah CLI
- Mode watch untuk hot-reload
- Mesin diff untuk perbandingan spesifikasi
- Mesin migrasi untuk versioning skema
- Generator dokumentasi
- Penguji multi-bahasa
- Server MCP untuk integrasi agen AI
- Dashboard WebSocket untuk monitoring real-time
- Eksekusi tugas terdistribusi
- Event sourcing dan logging audit