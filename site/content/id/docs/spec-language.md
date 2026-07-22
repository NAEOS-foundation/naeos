---
title: Bahasa Spesifikasi
description: Bahasa Spesifikasi NAEOS v2 — sintaks, fitur, dan kemampuan.
---

## Ikhtisar

Bahasa Spesifikasi NAEOS (NSL) v2 adalah bahasa deklaratif berbasis YAML untuk mendeskripsikan sistem rekayasa. Satu file spesifikasi menangkap struktur proyek, arsitektur, modul, layanan, dependensi, dan target deployment.

## Fitur Sintaks

### Interpolasi Variabel

Gunakan sintaks `${var}` untuk mereferensikan nilai dalam spesifikasi yang sama:

```yaml
project: my-service
version: "1.0"
modules:
  - name: api
    path: ./services/${project}
```

### Variabel Lingkungan

Resolusi variabel lingkungan saat runtime pipeline dengan `$env{VAR}`:

```yaml
services:
  - name: api
    port: $env{API_PORT}
    host: $env{HOST}
```

### Referensi Silang

Referensi bagian lain dari spesifikasi dengan `$ref{path}`:

```yaml
modules:
  - name: auth
    path: ./auth
  - name: api
    path: ./api
    dependencies: [$ref{modules[0].name}]
```

### Komposisi Multi-File

Komposisi spesifikasi dari banyak file dengan `$include{file}`:

```yaml
project: my-app
$include{./shared/modules.yaml}
$include{./shared/services.yaml}
```

### Fungsi Kustom

Transformasi nilai dengan fungsi bawaan menggunakan `$fn{name(args)}`:

| Fungsi | Deskripsi | Contoh |
|--------|-----------|--------|
| `upper` | Huruf kapital | `$fn{upper(name)}` |
| `lower` | Huruf kecil | `$fn{lower(name)}` |
| `slug` | Slug URL | `$fn{slug(project)}` |
| `default` | Nilai default | `$fn{default(port, 8080)}` |
| `len` | Panjang list/string | `$fn{len(modules)}` |
| `coalesce` | Non-null pertama | `$fn{coalesce(host, localhost)}` |

### Bagian Bersyarat

Gunakan `$if{condition}` dan `$endif` untuk blok bersyarat:

```yaml
services:
  - name: api
$if{env == production}
    replicas: 3
    resources:
      cpu: "2"
      memory: 4Gi
$endif
```

### Versioning Skema

Setiap spesifikasi menyertakan bidang versi dengan validasi otomatis:

```yaml
project: my-app
version: "1.0"
```

Pipeline memvalidasi bahwa versi spesifikasi memenuhi versi minimum yang diperlukan (v0.1.0).

## Contoh Lengkap

```yaml
project: ecommerce-platform
version: "2.0"
$include{./profiles/saas.yaml}

modules:
  - name: api-gateway
    path: ./gateway
    dependencies: [user-service, product-service]
  - name: user-service
    path: ./services/users
    dependencies: [database]
  - name: product-service
    path: ./services/products
    dependencies: [database, search-engine]
  - name: database
    path: ./infra/db
  - name: search-engine
    path: ./infra/search

services:
  - name: gateway
    kind: reverse-proxy
    port: $fn{default($env{GATEWAY_PORT}, 8080)}
  - name: user-api
    kind: rest
    port: 9001
  - name: product-api
    kind: rest
    port: 9002

architecture:
  pattern: microservices

generation:
  languages: [go, typescript, python]
  output_dir: ./generated
```

## Praktik Terbaik

- Gunakan `$include` untuk memisahkan spesifikasi besar menjadi modul yang mudah dikelola
- Gunakan `$env{VAR}` untuk nilai spesifik lingkungan
- Gunakan `$ref` daripada menduplikasi nilai
- Tambahkan versioning skema ke semua spesifikasi
- Jaga spesifikasi tetap mudah dibaca — komentar YAML dipertahankan
