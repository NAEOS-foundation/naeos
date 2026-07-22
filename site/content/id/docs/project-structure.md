---
title: Struktur Proyek
description: Layout repositori dan organisasi direktori untuk kode sumber NAEOS.
weight: 24
---

Dokumen ini mendeskripsikan struktur repositori NAEOS dan bagaimana kode sumber diorganisir.

## Root Level

```text
naeos/
├── cmd/naeos/           # CLI commands (35+ file)
├── internal/            # Package internal (60+ file)
├── pkg/                 # Package publik (4 package)
├── docs/                # Spesifikasi NES (56 file)
├── site/                # Website Hugo
├── governance/          # Dokumen governance (8 file)
├── constitution/        # Konstitusi engineering
├── specification/       # Dokumen spesifikasi (10 file)
├── kernel/              # Spesifikasi kernel (4 file)
├── policy/              # Dokumen kebijakan (7 file)
├── prompts/             # Template prompt AI
├── templates/           # Template ADR/RFC
├── examples/            # Dokumen contoh
├── Reference Architecture/ # Dokumen arsitektur referensi
├── go.mod               # Definisi module Go
├── go.sum               # Checksum module Go
├── Makefile             # Otomasi build
├── README.md            # Readme proyek
├── CHANGELOG.md         # Riwayat versi
├── CONTRIBUTING.md      # Panduan kontribusi
├── LICENSE              # Lisensi Apache 2.0
├── .github/             # Workflow dan konfigurasi GitHub
├── .gitignore           # Aturan ignore Git
├── .golangci.yml        # Konfigurasi linter
├── .goreleaser.yaml     # Konfigurasi rilis
└── Dockerfile           # Build Docker multi-stage
```

## cmd/naeos/

CLI commands diorganisir berdasarkan area fitur:

```text
cmd/naeos/
├── main.go              # Root command dan entry point
├── run_cmd.go           # Eksekusi pipeline
├── validate_cmd.go      # Validasi spesifikasi
├── lint_cmd.go          # Pengecekan gaya spesifikasi
├── create_cmd.go        # Pembuatan proyek
├── scaffold_cmd.go      # Scaffolding kode
├── export_cmd.go        # Ekspor artifact
├── compile_cmd.go       # Kompilasi AI
├── context_cmd.go       # Context bundles
├── profile_cmd.go       # Profil industri
├── artifacts_cmd.go     # Manajemen artifact
├── migrate_cmd.go       # Migrasi skema
├── doctor_cmd.go        # Pemeriksaan kesehatan sistem
├── diff_cmd.go          # Perbandingan spesifikasi
├── watch_cmd.go         # Pemantauan file
├── workspace_cmd.go     # Manajemen workspace
├── kernel_cmd.go        # Inspeksi kernel
├── plugin_cmd.go        # Manajemen plugin
├── template_cmd.go      # Manajemen template
├── lock_cmd.go          # Penguncian dependensi
├── rollback_cmd.go      # Rollback perubahan
├── audit_cmd.go         # Audit spesifikasi
├── marketplace_cmd.go   # Marketplace profil
├── status_cmd.go        # Status pipeline
├── docs_cmd.go          # Generasi dokumentasi
├── ai_cmd.go            # Bantuan AI
├── init_cmd.go          # Inisialisasi konfigurasi
├── version_cmd.go       # Info versi
├── completion_cmd.go    # Penyelesaian shell
├── repair_cmd.go        # Perbaikan spesifikasi
├── preview_cmd.go       # Mode preview
├── test_cmd.go          # Test runner
├── mcp_cmd.go           # Server MCP
├── helpers.go           # Helper bersama
├── e2e_test.go          # Test end-to-end
└── main_test.go         # Unit test
```

## internal/

```text
internal/
├── specification/       # Pemrosesan spesifikasi
│   ├── parser/          # Parser YAML/JSON dengan interpolasi variabel
│   ├── normalizer/      # Normalisasi data
│   └── resolver/        # Resolusi cross-reference
├── neir/                # Model NEIR
│   ├── model/           # Definisi model (Project, Module, Service, dll)
│   ├── builder/         # Builder NEIR dari spesifikasi yang di-parse
│   └── validator/       # Validator model NEIR
├── compiler/            # Kompiler instruksi AI
│   └── adapters/        # 6 adapter output (Copilot, Claude, Cursor, dll)
├── context/             # Context bundles
│   └── bundle/          # Generator bundle (markdown, plain text, JSON)
├── generation/          # Generasi kode
│   ├── engine/          # Mesin generasi
│   ├── adapters/        # Adapter bahasa (Go, TS, Python, Java, Rust)
│   └── renderers/       # Renderer template
├── governance/          # Governance
│   ├── policy/          # Evaluator kebijakan
│   └── review/          # Review artifact
├── artifacts/           # Artifact store dengan deduplikasi content-hash
├── profiles/            # Profil industri (5 built-in)
├── migration/           # Mesin migrasi skema
├── security/            # Aturan keamanan dan pemindaian
├── marketplace/         # Marketplace profil & plugin
├── pluginsdk/           # Plugin SDK dengan runtime WASM
├── ai/                  # Layanan AI dan integrasi LLM
├── mcp/                 # Implementasi server MCP
├── knowledge/           # Knowledge graph
├── database/            # Layer database (PostgreSQL, MySQL, SQLite)
├── websocket/           # Komunikasi real-time WebSocket
├── eventsourcing/       # Event sourcing dan snapshot aggregate
├── distributed/         # Eksekusi tugas terdistribusi
├── configreload/        # Hot-reload konfigurasi
├── pipelinecache/       # Cache hasil pipeline
├── pipelinemiddleware/  # Middleware pipeline yang dapat dikomposisi
├── audit/               # Layer logging audit
├── hcl/                 # Parser konfigurasi HCL
├── profiledetect/       # Deteksi otomatis bahasa/framework
├── testrunner/          # Test runner multi-bahasa
├── docgen/              # Generator dokumentasi
├── diff/                # Mesin diff dengan output berwarna
├── watch/               # File watcher untuk hot-reload
├── lock/                # Penguncian dependensi
├── rollback/            # Manajemen rollback
├── workspace/           # Manajemen workspace
├── templates/           # Mesin template
├── planner/             # Penjadwalan tugas berbasis DAG
├── runtime/             # Mesin runtime
├── profiling/           # Profiling performa
├── registry/            # Service registry
├── lint/                # Aturan lint
├── create/              # Pembuatan proyek
└── shared/              # Utilitas bersama
    ├── log/             # Logging terstruktur (slog)
    ├── strutil/         # Utilitas string
    └── contracts/       # Kontrak bersama
```

## pkg/

Package publik yang dapat di-import oleh konsumen eksternal:

```text
pkg/
├── pipeline/            # Orkestrasi pipeline utama
├── kernel/              # Kernel sistem (registry, event bus, telemetry)
├── config/              # Manajemen konfigurasi
└── plugin/              # Interface sistem plugin
```

## Jumlah File

| Direktori | File | Deskripsi |
|-----------|------|-----------|
| `cmd/naeos/` | 35 | CLI commands dan entry point |
| `internal/` | 60+ | Package internal |
| `pkg/` | 4 | Package API publik |
| `docs/` | 56 | Spesifikasi NES |
| `site/` | 100+ | Konten dan layout website Hugo |
| `governance/` | 8 | Dokumen governance |
| `specification/` | 10 | Dokumen spesifikasi |
| **Total** | **270+** | |

Lihat juga: [Arsitektur](/id/docs/architecture/), [Pipeline Engine](/id/docs/pipeline-engine/)
