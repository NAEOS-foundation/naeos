📄 NAEOS-GOV-008
Document ID: NAEOS-GOV-008
Title: Versioning Policy
Version: 1.0.0
Status: Stable
Category: Governance
Owner: NAEOS Foundation
Priority: Critical

Motto:
  Specify Once. Build Anywhere.

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

Versioning Policy mendefinisikan bagaimana seluruh artefak NAEOS berkembang secara konsisten dan dapat diprediksi.

Kebijakan ini berlaku untuk:

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

Dokumen ini bertujuan untuk:

menjaga kompatibilitas,
mengendalikan perubahan,
mempermudah migrasi,
memastikan stabilitas ekosistem.
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

MAJOR berubah apabila terdapat:

breaking changes,
perubahan struktur specification,
perubahan metadata wajib,
perubahan compiler yang tidak kompatibel.

Contoh:

1.x.x

↓

2.0.0
4. MINOR Version

MINOR bertambah untuk:

fitur baru,
penambahan standar,
extension,
backward compatible improvements.

Contoh:

1.2.0

↓

1.3.0
5. PATCH Version

PATCH digunakan untuk:

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

Setiap dokumen memiliki status berikut:

Status	Deskripsi
Draft	Sedang dikembangkan
Review	Menunggu evaluasi
Proposed	Diusulkan
Accepted	Disetujui
Stable	Siap digunakan
Deprecated	Tidak direkomendasikan
Archived	Tidak dipelihara
7. Release Channels

NAEOS memiliki empat jalur rilis.

Alpha

Eksperimental.

Belum stabil.

Beta

Fitur lengkap.

Masih dapat berubah.

Release Candidate (RC)

Hampir final.

Hanya menerima bug fix.

Stable

Direkomendasikan untuk produksi.

8. Compatibility Rules

Semua komponen NAEOS:

MUST:

mendeklarasikan versi,
menyatakan kompatibilitas,
mengikuti SemVer.

Compiler:

MUST mampu menolak specification yang tidak kompatibel.

9. Deprecation Policy

Fitur yang akan dihapus:

Ditandai sebagai Deprecated.
Tetap didukung minimal satu rilis MAJOR.
Memiliki panduan migrasi.
Baru dihapus pada rilis MAJOR berikutnya.
10. Migration Policy

Setiap breaking change wajib menyediakan:

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
Versi lebih lama: Community Support
13. Version Metadata

Setiap dokumen wajib memiliki metadata berikut:

version: 1.0.0
status: Stable
last_updated: 2026-07-09
owner: NAEOS Foundation
review_cycle: 12 months
14. Release Artifacts

Setiap rilis NAEOS harus menghasilkan:

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

NAEOS mengikuti prinsip:

Predictable Releases
Backward Compatibility
Explicit Breaking Changes
Transparent Migration
Long-Term Stability
17. Conformance Requirements

Implementasi NAEOS:

MUST:

menggunakan versi resmi,
mengikuti kebijakan kompatibilitas,
menyediakan metadata versi.

SHOULD:

menyediakan changelog,
menyediakan migration guide.
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
