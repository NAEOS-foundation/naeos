# NAEOS Presentation — Side-by-Side Brief

> **Tujuan:** panduan presentasi yang memasangkan pesan untuk audiens dengan bukti teknis yang dapat diperiksa langsung di repository. Dokumen ini adalah materi pendamping, bukan pengganti spesifikasi normatif.

## Cara menggunakan dokumen ini

Setiap bagian memiliki dua sisi:
- **Sisi presentasi:** kalimat ringkas yang dapat ditampilkan atau disampaikan kepada audiens.
- **Sisi bukti repository:** artefak yang dapat dibuka untuk memvalidasi penjelasan.

Gunakan klaim yang sesuai dengan bukti. Demo lokal membuktikan jalur yang diuji pada kondisi tersebut; demo bukan bukti otomatis atas kesiapan produksi, kepatuhan, adopsi pelanggan, atau keamanan seluruh integrasi eksternal.

## 1. Pesan utama

| Sisi presentasi | Sisi bukti repository |
|---|---|
| **NAEOS adalah engineering control plane untuk AI coding agents.** NAEOS membantu tim menyatakan intent engineering, memeriksa kebijakan, membatasi eksekusi, serta mengumpulkan bukti yang dapat ditinjau. | [README.md](../README.md) — positioning, model konseptual, dan trust principles. |
| **Bukan sekadar AI coding assistant.** Fokusnya adalah lapisan engineering di sekitar agent: aturan, otorisasi, observasi, evidence, dan verifikasi. | [README.md](../README.md) bagian “Why NAEOS?”, “The model”, dan “Trust model”. |

**Narasi pembicara:** “AI dapat mempercepat penulisan kode. Namun, tim engineering tetap perlu mengetahui intent yang dinyatakan, tindakan yang diizinkan, apa yang benar-benar terjadi, dan bukti yang mendukung hasilnya. NAEOS dirancang untuk menghubungkan kontrol-kontrol itu.”

## 2. Masalah → pendekatan

| Tanpa control plane yang eksplisit | Dengan model NAEOS |
|---|---|
| Prompt → AI agent → kode/tindakan → status dan bukti mungkin tidak terhubung dengan jelas. | Specification → NEIR → validation + policy → agent context/intent → authorized execution → observation/evidence → independent verification. |
| Sulit menelusuri hubungan antara intent, keputusan kebijakan, eksekusi, dan output jika jejaknya terpisah. | Artefak run dapat menyediakan identifier dan metadata untuk menelusuri input, representasi, tahapan, dan output. |

**Bukti:** [README.md](../README.md), [docs/REFERENCE-DEMO.md](REFERENCE-DEMO.md).

**Catatan presisi:** jangan menyatakan bahwa semua eksekusi agent atau integrasi eksternal otomatis aman. NAEOS menyediakan model kontrol dan bukti; setiap deployment tetap perlu dievaluasi pada batas kepercayaan dan integrasi yang digunakan.

## 3. Arsitektur secara sederhana

| Sisi presentasi | Sisi bukti repository |
|---|---|
| **1. Declare** — tim mendefinisikan engineering intent dalam specification YAML/JSON. | [site/content/docs/architecture.md](../site/content/docs/architecture.md) — Input Layer. |
| **2. Normalize** — intent diproses menjadi representasi engineering kanonis, yaitu NEIR. | [GLOSSARY.md](../GLOSSARY.md) — definisi Specification, NEIR, dan Pipeline. |
| **3. Validate & govern** — spesifikasi divalidasi dan policy dievaluasi sebelum tahapan berikutnya. | [docs/REFERENCE-DEMO.md](REFERENCE-DEMO.md) — evidence untuk validation dan deliberate policy rejection. |
| **4. Prepare agent context / generate** — konteks dan artefak disusun melalui pipeline. | [site/content/docs/architecture.md](../site/content/docs/architecture.md) — Generation dan Output Layers. |
| **5. Observe & verify** — hasil dan metadata run ditinjau; klaim harus sesuai dengan evidence yang tersedia. | [docs/REFERENCE-DEMO.md](REFERENCE-DEMO.md) — traceability contract dan reviewer checklist. |

### Diagram yang dapat dipakai pada slide

```text
Engineering intent
      |
      v
Specification (YAML / JSON)
      |
      v
NEIR — canonical engineering representation
      |
      v
Validation + policy evaluation
      |
      v
Agent context / authorized generation
      |
      v
Run metadata + generated artifacts
      |
      v
Reviewable evidence / independent verification
```

## 4. Demo: tunjukkan, jangan hanya klaim

| Apa yang disampaikan | Apa yang ditunjukkan |
|---|---|
| “Kita dapat memulai dari specification yang dapat dibaca.” | `spec.yaml` |
| “NAEOS membentuk representasi engineering yang dapat diperiksa.” | `inspect.json` |
| “Validasi menghasilkan output yang dapat ditinjau.” | `validate.json` |
| “Konfigurasi yang melanggar policy dapat ditolak pada jalur demo.” | `invalid-policy.log` |
| “Run menghasilkan metadata traceability.” | `run.json` — periksa `run_id`, `specification_hash`, `neir_hash`, serta metadata validation/policy/context/audit/stage yang tersedia. |
| “Hasilnya bukan hanya narasi; ada artefak yang bisa diperiksa.” | Folder `generated/` dan `summary.md`. |

Sumber kanonis: [docs/REFERENCE-DEMO.md](REFERENCE-DEMO.md) dan [docs/GOLDEN-PATH.md](GOLDEN-PATH.md).

