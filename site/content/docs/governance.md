---
title: Governance
description: Policy system, review workflows, and audit trails for NAEOS.
weight: 18
---

NAEOS provides a comprehensive governance system to enforce policies, validate artifacts, and maintain audit trails across your engineering workflow.

## Governance Structure

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

## Document Hierarchy

NAEOS governance follows a strict document hierarchy:

| Document | ID | Description |
|----------|----|-------------|
| Engineering Constitution | NAEOS-CON-001 | Foundational engineering rules |
| Project Charter | NAEOS-GOV-001 | Project scope and objectives |
| Vision | NAEOS-GOV-002 | Long-term direction |
| Mission | NAEOS-GOV-003 | How we achieve the vision |
| Manifesto | NAEOS-GOV-004 | Core beliefs and values |
| Core Principles | NAEOS-GOV-005 | Engineering principles |
| Governance Model | NAEOS-GOV-006 | Decision-making framework |
| Roadmap | NAEOS-GOV-007 | Development milestones |
| Versioning Policy | NAEOS-GOV-008 | SemVer rules and conventions |

## Policy System

### Default Rules

NAEOS ships with 5 default policy rules:

| Rule ID | Condition | Priority | Action |
|---------|-----------|----------|--------|
| `project-required` | `exists:project` | 1 | block |
| `modules-required` | `exists:modules` | 1 | block |
| `architecture-pattern-valid` | `in:architecture.pattern,...` | 2 | warn |
| `deployment-strategy-valid` | `in:deployment.strategy,...` | 2 | warn |
| `service-port-positive` | `gt:service.port,0` | 3 | warn |

### Condition Operators

| Operator | Example | Description |
|----------|---------|-------------|
| `exists` | `exists:project` | Check if a key exists |
| `not_empty` | `not_empty:project` | Check if a key is not empty |
| `contains` | `contains:project,app` | Check if value contains a substring |
| `gt` | `gt:port,0` | Greater than |
| `lt` | `lt:port,65535` | Less than |
| `in` | `in:kind,http,grpc` | Value is in the allowed list |

### Custom Rules

Define custom rules in your Go code:

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

## Review System

The review system checks generated artifacts against quality rules:

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

Every significant action is logged for traceability:

```bash
# Run audit on a specification
naeos audit --input-file spec.yaml

# Export audit log
naeos audit --input-file spec.yaml --output audit-report.json
```

The audit trail includes:
- Specification changes (create, update, migrate)
- Pipeline executions (start, complete, fail)
- Policy evaluations (pass, block, warn)
- Artifact generations (create, review, export)

## CI/CD Integration

Enforce governance in your CI pipeline:

```yaml
# .github/workflows/ci.yml
- name: Validate Spec
  run: naeos validate --input-file spec.yaml

- name: Lint Spec
  run: naeos lint --input-file spec.yaml

- name: Audit
  run: naeos audit --input-file spec.yaml
```

## Pipeline Integration

Enable governance in your pipeline configuration:

```go
cfg := pipeline.Config{
    Policies: policy.DefaultRules(),
    Review:   review.Config{Enabled: true},
    Audit:    audit.Config{Enabled: true},
}
```

See also: [Core Principles](/docs/core-principles/), [Validation](/docs/validation/)
