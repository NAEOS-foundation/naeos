---
title: Validation
description: Multi-level validation system for specifications, modules, services, and policies.
weight: 16
---

NAEOS provides comprehensive validation at multiple levels to catch errors early and enforce consistency across your project.

## Validation Levels

```text
┌─────────────────────────────────────────────┐
│              Validation System               │
├──────────┬──────────┬──────────┬────────────┤
│   Spec   │  Module  │ Service  │  Policy    │
│  Valid.  │  Valid.  │  Valid.  │  Valid.   │
└──────────┴──────────┴──────────┴────────────┘
```

## Spec Validation

### CLI Usage

```bash
# Validate a specification
naeos validate --input-file spec.yaml

# Lint a specification (additional style checks)
naeos lint --input-file spec.yaml

# Output as JSON
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

## Module Validation

### Duplicate Detection

Detects modules with the same name:

```go
validator := parser.NewSpecValidator()
issues := validator.ValidateModules(modules)

// Issues will include "module-duplicate" if duplicates exist
```

### Circular Dependency Detection

Detects circular dependencies between modules:

```go
modules := []parser.Module{
    {Name: "a", Dependencies: []string{"b"}},
    {Name: "b", Dependencies: []string{"a"}},
}

issues := validator.ValidateModules(modules)
// Issues will include "circular-dependency"
```

### Dangling Dependency Detection

Detects references to non-existent modules:

```go
modules := []parser.Module{
    {Name: "auth", Dependencies: []string{"nonexistent"}},
}

issues := validator.ValidateModules(modules)
// Issues will include "dependency-not-found"
```

## Service Validation

### Port Conflict Detection

Detects multiple services using the same port:

```go
services := []parser.Service{
    {Name: "api", Port: 8080},
    {Name: "web", Port: 8080},
}

issues := validator.ValidateServices(services)
// Issues will include "service-port-conflict"
```

### Port Range Validation

Ensures ports are within valid range (1–65535):

```go
services := []parser.Service{
    {Name: "bad", Port: 99999},
}

issues := validator.ValidateServices(services)
// Issues will include "service-port-range"
```

## Schema Version Validation

### Automatic Check on Parse

The parser automatically validates the schema version:

```yaml
# This spec will be checked against minimum version 0.1.0
version: 0.3.0
```

### Manual Check

```go
result := parser.CheckSpecVersion("0.3.0")
if !result.Valid {
    fmt.Println(result.Message)
    // "Schema version 0.3.0 is not supported. Minimum: 0.1.0"
}
```

## Policy Validation

### Default Rules

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

### Custom Rules

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

## Validation Result

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

## Pipeline Integration

Validation runs automatically as part of the pipeline:

```bash
# Full pipeline includes validation
naeos run --input spec.yaml

# Validation-only mode
naeos validate --input-file spec.yaml
```

### Skip Specific Validation

```bash
# Skip module validation
naeos run --input spec.yaml --skip-validation

# Skip specific checks
naeos validate --input-file spec.yaml --skip circular-deps
```

See also: [Governance](/docs/governance/), [Pipeline Engine](/docs/pipeline-engine/), [Spec Language](/docs/spec-language/)
