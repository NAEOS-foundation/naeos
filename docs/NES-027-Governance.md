# NAEOS Governance Documentation

## Status
- Status: Stable
- Version: 1.0.0
- Owner: NAEOS Foundation
- Last Updated: 2026-07-10

---

## 1. Overview

NAEOS Governance menyediakan mekanisme policy evaluation dan artifact review untuk memastikan kualitas dan kepatuhan seluruh artefak yang dihasilkan oleh pipeline.

---

## 2. Package Structure

```
internal/governance/
├── policy/
│   ├── evaluator.go       # Policy rule evaluation
│   └── evaluator_test.go  # Tests
└── review/
    ├── reviewer.go        # Artifact review
    └── reviewer_test.go   # Tests
```

---

## 3. Policy Evaluator

### 3.1 Types

```go
type Rule struct {
    RuleID    string
    Condition string
    Priority  int
    Action    string
    Scope     string
    Enabled   bool
}

type EvaluationResult struct {
    Passed    bool
    RuleID    string
    Message   string
    Action    string
    Priority  int
}
```

### 3.2 Interface

```go
type Evaluator interface {
    Evaluate(ctx map[string]any) error
    EvaluateRules(rules []Rule, ctx map[string]any) ([]EvaluationResult, error)
}
```

### 3.3 Condition Operators

| Operator | Syntax | Description |
|----------|--------|-------------|
| `exists` | `exists:key` | Check if key exists in context |
| `not_empty` | `not_empty:key` | Check if key value is not empty |
| `contains` | `contains:key,substr` | Check if value contains substring |
| `gt` | `gt:key,value` | Greater than comparison |
| `lt` | `lt:key,value` | Less than comparison |
| `in` | `in:key,v1,v2,...` | Check if value is in allowed list |
| `equals` | `key:expected` | Exact equality match (default) |

### 3.4 Default Rules

```go
func DefaultRules() []Rule {
    return []Rule{
        {RuleID: "project-required", Condition: "exists:project", Priority: 1, Action: "block"},
        {RuleID: "modules-required", Condition: "exists:modules", Priority: 1, Action: "block"},
        {RuleID: "architecture-pattern-valid", Condition: "in:architecture_pattern,hexagonal,layered,clean,event-driven,cqrs,microkernel,monolith,monolithic,microservices,serverless", Priority: 2, Action: "warn"},
        {RuleID: "deployment-strategy-valid", Condition: "in:deployment_strategy,rolling,blue-green,canary,recreate", Priority: 2, Action: "warn"},
        {RuleID: "service-port-positive", Condition: "gt:service_port,0", Priority: 3, Action: "warn"},
    }
}
```

### 3.5 Example Usage

```go
e := policy.NewEvaluator()

rules := []policy.Rule{
    {RuleID: "require-project", Condition: "exists:project", Priority: 1, Action: "block", Enabled: true},
    {RuleID: "valid-env", Condition: "in:env,staging,production", Priority: 2, Action: "warn", Enabled: true},
}

ctx := map[string]any{
    "project": "my-api",
    "env":     "production",
}

results, err := e.EvaluateRules(rules, ctx)
for _, r := range results {
    if !r.Passed {
        fmt.Printf("Rule %s failed: %s\n", r.RuleID, r.Message)
    }
}
```

---

## 4. Artifact Reviewer

### 4.1 Types

```go
type ReviewStatus string

const (
    StatusApproved  ReviewStatus = "approved"
    StatusRejected  ReviewStatus = "rejected"
    StatusPending   ReviewStatus = "pending"
    StatusChanges   ReviewStatus = "changes_requested"
)

type ReviewComment struct {
    RuleID  string
    Message string
}

type ReviewResult struct {
    Status   ReviewStatus
    Comments []ReviewComment
    Summary  string
}
```

### 4.2 Interface

```go
type Reviewer interface {
    Review(input any) error
    ReviewArtifact(name, content string, rules []string) (*ReviewResult, error)
}
```

