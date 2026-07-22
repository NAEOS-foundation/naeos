---
title: Pemecahan Masalah
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
    kind: rest

# BENAR
services:
  - name: api
    kind: rest
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

# Atau gunakan direktori output berbeda
naeos run --input-file spec.yaml --output-dir /tmp/output
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

Periksa adapter yang tersedia:

```bash
naeos gen --list
```

Instal adapter yang hilang melalui sistem plugin:

```bash
naeos plugin install rust-adapter
```

## AI Compiler

### `LLM API key tidak dikonfigurasi`

AI compiler memerlukan kunci API untuk provider pilihan Anda:

```bash
# Atur variabel lingkungan
export NAEOS_LLM_API_KEY="kunci-api-anda"

# Atau passing langsung
naeos compile --target copilot --api-key "kunci-api-anda"
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

Spesifikasi besar dengan banyak modul bisa memakan waktu. Gunakan kompilasi inkremental:

```bash
# Hanya kompilasi modul yang berubah
naeos compile --incremental --input-file spec.yaml
```

## Server API

### `address already in use`

Proses lain menggunakan port 8080:

```bash
# Cari prosesnya
lsof -i :8080

# Gunakan port berbeda
naeos serve --port 9090
```

### Error CORS di browser

Server API mengizinkan localhost secara default. Untuk origin lain, konfigurasikan CORS:

```bash
naeos serve --cors-origins "https://myapp.com,https://staging.myapp.com"
```

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
# Reset database (PERINGATAN: menghancurkan data)
naeos db reset --confirm

# Atau jalankan migrasi secara manual
naeos db migrate --verbose
```

## Performa

### Pipeline lambat

Untuk spesifikasi besar, coba optimasi ini:

```bash
# Gunakan caching (melewati stage yang tidak berubah)
naeos run --input-file spec.yaml --cache

# Generate hanya bahasa tertentu
naeos run --input-file spec.yaml --languages go,typescript

# Generasi paralel (jika didukung)
naeos run --input-file spec.yaml --parallel
```

### Penggunaan memori tinggi

Spesifikasi besar dengan 100+ modul mungkin menggunakan memori signifikan:

```bash
# Pantau penggunaan memori
naeos run --input-file spec.yaml --profile memory

# Proses modul dalam batch
naeos run --input-file spec.yaml --batch-size 20
```

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
