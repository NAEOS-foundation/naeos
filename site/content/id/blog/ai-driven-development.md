---
title: "Pengembangan Berbasis AI dengan NAEOS: Mengajarkan Asisten Anda Berpikir dalam Arsitektur"
description: "Bagaimana NAEOS mengkompilasi NEIR menjadi set instruksi AI yang memberikan konteks arsitektur nyata ke Copilot, Claude, dan Cursor."
date: 2026-07-15
categories: ["tutorial"]
---

Asisten AI coding sangat powerful — tapi mereka bekerja dalam kegelapan. Saat Anda meminta GitHub Copilot untuk menghasilkan fungsi, ia melihat file yang Anda buka. Saat Anda meminta Claude Code untuk me-refactor modul, ia melihat direktori yang Anda tunjuk. Tidak ada dari mereka yang melihat arsitektur.

NAEOS mengubah ini dengan **AI Compiler** — sistem yang mengubah model NEIR Anda menjadi set instruksi yang bisa dikonsumsi asisten coding. Hasilnya: alat AI Anda tidak hanya melihat kode, mereka melihat sistem.

## Masalah Konteks

Asisten AI coding modern unggul dalam tugas lokal. Mereka bisa melengkapi fungsi, menyarankan tipe, atau me-refactor kelas dengan akurasi yang mengesankan. Tapi mereka gagal dalam **tugas arsitektural** — jenis yang memerlukan pemahaman bagaimana modul terhubung, pola apa yang tim Anda gunakan, dan batasan apa yang berlaku.

Pertimbangkan meminta AI: "Tambahkan endpoint baru ke layanan order yang memanggil layanan payment."

Tanpa konteks, AI mungkin:
- Menghasilkan klien HTTP yang tidak cocok dengan pola RPC yang ada
- Melewatkan kontrak penanganan error antar layanan
- Mengabaikan kebijakan rate limiting yang didefinisikan dalam aturan governance
- Menggunakan format logging yang berbeda dari sisa codebase

Dengan konteks NAEOS, AI tahu:
- Layanan order menggunakan gRPC, bukan REST
- Error payment harus dibungkus dalam `PaymentError` dengan kode error spesifik
- Semua panggilan antar-layanan melalui service mesh dengan kebijakan retry
- Logging terstruktur menggunakan paket `slog` dengan field request ID

## Cara Kerja AI Compiler

AI Compiler mengambil model NEIR Anda dan menghasilkan **context bundle** — set instruksi terstruktur yang disesuaikan untuk setiap platform AI:

```bash
# Generate konteks untuk semua platform AI yang didukung
naeos compile --all --input-file spec.yaml

# Generate untuk platform spesifik
naeos compile --target copilot --input-file spec.yaml
```

Compiler mendukung enam platform:

| Platform | Format Output | Cara Penggunaan |
|----------|--------------|-----------------|
| GitHub Copilot | `copilot-instructions.md` | Dimuat otomatis oleh Copilot |
| Claude Code | `CLAUDE.md` | Dibaca Claude saat sesi dimulai |
| Cursor | `.cursorrules` | Diterapkan ke semua sesi Cursor |
| Gemini CLI | `GEMINI.md` | Konteks untuk Gemini CLI |
| Codex | `AGENTS.md` | Instruksi untuk OpenAI Codex |
| OpenCode | `AGENTS.md` | Instruksi untuk OpenCode |

## Apa yang Ada di Context Bundle

Context bundle bukan sekadar dump README. Ini adalah dokumen terstruktur yang menangkap:

### Ikhtisar Arsitektur
```markdown
## Pola Arsitektur: Microservices

### Service Mesh
- Protocol: gRPC dengan HTTP/2
- Load balancing: Round-robin
- Circuit breaking: 5 kegagalan berturut-turut → buka selama 30 detik

### Layanan
- order-service (port 9001): Menangani CRUD order dan lifecycle
- payment-service (port 9002): Memproses pembayaran via API Stripe
- inventory-service (port 9003): Mengelola tingkat stok
```

### Grafik Dependensi
```markdown
## Dependensi Modul

order-service
  ├── payment-service (gRPC)
  └── inventory-service (gRPC)

payment-service
  └── (eksternal: API Stripe)

inventory-service
  └── postgres-db (connection pool)
```

