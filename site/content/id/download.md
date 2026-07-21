---
title: Unduh
description: Pasang NAEOS dan mulai rekayasa proyek Anda.
---

## Metode Instalasi

<div class="download-grid">
<div class="download-card">
<h3>Go Install</h3>
<p>Pasang langsung menggunakan Go. Membutuhkan Go 1.25+.</p>
<div class="code-block">
<div class="code-block-header"><span>bash</span><button class="copy-btn">Copy</button></div>
<pre><code>go install github.com/NAEOS-foundation/naeos/cmd/naeos@latest</code></pre>
</div>
</div>

<div class="download-card">
<h3>Docker</h3>
<p>Jalankan menggunakan kontainer Docker.</p>
<div class="code-block">
<div class="code-block-header"><span>bash</span><button class="copy-btn">Copy</button></div>
<pre><code>docker pull ghcr.io/naeos-foundation/naeos:latest
docker run --rm ghcr.io/naeos-foundation/naeos:latest naeos version</code></pre>
</div>
</div>

<div class="download-card">
<h3>Rilis Biner</h3>
<p>Unduh biner terbaru dari GitHub Releases.</p>
<a href="https://github.com/NAEOS-foundation/naeos/releases" class="btn btn-primary" target="_blank" rel="noopener">Lihat Rilis</a>
</div>

<div class="download-card">
<h3>Bangun dari Sumber</h3>
<p>Kloning repositori dan bangun secara manual.</p>
<div class="code-block">
<div class="code-block-header"><span>bash</span><button class="copy-btn">Copy</button></div>
<pre><code>git clone https://github.com/NAEOS-foundation/naeos.git
cd naeos
go build ./cmd/naeos/</code></pre>
</div>
</div>
</div>

## Dukungan Platform

| Platform | Dukungan |
|----------|----------|
| Linux (amd64) | ✅ |
| Linux (arm64) | ✅ |
| macOS (amd64) | ✅ |
| macOS (arm64) | ✅ |
| Windows (amd64) | ✅ |

## Verifikasi Instalasi

```bash
naeos version
```

## Mulai Cepat

Setelah instalasi, inisialisasi proyek pertama Anda:

```bash
naeos init
naeos run --help
```