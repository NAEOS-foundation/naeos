---
title: "Governance & Compliance: Policy, Authorization, dan Batas Evidence"
description: "Pelajari policy, authorization, execution evidence, dan cara mengevaluasi capability governance NAEOS tanpa menganggapnya sebagai compliance certification."
date: 2026-07-25
author: "NAEOS Foundation"
categories: ["tutorial"]
---

Governance engineering membutuhkan lebih dari policy file atau laporan compliance yang dihasilkan otomatis. Tim membutuhkan authorization boundary yang jelas, execution yang dapat ditelusuri, evidence yang dapat diperiksa, dan independent verification.

## Governance di NAEOS

NAEOS adalah engineering control plane open-source dan vendor-neutral untuk AI coding agents. Public proof path saat ini menunjukkan policy dan authorization boundary, controlled execution, durable evidence, serta independent verification.

```
Specification → NEIR → Validation + Policy
→ Agent Intent → Authorized Execution
→ Observation → Evidence → Independent Verification
```

## Evaluasi control berdasarkan evidence

Saat mengevaluasi governance, pisahkan:
- **Policy evaluation** — rule dan input yang dievaluasi serta decision yang dihasilkan.
- **Authorization** — action, identity, scope, dan capability yang diizinkan atau ditolak.
- **Execution** — apakah action terotorisasi benar-benar berjalan dan outcome yang diamati.
- **Evidence** — record yang disimpan dan cara integrity serta identity binding diverifikasi.
- **Compliance mapping** — pemetaan implemented control terhadap framework, termasuk gap dan scope.

Contoh policy atau framework mapping bukan bukti implementasi control atau certification dengan sendirinya. NAEOS tidak mengklaim blanket compliance SOC 2, HIPAA, atau GDPR. Organisasi harus menilai scope, konfigurasi, proses operasional, dan independent evidence pada environment mereka sendiri.

## Evaluasi Teknis

Mulai dari Control Plane publik dan Golden Path yang reproducible. Periksa decision dan execution records, lalu jalankan independent verifier terhadap evidence terserialisasi. Gunakan dokumentasi serta release baseline repository saat ini, bukan asumsi dari deskripsi fitur lama.

- [FAQ](/id/faq/)
- [Tentang NAEOS](/id/about/)
- [Whitepaper](/id/whitepaper/)
- [Repository](https://github.com/NAEOS-foundation/naeos)
