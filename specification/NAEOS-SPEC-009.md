Document ID: NAEOS-SPEC-009

Title: Engineering Reasoning Graph

Short Name: ERG

Version: 1.0.0

Status: Stable

Category: Core Specification

Normative: true

Priority: CRITICAL

Owner: NAEOS Foundation

Motto:
"Every Decision Has a Reason."
Executive Summary

Engineering Reasoning Graph (ERG) adalah model yang merepresentasikan alasan di balik keputusan engineering.

ERG menghubungkan Requirement, Architecture Decision, Policy, Evidence, Risk, dan Outcome sehingga AI maupun engineer dapat menelusuri mengapa suatu keputusan dibuat.

Core Concept

NAEOS memiliki empat graph inti:

Knowledge Graph
        │
        ▼
Dependency Graph
        │
        ▼
Policy Graph
        │
        ▼
Evidence Graph
        │
        ▼
Reasoning Graph

Reasoning Graph berada di atas graph lainnya karena memanfaatkan seluruh informasi tersebut.

Reasoning Node

Setiap node dapat berupa:

Requirement
Assumption
Constraint
Risk
Decision
Alternative
Trade-off
Evidence
Outcome

Contoh:

decision:
  id: ADR-012
  title: Use Event Bus
Relationship Types

Hubungan yang didukung:

supports
justifies
depends_on
conflicts_with
supersedes
mitigates
causes
results_in
derived_from
Example Flow
Requirement
      │
      ▼
Decision

      │
      ▼
Policy

      │
      ▼
Evidence

      │
      ▼
Outcome
AI Integration

AI tidak hanya mengambil konteks, tetapi juga jalur penalaran.

Alur:

Question
      │
      ▼
Knowledge Graph
      │
      ▼
Reasoning Graph
      │
      ▼
Evidence Graph
      │
      ▼
Policy Graph
      │
      ▼
Answer + Justification

Dengan demikian AI dapat menjelaskan:

aturan yang digunakan,
keputusan yang diambil,
bukti yang mendukung,
alternatif yang dipertimbangkan.
Architectural Decision Records (ADR)

Setiap ADR menjadi bagian dari Reasoning Graph.

Contoh relasi:

Requirement
      │
      ▼
ADR
      │
      ▼
Implementation
      │
      ▼
Evidence
Risk Reasoning

Graph juga dapat menghubungkan:

Threat
      │
      ▼
Risk
      │
      ▼
Mitigation
      │
      ▼
Evidence

Ini memperkuat integrasi dengan Security Constitution.

Evolution

Ketika keputusan berubah:

ADR-001
      │
 superseded_by
      ▼
ADR-014

Riwayat penalaran tetap terjaga.

Compiler Support

Compiler dapat menghasilkan:

Decision Reports
Architecture Rationale
Compliance Justification
AI Context Bundle
Executive Summary

berdasarkan Reasoning Graph.

Validation

Validator memeriksa:

keputusan tanpa evidence,
evidence tanpa decision,
requirement tanpa rationale,
risiko tanpa mitigasi.
Conformance

Implementasi dianggap sesuai apabila:

keputusan penting memiliki rationale,
rationale memiliki evidence,
perubahan keputusan terdokumentasi,
graph dapat ditelusuri.
Status
NAEOS-SPEC-010

APPROVED

Engineering Reasoning Graph Established
