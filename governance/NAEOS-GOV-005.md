Document ID : NAEOS-GOV-005
Title       : Core Principles
Version     : 1.0.0
Status      : Stable
Owner       : NAEOS OSS
Category    : Governance
Priority    : Critical

Motto:
  Architecture Drives Engineering.

Depends On:
  - NAEOS-GOV-001 Project Charter
  - NAEOS-GOV-002 Vision
  - NAEOS-GOV-003 Mission
  - NAEOS-GOV-004 Manifesto

Referenced By:
  - NAEOS-GOV-006 Governance Model
  - NAEOS-CON-001 Engineering Constitution
  - NAEOS-SPEC-001 NAEOS Overview

NAEOS Core Principles
Executive Summary

Core Principles defines prinsip fundamental that become dasar all decisions teknis and arsitektural in ekosistem NAEOS.

Prinsip-prinsip ini bersifat normatif — setiap komponen NAEOS must konsisten with prinsip-prinsip following.

1. Purpose

This document menjawab pertanyaan:

"Prinsip apa that mengatur all decisions in NAEOS?"

Core Principles become filter to evaluasi setiap proposal, RFC, and ADR in ekosistem NAEOS.

2. The Principles

Principle 01
Specification is the Single Source of Truth

English: The specification defines what exists, what is valid, and what is required. No other artifact holds higher authority.

Indonesia: Spesifikasi defines apa that ada, apa that valid, and apa that required. Not ada artefak lain that have otoritas more tinggi.

Implikasi:

Kode must sinkron with spesifikasi.
Dokumentasi must dihasilkan from spesifikasi.
Decisions engineering must terdokumentasi in spesifikasi.
Prompt or instruksi AI not menggantikan spesifikasi.

Principle 02
Architecture Precedes Implementation

English: Design decisions must be made and documented before code is written.

Indonesia: Decisions desain must dibuat and didokumentasikan sebelum kode ditulis.

Implikasi:

Setiap modul must have arsitektur that didefinisikan.
Changes arsitektur must through proses governance.
AI not bby produce kode without memahami konteks arsitektural.

Principle 03
Knowledge is Reusable

English: Engineering knowledge must be structured for discovery and reuse across projects and teams.

Indonesia: Pengetahuan engineering must terstruktur to ditemukan and used ulang across proyek and tim.

Implikasi:

Knowledge must disimpan in format that can diquery.
Knowledge must terhubung with konteks (decisions, komponen, implementasi).
Knowledge must versi and can traced (traceable).

Principle 04
Documentation is Part of the Product

English: Documentation is not an afterthought; it is an artifact that must be generated, validated, and maintained alongside code.

Indonesia: Dokumentasi not hal that dipikirkan belakangan; dokumentasi is artefak that must dihasilkan, divalidasi, and maintained bersama kode.

Implikasi:

Dokumentasi must dihasilkan from spesifikasi.
Dokumentasi must divalidasi by governance.
Dokumentasi must versioned and terlacak.

Principle 05
Automation Reinforces Engineering

English: Manual processes that can be automated should be automated to ensure consistency and reduce human error.

Indonesia: Proses manual that can diotomasi must diotomasi to ensure konsistensi and mengurangi human error.

Implikasi:

Validator must berjalan in otomatis.
Pipeline must terotomasi from spesifikasi hingga artefak.
Governance rules must can dievaluasi in programmatic.

Principle 06
Every Rule Must Be Explainable

English: Every policy rule, governance decision, and validation constraint must have a clear rationale.

Indonesia: Setiap aturan policy, decisions governance, and kendala validasi must have alasan that jelas.

Implikasi:

Setiap rule must have deskripsi and rationale.
Decisions governance must can traced ke prinsip.
Error messages must jelas and actionable.

Principle 07
Every Artifact Must Be Traceable

English: Every generated artifact must maintain provenance back to its source specification and the decisions that shaped it.

Indonesia: Setiap artefak that dihasilkan must mempertahankan provenance kembali ke spesifikasi sumber and decisions that membentuknya.

Implikasi:

Provenance tracking must tercatat to setiap artefak.
Lineage must can traced mundur (backward tracing).
Metadata provenance must available to auditing.

Principle 08
Every Decision Should Be Reviewable

English: Engineering decisions should be captured in a format that allows future review, challenge, and evolution.

Indonesia: Decisions engineering must ditangkap in format that memungkinkan review, tantangan, and evolusi di masa depan.

Implikasi:

Decisions must didokumentasikan as ADR or RFC.
Proses review must terbuka and transparan.
Evolusi decisions must terlacak.

3. Principle Hierarchy

Prinsip-prinsip di atas have hierarki:

1. Specification is the Single Source of Truth (fondasi)
2. Architecture Precedes Implementation (desain)
3. Knowledge is Reusable (pengetahuan)
4. Documentation is Part of the Product (dokumentasi)
5. Automation Reinforces Engineering (otomasi)
6. Every Rule Must Be Explainable (transparansi)
7. Every Artifact Must Be Traceable (auditability)
8. Every Decision Should Be Reviewable (evolusi)

Jika terjadi konflik antar prinsip, prinsip with nomor more rendah have prioritas more tinggi.

4. Application in Practice

4.1 RFC Process

Setiap RFC must mendemonstrasikan konsistensi with prinsip-prinsip di atas.

4.2 Policy Design

Setiap policy rule must have rationale that merujuk ke prinsip that relevan.

4.3 Architecture Review

Setiap review arsitektur must mengevaluasi konsistensi with prinsip-prinsip di atas.

5. Exceptions

Pengecualian to prinsip hanya can diberikan through:

RFC resmi with justifikasi kuat,
perseobjective governance board,
dokumentasi in ADR terkait.

6. Revision Policy

Changes on Core Principles hanya can performed through RFC resmi and perseobjective governance.

7. References

- NAEOS-GOV-001 Project Charter
- NAEOS-GOV-002 Vision
- NAEOS-GOV-003 Mission
- NAEOS-GOV-004 Manifesto
- NAEOS-CON-001 Engineering Constitution
