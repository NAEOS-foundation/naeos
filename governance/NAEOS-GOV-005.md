Document ID : NAEOS-GOV-005
Title       : Core Principles
Version     : 1.0.0
Status      : Stable
Owner       : NAEOS Foundation
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

Core Principles mendefinisikan prinsip fundamental yang menjadi dasar seluruh keputusan teknis dan arsitektural dalam ekosistem NAEOS.

Prinsip-prinsip ini bersifat normatif — setiap komponen NAEOS harus konsisten dengan prinsip-prinsip berikut.

1. Purpose

Dokumen ini menjawab pertanyaan:

"Prinsip apa yang mengatur seluruh keputusan dalam NAEOS?"

Core Principles menjadi filter untuk evaluasi setiap proposal, RFC, dan ADR dalam ekosistem NAEOS.

2. The Principles

Principle 01
Specification is the Single Source of Truth

English: The specification defines what exists, what is valid, and what is required. No other artifact holds higher authority.

Indonesia: Spesifikasi mendefinisikan apa yang ada, apa yang valid, dan apa yang diperlukan. Tidak ada artefak lain yang memiliki otoritas lebih tinggi.

Implikasi:

Kode harus sinkron dengan spesifikasi.
Dokumentasi harus dihasilkan dari spesifikasi.
Keputusan engineering harus terdokumentasi dalam spesifikasi.
Prompt atau instruksi AI tidak menggantikan spesifikasi.

Principle 02
Architecture Precedes Implementation

English: Design decisions must be made and documented before code is written.

Indonesia: Keputusan desain harus dibuat dan didokumentasikan sebelum kode ditulis.

Implikasi:

Setiap modul harus memiliki arsitektur yang didefinisikan.
Perubahan arsitektur harus melalui proses governance.
AI tidak boleh menghasilkan kode tanpa memahami konteks arsitektural.

Principle 03
Knowledge is Reusable

English: Engineering knowledge must be structured for discovery and reuse across projects and teams.

Indonesia: Pengetahuan engineering harus terstruktur untuk ditemukan dan digunakan ulang lintas proyek dan tim.

Implikasi:

Knowledge harus disimpan dalam format yang dapat diquery.
Knowledge harus terhubung dengan konteks (keputusan, komponen, implementasi).
Knowledge harus versi dan dapat ditelusuri (traceable).

Principle 04
Documentation is Part of the Product

English: Documentation is not an afterthought; it is an artifact that must be generated, validated, and maintained alongside code.

Indonesia: Dokumentasi bukan hal yang dipikirkan belakangan; dokumentasi adalah artefak yang harus dihasilkan, divalidasi, dan dipelihara bersama kode.

Implikasi:

Dokumentasi harus dihasilkan dari spesifikasi.
Dokumentasi harus divalidasi oleh governance.
Dokumentasi harus versioned dan terlacak.

Principle 05
Automation Reinforces Engineering

English: Manual processes that can be automated should be automated to ensure consistency and reduce human error.

Indonesia: Proses manual yang dapat diotomasi harus diotomasi untuk memastikan konsistensi dan mengurangi human error.

Implikasi:

Validator harus berjalan secara otomatis.
Pipeline harus terotomasi dari spesifikasi hingga artefak.
Governance rules harus dapat dievaluasi secara programmatic.

Principle 06
Every Rule Must Be Explainable

English: Every policy rule, governance decision, and validation constraint must have a clear rationale.

Indonesia: Setiap aturan policy, keputusan governance, dan kendala validasi harus memiliki alasan yang jelas.

Implikasi:

Setiap rule harus memiliki deskripsi dan rationale.
Keputusan governance harus dapat ditelusuri ke prinsip.
Error messages harus jelas dan actionable.

Principle 07
Every Artifact Must Be Traceable

English: Every generated artifact must maintain provenance back to its source specification and the decisions that shaped it.

Indonesia: Setiap artefak yang dihasilkan harus mempertahankan provenance kembali ke spesifikasi sumber dan keputusan yang membentuknya.

Implikasi:

Provenance tracking harus tercatat untuk setiap artefak.
Lineage harus dapat ditelusuri mundur (backward tracing).
Metadata provenance harus tersedia untuk auditing.

Principle 08
Every Decision Should Be Reviewable

English: Engineering decisions should be captured in a format that allows future review, challenge, and evolution.

Indonesia: Keputusan engineering harus ditangkap dalam format yang memungkinkan review, tantangan, dan evolusi di masa depan.

Implikasi:

Keputusan harus didokumentasikan sebagai ADR atau RFC.
Proses review harus terbuka dan transparan.
Evolusi keputusan harus terlacak.

3. Principle Hierarchy

Prinsip-prinsip di atas memiliki hierarki:

1. Specification is the Single Source of Truth (fondasi)
2. Architecture Precedes Implementation (desain)
3. Knowledge is Reusable (pengetahuan)
4. Documentation is Part of the Product (dokumentasi)
5. Automation Reinforces Engineering (otomasi)
6. Every Rule Must Be Explainable (transparansi)
7. Every Artifact Must Be Traceable (auditability)
8. Every Decision Should Be Reviewable (evolusi)

Jika terjadi konflik antar prinsip, prinsip dengan nomor lebih rendah memiliki prioritas lebih tinggi.

4. Application in Practice

4.1 RFC Process

Setiap RFC harus mendemonstrasikan konsistensi dengan prinsip-prinsip di atas.

4.2 Policy Design

Setiap policy rule harus memiliki rationale yang merujuk ke prinsip yang relevan.

4.3 Architecture Review

Setiap review arsitektur harus mengevaluasi konsistensi dengan prinsip-prinsip di atas.

5. Exceptions

Pengecualian terhadap prinsip hanya dapat diberikan melalui:

RFC resmi dengan justifikasi kuat,
persetujuan governance board,
dokumentasi dalam ADR terkait.

6. Revision Policy

Perubahan pada Core Principles hanya dapat dilakukan melalui RFC resmi dan persetujuan governance.

7. References

- NAEOS-GOV-001 Project Charter
- NAEOS-GOV-002 Vision
- NAEOS-GOV-003 Mission
- NAEOS-GOV-004 Manifesto
- NAEOS-CON-001 Engineering Constitution
