---
title: Tentang NAEOS
description: NAEOS adalah engineering control plane open-source untuk AI coding agents, dengan governance, authorized execution, evidence, dan verification.
---

## Apa itu NAEOS

NAEOS (Nusantara AI Engineering Operating System) adalah engineering control plane open-source dan vendor-neutral untuk AI coding agents.

Fokus NAEOS adalah satu batas engineering yang jelas: **AI agent boleh mengusulkan pekerjaan, tetapi sistem engineering yang menentukan apa yang diizinkan, apa yang dieksekusi, dan evidence apa yang tersisa.**

Model canonical saat ini:

```text
Specification
    ↓
NEIR
    ↓
Validation + Policy
    ↓
Agent Intent
    ↓
Authorized Execution
    ↓
Observation
    ↓
Evidence
    ↓
Independent Verification
```

## Mengapa kami membangunnya

AI-assisted software development mengubah kecepatan dan bentuk pekerjaan engineering. Tantangannya bukan hanya menghasilkan kode; tim juga membutuhkan authorization boundary yang eksplisit, execution yang reproducible, traceability, dan evidence yang dapat diperiksa setelah agent selesai.

NAEOS dibangun sebagai layer tersebut tanpa bergantung pada satu AI vendor, agent runtime, atau cloud provider.

## Proof saat ini

Repository telah melampaui tahap konsep. Public proof path menggabungkan Control Plane live, Golden Path, Reference Demo, evidence records, dan independent verifier.

P1.6–P1.10 adalah urutan proof utama Golden Path. P1.11 menyediakan independent verification terhadap evidence yang telah diserialisasi.

## Prinsip proyek

- **Governance sebelum execution** — policy dan authorization menentukan batas execution.
- **Evidence di atas claim** — perilaku penting harus meninggalkan evidence yang dapat diperiksa.
- **Independent verification** — verification tidak bergantung pada agent yang melakukan pekerjaan.
- **Vendor neutrality** — control boundary tidak bergantung pada satu AI provider.
- **Reproducibility** — evaluator teknis harus dapat mengulang proof yang didokumentasikan.
- **Human accountability** — AI membantu engineering; manusia tetap bertanggung jawab atas perubahan konsekuensial.

## Arah eksekusi

1. **P0 — Public consistency:** sinkronkan website, SEO, whitepaper, FAQ, About, dan dokumentasi publik dengan repository truth.
2. **P1 — Golden Path:** jadikan P1.6–P1.10 sebagai demo developer yang reproducible.
3. **P2 — External adoption:** bekerja dengan 5–10 developer atau repository pertama dan mengumpulkan evidence serta failure modes.
4. **P3 — Ecosystem:** prioritaskan SDK, integrations, dan marketplace berdasarkan evidence adoption.
5. **P4 — Commercialization:** evaluasi Cloud/Enterprise setelah tersedia evidence technical adoption.