### 4.3 Review Rules

| Rule ID | Description | Severity |
|---------|-------------|----------|
| `no-todo` | Checks for TODO comments in code | changes_requested |
| `no-placeholder` | Checks for placeholder text (TODO, FIXME, XXX, PLACEHOLDER, CHANGEME, REPLACE_ME) | changes_requested |
| `has-package-declaration` | Validates Go files have package declaration | rejected |
| `has-license-header` | Checks for license header (License, Apache, MIT) | changes_requested |

### 4.4 Example Usage

```go
r := review.NewReviewer()

// Review a Go file
result, err := r.ReviewArtifact(
    "internal/auth/handler.go",
    `package auth

type Handler struct{}

func (h *Handler) Handle() {
    // TODO: implement
}
`,
    []string{"no-todo", "no-placeholder", "has-package-declaration"},
)

fmt.Printf("Status: %s\n", result.Status)
for _, c := range result.Comments {
    fmt.Printf("  - [%s] %s\n", c.RuleID, c.Message)
}
```

---

## 5. Integration with Pipeline

```mermaid
sequenceDiagram
    participant Pipeline
    participant PE as Policy Evaluator
    participant AR as Artifact Reviewer

    Pipeline->>PE: Evaluate NEIR with policy rules
    PE->>PE: Check each rule against context
    PE-->>Pipeline: Return EvaluationResults
    Pipeline->>AR: Review generated artifacts
    AR->>AR: Apply review rules per file
    AR-->>Pipeline: Return ReviewResults
    Pipeline->>Pipeline: Aggregate results and decide pass/fail
```

Governance components are automatically integrated into the pipeline:

```go
p, _ := pipeline.New(pipeline.Config{
    Policies: []policy.Rule{
        {RuleID: "require-auth", Condition: "exists:modules", Priority: 1, Action: "block", Enabled: true},
    },
})

result, err := p.Run(spec)

// Check review results
for _, review := range result.Reviews {
    if review.Status != review.StatusApproved {
        fmt.Printf("Review issues for %s:\n", review.Summary)
        for _, c := range review.Comments {
            fmt.Printf("  - %s\n", c.Message)
        }
    }
}
```

---

## 6. Custom Rules

### 6.1 Creating Custom Policy Rules

```go
rules := []policy.Rule{
    {
        RuleID:    "max-modules",
        Condition: "lt:modules,10",
        Priority:  1,
        Action:    "warn",
        Scope:     "spec",
        Enabled:   true,
    },
    {
        RuleID:    "require-description",
        Condition: "not_empty:description",
        Priority:  2,
        Action:    "warn",
        Scope:     "spec",
        Enabled:   true,
    },
}
```

### 6.2 Creating Custom Review Rules

To add custom review rules, extend the `ReviewArtifact` function in `reviewer.go`:

```go
case "no-hardcoded-secrets":
    if containsHardcodedSecret(content) {
        result.Status = StatusRejected
        result.Comments = append(result.Comments, ReviewComment{
            RuleID:  rule,
            Message: fmt.Sprintf("file %s contains hardcoded secrets", name),
        })
    }
```

---

## 7. Review Status Flow

```
                    ┌─────────────┐
                    │   Pending   │
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │   Review    │
                    └──────┬──────┘
                           │
          ┌────────────────┼────────────────┐
          │                │                │
   ┌──────▼──────┐  ┌──────▼──────┐  ┌──────▼──────┐
   │  Approved   │  │  Changes    │  │  Rejected   │
   └─────────────┘  │  Requested  │  └─────────────┘
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │   Fix &     │
                    │   Resubmit  │
                    └─────────────┘
```

---

## 8. References

- [NES-012-Policy.md](NES-012-Policy.md) - Policy Specification
- [NES-026-Pipeline.md](NES-026-Pipeline.md) - Pipeline Documentation
- [NAEOS-POL-001.md](../policy/NAEOS-POL-001.md) - Policy System
