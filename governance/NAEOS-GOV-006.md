Document ID: NAEOS-GOV-006
Title: Governance Model
Version: 1.0.0
Status: Stable
Category: Governance
Owner: NAEOS OSS
Priority: Critical

Motto:
  Architecture Drives Engineering.

Depends On:
  - NAEOS-GOV-001 Project Charter
  - NAEOS-GOV-002 Vision
  - NAEOS-GOV-003 Mission
  - NAEOS-GOV-005 Core Principles

Referenced By:
  - NAEOS-GOV-007 Roadmap
  - NAEOS-GOV-008 Versioning Policy
  - NAEOS-SPEC-008 Compiler Model
  - NAEOS-RFC-*
  - NAEOS-ADR-*
NAEOS Governance Model
Executive Summary

Governance Model defines struktur pengelolaan NAEOS so that proyek can berkembang in terbuka, transparan, and berkelanjutan.

NAEOS menggunakan model governance that menggabungkan:

open source governance,
technical leadership,
community contribution,
formal proposal process.

The primary objective governance:

Ensure setiap changes to NAEOS have alasan, dampak, and proses evaluasi that jelas.

1. Purpose

This document defines:

struktur organisasi NAEOS,
peran and tanggung jawab,
proses pengambilan decisions,
proses RFC,
proses ADR,
kontribusi komunitas,
pengelolaan release.
2. Governance Philosophy

NAEOS follows prinsip:

Open Contribution

+

Technical Excellence

+

Transparent Decision Making

+

Long-Term Sustainability
3. Governance Structure

Struktur governance NAEOS:

4. NAEOS OSS
Purpose

NAEOS OSS bertanggung jawab maintain visi, nilai, and keberlanjutan proyek.

Responsibilities

Foundation:

MUST:

maintain governance,
maintain trademark and identitas,
ensure netralitas proyek.

SHOULD:

support komunitas,
provide dokumentasi,
mengembangkan ekosistem.
5. Technical Steering Committee (TSC)

TSC is baand pengambil decisions teknis tertinggi.

Responsibilities

TSC bertanggung jawab to:

architecture decision,
specification approval,
standard approval,
major release.
Authority

TSC can:

menerima RFC,
menolak RFC,
mengubah specification,
menetapkan roadmap teknis.
6. Maintainer

Maintainer is engineer that bertanggung jawab maintain area tertentu.

Contoh:

Specification Maintainer

Compiler Maintainer

CLI Maintainer

Documentation Maintainer

Security Maintainer
Maintainer Responsibilities

MUST:

melakukan review,
maintain kualitas,
membantu contributor.
7. Contributor

Contributor is individu or organisasi that berkontribusi.

Kontribusi can berupa:

code,
documentation,
specification,
example,
research,
testing.
8. Decision Making Model

NAEOS menggunakan model:

Diagram not valid or not didukung.
9. RFC Process
RFC

(Request For Comments)

used to changes besar.

RFC Required For

MUST menggunakan RFC:

changes specification,
fitur baru,
changes architecture,
changes governance.
RFC Lifecycle
Draft

↓

Review

↓

Accepted

↓

Implementation

↓

Completed
RFC Template
rfc:

  id: RFC-XXXX

  title:

  author:

  status:

  motivation:

  proposal:

  impact:

  alternatives:

  decision:
10. ADR Process
ADR

Architecture Decision Record.

Digunakan to decisions teknis.

Contoh:

ADR-0001

Decision:
Use Go for NAEOS CLI

Reason:
Cross-platform binary distribution
ADR Lifecycle
Proposed

↓

Accepted

↓

Implemented

↓

Superseded
11. Change Management

Semua changes must follows:

12. Community Governance

NAEOS mendorong komunitas through:

Discussion

For:

ide,
feedback,
pertanyaan.
Issue

For:

bug,
improvement,
task.
Pull Request

For:

changes nyata.
13. Code of Conduct

Semua contributor wajib:

menghormati kontribusi,
memberikan kritik konstruktif,
maintain komunikasi profesional.
14. Security Governance

Security issue have proses khusus.

Security report:

MUST:

ditangani in privat,
performed triage,
diperbaiki sebelum disclosure.
15. Release Governance

Release have tiga kategori.

Patch Release

Contoh:

1.0.1

For:

bug fix,
dokumentasi.
Minor Release

Contoh:

1.1.0

For:

fitur baru,
extension.
Major Release

Contoh:

2.0.0

For:

breaking change.
16. Governance Principles

Governance NAEOS follows:

Transparency

Accountability

Merit-Based Contribution

Technical Excellence

Community Respect

Long-Term Thinking
17. Conflict Resolution

Jika terjadi konflik:

Prioritas:

NAEOS Principles

↓

Specification

↓

Technical Evidence

↓

Community Feedback

↓

Decision Authority
18. Governance Anti-Patterns

NAEOS menolak:

Single Person Control

Not bby bergantung on satu individu.

Hidden Decisions

Decisions penting must tercatat.

Unreviewed Standards

Standar must through proses review.

Vendor Influence

Not bby ada dominasi vendor tertentu.

19. Compliance Checklist
Requirement	Level
Have RFC Process	MUST
Have ADR Process	MUST
Have Maintainer	MUST
Transparansi decisions	MUST
Community contribution	SHOULD
Public roadmap	SHOULD
20. Related Documents
ID	Document
NAEOS-GOV-001	Project Charter
NAEOS-GOV-005	Core Principles
NAEOS-GOV-007	Roadmap
NAEOS-GOV-008	Versioning Policy
NAEOS-ADR-*	Architecture Decisions
NAEOS-RFC-*	Feature Proposals
Revision History
Version	Date	Change
1.0.0	2026	Initial Governance Model
Status
NAEOS-GOV-006

APPROVED

Governance Framework Established
