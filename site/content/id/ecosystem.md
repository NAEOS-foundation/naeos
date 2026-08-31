---
title: Ekosistem
description: Profil, plugin, integrasi, dan ekstensi komunitas untuk NAEOS.
---

## Profil Bawaan

NAEOS hadir dengan 5 profil industri yang menyediakan pola, aturan, dan template yang sudah dikonfigurasi sebelumnya untuk domain umum.

<div class="eco-card" style="margin-bottom:1rem;">
<h3>Profil SaaS</h3>
<p>Arsitektur multi-tenant, manajemen langganan, rate limiting API, dan pola RBAC untuk aplikasi Software-as-a-Service.</p>
<ul class="eco-list">
<li><span class="eco-dot" style="background:#08d6ff;"></span> Pola database multi-tenant</li>
<li><span class="eco-dot" style="background:#60a5fa;"></span> Integrasi langganan & penagihan</li>
<li><span class="eco-dot" style="background:#fbbf24;"></span> Manajemen kunci API & rate limiting</li>
</ul>
</div>

<div class="eco-card" style="margin-bottom:1rem;">
<h3>Profil AI Agent</h3>
<p>Arsitektur berbasis agen, pola integrasi LLM, scaffold penggunaan alat, dan manajemen konteks untuk aplikasi berbasis AI.</p>
<ul class="eco-list">
<li><span class="eco-dot" style="background:#08d6ff;"></span> Pola orkestrasi agen</li>
<li><span class="eco-dot" style="background:#60a5fa;"></span> Lapisan abstraksi penyedia LLM</li>
<li><span class="eco-dot" style="background:#fbbf24;"></span> Penggunaan alat dan function calling</li>
</ul>
</div>

<div class="eco-card" style="margin-bottom:1rem;">
<h3>Profil FinTech</h3>
<p>Pola domain keuangan, pemrosesan transaksi, jejak audit, dan aturan kepatuhan untuk aplikasi teknologi keuangan.</p>
<ul class="eco-list">
<li><span class="eco-dot" style="background:#08d6ff;"></span> Pemrosesan transaksi & buku besar</li>
<li><span class="eco-dot" style="background:#60a5fa;"></span> Jejak audit & logging kepatuhan</li>
<li><span class="eco-dot" style="background:#fbbf24;"></span> Penegakan aturan regulasi</li>
</ul>
</div>

<div class="eco-card" style="margin-bottom:1rem;">
<h3>Profil Kesehatan</h3>
<p>Pola kepatuhan HIPAA, integrasi API FHIR, manajemen data pasien, dan kontrol keamanan untuk aplikasi kesehatan.</p>
<ul class="eco-list">
<li><span class="eco-dot" style="background:#08d6ff;"></span> Scaffolding kepatuhan HIPAA</li>
<li><span class="eco-dot" style="background:#60a5fa;"></span> Definisi sumber daya FHIR</li>
<li><span class="eco-dot" style="background:#fbbf24;"></span> Pola penanganan data PHI</li>
</ul>
</div>

<div class="eco-card">
<h3>Profil Pemerintah</h3>
<p>Pola sistem pemerintah, kepatuhan regulasi, manajemen dokumen, dan standar keamanan untuk aplikasi sektor publik.</p>
<ul class="eco-list">
<li><span class="eco-dot" style="background:#08d6ff;"></span> Kerangka kepatuhan regulasi</li>
<li><span class="eco-dot" style="background:#60a5fa;"></span> Otomatisasi alur kerja dokumen</li>
    <li><span class="eco-dot" style="background:#fbbf24;"></span> Penegakan standar keamanan</li>
</ul>
</div>

## Sistem Plugin

Perluas NAEOS dengan WASM dan plugin native. Plugin SDK memudahkan pembuatan:

- **Generator kode** — Tambah adapter bahasa baru
- **Validator** — Aturan validasi kustom
- **Deployer** — Deploy ke platform apa pun
- **Analyzers** — Analisis dan pelaporan kustom

## Marketplace

Publikasikan dan temukan profil, plugin, dan template melalui NAEOS Marketplace.

- Cari dan instal dengan `naeos marketplace`
- Terverifikasi SHA-256 untuk keamanan
- Manajemen versi dan resolusi dependensi
- Rating dan ulasan komunitas