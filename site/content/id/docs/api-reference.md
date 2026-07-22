---
title: Referensi API
description: Dokumentasi API interaktif untuk REST API NAEOS.
weight: 17
---

REST API NAEOS memungkinkan Anda berinteraksi dengan runtime NAEOS secara programatis. Halaman ini menyediakan referensi interaktif yang dihasilkan dari spesifikasi OpenAPI 3.0.

## Ikhtisar

| Info | Nilai |
|------|-------|
| **Base URL** | `http://localhost:8080` |
| **Versi API** | 1.5.0 |
| **Format** | JSON |
| **Autentikasi** | Tidak ada (pengembangan lokal) |

## Referensi Cepat

### Sistem

| Metode | Endpoint | Deskripsi |
|--------|----------|-----------|
| GET | `/api/v1/health` | Pemeriksaan kesehatan |
| GET | `/api/v1/version` | Info versi |
| GET | `/api/v1/config/schema` | Skema konfigurasi |
| GET | `/healthz` | Liveness probe |
| GET | `/readyz` | Readiness probe |
| GET | `/metrics` | Metrik Prometheus |

### Spesifikasi

| Metode | Endpoint | Deskripsi |
|--------|----------|-----------|
| GET | `/api/v1/specs` | Daftar semua spesifikasi |
| POST | `/api/v1/specs` | Buat spesifikasi baru |
| POST | `/api/v1/specs/validate` | Validasi spesifikasi |
| POST | `/api/v1/specs/compile` | Kompilasi spesifikasi |

### Pipeline

| Metode | Endpoint | Deskripsi |
|--------|----------|-----------|
| POST | `/api/v1/pipeline/run` | Jalankan pipeline |
| GET | `/api/v1/pipeline/status` | Status pipeline |
| GET | `/api/v1/pipelines` | Daftar semua pipeline |

### AI (Streaming SSE)

| Metode | Endpoint | Deskripsi |
|--------|----------|-----------|
| POST | `/api/v1/ai/enrich/stream` | Pengayaan AI (streaming) |
| POST | `/api/v1/ai/explain/stream` | Penjelasan AI (streaming) |
| POST | `/api/v1/ai/compile/stream` | Kompilasi AI (streaming) |

### Lainnya

| Metode | Endpoint | Deskripsi |
|--------|----------|-----------|
| GET | `/api/v1/artifacts` | Daftar artifact |
| POST | `/api/v1/context/generate` | Hasilkan context bundle |
| GET/POST | `/api/v1/plugins` | Manajemen plugin |
| POST | `/api/v1/mcp/message` | Pesan MCP |
| POST | `/api/v1/cloud/plan` | Rencana deployment cloud |
| POST | `/api/v1/cloud/deploy` | Deploy ke cloud |
| POST | `/api/v1/cloud/destroy` | Hancurkan resource cloud |
| GET | `/api/v1/cloud/status` | Status cloud |

## Dokumentasi Interaktif

Dokumentasi API interaktif lengkap tersedia di bawah ini. Anda dapat menguji panggilan API langsung dari browser.

{{< swagger-ui >}}

## Penggunaan CLI

Anda juga dapat berinteraksi dengan API melalui CLI:

```bash
# Jalankan server API
naeos serve --port 8080

# Validasi melalui API
curl -X POST http://localhost:8080/api/v1/specs/validate \
  -H "Content-Type: application/json" \
  -d '{"spec": "project: my-app"}'

# Jalankan pipeline melalui API
curl -X POST http://localhost:8080/api/v1/pipeline/run \
  -H "Content-Type: application/json" \
  -d '{"spec_path": "spec.yaml"}'
```

## Spesifikasi OpenAPI

Unduh spesifikasi OpenAPI 3.0 lengkap:

- [openapi.yaml](/openapi.yaml)

Lihat juga: [Referensi CLI](/id/docs/cli-reference/), [Pipeline Engine](/id/docs/pipeline-engine/)
