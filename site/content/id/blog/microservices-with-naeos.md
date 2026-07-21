---
title: Membangun Microservices dengan NAEOS — Panduan Langkah demi Langkah
description: Pelajari cara mendesain, menghasilkan, dan men-deploy microservices menggunakan spesifikasi deklaratif NAEOS.
date: 2026-06-15
---

Dalam tutorial ini, kita akan membangun aplikasi microservices lengkap menggunakan NAEOS. Anda akan belajar cara mendefinisikan layanan, mengelola dependensi, menghasilkan kode, dan melakukan deployment.

## Prasyarat

- NAEOS terinstal ([lihat panduan instalasi](/id/docs/installation/))
- Pengetahuan dasar tentang YAML dan konsep microservices

## Langkah 1: Tentukan Spesifikasi

Buat file `spec.yaml` yang mendeskripsikan arsitektur microservices Anda:

```yaml
project: ecommerce-platform
version: "1.0"
modules:
  - name: api-gateway
    path: ./api-gateway
    description: API Gateway entry point
    dependencies: [user-service, order-service, product-service]
  - name: user-service
    path: ./services/users
    description: User management and authentication
    dependencies: [shared-db]
  - name: order-service
    path: ./services/orders
    description: Order processing
    dependencies: [user-service, payment-service]
  - name: product-service
    path: ./services/products
    description: Product catalog
    dependencies: [shared-db]
  - name: payment-service
    path: ./services/payments
    description: Payment processing
  - name: shared-db
    path: ./infra/db
    description: Shared database layer
services:
  - name: gateway
    kind: reverse-proxy
    port: 8080
  - name: user-api
    kind: rest
    port: 9001
  - name: order-api
    kind: rest
    port: 9002
  - name: product-api
    kind: rest
    port: 9003
architecture:
  pattern: microservices
generation:
  languages: [go, typescript]
```

## Langkah 2: Jalankan Pipeline

```bash
naeos run --input-file spec.yaml
```

## Kesimpulan

Anda sekarang memiliki proyek microservices lengkap dengan kode yang dihasilkan — dari satu file spesifikasi.