---
title: Referensi Cepat
slug: quick-reference
description: Perintah, pola, dan konfigurasi umum secara sekilas.
weight: 2
---

Kartu referensi cepat untuk NAEOS — ideal untuk pengguna berpengalaman yang membutuhkan referensi cepat.

## Perintah Penting

```bash
# Instal
go install github.com/NAEOS-foundation/naeos/cmd/naeos@latest

# Buat proyek baru
naeos init my-project

# Jalankan pipeline lengkap
naeos run --input-file spec.yaml

# Validasi spesifikasi
naeos validate --input-file spec.yaml

# Kompilasi untuk AI
naeos ai compile --input-file spec.yaml --target opencode

# Jalankan server API REST
naeos api --port 8080

# Jalankan dashboard
naeos dashboard
```

## Contoh Spesifikasi Minimal

```yaml
project: my-service
modules:
  - name: api
    path: ./api
    dependencies: [database]
  - name: database
    path: ./db
services:
  - name: rest-api
    kind: http
    port: 8080
architecture:
  pattern: microservices
generation:
  languages: [go, typescript]
```

## Opsi Modul

| Field | Tipe | Deskripsi |
|-------|------|-----------|
| `name` | string | Pengenal modul (wajib) |
| `path` | string | Path filesystem (wajib) |
| `description` | string | Deskripsi yang mudah dibaca |
| `dependencies` | list | Nama modul lain |
| `kind` | string | Tipe modul (default: service) |

## Jenis Layanan

| Jenis | Protokol | Kasus Penggunaan |
|-------|----------|------------------|
| `http` | HTTP/JSON | API REST, gateway, endpoint WebSocket |
| `grpc` | gRPC/Protobuf | Komunikasi layanan internal |
| `worker` | — | Pekerjaan latar belakang / tugas serverless |
| `cli` | — | Alat baris perintah |
| `job` | — | Pekerjaan terjadwal sekali jalan |

## Pola Arsitektur

| Pola | Deskripsi | Cocok Untuk |
|------|-----------|-------------|
| `microservices` | Layanan independen, longgar terikat | Tim besar, domain kompleks |
| `monolithic` | Unit deploy tunggal | Tim kecil, domain sederhana |
| `serverless` | Function-as-a-service | Event-driven, beban variabel |
| `event-driven` | Pesan async | Throughput tinggi, dekoupling |

## Bahasa Generasi

| Bahasa | Adapter | Output |
|--------|---------|--------|
| Go | `go` | File `.go` dengan modul, paket |
| TypeScript | `typescript` | File `.ts` dengan interface |
| Python | `python` | File `.py` dengan kelas |
| Java | `java` | File `.java` dengan paket |
| Rust | `rust` | File `.rs` dengan crate |

## Referensi Cepat CLI

| Perintah | Deskripsi |
|----------|-----------|
| `naeos init` | Buat proyek baru |
| `naeos run` | Jalankan pipeline lengkap |
| `naeos validate` | Validasi spesifikasi |
| `naeos ai compile` | Kompilasi spesifikasi untuk agen AI |
| `naeos build` | Build artifact dari spesifikasi |
| `naeos api` | Jalankan server API REST |
| `naeos dashboard` | Jalankan dashboard web |
| `naeos cloud plan` | Generate rencana deployment cloud |
| `naeos cloud deploy` | Deploy ke provider cloud |
| `naeos cloud status` | Tampilkan status resource ter-deploy |
| `naeos plugin install` | Instal plugin |
| `naeos plugin list` | Daftar plugin terinstal |
| `naeos db migrate` | Jalankan migrasi database |
| `naeos db connect` | Hubungkan ke database |
| `naeos version` | Tampilkan info versi |

## Endpoint API

| Metode | Endpoint | Deskripsi |
|--------|----------|-----------|
| GET | `/api/v1/health` | Pemeriksaan kesehatan |
| GET | `/api/v1/version` | Info versi |
| POST | `/api/v1/specs/validate` | Validasi spesifikasi |
| POST | `/api/v1/specs/compile` | Kompilasi spesifikasi |
| POST | `/api/v1/pipeline/run` | Jalankan pipeline |
| GET | `/api/v1/pipeline/status` | Status pipeline |
| GET | `/api/v1/artifacts` | Daftar artifact |
| POST | `/api/v1/context/generate` | Hasilkan konteks |
| POST | `/api/v1/ai/enrich/stream` | Pengayaan AI (SSE) |
| POST | `/api/v1/ai/compile/stream` | Kompilasi AI (SSE) |
| GET | `/api/v1/plugins` | Daftar plugin |
| WS | `/ws` | Event real-time WebSocket |

## Variabel Lingkungan

| Variabel | Deskripsi | Default |
|----------|-----------|---------|
| `NAEOS_LLM_API_KEY` | Kunci API untuk provider LLM (`naeos ai compile`) | — |
| `NAEOS_LLM_PROVIDER` | Provider LLM (openai, anthropic, gemini, ...) | — |
| `NAEOS_ENCRYPTION_KEY` | Passphrase untuk user store auth server API (`naeos api`) | — |
| `NAEOS_PIPELINES_FILE` | Path ke file pipelines (default: `~/.naeos/pipelines.json`) | — |

## Struktur Direktori Output

```
output/
├── go/                    # Kode Go yang di-generate
│   ├── cmd/
│   ├── internal/
│   └── go.mod
├── typescript/            # TypeScript yang di-generate
│   ├── src/
│   ├── package.json
│   └── tsconfig.json
├── ai/                    # Set instruksi AI
│   ├── copilot-instructions.md
│   ├── CLAUDE.md
│   ├── .cursorrules
│   └── .gemini/CONFIG.md
├── context/               # Context bundle
│   └── summary.md
└── terraform/             # Deployment cloud (jika dikonfigurasi)
    ├── main.tf
    ├── variables.tf
    └── outputs.tf
```

Lihat juga: [Referensi CLI](/docs/cli-reference/), [Bahasa Spesifikasi](/docs/spec-language/), [Arsitektur](/docs/architecture/)
