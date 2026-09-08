# Demo NAEOS untuk Kebutuhan Marketing

## Tujuan demo
Menunjukkan bahwa NAEOS bukan sekadar generator kode, melainkan platform engineering deklaratif yang mengubah spesifikasi menjadi sistem yang tervalidasi dan dapat dipakai oleh AI serta tim engineering.

## Narasi utama
"AI bisa menghasilkan kode. NAEOS membantu menyusun sistem engineering di sekitarnya."

## Durasi yang disarankan
- 45 detik: versi social / teaser
- 2 menit: versi pitch deck
- 5 menit: versi demo teknis untuk komunitas atau investor

---

## Versi 45 detik (teaser / LinkedIn / X)

### Hook
"Ketika tim engineering bekerja dengan AI, masalahnya bukan hanya kecepatan menulis kode. Masalahnya adalah konsistensi sistem."

### Script
"NAEOS mengubah satu spesifikasi menjadi model engineering yang dapat divalidasi, diatur, dan dipakai kembali."

"Mulai dari spesifikasi YAML, NAEOS melakukan parsing, normalisasi, resolusi, membangun NEIR, memvalidasi, lalu menghasilkan artefak dan konteks AI."

"Jadi, bukan hanya kode yang dihasilkan. Yang dihasilkan adalah sistem yang lebih terstruktur, lebih konsisten, dan lebih siap untuk governance."

"Tentukan sekali. Bangun di mana saja."

### CTA
"Lihat repositori NAEOS dan coba quick start."

---

## Versi 2 menit (pitch marketing)

### Visual yang muncul
1. Judul: "One specification. One engineering model."
2. Kode YAML sample di sisi kiri
3. Pipeline visual: Parse → Normalize → Resolve → NEIR → Validate → Generate → AI Context
4. Hasil output: project skeleton, CI/CD config, docs, and AI instructions

### Voiceover
"NAEOS adalah platform engineering deklaratif. Tim hanya mendefinisikan sistem sekali, lalu NAEOS membangun model engineering internal yang konsisten."

"Dalam praktiknya, NAEOS membaca spesifikasi, memvalidasi struktur, membangun representasi NEIR, lalu menghasilkan artefak yang siap dipakai."

"Tidak hanya menulis kode. NAEOS membantu mesin dan tim memahami arsitektur, kebijakan, dependency, dan konteks proyek di seluruh siklus hidupnya."

"Jadi, saat AI membantu pengembangan, tim tidak bergantung pada prompt yang tersebar. Mereka punya satu sumber kebenaran: spesifikasi."

"Dari sana, NAEOS dapat menghasilkan kode, plugin, dokumentasi, dan konteks AI yang siap dikirim ke alat seperti Copilot, Claude Code, Cursor, Gemini CLI, Codex, OpenCode, dan Windsurf."

"Itu yang membedakan NAEOS dari scaffolding sederhana. NAEOS berpikir dalam istilah sistem, bukan hanya template."

### Closing statement
"NAEOS membawa kejelasan, traceability, dan governance ke dalam pengembangan berbasis AI."

---

## Versi demo teknis 5 menit

### Skenario
Tampilkan contoh proyek sederhana: aplikasi dengan modul `auth`, `api`, dan `gateway`.

### Alur demo

#### 1. Mulai dari spesifikasi
Tampilkan file `spec.yaml` seperti ini:

```yaml
project: my-app
modules:
  - name: auth
    path: ./auth
  - name: api
    path: ./api
    dependencies: [auth]
services:
  - name: gateway
    kind: http
    port: 8080
architecture:
  pattern: hexagonal
generation:
  languages: [go, typescript]
```

Narasi:
"Tim hanya mendefinisikan niat engineering sekali. Ini menjadi sumber kebenaran untuk seluruh pipeline."

#### 2. Pipeline internal
Tampilkan diagram:

Parse → Normalize → Resolve → Build NEIR → Validate → Schedule → Generate

Narasi:
"NAEOS tidak membuat code secara acak. Ia membangun model internal yang mewakili struktur, dependency, dan arsitektur proyek."

#### 3. Validasi dan governance
Tampilkan hasil validasi:
- circular dependency tidak ditemukan
- port conflict tidak ditemukan
- module boundaries sesuai rencana

Narasi:
"Sebelum output dibuat, NAEOS mengevaluasi kebijakan dan validasi untuk mengurangi drift dan kesalahan arsitektur."

#### 4. Hasil generated project
Tampilkan output yang dihasilkan:
- folder `auth/`
- folder `api/`
- dokumen API
- konfigurasi deployment
- file untuk AI instruction bundle

Narasi:
"Hasilnya bukan hanya template. Ini adalah artefak yang siap dipakai oleh tim dan AI."

#### 5. AI context bundle
Tampilkan contoh AI bundles untuk:
- GitHub Copilot
- Claude Code
- Cursor
- Gemini CLI
- Codex
- OpenCode
- Windsurf

Narasi:
"NAEOS menyiapkan konteks AI berdasarkan satu model yang konsisten, bukan prompt yang disalin berulang kali."

#### 6. Closing
"Ketika engineering menjadi spesifikasi, project bisa lebih konsisten, lebih valid, dan lebih mudah ditelusuri."

---

## Kutipan yang bisa dipakai di slide / pitch

- "AI bisa menghasilkan kode. NAEOS membantu struktur sistem di sekitarnya."
- "Specify once. Build anywhere."
- "NAEOS turns engineering intent into a validated, reusable system model."
- "From specification to AI context, governance, and generated artifacts."
- "The real bottleneck in AI-assisted development is not code generation — it is engineering consistency."

---

## CTA pada akhir demo
- Explore GitHub
- Try NAEOS
- Read the architecture
- Build your first spec

---

## Script ringkas untuk presenter

"Hari ini saya akan menunjukkan bagaimana NAEOS bekerja sebagai platform engineering deklaratif. Kita mulai dari satu spesifikasi, lalu NAEOS memprosesnya melalui parsing, normalisasi, resolusi, dan pembuatan NEIR. Setelah validasi, sistem menghasilkan artefak dan konteks AI yang konsisten. Yang membedakan NAEOS dari scaffolding biasa adalah pendekatannya terhadap sistem, bukan hanya kode. Dengan satu model, tim bisa menjaga alignment, governance, dan traceability di seluruh lifecycle yang lebih rapi."