### Konvensi Coding
```markdown
## Konvensi

### Penanganan Error
- Semua error layanan mengembalikan struct `ServiceError`
- Kode error mengikuti: `{module}.{action}.{reason}`
- Contoh: `order.payment.insufficient_funds`

### Logging
- Gunakan `slog` dengan atribut terstruktur
- Setiap entri log harus menyertakan `request_id` dan `service_name`
```

### Aturan Governance
```markdown
## Kebijakan

- Tidak ada akses database langsung antar layanan (gunakan API)
- Semua endpoint API harus memiliki rate limiting
- Autentikasi via JWT dengan penandatanganan RS256
```

## Contoh Nyata

Mari kita lalui contoh konkreto. Katakan Anda memiliki spesifikasi untuk aplikasi chat:

```yaml
project: chat-app
modules:
  - name: gateway
    path: ./gateway
    dependencies: [message-service, user-service]
  - name: message-service
    path: ./services/messages
    dependencies: [redis-cache, postgres-db]
  - name: user-service
    path: ./services/users
    dependencies: [postgres-db]
services:
  - name: ws-gateway
    kind: websocket
    port: 8080
  - name: message-api
    kind: rest
    port: 9001
  - name: user-api
    kind: rest
    port: 9002
architecture:
  pattern: microservices
generation:
  languages: [go, typescript]
```

Menjalankan `naeos compile --all` menghasilkan enam file instruksi berbeda, masing-masing dioptimasi untuk platform targetnya. Saat Anda membuka proyek di Cursor, file `.cursorrules` memberitahu Claude tentang WebSocket gateway, strategi caching Redis layanan pesan, dan alur validasi JWT layanan user — sebelum Anda mengetik satu prompt pun.

## Alur Kerja

Alur kerja lengkapnya seperti ini:

1. **Tulis spesifikasi** — Definisikan sistem Anda dalam YAML
2. **Jalankan pipeline** — Generate kode + context bundle
3. **Buka IDE Anda** — Asisten AI memuat konteks secara otomatis
4. **Coding dengan konteks** — Saran AI menghormati arsitektur Anda

```bash
# Satu perintah menghasilkan semua
naeos run --input-file spec.yaml --compile-all

# Proyek Anda sekarang memiliki:
# - Kode Go dan TypeScript yang di-generate
# - Set instruksi AI untuk 6 platform
# - Grafik dependensi tervalidasi
# - Struktur yang patuh governance
```

## Tips untuk Konteks AI yang Lebih Baik

Kualitas output AI tergantung pada kualitas spesifikasi Anda. Berikut beberapa tips:

**Jelaskan pola secara eksplisit.** Alih-alih hanya mendaftar layanan, jelaskan pola interaksi mereka:

```yaml
architecture:
  pattern: microservices
  description: |
    Arsitektur event-driven dengan:
    - gRPC sinkron untuk request/response
    - Redis Stream untuk event async
    - Circuit breaker pada semua panggilan eksternal
```

**Dokumentasikan konvensi dalam spesifikasi.** AI compiler menyertakan field `description` spesifikasi Anda dalam context bundle:

```yaml
modules:
  - name: order-service
    description: |
      Manajemen lifecycle order.
      Menggunakan pola CQRS: perintah menuju model tulis,
      query menghitung model baca (Elasticsearch).
      Semua perubahan state mengemit event domain ke Redis Stream.
```

**Pertahankan grafik dependensi yang akurat.** AI menggunakan ini untuk memahami siapa yang bisa memanggil siapa. Dependensi yang hilang mengarah ke kode yang di-generate yang melewati batas layanan Anda.

## Apa Selanjutnya

Kami sedang mengerjakan pembuatan context bundle lebih kaya:

- **Kompilasi inkremental** — Hanya regenerate konteks untuk modul yang berubah
- **Template instruksi kustom** — Override format default untuk konvensi tim Anda
- **Sinkronisasi konteks live** — Context bundle berubah saat spesifikasi Anda berkembang

Tujuannya sederhana: buat asisten AI coding Anda sepengetahuan tentang sistem Anda seperti engineer senior Anda. Spesifikasi deklaratif adalah sumber pengetahuan yang sempurna — presisi, lengkap, dan bisa dibaca mesin.
