Document ID: NAEOS-CON-006

Title: Testing Constitution

Short Name: NTC

Version: 1.0.0

Status: Stable

Category: Constitution

Normative: true

Priority: CRITICAL

Owner: NAEOS Foundation

Motto:
"Every Claim Requires Evidence."

Depends On:

- NAEOS-CON-001
- NAEOS-CON-003
- NAEOS-SPEC-006
- NAEOS-SPEC-007

Referenced By:

- Test Engine
- Validation Engine
- Compiler
- AI Runtime
- CI/CD
Testing Constitution
Executive Summary

Testing Constitution menetapkan bahwa seluruh aktivitas pengujian dalam NAEOS bertujuan menghasilkan Engineering Evidence.

Evidence digunakan untuk membuktikan bahwa Requirement, Specification, Architecture, Implementation, dan Deployment memenuhi aturan yang berlaku.

Testing bukan tahap terakhir, tetapi bagian dari siklus Engineering Knowledge.

Article I — Evidence First

Setiap klaim engineering MUST memiliki bukti yang dapat diverifikasi.

Contoh klaim:

Requirement telah dipenuhi.
API sesuai kontrak.
Sistem aman.
Performa memenuhi target.
Deployment berhasil.

Tanpa evidence, klaim dianggap belum terbukti.

Article II — Traceable Verification

Setiap aktivitas pengujian harus dapat ditelusuri.

Requirement
↓

Specification
↓

Architecture

↓

Implementation

↓

Test Case

↓

Evidence

Compiler harus dapat membangun hubungan ini secara otomatis.

Article III — Continuous Verification

Verifikasi harus dilakukan secara berkelanjutan.

Minimal pada:

Commit
Pull Request
Build
Release
Deployment
Scheduled Validation
Article IV — Multi-Level Testing

Setiap proyek harus memiliki strategi pengujian berlapis sesuai kebutuhan.

Jenis pengujian dapat mencakup:

Unit
Integration
Contract
End-to-End
Performance
Security
Resilience
AI Evaluation
Infrastructure Validation

Validator dapat memeriksa keberadaan strategi yang sesuai dengan profil proyek.

Article V — Contract Verification

Semua kontrak harus diverifikasi.

Contoh:

API Contract
Event Contract
Database Schema
Prompt Contract
Tool Interface

Perubahan kontrak harus memicu pengujian terkait.

Article VI — Deterministic Testing

Pengujian SHOULD menghasilkan hasil yang konsisten.

Jika pengujian bersifat nondeterministik (misalnya melibatkan model AI), artefak pengujian harus mendokumentasikan:

parameter,
model,
konfigurasi,
toleransi hasil.
Article VII — Automation First

Seluruh pengujian yang dapat diotomatisasi SHOULD diotomatisasi.

Evidence harus dapat dihasilkan oleh:

CI/CD,
Validation Engine,
AI Review Engine,
Runtime Monitoring.
Article VIII — Quality Gates

Pipeline harus memiliki Quality Gate.

Contoh:

Commit
↓

Validate
↓

Test
↓

Evidence
↓

Quality Gate
↓

Release

Release tidak boleh dilakukan jika Quality Gate gagal.

Article IX — Observability as Evidence

Log, metric, trace, dan health check merupakan bagian dari evidence operasional.

Evidence runtime harus dapat dikaitkan dengan Deployment dan Version yang relevan.

Article X — AI Evaluation

Komponen AI harus memiliki mekanisme evaluasi.

Evaluasi dapat mencakup:

kualitas jawaban,
kepatuhan terhadap Rule Model,
keamanan,
konsistensi,
penggunaan konteks.

Hasil evaluasi menjadi bagian dari Engineering Knowledge Graph.

Article XI — Evidence Preservation

Evidence harus:

memiliki metadata,
dapat ditelusuri,
memiliki masa retensi sesuai kebijakan,
dapat digunakan untuk audit.

Evidence diperlakukan sebagai Artifact dalam Universal Artifact Model.

Article XII — Continuous Improvement

Temuan dari pengujian harus digunakan untuk memperbaiki:

Requirement,
Specification,
Architecture,
Standards,
Playbooks,
Rule Model.

Testing bukan hanya mendeteksi masalah, tetapi juga memperkaya Engineering Knowledge.

Constitutional Compliance

Suatu proyek dinyatakan Testing Compliant apabila:

seluruh Requirement memiliki mekanisme verifikasi,
evidence tersedia untuk artefak yang diwajibkan,
Quality Gate terpenuhi,
hasil pengujian dapat ditelusuri,
lolos Validation Engine.
Enforcement

Validation Engine dan Compiler harus mampu:

menghubungkan Requirement dengan Evidence,
memverifikasi keberadaan Quality Gate,
menghasilkan Traceability Matrix,
menghitung Quality Score berdasarkan evidence yang tersedia.
Related Documents
ID	Document
NAEOS-CON-001	Engineering Constitution
NAEOS-CON-003	Architecture Constitution
NAEOS-SPEC-006	Dependency Graph
NAEOS-SPEC-007	Validation Model
NAEOS-SPEC-008	Compiler Model
Revision History
Version	Date	Change
1.0.0	2026-07-09	Initial Testing Constitution
Status
NAEOS-CON-006

APPROVED

Testing Constitution Established
