Document ID: NAEOS-SPEC-005

Title: Rule Model

Short Name: NRL (NAEOS Rule Language)

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

Referenced By:

- Validator
- Compiler
- CLI
- AI Runtime
- Compliance Engine
NAEOS Rule Model (NRL)
Executive Summary

NAEOS Rule Model mendefinisikan bagaimana aturan engineering ditulis, diwariskan, divalidasi, dan dievaluasi secara otomatis.

Setiap standar, constitution, policy, dan playbook pada akhirnya diterjemahkan menjadi Rule.

Rule bukan sekadar teks dokumentasi, tetapi objek formal yang dapat diproses oleh mesin.

1. Purpose

Rule Model bertujuan untuk:

mendefinisikan aturan secara konsisten,
memungkinkan validasi otomatis,
mendukung AI reasoning,
menyediakan dasar bagi compliance engine.
2. Rule Architecture

```
┌─────────────────────────────────────────────────────────┐
│                  Rule Architecture                      │
│                                                         │
│  ┌──────────┐                                          │
│  │ Artifact  │                                         │
│  └─────┬────┘                                          │
│        │                                                │
│  ┌─────▼────┐                                          │
│  │ Metadata  │                                         │
│  └─────┬────┘                                          │
│        │                                                │
│  ┌─────▼────┐                                          │
│  │  Rules   │                                          │
│  └─────┬────┘                                          │
│        │                                                │
│  ┌─────▼──────────┐  ┌──────────────┐                 │
│  │   Validator     │  │ Policy Engine │                 │
│  └─────┬──────────┘  └──────┬───────┘                 │
│        │                    │                          │
│        └────────┬───────────┘                          │
│                 ▼                                       │
│  ┌──────────────────────────┐                          │
│  │   Compiler / Adapter     │                          │
│  │   (generates artifacts)  │                          │
│  └──────────────┬───────────┘                          │
│                 ▼                                       │
│  ┌──────────────────────────┐                          │
│  │   AI Review              │                          │
│  └──────────────┬───────────┘                          │
│                 ▼                                       │
│  ┌──────────────────────────┐                          │
│  │   Compliance             │                          │
│  └──────────────────────────┘                          │
└─────────────────────────────────────────────────────────┘
```

Rule menjadi lapisan logika yang menghubungkan spesifikasi dengan implementasi.

```mermaid
graph TB
    A[Artifact] --> B[Metadata]
    B --> C[Rules]
    C --> D[Validator]
    C --> E[Policy Engine]
    D --> F[Compiler / Adapter]
    E --> F
    F --> G[AI Review]
    G --> H[Compliance]
```

3. Rule Definition

Setiap Rule terdiri dari:

Identifier
Scope
Condition
Constraint
Severity
Action
Message
Reference
4. Rule Lifecycle
5. Canonical Rule Schema
rule:

  id:

  title:

  description:

  scope:

  applies_to:

  condition:

  constraint:

  severity:

  action:

  message:

  references:
6. Rule Severity

Empat tingkat severity:

Level	Arti
Info	Informasi
Warning	Perlu perhatian
Error	Melanggar spesifikasi
Critical	Harus dihentikan

Compiler MUST menghentikan proses jika terdapat pelanggaran Critical.

7. Rule Scope

Rule dapat diterapkan pada:

Repository
Project
Module
Component
Document
API
Database
Workflow
AI Agent
8. Rule Categories

Kategori standar:

Architecture

Documentation

Security

API

Database

Testing

Performance

AI

Compliance

Governance

Generation (Multi-Language SDK)

### 8.1 Generation Rules

Kategori `Generation` mendefinisikan aturan untuk produksi artefak multi-bahasa:

| Rule | Scope | Condition | Severity |
|------|-------|-----------|----------|
| GEN-001 | Project | `generation.languages` contains only supported languages | Error |
| GEN-002 | Project | At least one adapter registered for each target language | Warning |
| GEN-003 | Module | Module path valid per target language conventions | Error |
| GEN-004 | Service | Service port not conflict across adapters | Error |
| GEN-005 | Project | Output directory writable | Error |

