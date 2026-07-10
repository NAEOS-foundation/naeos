# NAEOS-POL-005: Policy Compiler & Engine Deep Dive

## Status
- Status: Draft
- Version: 1.0
- Owner: NAEOS Foundation
- Last Updated: 2026-07-10

---

## 1. Executive Summary

Deep technical dive into NAEOS Policy Compiler - mesin yang mengubah declarative policies menjadi executable rules dan mengevaluasi compliance.

Mencakup:
- Architecture & components
- Policy graph construction
- Rule compilation & optimization
- Dependency resolution
- Conflict detection & resolution

---

## 2. Architecture Overview

### 2.1 High-Level Architecture

```
Constitution
     ↓
Governance
     ↓
Profiles
     ↓
Policies ──→ Policy Parser
              ↓
         Policy Normalizer
              ↓
         Policy Resolver
              ↓
         Policy Graph Builder
              ↓
         Dependency Analyzer
              ↓
         Conflict Detector
              ↓
         Rule Compiler
              ↓
         Executable Policy Graph
              ↓
         Policy Engine
              ↓
         Evaluation Results
```

### 2.2 Components

| Component | Responsibility |
|-----------|-----------------|
| **Parser** | Parse policy YAML/JSON into AST |
| **Normalizer** | Normalize policy to canonical form |
| **Resolver** | Resolve references between policies |
| **Graph Builder** | Build policy dependency graph |
| **Analyzer** | Analyze dependencies & relationships |
| **Conflict Detector** | Detect policy conflicts |
| **Compiler** | Compile to executable rules |
| **Engine** | Evaluate rules against targets |

---

## 3. Policy Graph

### 3.1 Graph Structure

Policy graph adalah DAG (Directed Acyclic Graph) dengan nodes = policies, edges = relationships.

```
        sec-base
        /  |  \
       /   |   \
      /    |    \
  sec-auth sec-enc sec-audit
    / \       |       |
   /   \      |       |
 mfa  oauth2  tls13   logs
```

### 3.2 Graph Relationships

```go
type PolicyNode struct {
    ID          string
    Name        string
    Category    string
    Priority    int
    
    // Relationships
    Inherits    []string     // id of parent policies
    Requires    []string     // hard dependencies
    Suggests    []string     // soft recommendations
    Conflicts   []string     // incompatible policies
    Overrides   []string     // policies this overrides
}

type PolicyEdge struct {
    From        string
    To          string
    Type        string  // "inherits", "requires", "suggests", "conflicts", "overrides"
    Weight      int     // for priority
    Condition   string  // conditional relationship
}
```

### 3.3 Building the Graph

```
1. Load all policies
2. Create node for each policy
3. Parse relationships
4. Create edges based on relationships
5. Validate no cycles
6. Order nodes by dependency
```

```go
func BuildPolicyGraph(policies []Policy) (*Graph, error) {
    g := NewGraph()
    
    // 1. Add all nodes
    for _, p := range policies {
        g.AddNode(p.ID, p)
    }
    
    // 2. Add edges
    for _, p := range policies {
        for _, dep := range p.Requires {
            g.AddEdge(p.ID, dep, "requires")
        }
        for _, parent := range p.Inherits {
            g.AddEdge(p.ID, parent, "inherits")
        }
        // ... etc for other relationships
    }
    
    // 3. Validate DAG (no cycles)
    if err := g.ValidateDAG(); err != nil {
        return nil, err
    }
    
    // 4. Topological sort
    g.TopologicalSort()
    
    return g, nil
}
```

---

## 4. Policy Resolution

### 4.1 Resolution Algorithm

Resolve all policy references to concrete rules:

```
Input: Set of policy IDs + configuration
Process:
  1. Load each policy from registry
  2. Trace inheritance chain
  3. Merge inherited rules
  4. Apply profile-level overrides
  5. Apply project-level overrides
  6. Detect conflicts
  7. Apply resolution rules
  8. Output: Resolved policy set
```

### 4.2 Inheritance Resolution

```yaml
# base-policy.yaml
policy:
  id: base-security
  rules:
    tls_version: "1.2"           # Base default
    encryption: "AES-128"
    audit_logging: true

# enterprise-policy.yaml
policy:
  id: enterprise-security
  inherits_from: base-security
  overrides:
    tls_version: "1.3"           # Override base

# fintech-policy.yaml
policy:
  id: fintech-security
  inherits_from: enterprise-security
  overrides:
    encryption: "AES-256"        # Override enterprise
```

**Resolution process:**

```
fintech-security
  ├── inherits enterprise-security
  │     ├── inherits base-security
  │     │     ├── tls_version: "1.2"
  │     │     ├── encryption: "AES-128"
  │     │     └── audit_logging: true
  │     └── overrides tls_version: "1.3"
  └── overrides encryption: "AES-256"

Result (fintech-security resolved):
  ├── tls_version: "1.3"        (from enterprise override)
  ├── encryption: "AES-256"      (from fintech override)
  └── audit_logging: true        (inherited from base)
```

### 3.3 Conflict Resolution

**Conflict types:**

