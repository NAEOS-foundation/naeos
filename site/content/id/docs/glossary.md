---
title: Glosarium
description: Istilah-istilah penting dan definisinya yang digunakan di seluruh platform NAEOS.
weight: 22
---

| Istilah | Definisi |
|---------|----------|
| **NAEOS** | Nusantara Engineering & Architecture Operating System — platform engineering deklaratif open-source |
| **NEIR** | NAEOS Engineering Intermediate Representation — model perantara terpadu yang merepresentasikan seluruh proyek |
| **Spec** | Spesifikasi — dokumen YAML atau JSON yang mendefinisikan proyek, modul, layanan, dan arsitektur |
| **Pipeline** | Rantai pemrosesan: parse → normalize → resolve → build → validate → schedule → generate → compile → export |
| **Kernel** | Komponen inti runtime yang mengelola service registry, event bus, telemetry, dan lifecycle |
| **Artifact** | Output apa pun yang dihasilkan oleh pipeline: kode, konfigurasi, dokumentasi, atau konteks AI |
| **Profile** | Konfigurasi preset spesifik industri (SaaS, AI Agent, FinTech, Healthcare, Government) |
| **Adapter** | Generator output untuk target tertentu (Copilot, Claude, Cursor, Gemini, Codex, OpenCode) |
| **Compiler** | Komponen yang mentransformasikan NEIR ke instruction sets untuk asisten coding |
| **Context Bundle** | Ringkasan proyek dalam format markdown atau plain text, dioptimalkan untuk konsumsi LLM |
| **Governance** | Sistem kebijakan, aturan validasi, dan alur kerja review yang menerapkan standar |
| **Policy** | Aturan yang dievaluasi selama pipeline berjalan (operator: exists, not_empty, contains, gt, lt, in) |
| **Schema Version** | Versi SemVer dari format spesifikasi (minimum: 0.1.0) |
| **Module** | Unit kode dalam proyek, didefinisikan oleh nama, path, dan dependensi |
| **Service** | Komponen runtime (http, grpc, worker, cli, job) dengan endpoint dan konfigurasi |
| **Endpoint** | Titik masuk API yang didefinisikan oleh method, path, dan action |
| **DAG** | Directed Acyclic Graph — struktur data yang digunakan untuk resolusi dependensi dan penjadwalan tugas |
| **Artifact Store** | Penyimpanan persisten untuk artifact pipeline dengan deduplikasi content-hash |
| **Migration** | Proses meng-upgrade spesifikasi dari satu versi skema ke versi lainnya |
| **LSP** | Language Server Protocol — menyediakan fitur IDE seperti autocomplete dan diagnosa untuk file spesifikasi |
| **MCP** | Model Context Protocol — memungkinkan integrasi agen AI dengan runtime NAEOS |
