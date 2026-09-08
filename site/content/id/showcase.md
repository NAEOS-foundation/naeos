---
title: Showcase
description: Proyek dan sistem nyata yang dibangun dengan NAEOS.
---

Contoh berikut menunjukkan pola yang dapat dimodelkan dengan NAEOS. Ini bersifat ilustratif, bukan referensi pelanggan. Ingin menambahkan proyek Anda? [Kirim showcase](https://github.com/NAEOS-foundation/naeos/discussions/new?category=show-and-tell).

<div class="showcase-grid">
  <div class="showcase-card">
    <div class="showcase-card-header">
      <span class="showcase-badge">Microservices</span>
      <span class="showcase-badge">Spec-driven</span>
    </div>
    <h3>Platform Microservices</h3>
    <p>Satu spesifikasi NAEOS dapat mendeskripsikan batas layanan, kontrak API, dan konfigurasi deployment untuk platform multi-service.</p>
    <div class="showcase-meta">
      <span class="showcase-stat"><strong>Satu</strong> spesifikasi</span>
      <span class="showcase-stat"><strong>Go, TS</strong> output</span>
      <span class="showcase-stat"><strong>Model</strong> bersama</span>
    </div>
  </div>

  <div class="showcase-card">
    <div class="showcase-card-header">
      <span class="showcase-badge">AI</span>
      <span class="showcase-badge">Context</span>
    </div>
    <h3>Arsitektur Produk AI</h3>
    <p>NAEOS dapat mengompilasi model NEIR tunggal menjadi bundel konteks AI untuk asisten pengkodean, menjaga arsitektur tetap selaras dengan implementasi.</p>
    <div class="showcase-meta">
      <span class="showcase-stat"><strong>Model</strong> NEIR</span>
      <span class="showcase-stat"><strong>Konteks</strong> AI</span>
      <span class="showcase-stat"><strong>Niat</strong> bersama</span>
    </div>
  </div>

  <div class="showcase-card">
    <div class="showcase-card-header">
      <span class="showcase-badge">Serverless</span>
      <span class="showcase-badge">Event-driven</span>
    </div>
    <h3>Sistem Berbasis Event</h3>
    <p>NAEOS dapat memodelkan alur asinkron, batas dependensi, dan artefak yang dihasilkan untuk layanan event-driven dan integrasi.</p>
    <div class="showcase-meta">
      <span class="showcase-stat"><strong>Alur</strong> async</span>
      <span class="showcase-stat"><strong>Artefak</strong> siap dibuat</span>
      <span class="showcase-stat"><strong>Validasi</strong> kebijakan</span>
    </div>
  </div>

  <div class="showcase-card">
    <div class="showcase-card-header">
      <span class="showcase-badge">Governance</span>
      <span class="showcase-badge">Kepatuhan</span>
    </div>
    <h3>Platform Berbasis Kebijakan</h3>
    <p>Aturan tata kelola, audit trail, dan template kebijakan yang berorientasi pada kepatuhan dapat diterapkan melalui pipeline validasi dan review NAEOS.</p>
    <div class="showcase-meta">
      <span class="showcase-stat"><strong>Rule</strong> engine</span>
      <span class="showcase-stat"><strong>Audit</strong> trail</span>
      <span class="showcase-stat"><strong>Review</strong> kebijakan</span>
    </div>
  </div>

  <div class="showcase-card">
    <div class="showcase-card-header">
      <span class="showcase-badge">Architecture</span>
      <span class="showcase-badge">NEIR</span>
    </div>
    <h3>Proyek Arsitektur Bersih</h3>
    <p>Model domain, batas adaptor, dan kontrak layanan dapat direpresentasikan dalam satu model NEIR dan didistribusikan ke beberapa target keluaran.</p>
    <div class="showcase-meta">
      <span class="showcase-stat"><strong>Satu</strong> model</span>
      <span class="showcase-stat"><strong>Banyak</strong> output</span>
      <span class="showcase-stat"><strong>Batas</strong> jelas</span>
    </div>
  </div>

  <div class="showcase-card showcase-card-add">
    <h3>Proyek Anda di Sini</h3>
    <p>Bangun sesuatu dengan NAEOS? Bagikan cerita Anda dengan komunitas.</p>
    <a href="https://github.com/NAEOS-foundation/naeos/discussions/new?category=show-and-tell" class="btn btn-primary btn-sm">Kirim Proyek</a>
  </div>
</div>

## Coba Sendiri — Proyek Demo

Proyek demo resmi tersedia di repositori: [`cmd/naeos/demo-app/`](https://github.com/NAEOS-foundation/naeos/tree/main/cmd/naeos/demo-app). Ini adalah aplikasi Go berarsitektur heksagonal yang sepenuhnya dijelaskan oleh spesifikasi.

```bash
# 1. Build artifact dari spesifikasi demo
naeos build --config cmd/naeos/demo-app/config.yaml --input cmd/naeos/demo-app/spec.yaml

# 2. Jalankan pipeline lengkap dengan output yang dapat dilacak
naeos run --config cmd/naeos/demo-app/config.yaml --input cmd/naeos/demo-app/spec.yaml

# 3. Distribusikan build ke banyak worker
naeos distributed --config cmd/naeos/demo-app/config.yaml --workers 4
```

Lebih suka mulai dari template? Scaffold starter microservices lengkap:

```bash
naeos template init microservices-go -o .
naeos build --config naeos.yaml --input spec.yaml
```

Kedua proyek bersifat open source dan cukup kecil untuk dibaca dalam hitungan menit — cara terbaik untuk melihat NAEOS beraksi.
