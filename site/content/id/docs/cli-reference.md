---
title: Referensi CLI
slug: cli-reference
description: Referensi perintah lengkap untuk semua perintah CLI NAEOS.
---

NAEOS menyediakan CLI lengkap dengan perintah untuk setiap tahap pipeline rekayasa.

## Perintah Inti

| Perintah | Deskripsi |
|----------|-----------|
| `naeos run` | Jalankan pipeline lengkap |
| `naeos validate` | Validasi spesifikasi menggunakan pipeline NAEOS |
| `naeos ai compile` | Kompilasi spesifikasi ke set instruksi AI |
| `naeos context` | Generate bundel konteks AI dari spesifikasi |
| `naeos test` | Jalankan tes untuk kode yang dihasilkan |
| `naeos docgen` | Generate dokumentasi dari spesifikasi |

## Perintah Proyek & Spesifikasi

| Perintah | Deskripsi |
|----------|-----------|
| `naeos init` | Inisialisasi proyek NAEOS baru atau generate config |
| `naeos create` | Wizard pembuatan proyek interaktif |
| `naeos scaffold` | Generate scaffold proyek starter |
| `naeos import` | Impor spesifikasi dari format HCL ke NAEOS YAML/JSON |
| `naeos export` | Ekspor artefak yang dihasilkan ke direktori |
| `naeos audit` | Audit keamanan file yang dihasilkan atau sumber |
| `naeos diff` | Bandingkan artefak yang dihasilkan dengan direktori output |
| `naeos repair` | Perbaiki direktori output NAEOS |
| `naeos lint` | Lint file spesifikasi |
| `naeos inspect` | Periksa hasil pipeline NAEOS |
| `naeos preview` | Pratinjau artefak yang dihasilkan tanpa menulisnya |
| `naeos distributed` | Jalankan tugas pipeline dalam mode terdistribusi di banyak worker |

## Build, Deploy & CI/CD

| Perintah | Deskripsi |
|----------|-----------|
| `naeos build` | Build artefak dari spesifikasi |
| `naeos cicd` | Generate konfigurasi pipeline CI/CD |
| `naeos cloud` | Perintah deployment cloud |
| `naeos deploy` | Deploy output pipeline ke lingkungan target |
| `naeos gateway` | Manajemen API gateway |
| `naeos graphql` | Mulai server API GraphQL |
| `naeos helm` | Scaffolding dan validasi Helm chart |
| `naeos airgap` | Buat, periksa, dan impor bundle distribusi air-gapped |

## Perintah Manajemen

| Perintah | Deskripsi |
|----------|-----------|
| `naeos mcp` | Mulai server MCP (Model Context Protocol) |
| `naeos marketplace` | Jelajahi dan instal template, profil, dan plugin |
| `naeos profile` | Kelola profil industri spesifik |
| `naeos plugin` | Kelola plugin NAEOS |
| `naeos template` | Kelola template generasi, pustaka prompt, dan marketplace template |
| `naeos workspace` | Kelola workspace multi-modul |
| `naeos artifacts` | Kelola artefak proyek yang dihasilkan |
| `naeos migrate` | Kelola migrasi skema spec |
| `naeos migration` | Manajemen migrasi database |
| `naeos rollback` | Kembalikan ke snapshot artefak sebelumnya |
| `naeos lock` | Kelola file lock untuk build yang reproduktif |

## Monitoring & Operasi

| Perintah | Deskripsi |
|----------|-----------|
| `naeos status` | Tampilkan status pipeline, sistem, dan proyek |
| `naeos doctor` | Jalankan diagnostik pada lingkungan dan konfigurasi NAEOS |
| `naeos watch` | Pantau perubahan spesifikasi dan jalankan ulang pipeline |
| `naeos kernel` | Periksa kernel NAEOS dan service registry |
| `naeos dashboard` | Mulai dashboard web NAEOS |
| `naeos api` | Mulai server REST API NAEOS |
| `naeos health` | Jalankan pemeriksaan kesehatan sistem dan diagnostik |
| `naeos monitor` | Mulai server monitoring dengan metrik Prometheus |
| `naeos observability` | Manajemen observability dan telemetri |
| `naeos events` | Perintah event sourcing untuk audit trail dan replay pipeline |
| `naeos history` | Tampilkan riwayat eksekusi pipeline dari event tersimpan |
| `naeos serve` | Jalankan NAEOS sebagai daemon produksi (HTTP/TLS, systemd) |

## Keamanan & Kepatuhan

| Perintah | Deskripsi |
|----------|-----------|
| `naeos auth` | Manajemen autentikasi dan otorisasi |
| `naeos security` | Manajemen keamanan dan rahasia (secrets) |
| `naeos compliance` | Pelaporan kepatuhan dan ekspor audit log |
| `naeos config` | Manajemen dan resolusi konfigurasi (env/file/K8s secret/Vault) |
| `naeos policy` | Kelola kebijakan governance |
| `naeos control` | Evaluasi control plane governance |
| `naeos evidence` | Evidence store immutable — audit trail governance |
| `naeos verify` | Verifikasi independen bukti governance |
| `naeos runtime` | Runtime gateway untuk eksekusi agen terotorisasi |
| `naeos sign` | Tanda tangani dan verifikasi artefak (Ed25519) |
| `naeos sbom` | Generasi dan verifikasi Software Bill of Materials |

## Data & Integrasi

| Perintah | Deskripsi |
|----------|-----------|
| `naeos db` | Manajemen koneksi dan migrasi database |
| `naeos broker` | Manajemen message broker |
| `naeos ws` | Mulai server WebSocket untuk pembaruan real-time |
| `naeos lsp` | Mulai server NEIR Language Server Protocol |
| `naeos supabase` | Manajemen backend Supabase |

## Pengalaman Developer & Utilitas

| Perintah | Deskripsi |
|----------|-----------|
| `naeos ai` | Perintah bantuan AI |
| `naeos dx` | Alat pengalaman developer |
| `naeos tui` | Alat antarmuka terminal (TUI) |
| `naeos schema` | Operasi NEIR schema registry |
| `naeos search` | Manajemen mesin pencarian full-text |
| `naeos perf` | Alat optimasi performa |
| `naeos benchmark` | Jalankan benchmark pipeline |
| `naeos docs` | Generate dokumentasi proyek |
| `naeos completion` | Generate skrip penyelesaian shell |
| `naeos version` | Tampilkan versi NAEOS |

Untuk penggunaan detail setiap perintah, jalankan `naeos <command> --help`.

## Unduhan

- [PDF Referensi CLI](/downloads/naeos-cli-reference.pdf)