1. **Type Conflict**: Different value types
   ```yaml
   policy1:
     rule: tls_version: "1.3"    # string
   policy2:
     rule: tls_version: 1.3      # number
   ```

2. **Value Conflict**: Incompatible values
   ```yaml
   policy1:
     rule: tls_version: "1.3"
   policy2:
     rule: tls_version: "1.2"    # Conflict!
   ```

3. **Dependency Conflict**: Circular dependencies
   ```yaml
   policy1:
     requires: policy2
   policy2:
     requires: policy1            # Circular!
   ```

**Resolution strategies:**

```
Priority:  Project > Team > Organization > Base
Method:    Last-wins (later override earlier)
Action:    If unresolvable, error with report
```

```go
func ResolveConflicts(policies []Policy) (Policy, []Conflict, error) {
    var conflicts []Conflict
    result := NewMergedPolicy()
    
    for _, p := range policies {
        for key, value := range p.Rules {
            if existing, ok := result.Rules[key]; ok {
                if !Equals(existing, value) {
                    conflicts = append(conflicts, Conflict{
                        Policy1: result.ID,
                        Policy2: p.ID,
                        Key: key,
                        Value1: existing,
                        Value2: value,
                        Resolution: ApplyPriority(p),
                    })
                    // Apply resolution
                    result.Rules[key] = value
                }
            } else {
                result.Rules[key] = value
            }
        }
    }
    
    return result, conflicts, nil
}
```

---

## 5. Rule Compilation

### 5.1 Compilation Process

Convert policy to executable rules:

```
Input:  Resolved Policy
        ↓
Step 1: Validate policy structure
        ↓
Step 2: Expand templates (if any)
        ↓
Step 3: Optimize rules
        ↓
Step 4: Generate evaluation code
        ↓
Step 5: Generate remediation code
        ↓
Output: Executable Rules
```

### 5.2 Example: Compiling TLS Policy

```yaml
# Input policy
policy:
  id: sec-tls-13
  rule:
    tls_minimum_version: "1.3"
    tls_ciphers:
      - TLS_AES_256_GCM_SHA384
      - TLS_CHACHA20_POLY1305_SHA256

# Compiled to:
rule:
  id: sec-tls-13-check-version
  target: network_connection
  condition: |
    tlsVersion < "1.3"
  action: |
    fail("TLS version must be 1.3 or higher, got: " + tlsVersion)
  remediation: |
    updateServerConfig(tls_version: "1.3")
  
rule:
  id: sec-tls-13-check-ciphers
  target: network_connection
  condition: |
    cipherSuite not in [
      "TLS_AES_256_GCM_SHA384",
      "TLS_CHACHA20_POLY1305_SHA256"
    ]
  action: |
    fail("Cipher suite not in allowed list: " + cipherSuite)
  remediation: |
    updateServerConfig(ciphers: allowedCiphers)
```

### 5.3 Compilation to Go Code

```go
// Generated code from policy
package policies

type RuleChecker interface {
    Check(ctx Context) bool
}

// sec-tls-13 compiled
type SecTls13Checker struct{}

func (c SecTls13Checker) Check(ctx Context) bool {
    conn := ctx.NetworkConnection
    
    // Check TLS version
    if conn.TLSVersion < "1.3" {
        ctx.Fail("TLS version must be 1.3 or higher")
        return false
    }
    
    // Check cipher suites
    allowedCiphers := map[string]bool{
        "TLS_AES_256_GCM_SHA384": true,
        "TLS_CHACHA20_POLY1305_SHA256": true,
    }
    
    if !allowedCiphers[conn.CipherSuite] {
        ctx.Fail("Cipher suite not allowed: " + conn.CipherSuite)
        return false
    }
    
    ctx.Pass("TLS configuration valid")
    return true
}
```

---

## 6. Optimization

### 6.1 Rule Optimization

Optimize compiled rules untuk performance:

```
Before optimization:
  - 1000 rules
  - Avg evaluation time: 500ms
  - Memory: 50MB

After optimization:
  - 800 rules (200 merged)
  - Avg evaluation time: 150ms
  - Memory: 30MB
```

### 6.2 Optimization Techniques

1. **Rule Merging**: Combine compatible rules

   ```yaml
   Before:
     - rule: tls_version >= 1.3
     - rule: cipher_in [AES_256, CHACHA20]
   
   After:
     - rule: (tls_version >= 1.3) AND (cipher_in [AES_256, CHACHA20])
   ```

2. **Dead Code Elimination**: Remove unreachable rules

   ```yaml
   Before:
     - rule: if (false) then fail()     # Dead code
     - rule: normal_check()
   
   After:
     - rule: normal_check()
   ```

3. **Constant Folding**: Precompute constants

   ```yaml
   Before:
     - rule: length(password) >= (8 + 2) * 1
   
   After:
     - rule: length(password) >= 10
   ```

4. **Caching**: Cache evaluation results

   ```go
   type CachedRule struct {
       Inner Rule
       Cache map[string]bool
   }
   
   func (r CachedRule) Check(ctx Context) bool {
       key := ctx.Hash()
       if result, ok := r.Cache[key]; ok {
           return result
       }
       result := r.Inner.Check(ctx)
       r.Cache[key] = result
       return result
   }
   ```

