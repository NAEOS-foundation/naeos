# Glossary

Berikut istilah utama yang sering digunakan dalam ekosistem NAEOS.

## Istilah inti
- NAEOS: Nusantara Engineering & Architecture Operating System — platform engineering deklaratif open-source.
- NEIR: NAEOS Engineering Intermediate Representation — model antara terpadu yang merepresentasikan seluruh proyek.
- Spec: Specification — dokumen YAML atau JSON yang mendefinisikan proyek, modul, layanan, dan arsitektur.
- Pipeline: rantai pemrosesan: parse → normalize → resolve → build NEIR → validate → build graph → policy evaluation → schedule → generate → review → write artifacts.
- Kernel: komponen runtime inti yang mengelola service registry, event bus, telemetry, dan lifecycle.
- Constitution: dokumen normatif tertinggi yang memuat aturan dasar.
- Governance: kerangka tata kelola, proses, dan tanggung jawab organisasi.
- Specification: dokumen yang menjelaskan kebutuhan, desain, dan batasan sistem.
- Policy: aturan yang dapat dieksekusi untuk mengarahkan perilaku sistem atau AI agent.
- Traceability: kemampuan menelusuri hubungan antar requirement, specification, implementation, dan test.
- Profile: definisi khusus yang menyesuaikan standar untuk konteks tertentu.
- Compiler: mesin yang mengubah policy dan spesifikasi menjadi artefak yang dapat dipakai.
- Validator: komponen yang memeriksa apakah artefak memenuhi aturan yang berlaku.

## Istilah teknis
- Artifact: dokumen atau output yang dihasilkan dari proses engineering.
- Adapter: generator output untuk target tertentu (Copilot, Claude, Cursor, Gemini, Codex, OpenCode).
- Context Bundle: ringkasan proyek dalam format markdown atau teks, dioptimalkan untuk konsumsi LLM.
- Module: unit kode dalam proyek, didefinisikan oleh nama, path, dan dependensi.
- Service: komponen runtime (http, grpc, worker, cli, job) dengan endpoint dan konfigurasi.
- Endpoint: titik masuk API yang didefinisikan oleh method, path, dan action.
- DAG: Directed Acyclic Graph — struktur data untuk resolusi dependensi dan penjadwalan tugas.
- Artifact Store: penyimpanan persisten untuk artefak pipeline dengan deduplikasi content-hash.
- Migration: proses upgrade spesifikasi dari satu schema version ke yang lain.
- Schema Version: versi SemVer dari format spesifikasi (minimum: 0.1.0).
- LSP: Language Server Protocol — menyediakan fitur IDE seperti autocomplete dan diagnostics untuk file spesifikasi.
- MCP: Model Context Protocol — memungkinkan integrasi AI agent dengan runtime NAEOS.
- ADR: Architecture Decision Record, catatan keputusan arsitektur.
- RFC: Request for Comments, mekanisme usulan perubahan.
- Compliance: keadaan di mana implementasi memenuhi standar dan aturan yang berlaku.