### 8.2 Adapter Validation Rules

Setiap adapter harus mematuhi aturan berikut:

| Rule | Description |
|------|-------------|
| ADP-001 | Adapter MUST implement `OutputAdapter` interface |
| ADP-002 | Adapter MUST register via `init()` in its package |
| ADP-003 | Adapter MUST return valid `language.Language` from `Language()` |
| ADP-004 | Adapter MUST NOT generate artifacts that conflict with default engine |
| ADP-005 | Adapter SHOULD generate artifacts following target language conventions |
9. Rule Evaluation

Urutan evaluasi:

Metadata

↓

Dependencies

↓

Artifact

↓

Rule Engine

↓

Validation Result
10. Rule Inheritance

Rule dapat diwariskan.

Contoh:

Engineering Constitution

↓

Security Standard

↓

Backend Standard

↓

Project Rule

Rule yang lebih spesifik dapat memperketat aturan, tetapi tidak boleh melemahkan aturan induknya.

11. Rule Conflict Resolution

Jika dua Rule bertentangan, prioritasnya:

Constitution

↓

Core Specification

↓

Standards

↓

Project Rules

↓

Local Rules

Rule dengan prioritas lebih tinggi selalu menang.

12. Rule Expressions

Rule dapat ditulis menggunakan ekspresi deklaratif.

Contoh:

condition:
  artifact.type == "API"

constraint:
  metadata.version exists

severity:
  Error

Rule Language harus mudah dibaca manusia sekaligus dapat diproses oleh mesin.

13. Validation Output

Validator menghasilkan laporan seperti:

result:

  status: Failed

  severity: Error

  rule: NAEOS-RULE-001

  artifact: API-001

  message: Missing version metadata
14. AI Integration

AI Agent menggunakan Rule Model untuk:

memeriksa hasil generate,
memberikan rekomendasi,
menjelaskan alasan pelanggaran,
mengusulkan perbaikan.

Dengan demikian AI tidak hanya menghasilkan kode, tetapi juga mematuhi aturan engineering.

### 14.1 Adapter Validation by AI

AI Agent dapat memvalidasi artefak yang dihasilkan oleh adapter:

```
Generated Artifact (Go / TypeScript / Python / Java / Rust)
       │
       ▼
Rule Engine (Generation Rules)
       │
       ├──→ Language convention check
       ├──→ Build file validity
       ├──→ Module structure check
       └──→ Dockerfile correctness
       │
       ▼
Validation Result
```

AI menggunakan rule kategori Generation dan adapter-specific rules untuk memastikan artefak output memenuhi standar bahasa target.

15. Compliance Engine

Rule Model menjadi dasar Compliance Engine.

Engine harus mampu:

mengevaluasi ribuan Rule,
menghasilkan laporan kepatuhan,
menghitung skor kualitas,
melacak tren pelanggaran.
16. Extensibility

Organisasi dapat membuat Rule baru menggunakan namespace.

Contoh:

ORG-RULE-SEC-001
ORG-RULE-API-005

Rule inti NAEOS tidak boleh diubah.

17. Conformance

Implementasi Rule Model MUST:

mendukung Rule Schema resmi,
mendukung Rule Inheritance,
mendukung Rule Evaluation,
menghasilkan Validation Report standar.
18. Related Documents
ID	Document
NAEOS-SPEC-002	Engineering Knowledge Graph
NAEOS-SPEC-003	Universal Artifact Model
NAEOS-SPEC-004	Metadata Specification
NAEOS-SPEC-006	Dependency Graph
NAEOS-SPEC-008	Compiler Model
NES-039	SDK Multi-Language Specification
NES-040	Output Adapter Architecture
Revision History
Version	Date	Change
1.0.0	2026-07-09	Initial Rule Model
1.1.0	2026-07-10	Expanded Rule Architecture diagram, added Generation Rules, Adapter Validation Rules, AI Adapter Validation
Status
NAEOS-SPEC-005

APPROVED

NAEOS Rule Language Established
