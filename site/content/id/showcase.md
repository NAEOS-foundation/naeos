---
title: Showcase
description: Proyek dan sistem nyata yang dibangun dengan NAEOS.
---

Lihat bagaimana tim menggunakan NAEOS untuk membangun, memvalidasi, dan mengembangkan sistem mereka. Ingin menambahkan proyek Anda? [Kirim showcase](https://github.com/NAEOS-foundation/naeos/discussions/new?category=show-and-tell).

<div class="showcase-grid">
  <div class="showcase-card">
    <div class="showcase-card-header">
      <span class="showcase-badge">Microservices</span>
      <span class="showcase-badge">Produksi</span>
    </div>
    <h3>Platform E-Commerce</h3>
    <p>12 microservices dalam satu spesifikasi. Generasi kode Go + TypeScript dengan manifes deployment Kubernetes.</p>
    <div class="showcase-meta">
      <span class="showcase-stat"><strong>12</strong> layanan</span>
      <span class="showcase-stat"><strong>Go, TS</strong> bahasa</span>
      <span class="showcase-stat"><strong>K8s</strong> deploy</span>
    </div>
  </div>

  <div class="showcase-card">
    <div class="showcase-card-header">
      <span class="showcase-badge">AI</span>
      <span class="showcase-badge">Open Source</span>
    </div>
    <h3>Platform Chat AI</h3>
    <p>Layanan GenAI multi-provider dengan OpenAI dan Anthropic. Menggunakan bundel konteks AI NAEOS.</p>
    <div class="showcase-meta">
      <span class="showcase-stat"><strong>4</strong> modul</span>
      <span class="showcase-stat"><strong>Python, TS</strong> bahasa</span>
      <span class="showcase-stat"><strong>2</strong> penyedia AI</span>
    </div>
  </div>

  <div class="showcase-card">
    <div class="showcase-card-header">
      <span class="showcase-badge">Serverless</span>
      <span class="showcase-badge">Startup</span>
    </div>
    <h3>Pipeline Analitik Serverless</h3>
    <p>Platform analitik berbasis event dengan Lambda, SQS, dan DynamoDB.</p>
    <div class="showcase-meta">
      <span class="showcase-stat"><strong>8</strong> fungsi</span>
      <span class="showcase-stat"><strong>Python</strong> bahasa</span>
      <span class="showcase-stat"><strong>AWS</strong> infra</span>
    </div>
  </div>

  <div class="showcase-card">
    <div class="showcase-card-header">
      <span class="showcase-badge">Enterprise</span>
      <span class="showcase-badge">Kepatuhan</span>
    </div>
    <h3>Platform FinTech</h3>
    <p>SOC 2 compliant dengan tata kelola NAEOS: jejak audit, RBAC, dan kebijakan enkripsi.</p>
    <div class="showcase-meta">
      <span class="showcase-stat"><strong>20+</strong> layanan</span>
      <span class="showcase-stat"><strong>Go, Java</strong> bahasa</span>
      <span class="showcase-stat"><strong>SOC 2</strong> compliant</span>
    </div>
  </div>

  <div class="showcase-card">
    <div class="showcase-card-header">
      <span class="showcase-badge">Hexagonal</span>
      <span class="showcase-badge">Greenfield</span>
    </div>
    <h3>CMS Arsitektur Bersih</h3>
    <p>Arsitektur hexagonal dengan domain-driven design. REST, gRPC, dan CLI dari model NEIR yang sama.</p>
    <div class="showcase-meta">
      <span class="showcase-stat"><strong>5</strong> lapis</span>
      <span class="showcase-stat"><strong>Go</strong> bahasa</span>
      <span class="showcase-stat"><strong>REST+gRPC</strong> API</span>
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
