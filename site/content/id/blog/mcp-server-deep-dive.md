---
title: "Deep Dive Server MCP NAEOS: Hubungkan Spesifikasi Anda ke Setiap Alat AI"
description: "Bagaimana server MCP NAEOS mengekspos tool parse, validate, compile, dan context lewat Model Context Protocol — dan cara menggunakannya dari editor dan agen."
date: 2026-08-19
author: "NAEOS Foundation"
categories: ["technical"]
---

Model Context Protocol (MCP) adalah cara alat AI menemukan dan memanggil kemampuan eksternal. NAEOS hadir dengan server MCP yang mengekspos mesin spesifikasinya — parsing, validasi, kompilasi, dan inspeksi artifact — sebagai sembilan tool yang bisa dipanggil klien MCP mana pun.

Ini deep dive-nya: apa yang dilakukan server, cara menjalankannya, dan cara memanggilnya langsung dengan curl.

## Kenapa MCP?

Kami mendokumentasikan keputusan ini di [ADR-003](https://github.com/NAEOS-foundation/naeos/blob/main/docs/adr/003-why-mcp-for-ai-integration.md): alih-alih membangun integrasi khusus per vendor (Copilot, Claude, Cursor, Gemini, Codex, OpenCode), MCP memberi kami satu antarmuka vendor-netral yang dipakai semua agen. Protokolnya self-describing — klien memanggil `tools/list` dan mempelajari apa yang tersedia — dan transport-agnostic.

Separuh lain dari strategi ini adalah AI compiler: `naeos ai compile` menulis file instruksi per-agen (`CLAUDE.md`, `.cursorrules`, `.github/copilot-instructions.md`, ...). MCP adalah antarmuka *live*; set instruksi adalah konteks *persisten*. Bersama-sama mereka mencakup kedua mode pengembangan berbantuan AI.

## Sembilan Tool

Server mengekspos tool berikut lewat JSON-RPC 2.0:

| Tool | Fungsi | Argumen |
|------|--------|---------|
| `parse_spec` | Parse spec, kembalikan ringkasan proyek/modul/layanan | `spec` (wajib) |
| `validate_spec` | Validasi spec; PASS/FAIL + isu | `spec` (wajib) |
| `generate_context` | Generate bundle konteks AI | `spec`, `format` (markdown/plain) |
| `compile_spec` | Kompilasi ke set instruksi untuk agen target | `spec`, `target` (copilot, claude, cursor, gemini, codex, opencode) |
| `explain_concept` | Jelaskan konsep NAEOS (pipeline, neir, kernel, policy, ...) | `concept` (wajib) |
| `list_artifacts` | Daftar artifact dari artifact store | — |
| `get_pipeline_status` | Status job pipeline berdasarkan ID | `job_id` (wajib) |
| `export_terraform` | Generate Terraform HCL dari spec | `spec` (wajib) |
| `list_plugins` | Daftar plugin terpasang dan statusnya | — |

## Menjalankan Server

```bash
naeos mcp                     # melayani di :3000
naeos mcp --port 8080         # atau port Anda sendiri
```

Anda akan melihat banner startup beserta URL health-check. Server mengimplementasikan metode `initialize`, `tools/list`, dan `tools/call` serta melaporkan protocol version `2024-11-05`.

Endpoint MCP yang sama juga terpasang di dalam server REST API di `POST /api/v1/mcp/message`:

```bash
naeos api                     # REST API di :8080
```

## Memanggilnya dengan curl

Health check:

```bash
curl http://localhost:3000/health
# {"status":"ok"}
```

Daftar tool:

```bash
curl -s -X POST http://localhost:3000/mcp \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","method":"tools/list","id":1}'
```

Validasi spesifikasi:

```bash
curl -s -X POST http://localhost:3000/mcp \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","method":"tools/call","id":2,
       "params":{"name":"validate_spec",
                 "arguments":{"spec":"project: shop\nmodules:\n  - name: core\n    path: ./core\n"}}}'
```

Parse dan ringkas:

```bash
curl -s -X POST http://localhost:3000/mcp \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","method":"tools/call","id":3,
       "params":{"name":"parse_spec",
                 "arguments":{"spec":"project: shop\nmodules:\n  - name: core\n    path: ./core\n"}}}'
# → "Project: shop · Modules: 1 · Services: 0"
```

Generate Terraform:

```bash
curl -s -X POST http://localhost:3000/mcp \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","method":"tools/call","id":4,
       "params":{"name":"export_terraform",
                 "arguments":{"spec":"project: shop\nservices:\n  - name: api\n    kind: http\n    port: 8080\n"}}}'
```

## Menggunakannya dari Server API

Varian REST API mengekspos protokol yang identik di `/api/v1/mcp/message`:

```bash
curl -s -X POST http://localhost:8080/api/v1/mcp/message \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","method":"initialize","id":1}'
```

Di dalam server API, layer MCP juga mendapat akses ke artifact store dan plugin manager, sehingga `list_artifacts` dan `list_plugins` berfungsi penuh di sana.

## Yang Dimungkinkan

Server MCP mengubah NAEOS menjadi bagian *live* dari alur kerja AI Anda, bukan generator sekali jalan:

- **Agen bisa memvalidasi spec sebelum generate.** Tidak perlu lagi minta Claude untuk "tulis service" lalu menemukan spec-nya salah.
- **Agen bisa bertanya apa arti sebuah konsep.** `explain_concept` memberi jawaban ter-grounding di level model tentang stage pipeline, NEIR, policy, dan kernel.
- **Tooling CI bisa cek status artifact** tanpa binary NAEOS — cukup panggilan HTTP.
- **Infra-as-code mengalir langsung dari model.** `export_terraform` menurunkan HCL dari model NEIR yang sama dengan kodenya.

## Trade-off yang Jujur

Server saat ini mengimplementasikan revisi tools-only dari MCP (`2024-11-05`) lewat HTTP — belum ada resources, prompts, atau streaming. Jika Anda butuh kemampuan MCP yang lebih dalam, roadmap melacak dukungan protokol yang lebih kaya bersama config klien editor.

Referensi CLI yang otoritatif ada di [`docs/cli/naeos_mcp.md`](https://github.com/NAEOS-foundation/naeos/blob/main/docs/cli/naeos_mcp.md), dan spec OpenAPI mendokumentasikan rute server API. Coba [tutorial end-to-end](/id/blog/ecommerce-end-to-end-tutorial/) dulu, lalu hubungkan spec Anda ke agen favorit dengan `naeos mcp`.