📄 NAEOS-GOV-008
Document ID: NAEOS-GOV-008
Title: Versioning Policy
Version: 1.0.0
Status: Stable
Category: Governance
Steward: NAEOS OSS
Priority: Critical

Motto:
  Architecture Drives Engineering.

Depends On:
  - NAEOS-GOV-001 Project Charter
  - NAEOS-GOV-006 Governance Model
  - NAEOS-GOV-007 Roadmap

Referenced By:
  - All NAEOS Documents
  - Compiler
  - CLI
  - Validator
NAEOS Versioning Policy
Executive Summary

Versioning Policy defines bagaimana all artefak NAEOS berkembang in konsisten and can diprediksi.

Kebijakan ini berlaku to:

Specification
Constitution
Standards
Playbooks
Templates
JSON Schema
Compiler
CLI
SDK
Website
Reference Platform
1. Purpose

This document berobjective to:

maintain kompatibilitas,
mengendalikan changes,
mempermudah migrasi,
ensure stabilitas ekosistem.
2. Versioning Model

NAEOS menggunakan Semantic Versioning (SemVer 2.0.0).

Format:

MAJOR.MINOR.PATCH

Contoh:

1.0.0
1.2.0
1.2.5
2.0.0
3. MAJOR Version

MAJOR berubah apabila tercan:

breaking changes,
changes struktur specification,
changes metadata wajib,
changes compiler that not kompatibel.

Contoh:

1.x.x

↓

2.0.0
4. MINOR Version

MINOR bertambah to:

fitur baru,
penambahan standar,
extension,
backward compatible improvements.

Contoh:

1.2.0

↓

1.3.0
5. PATCH Version

PATCH used to:

typo,
perbaikan dokumentasi,
bug compiler,
bug validator,
koreksi kecil.

Contoh:

1.3.2

↓

1.3.3
6. Document Lifecycle

Setiap dokumen have status following:

Status	Deskripsi
Draft	Seandg dikembangkan
Review	Menunggu evaluasi
Proposed	Diusulkan
Accepted	Disetujui
Stable	Siap used
Deprecated	Not direkomendasikan
Archived	Not maintained
7. Release Channels

NAEOS have empat jalur rilis.

Alpha

Eksperimental.

Belum stabil.

Beta

Fitur lengkap.

Masih can berubah.

Release Candidate (RC)

Hampir final.

Hanya menerima bug fix.

Stable

Direkomendasikan to produksi.

8. Compatibility Rules

Semua komponen NAEOS:

MUST:

mendeklarasikan versi,
menyatakan kompatibilitas,
follows SemVer.

Compiler:

MUST mampu menolak specification that not kompatibel.

9. Deprecation Policy

Fitur that akan dihapus:

Ditandai as Deprecated.
Tetap didukung minimal satu rilis MAJOR.
Have panduan migrasi.
Baru dihapus on rilis MAJOR followingnya.
10. Migration Policy

Setiap breaking change wajib provide:

Migration Guide
Compatibility Notes
Change Log
Contoh implementasi baru
11. Release Cadence
Release	Target
Patch	Sesuai kebutuhan
Minor	Setiap 3 bulan
Major	Setiap 12–18 bulan
12. Supported Versions

Kebijakan dukungan:

MAJOR terbaru: Full Support
MAJOR sebelumnya: Security & Critical Fixes
Versi more lama: Community Support
13. Version Metadata

Setiap dokumen wajib have metadata following:

version: 1.0.0
status: Stable
last_updated: 2026-07-09
owner: NAEOS OSS
review_cycle: 12 months
14. Release Artifacts

Setiap rilis NAEOS must produce:

Release Notes
Change Log
Migration Guide
Updated Specification
Updated JSON Schema
Compatibility Matrix
15. Version Compatibility Matrix
Component	Policy
Specification	SemVer
Constitution	SemVer
Standards	SemVer
Compiler	SemVer
CLI	SemVer
SDK	SemVer
Website	Rolling Release
16. Versioning Principles

NAEOS follows prinsip:

Predictable Releases
Backward Compatibility
Explicit Breaking Changes
Transparent Migration
Long-Term Stability
17. Conformance Requirements

Implementasi NAEOS:

MUST:

menggunakan versi resmi,
follows kebijakan kompatibilitas,
provide metadata versi.

SHOULD:

provide changelog,
provide migration guide.
18. Related Documents
ID	Document
NAEOS-GOV-001	Project Charter
NAEOS-GOV-006	Governance Model
NAEOS-GOV-007	Roadmap
NAEOS-SPEC-001	Overview
Revision History
Version	Date	Change
1.0.0	2026-07-09	Initial Versioning Policy
Status
NAEOS-GOV-008

APPROVED

Governance Foundation Complete
