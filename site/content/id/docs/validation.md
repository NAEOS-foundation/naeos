---
title: Validasi
description: Sistem validasi multi-level untuk spesifikasi, modul, layanan, dan kebijakan.
weight: 16
---

NAEOS menyediakan validasi komprehensif di beberapa level untuk menangkap error lebih awal dan menerapkan konsistensi di seluruh proyek Anda.

## Level Validasi

```text
┌─────────────────────────────────────────────┐
│              Sistem Validasi                 │
├──────────┬──────────┬──────────┬────────────┤
│   Spec   │  Module  │ Service  │  Policy    │
│  Valid.  │  Valid.  │  Valid.  │  Valid.   │
└──────────┴──────────┴──────────┴────────────┘
```

## Validasi Spesifikasi

### Penggunaan CLI

```bash
# Validasi spesifikasi
naeos validate --input-file spec.yaml

# Lint spesifikasi (pengecekan gaya tambahan)
naeos lint --input-file spec.yaml

# Output sebagai JSON
naeos validate --input-file spec.yaml --output json
```

### Go API

```go
import "github.com/NAEOS-foundation/naeos/internal/lint"

result, err := lint.ValidateSpec(specContent)
if err != nil {
    log.Fatal(err)
}

if !result.Valid {
    for _, issue := range result.Issues {
        fmt.Printf("[%s] %s\n", issue.Severity, issue.Message)
    }
}
```

## Validasi Module

### Deteksi Duplikat

Mendeteksi modul dengan nama yang sama:

```go
validator := parser.NewSpecValidator()
issues := validator.ValidateModules(modules)

// Issues akan menyertakan "module-duplicate" jika ada duplikat
```

### Deteksi Circular Dependency

Mendeteksi circular dependency antar modul:

```go
modules := []parser.Module{
    {Name: "a", Dependencies: []string{"b"}},
    {Name: "b", Dependencies: []string{"a"}},
}

issues := validator.ValidateModules(modules)
// Issues akan menyertakan "circular-dependency"
```

### Deteksi Dangling Dependency

Mendeteksi referensi ke modul yang tidak ada:

```go
modules := []parser.Module{
    {Name: "auth", Dependencies: []string{"nonexistent"}},
}

issues := validator.ValidateModules(modules)
// Issues akan menyertakan "dependency-not-found"
```

## Validasi Service

### Deteksi Konflik Port

Mendeteksi beberapa layanan menggunakan port yang sama:

```go
services := []parser.Service{
    {Name: "api", Port: 8080},
    {Name: "web", Port: 8080},
}

issues := validator.ValidateServices(services)
// Issues akan menyertakan "service-port-conflict"
```

### Validasi Range Port

Memastikan port berada dalam rentang valid (1–65535):

```go
services := []parser.Service{
    {Name: "bad", Port: 99999},
}

issues := validator.ValidateServices(services)
// Issues akan menyertakan "service-port-range"
```

## Validasi Versi Skema

### Pemeriksaan Otomatis Saat Parse

Parser secara otomatis memvalidasi versi skema:

```yaml
# Spesifikasi ini akan diperiksa terhadap versi minimum 0.1.0
version: 0.3.0
```

### Pemeriksaan Manual

```go
result := parser.CheckSpecVersion("0.3.0")
if !result.Valid {
    fmt.Println(result.Message)
    // "Schema version 0.3.0 is not supported. Minimum: 0.1.0"
}
```

## Validasi Kebijakan

### Aturan Default

```go
import "github.com/NAEOS-foundation/naeos/internal/governance/policy"

evaluator := policy.NewEvaluator()
rules := policy.DefaultRules()

ctx := map[string]any{
    "project":  "my-app",
    "modules":  3,
    "services": 2,
}

results, err := evaluator.EvaluateRules(rules, ctx)
for _, r := range results {
    fmt.Printf("[%s] %s: %s\n", r.RuleID, r.Action, r.Message)
}
```

### Aturan Kustom

```go
rules := []policy.Rule{
    {
        RuleID:    "require-architecture",
        Condition: "exists:architecture",
        Priority:  1,
        Action:    "block",
        Scope:     "spec",
        Enabled:   true,
    },
    {
        RuleID:    "max-modules",
        Condition: "lt:module_count,50",
        Priority:  2,
        Action:    "warn",
        Scope:     "spec",
        Enabled:   true,
    },
}
```

## Hasil Validasi

```go
type ValidationResult struct {
    Valid        bool
    Issues       []ValidationIssue
    Warnings     []ValidationIssue
    ModuleCount  int
    ServiceCount int
}

type ValidationIssue struct {
    Severity string  // "error", "warning", "info"
    Rule     string
    Message  string
    Line     int
}
```

## Integrasi Pipeline

Validasi berjalan secara otomatis sebagai bagian dari pipeline:

```bash
# Pipeline lengkap menyertakan validasi
naeos run --input spec.yaml

# Mode validasi saja
naeos validate --input-file spec.yaml
```

### Lewati Validasi Tertentu

```bash
# Lewati validasi module
naeos run --input spec.yaml --skip-validation

# Lewati pemeriksaan tertentu
naeos validate --input-file spec.yaml --skip circular-deps
```

Lihat juga: [Governance](/id/docs/governance/), [Pipeline Engine](/id/docs/pipeline-engine/), [Bahasa Spesifikasi](/id/docs/spec-language/)
