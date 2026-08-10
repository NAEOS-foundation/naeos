---
title: Dashboard
description: Pantau aktivitas pipeline, statistik, dan kesehatan komponen secara real time.
weight: 14
---

Dashboard web NAEOS memberikan tampilan langsung aktivitas pipeline,
statistik sistem, dan kesehatan komponen, dengan pembaruan real time melalui
WebSocket.

## Menjalankan Dashboard

```bash
naeos dashboard
```

Berjalan di `http://localhost:3000` secara default. Gunakan `--port` untuk
mengubahnya:

```bash
naeos dashboard --port 8080
```

## Fitur

- **Log aktivitas langsung** — event pipeline dan pesan log mengalir real
  time melalui WebSocket (`/ws`)
- **Statistik** — statistik pipeline disiarkan setiap 5 detik
- **Kesehatan komponen** — status server API, parser, kompiler, dan server MCP
- **Endpoint API** — `GET /api/stats`, `GET /api/activity`,
  `GET /api/health`

## Komponen Dashboard

| Komponen | Sumber Status |
|----------|---------------|
| Server API | `api.NewServer` dengan auth dinonaktifkan |
| Parser | Siap saat dimulai |
| Kompiler | Siap saat dimulai |
| Server MCP | Berhenti (degradasi) kecuali diaktifkan |

## Menggunakan Dashboard dengan Server API

Dashboard menggunakan kembali server API NAEOS, sehingga antarmuka dashboard
dan REST API berjalan di port yang sama. Upgrade WebSocket terjadi di `/ws`,
dan dashboard sendiri dilayani di `/`.

## Terkait

- [Referensi CLI](/id/docs/cli-reference/) — semua perintah CLI termasuk `dashboard`
- [Build Terdistribusi](/id/docs/distributed-builds/) — eksekusi pipeline paralel
