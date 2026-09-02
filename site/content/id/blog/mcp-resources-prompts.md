---
title: "Server MCP NAEOS Berkembang: Resources, Prompts, Completions, dan Pagination"
description: "Server MCP NAEOS kini mendukung penuh model layer terbaru — resources, prompts, completions, liveness ping, dan pagination berbasis cursor — sehingga agen bisa menemukan spesifikasi, dokumen, dan artifact sesuai permintaan."
date: 2026-08-30
author: "NAEOS Foundation"
categories: ["technical", "release"]
---

Saat pertama kali kami merilis [deep dive server MCP](/id/blog/mcp-server-deep-dive/), server hanya mengimplementasikan revisi tools-only dari Model Context Protocol (`2024-11-05`) — sembilan tool dan tidak lebih. Tanpa resources, tanpa prompts, tanpa streaming. Baris terakhir itu, "no resources, prompts, or streaming yet," kami tulis sebagai catatan jujur. Hari ini kami menutup celah itu.

Rilis NAEOS berikutnya mengubah server MCP menjadi pengalaman model layer yang lengkap: sembilan tool yang sudah ada kini digabung dengan **resources**, **prompts**, **argument completions**, sebuah **liveness ping**, dan **pagination berbasis cursor** di setiap operasi list.

## Terverifikasi terhadap daftar evaluasi

Di post sebelumnya kami menulis bahwa server "saat ini hanya mengimplementasikan revisi tools-only... belum ada resources, prompts, atau streaming." Berikut yang berubah:

- **Resources** — baca dokumen konsep NAEOS, isi artifact store, dan status pipeline job sebagai resource MCP.
- **Prompts** — empat template bawaan yang bisa diambil dan diinstansiasi agen.
- **Completions** — sajikan autocompletion argumen untuk klien di tengah percakapan.
- **Ping** — metode liveness ringan untuk pemeriksaan kesehatan koneksi.
- **Pagination** — list berbasis cursor di setiap metode `*_list`.

## Apa itu resource MCP?

Model Context Protocol tidak hanya memodelkan tool call. *Resource* adalah sepotong konten yang bisa dibaca klien sesuai permintaan — anggap sebagai sistem file hidup yang bisa dibuka agen, bukan fungsi yang dipanggilnya. NAEOS mengekspos tiga namespace:

| Namespace resource | Yang diekspos |
|--------------------|---------------|
| `naeos://docs/{concept}` | Dokumentasi konsep NAEOS (pipeline, NEIR, kernel, policy, ...) |
| `naeos://artifacts/{path}` | Isi artifact store berdasarkan path |
| `naeos://jobs/{id}` | Status pipeline job berdasarkan ID |

Jadi saat agen Anda ingin tahu maksud "kernel" dalam model NAEOS, atau memeriksa apakah sebuah build artifact sudah masuk, ia bisa *membaca resource*-nya alih-alih menebak.

```bash
# resources/list dengan pagination
curl -s -X POST http://localhost:3000/mcp \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","method":"resources/list","id":1,
       "params":{"cursor":"<opaque-cursor>"}}'

# membaca satu resource
curl -s -X POST http://localhost:3000/mcp \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","method":"resources/read","id":2,
       "params":{"uri":"naeos://docs/neir"}}'
```

## Template prompt bawaan

Resources menjawab "apa yang kamu tahu?"; prompts menjawab "apa yang sebaiknya kuminta kamu lakukan?". NAEOS menyediakan empat template yang bisa didaftarkan dan diinstansiasi klien, dengan substitusi argumen verbatim ke dalam pesan prompt:

- `review-spec` — berikan spesifikasi kepada agen untuk direview.
- `enrich-spec` — sarankan perbaikan pada sebuah spesifikasi.
- `explain-architecture` — jelaskan sebuah pattern arsitektur.
- `generate-spec` — buat spesifikasi dari deskripsi longgar.

```bash
curl -s -X POST http://localhost:3000/mcp \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","method":"prompts/list","id":1}'

curl -s -X POST http://localhost:3000/mcp \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","method":"prompts/get","id":2,
       "params":{"name":"explain-architecture",
                 "arguments":{"architecture":"microservices"}}}'
```

## Completions: autocomplete di dalam percakapan

Banyak agen membiarkan tool mengisi argumen saat Anda mengetik. MCP memodelkannya lewat `completion/complete`. NAEOS menyambungkan dua sumber completion:

- **`ref/prompt`** — melengkapi argumen `architecture` dari `explain-architecture` dengan seluruh sepuluh pattern arsitektur NEIR.
- **`ref/resource`** — melengkapi URI resource dengan memfilter daftar resource langsung (dokumen konsep, path artifact, ID pipeline job), dengan prefix matching case-insensitive dan batas 100 nilai.

Handshake `initialize` kini mengiklankan kapabilitas `completions`, sehingga klien tahu untuk menawarkan autocomplete.

## Sehat secara desain: ping

Pemeriksaan kesehatan koneksi butuh round trip yang murah dan bebas efek samping. `ping` menyediakannya:

```bash
curl -s -X POST http://localhost:3000/mcp \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","method":"ping","id":3}'
```

## Pagination cursor di mana-mana

Setiap metode list — `tools/list`, `resources/list`, `prompts/list` — kini di-paginasi dengan cursor alih-alih membanjiri klien dengan satu respons raksasa. Respons berisi 50 item per halaman; ketika masih ada lagi, server mengembalikan cursor base64 yang opaque:

```json
{ "resources": [ ...50 item... ], "nextCursor": "bmV4dDoxMDA=" }
```

Teruskan nilai itu sebagai parameter `cursor` untuk mengambil halaman berikutnya. Cursor yang rusak mengembalikan error `-32602` (Invalid Params).

## Yang ini buka untuk alur kerja agen Anda

Server MCP NAEOS tak lagi sekadar tool di rak. Ini adalah **model layer yang konversasional**:

- Agen bisa *membaca* artifact store untuk memeriksa hasil build, bukan sekadar memanggil fungsi yang meringkasnya.
- `initialize` yang lebih kaya mengumumkan resources, prompts, dan completions — sehingga klien menyiapkan UX yang tepat sejak awal.
- Registry besar (plugins, templates, schemas) bisa dijelajahi halaman demi halaman tanpa menghabiskan memori atau konteks token.

Satu hal yang belum kami kirim adalah event progres streaming/SSE. Ini di roadmap; hal lain di model layer kini sudah hidup.

## Coba

Ambil build terbaru, jalankan server, dan usik dengan curl:

```bash
naeos mcp --port 3000
```

Lalu jelajahi resources:

```bash
curl -s -X POST http://localhost:3000/mcp \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","method":"resources/list","id":1}'
```

Padukan dengan [panduan pengembangan berbasis AI](/id/blog/ai-driven-development/) dan [tutorial end-to-end](/id/blog/ecommerce-end-to-end-tutorial/) untuk melihat gambaran utuh. Referensi CLI resmi ada di [`docs/cli/naeos_mcp.md`](https://github.com/NAEOS-foundation/naeos/blob/main/docs/cli/naeos_mcp.md).
