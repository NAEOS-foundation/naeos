---
title: Model NEIR
description: NAEOS Engineering Intermediate Representation — model sistem kanonikal.
---

## Ikhtisar

NEIR (NAEOS Engineering Intermediate Representation) adalah model kanonikal yang merepresentasikan seluruh sistem yang sedang direkayasa. Model ini adalah sumber kebenaran tunggal yang mengalir melalui pipeline, memungkinkan semua pemrosesan hilir — generasi kode, kompilasi AI, dokumentasi, dan deployment.

## Komponen Inti

### Metadata Proyek

Informasi tingkat atas tentang sistem yang dibangun: nama, versi, deskripsi, pola arsitektur, model domain, dan informasi tim.

### Struktur Modul

Graf modul menangkap semua komponen kode dan hubungannya: definisi modul dengan jalur dan tipe, tepi dependensi dengan kendala versi, grup modul, dan ruang nama.

### Definisi Layanan

Layanan mewakili komponen yang dapat dijalankan: tipe layanan (REST, GraphQL, WebSocket, gRPC), pemetaan port dan protokol, definisi endpoint, middleware, dan konfigurasi health check.

### Kontrak API

Definisi API termasuk endpoint REST dengan metode dan jalur, tipe skema GraphQL, tipe event WebSocket, skema request/response, dan aturan autentikasi.

### Penyimpanan & Database

Konfigurasi lapisan data: engine database (PostgreSQL, Redis, MongoDB), skema tabel/koleksi, konfigurasi migrasi, pengaturan connection pool, kebijakan backup dan replikasi.

### Infrastruktur & Cloud

Definisi infrastruktur sebagai kode: sumber daya Kubernetes, konfigurasi Docker, sumber daya cloud (AWS, GCP, Azure), konfigurasi jaringan dan VPC, aturan load balancer dan auto-scaling.

### Keamanan & Kebijakan

Model keamanan: penyedia autentikasi dan metode, definisi peran RBAC, aturan dan kendala kebijakan, pengaturan enkripsi, konfigurasi jejak audit.

### Integrasi AI

Konfigurasi khusus AI: penyedia LLM, pemilihan model dan parameter, template prompt, definisi alat dan fungsi, pengaturan orkestrasi agen.

### Deployment & CI/CD

Konfigurasi deployment: definisi lingkungan (dev, staging, production), template pipeline CI/CD, pengaturan blue/green dan canary, kebijakan rollback.

## Mengakses Model NEIR

```bash
# Ekspor model NEIR sebagai JSON
naeos export --format json --output neir.json

# Periksa model
naeos kernel --model

# Validasi model
naeos validate --model
```

## NEIR dalam Konteks AI

Model NEIR juga digunakan untuk menghasilkan bundel konteks AI. Ketika dikompilasi, ia menghasilkan set instruksi yang sadar-arsitektur yang membantu asisten coding AI memahami konteks sistem penuh.