### Jalankan demo referensi

Dari checkout repository yang bersih:

```bash
go build -o naeos ./cmd/naeos
NAEOS_DEMO_OUTPUT_DIR=/tmp/naeos-demo ./examples/demo-cli/run-demo.sh
find /tmp/naeos-demo -maxdepth 3 -type f | sort
```

Demo default tidak memerlukan API key LLM; tahap kompilasi AI opsional hanya berjalan jika `NAEOS_LLM_API_KEY` dikonfigurasi. Ikuti petunjuk terbaru di dokumen demo bila command atau prasyarat berubah.

## 5. Nilai bagi tim engineering

| Pesan untuk audiens | Cara membuktikannya / batas klaim |
|---|---|
| **Kejelasan intent:** specification menjadi input yang eksplisit. | Tunjukkan `spec.yaml`; jangan menyamakan spesifikasi yang tertulis dengan kebenaran implementasi secara otomatis. |
| **Representasi konsisten:** NEIR menjadi representasi kanonis untuk pemrosesan berikutnya. | Tunjukkan `inspect.json` dan kaitannya dengan specification. |
| **Policy sebagai gate:** policy dapat menolak konfigurasi yang tidak diperbolehkan pada jalur yang diuji. | Tunjukkan `invalid-policy.log`; jangan menggeneralisasi satu tes menjadi bukti semua batas policy. |
| **Traceability:** metadata menghubungkan input, representasi, proses, dan output. | Tunjukkan field yang benar-benar tersedia pada `run.json`; catat commit dan hasil run saat evaluasi. |
| **Vendor-neutral engineering layer:** konteks engineering tidak dimaksudkan bergantung pada satu agent saja. | README mendokumentasikan target adapter/context; uji tiap integrasi yang relevan sebelum mengklaim kompatibilitas operasional. |

## 6. Siapa yang sebaiknya mengevaluasi NAEOS?

- **Engineering lead / platform team:** apakah model specification dan pipeline cocok dengan proses engineering tim?
- **AI engineering team:** apakah agent context dan adapter yang tersedia sesuai dengan toolchain mereka?
- **Security / governance reviewer:** apakah policy, otorisasi, audit, dan evidence memenuhi kebutuhan threat model mereka?
- **Potential design partner:** dapatkah tim menjalankan demo, merekam hasil, lalu menyepakati satu pilot dengan acceptance criteria yang terukur?

Rujukan awal untuk pilot: [docs/REFERENCE-DEMO.md](REFERENCE-DEMO.md) bagian “Partner-pilot use” dan [docs/EXTERNAL-VALIDATION.md](EXTERNAL-VALIDATION.md).

## 7. Klaim yang aman vs. klaim yang harus dihindari

| Aman bila sesuai bukti | Hindari tanpa bukti tambahan |
|---|---|
| “Demo lokal menunjukkan jalur dari specification sampai artefak dan metadata run.” | “NAEOS menjamin seluruh kode yang dihasilkan aman.” |
| “Demo menunjukkan penolakan pada konfigurasi policy yang sengaja dibuat tidak valid.” | “Semua tindakan agent selalu diblokir jika berisiko.” |
| “Repository menyediakan artefak untuk inspeksi dan evaluasi independen.” | “NAEOS sudah terbukti production-ready untuk semua skala.” |
| “Arsitektur dirancang untuk mendukung berbagai AI coding agents.” | “Semua agent dan integrasi telah teruji penuh.” |
| “Proyek menyediakan mekanisme dan eksperimen evidence/verification tertentu.” | “Semua mekanisme tersebut secara otomatis memenuhi standar compliance tertentu.” |

## 8. Alur presentasi 7 menit

1. **0:00–0:45 — Problem:** kecepatan coding meningkat, tetapi kontrol engineering harus tetap eksplisit.
2. **0:45–1:30 — Positioning:** NAEOS sebagai engineering control plane, bukan coding assistant baru.
3. **1:30–2:45 — Architecture:** specification → NEIR → validation/policy → context/execution → evidence.
4. **2:45–5:15 — Live demo:** jalankan Golden Path dan buka evidence output.
5. **5:15–6:15 — Trust boundary:** jelaskan apa yang dibuktikan demo dan apa yang masih perlu diuji.
6. **6:15–7:00 — Ask:** ajak audiens menjalankan evaluasi teknis kecil dengan acceptance criteria dan evidence yang disepakati.

## 9. Ajakan kolaborasi

**Usulan:** lakukan evaluasi teknis kecil berbasis repository, bukan langsung meminta komitmen besar. Sepakati use case, toolchain, batas akses, acceptance criteria, dan bukti yang harus diserahkan. Catat commit SHA, versi toolchain, command yang dijalankan, exit status, artefak yang dihasilkan, dan setiap penyimpangan dari kriteria.

## Sumber utama

- [Repository NAEOS](https://github.com/NAEOS-foundation/naeos)
- [README](../README.md)
- [Reference Demo & Evidence Story](REFERENCE-DEMO.md)
- [Golden Path](GOLDEN-PATH.md)
- [Architecture](../site/content/docs/architecture.md)
- [Glossary](../GLOSSARY.md)

---

**Prinsip presentasi:** tunjukkan klaim di satu sisi dan bukti yang bisa diperiksa di sisi lain. Jika bukti belum ada, sebutkan sebagai tujuan atau hipotesis, bukan kemampuan yang telah terbukti.
