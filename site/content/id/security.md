---
title: Keamanan
description: Kebijakan keamanan dan pelaporan kerentanan untuk NAEOS.
---

## Kebijakan Keamanan

Proyek NAEOS menganggap serius masalah keamanan. Kami menghargai upaya Anda dalam melaporkan kerentanan secara bertanggung jawab.

## Melaporkan Kerentanan

**Mohon jangan melaporkan kerentanan keamanan melalui issue GitHub publik.**

Sebagai gantinya, laporkan melalui salah satu metode berikut:

1. **Email** — Kirim detail ke `security@naeos.dev`
2. **GitHub Private Vulnerability Reporting** — Gunakan fitur "Report a vulnerability" di tab Security repositori

## Cakupan

Berikut adalah yang termasuk dalam cakupan laporan keamanan:

- Alat CLI NAEOS (`github.com/NAEOS-foundation/naeos`)
- Situs web NAEOS (`naeos.dev`)
- Paket dan distribusi NAEOS resmi

## Kebijakan Pengungkapan

Kami mengikuti proses pengungkapan terkoordinasi:

1. Pelapor mengirimkan detail kerentanan
2. Maintainer mengakui penerimaan dalam 48 jam
3. Maintainer menyelidiki dan mengembangkan perbaikan

## Tindakan Keamanan

NAEOS menerapkan beberapa tindakan keamanan:

- **Tanpa telemetri** — NAEOS CLI tidak mengirimkan data penggunaan
- **Eksekusi lokal** — Semua generasi kode berjalan secara lokal di mesin Anda
- **Pemindaian dependensi** — Pemindaian kerentanan otomatis melalui Dependabot
- **Review kode** — Semua kontribusi ditinjau sebelum digabungkan
