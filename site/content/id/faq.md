---
title: Pertanyaan Umum
description: Pertanyaan umum tentang NAEOS dan rekayasa deklaratif.
---

<div class="faq-list">
<div class="faq-item">
<button class="faq-question">
<span>Apa itu NAEOS?</span>
<span class="faq-arrow">▾</span>
</button>
<div class="faq-answer">
<p>NAEOS (Nusantara Engineering & Architecture Operating System) adalah platform rekayasa deklaratif yang mengubah spesifikasi YAML/JSON menjadi sistem perangkat lunak multi-bahasa yang tervalidasi. NAEOS adalah runtime rekayasa — bukan sekadar generator proyek — yang memahami spesifikasi, membangun model internal (NEIR), mengatur rencana eksekusi, menghasilkan artefak, memvalidasi hasil, dan menjaga proyek tetap selaras dengan spesifikasi sepanjang siklus hidupnya.</p>
</div>
</div>

<div class="faq-item">
<button class="faq-question">
<span>Apa bedanya NAEOS dengan generator proyek biasa?</span>
<span class="faq-arrow">▾</span>
</button>
<div class="faq-answer">
<p>Tidak seperti generator proyek statis (seperti create-react-app atau cookiecutters), NAEOS adalah runtime rekayasa lengkap. NAEOS tidak hanya membuat file sekali — ia memelihara hubungan berkelanjutan antara spesifikasi dan kode Anda. NAEOS memvalidasi, mengkompilasi ke set instruksi AI, menghasilkan dokumentasi, dan beradaptasi saat spesifikasi Anda berkembang.</p>
</div>
</div>

<div class="faq-item">
<button class="faq-question">
<span>Bahasa apa saja yang didukung?</span>
<span class="faq-arrow">▾</span>
</button>
<div class="faq-answer">
<p>NAEOS saat ini mendukung生成 kode untuk Go, TypeScript, Python, Java, dan Rust. Setiap bahasa memiliki adapter khusus yang mengikuti praktik terbaik dan pola idiomatis untuk bahasa tersebut.</p>
</div>
</div>

<div class="faq-item">
<button class="faq-question">
<span>Asisten coding AI apa yang didukung?</span>
<span class="faq-arrow">▾</span>
</button>
<div class="faq-answer">
<p>NAEOS mengkompilasi spesifikasi menjadi set instruksi AI untuk 6 platform: GitHub Copilot, Claude Code, Cursor, Gemini CLI, Codex, dan OpenCode. Setiap adapter menghasilkan file konteks khusus platform yang membantu asisten AI Anda memahami arsitektur proyek Anda.</p>
</div>
</div>

<div class="faq-item">
<button class="faq-question">
<span>Apakah saya perlu tahu Go untuk menggunakan NAEOS?</span>
<span class="faq-arrow">▾</span>
</button>
<div class="faq-answer">
<p>Tidak. NAEOS adalah alat CLI yang ditulis dalam Go, tetapi Anda hanya perlu menulis spesifikasi YAML/JSON. Outputnya bisa dalam salah satu dari 5 bahasa yang didukung. Anda tidak memerlukan pengetahuan Go untuk menggunakan NAEOS secara efektif.</p>
</div>
</div>

<div class="faq-item">
<button class="faq-question">
<span>Apakah NAEOS gratis?</span>
<span class="faq-arrow">▾</span>
</button>
<div class="faq-answer">
<p>Ya. NAEOS sepenuhnya gratis dan open source di bawah Lisensi Apache 2.0. Anda dapat menggunakannya untuk proyek pribadi, aplikasi komersial, atau deployment enterprise tanpa biaya lisensi apa pun.</p>

<p>Lihat <a href="/id/docs/getting-started/">Panduan Awal</a> untuk memulai. Pasang NAEOS via Go, Docker, atau unduh biner dari GitHub Releases. Buat spesifikasi YAML, jalankan <code>naeos run</code>, dan dalam hitungan menit Anda akan memiliki kode yang dihasilkan.</p>
</div>
</div>
</div>