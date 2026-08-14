---
title: Pemecahan Masalah
slug: troubleshooting
description: Masalah umum dan solusi saat bekerja dengan NAEOS.
weight: 16
---

Halaman ini membahas masalah umum yang dihadapi saat menggunakan NAEOS dan cara menyelesaikannya.

## Instalasi

### `naeos: command not found`

Setelah menginstal dengan `go install`, pastikan direktori Go bin ada di PATH Anda:

```bash
# Periksa apakah binary ada
ls ~/go/bin/naeos

# Tambahkan ke PATH (tambahkan ke profil shell Anda)
export PATH="$HOME/go/bin:$PATH"
```

### Permission denied di macOS

```bash
# Hapus attribute quarantine
xattr -d com.apple.quarantine ~/go/bin/naeos
```

### Ketidakcocokan versi

Jika Anda melihat konflik versi antara CLI dan API:

```bash
# Periksa versi CLI
naeos version

# Pastikan versi terbaru
go install github.com/NAEOS-foundation/naeos/cmd/naeos@latest
```

## Parsing Spesifikasi

### `yaml: unmarshal errors`

Ini biasanya berarti spesifikasi YAML Anda memiliki masalah struktural:

```yaml
# SALAH — services harus berupa list
services:
  api:
    kind: http

# BENAR
services:
  - name: api
    kind: http
```

### `module not found in dependency graph`

Modul merujuk dependensi yang tidak ada:

```yaml
# SALAH — user-service bergantung pada cache-service, tapi cache-service tidak didefinisikan
modules:
  - name: user-service
    dependencies: [cache-service]

# BENAR — definisikan cache-service terlebih dahulu
modules:
  - name: cache-service
    path: ./cache
  - name: user-service
    path: ./users
    dependencies: [cache-service]
```

### Siklus dependensi terdeteksi

NAEOS tidak mengizinkan siklus dependensi antar modul:

```
Error: siklus dependensi terdeteksi: A → B → C → A
```

Selesaikan dengan mengekstrak fungsionalitas bersama ke modul terpisah:

```
# SEBELUM (siklus)
A bergantung pada B
B bergantung pada C
C bergantung pada A

# SESUDAH (diselesaikan)
A bergantung pada D, B
B bergantung pada D, C
C bergantung pada D
D (modul bersama, tanpa dependensi)
```

## Generasi

### `permission denied` saat menulis output

Direktori output mungkin memiliki izin yang ketat:

```bash
# Perbaiki izin
chmod -R u+w ./output

# Atau arahkan ke direktori output berbeda di naeos.yaml
output_dir: /tmp/output
```

### Kode yang di-generate tidak bisa dikompilasi

Ini jarang terjadi tapi bisa terjadi dengan target generasi kustom. Coba:

```bash
# Jalankan ulang dengan output verbose untuk melihat apa yang di-generate
naeos run --input-file spec.yaml --verbose

# Periksa file yang di-generate
find ./generated -name "*.go" -o -name "*.ts" | head -20
```

### Adapter bahasa tidak ditemukan

Jika Anda menentukan bahasa yang tidak terinstal:

```
Error: adapter bahasa "rust" tidak ditemukan
```

Jika Anda menentukan bahasa yang tidak didukung:

```
Error: unsupported language: "rust"
```

Bahasa yang didukung: `go`, `typescript`, `python`, `java`, `rust`. Varian
framework (`fastapi` untuk python, `actix-web` untuk rust) dipilih otomatis
untuk bahasa dasarnya. Verifikasi target Anda:

```bash
naeos run --input-file spec.yaml --language go --dry-run
```

## AI Compiler

### `LLM API key tidak dikonfigurasi`

AI compiler memerlukan kunci API untuk provider pilihan Anda:

```bash
# Atur variabel lingkungan
export NAEOS_LLM_API_KEY="kunci-api-anda"
export NAEOS_LLM_PROVIDER="anthropic"   # opsional; default: openai

# Lalu kompilasi
naeos ai compile --input-file spec.yaml --target claude
```

### Context bundle kosong

Ini biasanya berarti spesifikasi Anda tidak memiliki cukup detail:

```yaml
# TERLALU SEDERHANA — konteks minimal yang dihasilkan
project: my-app

# LEBIH BAIK — lebih banyak konteks dalam bundle
project: my-app
description: Platform e-commerce dengan microservices event-driven
modules:
  - name: api-gateway
    path: ./gateway
    description: Reverse proxy dengan rate limiting dan validasi JWT
    dependencies: [user-service, order-service]
```

### Kompilasi terlalu lama

Spesifikasi besar dengan banyak modul bisa memakan waktu. Kurangi beban per eksekusi:

```bash
# Kompilasi untuk satu agen, bukan beberapa
naeos ai compile --input-file spec.yaml --target claude

# Jaga spesifikasi tetap fokus — spesifikasi yang lebih kecil kompilasi lebih cepat
```

## Server API

### `address already in use`

Proses lain menggunakan port 8080:

```bash
# Cari prosesnya
lsof -i :8080

# Gunakan port berbeda
naeos api --port 9090
```

### Error CORS di browser

Server API mengizinkan origin `localhost` dan `naeos.dev` secara default. Jika frontend Anda berjalan di origin berbeda, origin tersebut harus ditambahkan ke daftar yang diizinkan sebelum permintaan dapat lewat.

### Koneksi WebSocket gagal

Pastikan proxy/load balancer Anda mendukung WebSocket upgrade. Endpoint WebSocket ada di `/ws`.

## Database

### `connection refused`

Periksa apakah database Anda berjalan dan dapat diakses:

```bash
# PostgreSQL
psql -h localhost -U postgres -c "SELECT 1"

# MySQL
mysql -h localhost -u root -e "SELECT 1"

# SQLite (berbasis file)
ls -la ./naeos.db
```

### Error migrasi

Jika migrasi database gagal:

```bash
# Jalankan migrasi secara manual dengan output verbose
naeos db migrate --name <koneksi> --verbose

# Periksa direktori migrasi
ls migrations/
```

## Performa

### Pipeline lambat

Untuk spesifikasi besar, coba optimasi berikut:

```bash
# Gunakan caching antar eksekusi (aktif via --cache-dir)
naeos run --input-file spec.yaml --cache-dir .naeos-cache

# Generate hanya bahasa tertentu
naeos run --input-file spec.yaml --language go --language typescript
```

### Penggunaan memori tinggi

Spesifikasi besar dengan 100+ modul mungkin menggunakan memori signifikan:

```bash
# Profiling penggunaan memori
naeos run --input-file spec.yaml --profile --profile-out profile.json
```

Jika spesifikasi perlu dipecah, bagi menjadi beberapa file dan komposisikan dengan `$include` pada spesifikasi utama.

## Mendapatkan Bantuan

Jika masalah Anda tidak tercakup di sini:

1. Periksa [GitHub Issues](https://github.com/NAEOS-foundation/naeos/issues) untuk masalah serupa
2. Cari di [GitHub Discussions](https://github.com/NAEOS-foundation/naeos/discussions)
3. Tanya di [komunitas Discord](https://discord.gg/naeos)
4. Buka issue baru dengan:
   - Versi NAEOS (`naeos version`)
   - Sistem operasi dan arsitektur
   - Pesan error lengkap
   - Cuplikan spesifikasi yang relevan (sensor data sensitif)

Lihat juga: [Memulai](/docs/getting-started/), [Instalasi](/docs/installation/), [Referensi CLI](/docs/cli-reference/)
