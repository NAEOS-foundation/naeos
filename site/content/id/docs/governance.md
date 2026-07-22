---
title: Governance
description: Sistem kebijakan, alur kerja review, dan audit trail untuk NAEOS.
weight: 18
---

NAEOS menyediakan sistem governance komprehensif untuk menerapkan kebijakan, memvalidasi artifact, dan mempertahankan audit trail di seluruh alur kerja engineering Anda.

## Struktur Governance

```text
┌─────────────────────────────────────────┐
│            Governance Model             │
├─────────────┬─────────────┬─────────────┤
│  Charter    │  Policies   │  Reviews    │
│  Vision     │  Rules      │  Audit      │
│  Mission    │  Evaluator  │  Trail      │
│  Manifesto  │  Enforcer   │  Reports    │
└─────────────┴─────────────┴─────────────┘
```

## Hierarki Dokumen

Governance NAEOS mengikuti hierarki dokumen yang ketat:

| Dokumen | ID | Deskripsi |
|---------|----|-----------|
| Konstitusi Engineering | NAEOS-CON-001 | Aturan engineering fundamental |
| Piagam Proyek | NAEOS-GOV-001 | Cakupan dan tujuan proyek |
| Visi | NAEOS-GOV-002 | Arah jangka panjang |
| Misi | NAEOS-GOV-003 | Bagaimana kita mencapai visi |
| Manifesto | NAEOS-GOV-004 | Keyakinan dan nilai inti |
| Prinsip Inti | NAEOS-GOV-005 | Prinsip engineering |
| Model Governance | NAEOS-GOV-006 | Kerangka pengambilan keputusan |
| Peta Jalan | NAEOS-GOV-007 | Milestone pengembangan |
| Kebijakan Versioning | NAEOS-GOV-008 | Aturan SemVer dan konvensi |

## Sistem Kebijakan

### Aturan Default

NAEOS menyertakan 5 aturan kebijakan default:

| Rule ID | Kondisi | Prioritas | Aksi |
|---------|---------|-----------|------|
| `project-required` | `exists:project` | 1 | block |
| `modules-required` | `exists:modules` | 1 | block |
| `architecture-pattern-valid` | `in:architecture.pattern,...` | 2 | warn |
| `deployment-strategy-valid` | `in:deployment.strategy,...` | 2 | warn |
| `service-port-positive` | `gt:service.port,0` | 3 | warn |

### Operator Kondisi

| Operator | Contoh | Deskripsi |
|----------|--------|-----------|
| `exists` | `exists:project` | Periksa apakah key ada |
| `not_empty` | `not_empty:project` | Periksa apakah key tidak kosong |
| `contains` | `contains:project,app` | Periksa apakah value mengandung substring |
| `gt` | `gt:port,0` | Greater than |
| `lt` | `lt:port,65535` | Less than |
| `in` | `in:kind,http,grpc` | Value ada di daftar yang diizinkan |

### Kebijakan Kustom

Definisikan aturan kustom di kode Go Anda:

```go
import "github.com/NAEOS-foundation/naeos/internal/governance/policy"

rules := []policy.Rule{
    {
        RuleID:    "require-testing",
        Condition: "exists:testing",
        Priority:  1,
        Action:    "block",
        Scope:     "spec",
        Enabled:   true,
    },
    {
        RuleID:    "limit-modules",
        Condition: "lt:module_count,50",
        Priority:  2,
        Action:    "warn",
        Scope:     "spec",
        Enabled:   true,
    },
}
```

## Sistem Review

Sistem review memeriksa artifact yang dihasilkan terhadap aturan kualitas:

```go
import "github.com/NAEOS-foundation/naeos/internal/governance/review"

reviewer := review.NewReviewer()

result, err := reviewer.ReviewArtifact(
    "main.go",
    string(content),
    []string{"no-todo", "no-placeholder", "no-hardcoded-secrets"},
)

if !result.Passed {
    for _, issue := range result.Issues {
        fmt.Printf("[%s] %s\n", issue.Severity, issue.Message)
    }
}
```

## Audit Trail

Setiap aksi signifikan di-log untuk ketelusuran:

```bash
# Jalankan audit pada spesifikasi
naeos audit --input-file spec.yaml

# Ekspor log audit
naeos audit --input-file spec.yaml --output audit-report.json
```

Audit trail mencakup:
- Perubahan spesifikasi (create, update, migrate)
- Eksekusi pipeline (start, complete, fail)
- Evaluasi kebijakan (pass, block, warn)
- Generasi artifact (create, review, export)

## Integrasi CI/CD

Terapkan governance di pipeline CI Anda:

```yaml
# .github/workflows/ci.yml
- name: Validate Spec
  run: naeos validate --input-file spec.yaml

- name: Lint Spec
  run: naeos lint --input-file spec.yaml

- name: Audit
  run: naeos audit --input-file spec.yaml
```

## Integrasi Pipeline

Aktifkan governance di konfigurasi pipeline:

```go
cfg := pipeline.Config{
    Policies: policy.DefaultRules(),
    Review:   review.Config{Enabled: true},
    Audit:    audit.Config{Enabled: true},
}
```

Lihat juga: [Prinsip Inti](/id/docs/core-principles/), [Validasi](/id/docs/validation/)
