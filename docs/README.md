# NAEOS Documentation Structure

Dokumen ini memetakan struktur dokumentasi NAEOS berdasarkan skema modular yang dirancang untuk mendukung navigasi, traceability, dan evolusi artefak teknis.

## Dokumentasi inti
- NES-000 Foundation — prinsip fondasi dan asumsi arsitektural.
- NES-001 Repository — struktur repositori dan tata letak artefak.
- NES-002 Kernel — lapisan runtime inti dan service primitive.
- NES-002-Kernel-API — referensi API kernel Go (pkg/kernel).
- NES-003 Workspace — konteks eksekusi proyek dan state lokal.
- NES-004 Bootstrap — inisialisasi proyek dan pembentukan workspace.
- NES-005 Blueprint — model desain tingkat menengah.
- NES-006 Template — kerangka reusable untuk artefak.
- NES-007 Generator — transformasi desain menjadi artefak implementasi.
- NES-008 Registry — katalog metadata dan dependency.
- NES-009 Plugin — mekanisme ekstensi modular.
- NES-010 Knowledge — knowledge model dan provenance.
- NES-011 Graph — representasi relasional dependency dan alur eksekusi.
- NES-012 Policy — model aturan dan governance.
- NES-013 Compiler — pipeline transformasi model menjadi output.
- NES-014 Validator — pemeriksaan kualitas dan konsistensi.
- NES-015 Runtime — pelaksanaan artefak pada lingkungan target.
- NES-016 AI — peran AI assistant dan automation support.
- NES-017 Studio — antarmuka pengguna dan experience layer.
- NES-018 Cloud — target deployment dan operasi cloud.
- NES-019 SDK — integrasi programmatik dan plugin development.
- NES-020 Security — kontrol akses, audit, dan prinsip keamanan.
- NES-021 Testing — validation quality gate dan regression coverage.
- NES-022 Release — proses publikasi dan rollout.
- NES-023 NEIR — model engineering sentral, pipeline NEIR, dan versioning.
- NES-023-NEIR-Model — referensi model NEIR Go (internal/neir/model).
- NES-024 Internal Structure — draft struktur folder internal untuk implementasi teknis.
- NES-025 Implementation Skeletons — draft file-level skeleton untuk modul internal utama.
- NES-026 Pipeline — dokumentasi pipeline (pkg/pipeline).
- NES-027 Governance — dokumentasi governance policy dan review.
- NES-028 CLI Reference — referensi perintah CLI (cmd/naeos).
- NES-029 Configuration — referensi format konfigurasi pipeline.
- NES-030 Specification Language — bahasa spesifikasi NAEOS.
- NES-031 Errors — katalog kode error dan penanganannya.
- NES-032 Telemetry — referensi telemetry, metrik, dan observabilitas.
- NES-033 Testing Guide — panduan pengujian dan coverage.
- NES-034 Event Bus — event bus internal (pub/sub) untuk komunikasi antar komponen.
- NES-035 Version Management — manajemen versi SemVer untuk NEIR.
- NES-036 Template Renderer — engine rendering template berbasis text/template.
- NES-037 Knowledge Graph & Provenance — knowledge graph dan provenance tracking.
- NES-038 Shared Types & Contracts — tipe data dan kontrak bersama untuk seluruh komponen.
- NES-039 SDK Multi-Language — spesifikasi SDK multi-language (Go, TypeScript, Python, Java, Rust).
- NES-040 Output Adapter Architecture — arsitektur output adapter untuk ekstensi bahasa.
- NES-042 Database — database layer (PostgreSQL, MySQL, SQLite).
- NES-043 WebSocket — WebSocket real-time communication.
- NES-044 EventSourcing — event sourcing dan aggregate snapshots.
- NES-045 Distributed — distributed task execution.
- NES-046 Config Hot-Reload — configuration hot-reload.
- NES-047 Pipeline Cache — pipeline result caching.
- NES-048 Pipeline Middleware — composable pipeline middleware.
- NES-049 Audit Logging — audit logging layer.
- NES-050 HCL Parser — HCL configuration parser.
- NES-051 Profile Detection — automatic language/framework detection.
- NES-052 CI/CD — CI/CD pipeline automation.
- NES-053 WASM Plugin — WASM plugin sandboxed execution.

## Dokumentasi pendukung
- Kernel Architecture (NAEOS-KER-001) — arsitektur kernel.
- Kernel Implementation (NAEOS-KER-002) — panduan implementasi kernel.
- Kernel Examples (NAEOS-KER-003) — contoh penggunaan kernel.
- Kernel Best Practices (NAEOS-KER-004) — praktik terbaik kernel.

## Architecture Decision Records (ADRs)
- [adr/001-why-go-for-runtime.md](adr/001-why-go-for-runtime.md) — ADR-001: Why Go for the Runtime
- [adr/002-why-neir-as-central-model.md](adr/002-why-neir-as-central-model.md) — ADR-002: Why NEIR as the Central Model
- [adr/003-why-mcp-for-ai-integration.md](adr/003-why-mcp-for-ai-integration.md) — ADR-003: Why MCP for AI Integration
- [adr/004-database-layer.md](adr/004-database-layer.md) — ADR-004: Database Layer (PostgreSQL, MySQL, SQLite)
- [adr/005-websocket-communication.md](adr/005-websocket-communication.md) — ADR-005: WebSocket Communication
- [adr/006-distributed-task-execution.md](adr/006-distributed-task-execution.md) — ADR-006: Distributed Task Execution
- [adr/007-prompt-library.md](adr/007-prompt-library.md) — ADR-007: Prompt Library
- [adr/008-wasm-plugin-sandbox.md](adr/008-wasm-plugin-sandbox.md) — ADR-008: WASM Plugin Sandbox

## Document placement recommendations
Each area can have supporting documents in the following format:
- Overview
- Concepts
- Architecture
- Workflow
- Configuration
- Examples
- Troubleshooting
- Validation
