Document ID: NAEOS-SPEC-007

Title: Validation Model

Short Name: Validation Engine

Version: 1.0.0

Status: Stable

Category: Core Specification

Normative: true

Priority: Critical

Owner: NAEOS OSS

Depends On:

- SPEC-001
- SPEC-002
- SPEC-003
- SPEC-004
- SPEC-005
- SPEC-006

Referenced By:

- Compiler
- CLI
- Validator
- Studio
- SDK
- CI/CD
Validation Model
Executive Summary

Validation Model mendefinisikan bagaimana seluruh Artifact NAEOS diperiksa sebelum digunakan oleh Compiler, AI Agent, maupun pengguna.

Validasi dilakukan secara bertingkat sehingga kesalahan dapat ditemukan sedini mungkin.

Validation Engine menjadi gatekeeper seluruh ekosistem.

1. Purpose

Validation Model bertujuan untuk:

memastikan integritas artefak,
mencegah kesalahan sejak awal,
menjaga konsistensi knowledge,
meningkatkan kualitas engineering,
menyediakan feedback otomatis.
2. Validation Philosophy

NAEOS menggunakan prinsip:

Validate Early

↓

Validate Continuously

↓

Compile Safely

↓

Deploy Confidently

Validasi dilakukan pada setiap perubahan, bukan hanya sebelum rilis.

3. Validation Pipeline

```mermaid
flowchart TD
    Start([Artifact Submitted]) --> S1[Stage 1: Syntax Validation]
    S1 -->|Pass| S2[Stage 2: Metadata Validation]
    S1 -->|Fail| F1[Report Syntax Errors]
    S2 -->|Pass| S3[Stage 3: Schema Validation]
    S2 -->|Fail| F2[Report Metadata Errors]
    S3 -->|Pass| S4[Stage 4: Dependency Validation]
    S3 -->|Fail| F3[Report Schema Errors]
    S4 -->|Pass| S5[Stage 5: Rule Evaluation]
    S4 -->|Fail| F4[Report Dependency Errors]
    S5 -->|Pass| S6[Stage 6: Knowledge Validation]
    S5 -->|Fail| F5[Report Rule Violations]
    S6 -->|Pass| S7[Stage 7: Compliance Validation]
    S6 -->|Fail| F6[Report Knowledge Errors]
    S7 -->|Pass| S8[Stage 8: Quality Assessment]
    S7 -->|Fail| F7[Report Compliance Issues]
    S8 -->|Pass| Done([Validation Passed])
    S8 -->|Warning| Warn([Validation Passed with Warnings])
```

Setiap tahap hanya dijalankan jika tahap sebelumnya berhasil.

4. Validation Stages
Stage 1 — Syntax Validation

Memastikan dokumen memiliki format yang benar.

Contoh pemeriksaan:

YAML valid
Markdown valid
JSON valid
UTF-8
struktur file
Stage 2 — Metadata Validation

Memastikan seluruh metadata wajib tersedia.

Contoh:

ID
Version
Owner
Status
Category
Stage 3 — Schema Validation

Membandingkan Artifact dengan JSON Schema resmi.

Stage 4 — Dependency Validation

Memastikan:

dependency tersedia,
versi kompatibel,
tidak ada cycle,
referensi valid.
Stage 5 — Rule Evaluation

Menjalankan seluruh Rule yang berlaku.

Rule berasal dari:

Constitution
Standards
Project Rules
Organization Rules
Stage 6 — Knowledge Validation

Memastikan:

node graph valid,
relationship valid,
tidak ada orphan node,
traceability lengkap.
Stage 7 — Compliance Validation

Mengukur kepatuhan terhadap:

Core Principles
Constitution
Standards
Organization Policy
Stage 8 — Quality Assessment

Menghasilkan skor kualitas keseluruhan.

5. Validation Result Model
validation:

  artifact:

  status:

  score:

  duration:

  warnings:

  errors:

  critical:

  recommendations:
6. Validation Status

Status resmi:

Status	Arti
Passed	Lulus
Warning	Ada peringatan
Failed	Gagal
Blocked	Tidak dapat diproses
7. Severity Levels
Level	Dampak
Info	Informasi
Warning	Perlu perhatian
Error	Harus diperbaiki
Critical	Kompilasi dihentikan
8. Quality Score

Validation Engine menghasilkan skor 0–100.

Contoh:

Score	Grade
95–100	A
85–94	B
70–84	C
<70	Failed

Quality Score dapat digunakan sebagai syarat merge di CI/CD.

9. Validation Report

Contoh:

artifact:

  id: SPEC-004

status:

  Failed

issues:

- severity: Error

  message: Missing owner

- severity: Warning

  message: Missing review cycle

score:

 82
10. Continuous Validation

Validation dapat dijalankan:

saat editor menyimpan file,
saat commit Git,
saat Pull Request,
saat build,
saat release,
secara terjadwal.
11. AI Assisted Validation

AI dapat membantu:

menjelaskan error,
memberikan saran perbaikan,
membuat patch,
menghasilkan dokumentasi tambahan.

Namun keputusan akhir tetap berasal dari Validation Engine, bukan AI.

12. Compliance Score

Selain Quality Score, dihasilkan Compliance Score.

Contoh:

Area	Score
Metadata	100
Security	96
Standards	91
Documentation	88
Traceability	99
13. Performance Requirements

Validator harus:

mendukung validasi paralel,
mendukung incremental validation,
menggunakan cache bila memungkinkan,
memberikan hasil deterministik.
14. CLI Integration

Contoh penggunaan:

naeos validate

naeos validate docs/

naeos validate project.yaml

naeos validate --strict

naeos validate --json

naeos validate --graph
15. CI/CD Integration

Validation menjadi Quality Gate.

Commit

↓

Validate

↓

Quality Gate

↓

Build

↓

Deploy

Pipeline tidak boleh melanjutkan build jika terdapat pelanggaran Critical.

16. Studio Integration

NAEOS Studio harus menampilkan:

daftar error,
warning,
dependency yang bermasalah,
quality score,
compliance score,
rekomendasi AI.
17. Conformance

Implementasi Validation Engine MUST:

mendukung seluruh tahap validasi,
menghasilkan laporan standar,
mendukung Rule Model,
mendukung Dependency Graph,
kompatibel dengan Engineering Knowledge Graph.
18. Related Documents
ID	Document
NAEOS-SPEC-004	Metadata Specification
NAEOS-SPEC-005	Rule Model
NAEOS-SPEC-006	Dependency Graph
NAEOS-SPEC-008	Compiler Model
Revision History
Version	Date	Change
1.0.0	2026-07-09	Initial Validation Model
Status
NAEOS-SPEC-007

APPROVED

Validation Engine Established
