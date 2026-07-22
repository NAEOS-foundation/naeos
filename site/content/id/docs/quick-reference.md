---
title: Referensi Cepat
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
naeos compile --all --input-file spec.yaml

# Jalankan server API
naeos serve --port 8080

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
    kind: rest
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
| `rest` | HTTP/JSON | API REST |
| `grpc` | gRPC/Protobuf | Komunikasi layanan internal |
| `graphql` | HTTP/GraphQL | API query fleksibel |
| `websocket` | WebSocket | Komunikasi real-time |
| `worker` | — | Pekerjaan latar belakang |
| `lambda` | — | Fungsi serverless |
| `reverse-proxy` | HTTP | API gateway / load balancer |

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
| `naeos compile` | Kompilasi spesifikasi untuk asisten AI |
| `naeos gen` | Generate kode untuk bahasa spesifik |
| `naeos serve` | Jalankan server API |
| `naeos dashboard` | Jalankan dashboard web |
| `naeos cloud plan` | Generate rencana deployment cloud |
| `naeos cloud deploy` | Deploy ke provider cloud |
| `naeos cloud destroy` | Hancurkan resource cloud |
| `naeos plugin install` | Instal plugin |
| `naeos plugin list` | Daftar plugin terinstal |
| `naeos db migrate` | Jalankan migrasi database |
| `naeos db reset` | Reset database |
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
| `NAEOS_LLM_API_KEY` | Kunci API untuk provider LLM | — |
| `NAEOS_DB_DRIVER` | Driver database (postgres, mysql, sqlite) | sqlite |
| `NAEOS_DB_DSN` | String koneksi database | — |
| `NAEOS_PORT` | Port server API | 8080 |
| `NAEOS_LOG_LEVEL` | Level log (debug, info, warn, error) | info |

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
│   └── GEMINI.md
├── context/               # Context bundle
│   └── summary.md
└── terraform/             # Deployment cloud (jika dikonfigurasi)
    ├── main.tf
    ├── variables.tf
    └── outputs.tf
```

Lihat juga: [Referensi CLI](/docs/cli-reference/), [Bahasa Spesifikasi](/docs/spec-language/), [Arsitektur](/docs/architecture/)
