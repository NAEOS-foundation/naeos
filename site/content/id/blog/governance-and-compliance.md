---
title: "Tata Kelola & Kepatuhan di NAEOS: Kebijakan, Audit, dan SSO Terintegrasi"
description: "Bagaimana NAEOS membantu tim menerapkan kebijakan, menjaga jejak audit, dan memenuhi persyaratan SOC 2, HIPAA, dan GDPR."
date: 2026-07-25
author: "NAEOS Foundation"
categories: ["tutorial"]
---

Proyek perangkat lunak enterprise menghadapi berbagai persyaratan tata kelola: penegakan kebijakan, jejak audit, kontrol akses, dan kepatuhan regulasi. NAEOS menanamkan tata kelola langsung ke dalam pipeline.

## Sistem Kebijakan

NAEOS hadir dengan lima aturan kebijakan bawaan yang berjalan selama tahap validasi: `project-required`, `modules-required`, `architecture-pattern-valid`, `deployment-strategy-valid`, dan `service-port-positive`.

Aturan kustom dapat didefinisikan dalam Go:

```go
rules := []policy.Rule{
    {
        RuleID:    "require-testing",
        Condition: "exists:testing",
        Priority:  1,
        Action:    "block",
    },
}
```

## RBAC

NAEOS menyertakan sistem RBAC hierarkis dengan tiga peran bawaan: **admin** (akses penuh), **developer** (menjalankan pipeline, mengelola spesifikasi), dan **viewer** (akses baca-saja). Mendukung rantai induk dan aturan tolak.

## Jejak Audit

Setiap eksekusi pipeline, evaluasi kebijakan, dan pembuatan artefak dicatat. Tiga backend tersedia: **Standard Auditor** (JSON log), **Hashed Chain Auditor** (SHA256 tamper-evident chain), dan **Encrypted Auditor** (AES-256-GCM).

## Kerangka Kepatuhan

NAEOS menyediakan evaluasi kepatuhan bawaan untuk SOC 2 (8 kontrol), HIPAA (11 kontrol), dan GDPR (8 pasal):

```bash
naeos compliance report --framework soc2 --output report.json
naeos compliance report --framework hipaa --output-format json
```

## Integrasi SSO

Tiga protokol SSO enterprise didukung: OIDC, SAML 2.0, dan LDAP. Konfigurasi melalui CLI:

```bash
naeos auth sso configure \
  --provider-type oidc \
  --issuer https://accounts.google.com \
  --client-id your-client-id
```

## Tata Kelola di CI/CD

Integrasikan pemeriksaan tata kelola ke pipeline CI Anda: validasi spesifikasi, lint, audit, dan verifikasi kepatuhan — semuanya dalam satu workflow.

Untuk dokumentasi lengkap, lihat [dokumentasi Tata Kelola](/docs/governance/).
