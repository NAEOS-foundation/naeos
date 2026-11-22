---
title: Peta Jalan
description: Strategi eksekusi NAEOS dari konsistensi publik, Golden Path yang dapat direproduksi, adopsi eksternal, hingga ekosistem dan komersialisasi.
---

## Strategi Saat Ini

NAEOS mengurutkan eksekusi dengan prinsip **proof first, adoption second, ecosystem third, commercialization last**.

Arah produk canonical saat ini:

> **The Engineering Control Plane for AI Coding Agents**

Model control canonical:

```
Specification
→ NEIR
→ Validation + Policy
→ Agent Intent
→ Authorized Execution
→ Observation
→ Evidence
→ Independent Verification
```

## P0 — Konsistensi Publik

**Status: ACTIVE**

Sinkronkan seluruh permukaan publik dengan kondisi engineering repository yang sebenarnya.

- Positioning dan navigasi website
- SEO titles, descriptions, canonical messaging, dan structured copy
- Whitepaper
- FAQ
- About
- Roadmap dan referensi dokumentasi publik
- Pemisahan versi software release dengan milestone engineering/experiment

## P1 — Golden Path

**Status: NEXT**

Jadikan control-plane proof yang sudah ada sebagai demonstrasi developer utama.

```
Control Plane
→ Golden Path
→ P1.6–P1.10 Reference Demo
→ Evidence
→ Independent Verification (P1.11)
```

Targetnya adalah alur fresh-checkout yang dapat dijalankan, diperiksa, dan dipahami technical evaluator dalam beberapa menit.

## P2 — Adopsi Eksternal

**Status: GATED BY P1**

Bekerja dengan 5–10 developer atau repository pertama.

Capture:
- jalur evaluasi yang berhasil;
- friction integrasi;
- deviasi dari Golden Path terdokumentasi;
- failure modes;
- dokumentasi yang kurang;
- kualitas evidence;
- kebutuhan berulang dari evaluator independen.

Observasi ini menjadi input prioritas engineering berikutnya.

## P3 — Ekosistem

**Status: GATED BY P2**

Prioritaskan SDK, integrasi, dan marketplace hanya jika evidence adopsi eksternal menunjukkan kebutuhan yang berulang.

Tujuannya memperluas control boundary tanpa memecah governance atau model evidence.

## P4 — Komersialisasi

**Status: GATED BY P2**

Packaging Cloud dan Enterprise mengikuti evidence adopsi teknis, bukan mendahuluinya.

Arah komersial ditentukan oleh deployment boundary, operational requirements, governance needs, dan evidence dari penggunaan eksternal.

## Engineering Proof yang Selesai

- **P1.6** — Public Control-Plane Golden Path — DONE
- **P1.7** — Policy Change Mid-Run / Stale Authorization — DONE
- **P1.8** — Atomic Execution Commit Boundary — DONE
- **P1.9** — Evidence & Verification Plane — DONE
- **P1.10** — Evidence Query & Audit API — DONE
- **P1.11** — Independent Verifier CLI — DONE

P1.11 adalah capability verifikasi yang mendukung proof path, bukan alasan membuka siklus feature expansion baru.

## Urutan Strategis

```
Public consistency
→ Reproducible Golden Path
→ 5–10 external evaluators
→ Evidence-backed ecosystem priorities
→ Cloud / Enterprise
```

NAEOS sengaja belum memprioritaskan ekspansi dashboard, fitur AI provider-specific, pertumbuhan marketplace, atau packaging komersial sebelum tersedia evidence adopsi teknis.