---

## 7. Policy Engine

### 7.1 Engine Architecture

```go
type PolicyEngine struct {
    rules        []Rule
    evaluators   map[string]Evaluator
    cache        *Cache
    audit        *AuditLog
}

func (e *PolicyEngine) Evaluate(ctx Context) Result {
    result := NewResult()
    
    for _, rule := range e.rules {
        evaluator := e.evaluators[rule.Type]
        ruleResult := evaluator.Evaluate(ctx, rule)
        
        result.Add(ruleResult)
        e.audit.Log(ruleResult)
        
        if ruleResult.Severity == Critical && !ruleResult.Pass {
            return result  // Short-circuit on critical failure
        }
    }
    
    return result
}
```

### 7.2 Evaluation Methods

```go
// Method 1: Direct evaluation
result := engine.Evaluate(context)

// Method 2: Streaming evaluation (for large contexts)
for event := range engine.EvaluateStream(context) {
    handleEvent(event)
}

// Method 3: Parallel evaluation
results := engine.EvaluateParallel(contexts, numWorkers)

// Method 4: Conditional evaluation (short-circuit)
result := engine.EvaluateShortCircuit(context)  // Stop on first failure
```

---

## 8. Execution Modes

### 8.1 Blocking Mode

Policy violation blocks action:

```go
type BlockingEvaluator struct {
    rules []Rule
}

func (e BlockingEvaluator) Evaluate(ctx Context) Result {
    for _, rule := range e.rules {
        if !rule.Check(ctx) {
            return Result{
                Status: BLOCKED,
                Reason: rule.ID + ": " + rule.Message,
            }
        }
    }
    return Result{Status: ALLOWED}
}
```

### 8.2 Reporting Mode

Policy violations logged but don't block:

```go
type ReportingEvaluator struct {
    rules []Rule
    log   Logger
}

func (e ReportingEvaluator) Evaluate(ctx Context) Result {
    violations := []string{}
    
    for _, rule := range e.rules {
        if !rule.Check(ctx) {
            violations = append(violations, rule.ID)
            e.log.Warn(rule.ID + ": " + rule.Message)
        }
    }
    
    return Result{
        Status: REPORTED,
        Violations: violations,
    }
}
```

### 8.3 Remediation Mode

Auto-fix violations:

```go
type RemediationEvaluator struct {
    rules []Rule
}

func (e RemediationEvaluator) Evaluate(ctx Context) Result {
    for _, rule := range e.rules {
        if !rule.Check(ctx) {
            if rule.Remediation != nil {
                rule.Remediation(ctx)
                // Re-check
                if !rule.Check(ctx) {
                    return Result{
                        Status: REMEDIATION_FAILED,
                        Rule: rule.ID,
                    }
                }
            }
        }
    }
    return Result{Status: SUCCESS}
}
```

---

## 9. Audit & Logging

### 9.1 Audit Trail

```go
type AuditEntry struct {
    Timestamp    time.Time
    PolicyID     string
    RuleID       string
    Target       string
    Result       bool
    Reason       string
    User         string
    Environment  string
}

// Audit log entry structure
{
    "timestamp": "2026-07-10T14:32:00Z",
    "policy": "sec-tls-13",
    "rule": "sec-tls-13-check-version",
    "target": "api-gateway:443",
    "result": "PASS",
    "tls_version": "1.3",
    "evaluated_by": "ci-pipeline",
    "duration_ms": 45
}
```

### 9.2 Compliance Report

Generate compliance report dari audit trail:

```json
{
  "policy": "sec-tls-13",
  "compliance_percentage": 92.5,
  "total_checks": 120,
  "passed": 111,
  "failed": 9,
  "errors": 0,
  
  "failures": [
    {
      "target": "legacy-payment-api:443",
      "reason": "TLS 1.2 in use",
      "last_checked": "2026-07-10T14:30:00Z",
      "remediation": "Update server configuration"
    }
  ],
  
  "trend": {
    "day": 92.5,
    "week": 85.2,
    "month": 78.1
  }
}
```

---

## 10. Performance Considerations

### 10.1 Benchmarking

```bash
# Benchmark policy evaluation
go test -bench=BenchmarkPolicyEvaluation -benchmem

# Output:
# BenchmarkPolicyEvaluation-8
#   1000000   1234 ns/op   512 B/op   8 allocs/op
```

### 10.2 Optimization Tips

1. **Use caching** for repeated evaluations
2. **Parallelize** independent rule checks
3. **Short-circuit** on critical failures
4. **Lazy-load** policies on demand
5. **Pre-compile** policies offline
6. **Index** rules by type for faster lookup

---

## 11. References

- [NAEOS-POL-001.md](NAEOS-POL-001.md) - Policy Compiler Overview
- [NAEOS-POL-002.md](NAEOS-POL-002.md) - Policy Definition
- [NAEOS-POL-003.md](NAEOS-POL-003.md) - Policy Examples
- [docs/NES-012-Policy.md](docs/NES-012-Policy.md) - Policy Model
