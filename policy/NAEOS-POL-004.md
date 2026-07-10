# NAEOS-POL-004: Policy Best Practices

## Status
- Status: Draft
- Version: 1.0
- Owner: NAEOS Foundation
- Last Updated: 2026-07-10

---

## 1. Executive Summary

Best practices untuk design, implementation, dan management policy di NAEOS. Mengikuti praktik ini akan memastikan policy Anda efektif, maintainable, dan scalable.

---

## 2. Design Principles

### 2.1 Principle 1: Single Responsibility

Setiap policy harus memiliki satu purpose yang jelas.

**❌ BAD**: Policy yang mencakup terlalu banyak
```yaml
policy:
  id: mega-security
  name: "Security and Testing and Deployment"
  rule:
    - tls_version: "1.3"
    - test_coverage: 80
    - approval_required: true
    # Too many concerns!
```

**✅ GOOD**: Focused policies
```yaml
# policy-sec-tls.yaml
policy:
  id: sec-tls-13
  name: "TLS 1.3 Minimum"
  rule:
    tls_version: "1.3"

# policy-test-coverage.yaml
policy:
  id: test-coverage-80
  name: "80% Test Coverage"
  rule:
    coverage_threshold: 80

# policy-deploy-approval.yaml
policy:
  id: deploy-approval
  name: "Deployment Approval"
  rule:
    approval_required: true
```

### 2.2 Principle 2: Clarity First

Policies must be crystal clear.

**❌ BAD**: Vague
```yaml
policy:
  rule:
    security: "good"
    testing: "enough"
    code: "quality"
```

**✅ GOOD**: Specific and measurable
```yaml
policy:
  rule:
    tls_version: "1.3"          # Specific version
    coverage_threshold: 80       # Measurable threshold
    code_complexity: "< 10"      # Quantified
```

### 2.3 Principle 3: Actionable Policies

Policy harus dapat dieksekusi.

**❌ BAD**: Not actionable
```yaml
policy:
  rule:
    "systems should be secure"
```

**✅ GOOD**: Clear actions
```yaml
policy:
  rule:
    tls_minimum_version: "1.3"
  
  evaluation:
    tools:
      - testssl.sh
      - nmap
    baseline:
      success_criteria: "TLS version >= 1.3"
  
  on_violation:
    action: "block"
    prevents: "deployment"
```

### 2.4 Principle 4: Documented Rationale

Setiap policy harus explain "mengapa".

```yaml
policy:
  id: sec-tls-13
  
  # Always include rationale
  rationale: |
    TLS 1.0-1.2 have known vulnerabilities.
    TLS 1.3 is more secure and performant.
    Required for PCI DSS and SOC2 compliance.
    Industry standard since 2023.
```

---

## 3. Policy Organization

### 3.1 Directory Structure

```
policy/
├── base/
│   ├── sec-base.yaml            # Base security policies
│   ├── test-base.yaml           # Base testing policies
│   └── code-base.yaml           # Base code quality
├── security/
│   ├── authentication/
│   │   ├── sec-mfa.yaml
│   │   └── sec-oauth2.yaml
│   ├── encryption/
│   │   ├── sec-tls-13.yaml
│   │   └── sec-encryption-rest.yaml
│   └── monitoring/
│       ├── sec-audit-logging.yaml
│       └── sec-vulnerability-scan.yaml
├── testing/
│   ├── test-coverage.yaml
│   ├── test-integration.yaml
│   └── test-security.yaml
├── compliance/
│   ├── comp-gdpr.yaml
│   ├── comp-pci-dss.yaml
│   └── comp-hipaa.yaml
└── deployment/
    ├── deploy-approval.yaml
    ├── deploy-canary.yaml
    └── deploy-health-check.yaml
```

### 3.2 Naming Convention

Follow consistent naming:

```
<category>-<specific>[-<variant>].yaml

Examples:
sec-tls-13              # Security > TLS 1.3
test-coverage           # Testing > Coverage
code-format             # Code Quality > Formatting
comp-pci-dss            # Compliance > PCI DSS
deploy-approval         # Deployment > Approval
```

### 3.3 Version Control

Treat policies as code:

```bash
# Track in git
git add policy/
git commit -m "feat(policy): add TLS 1.3 requirement"

# Tag versions
git tag policy-v1.0.0

# View history
git log --oneline policy/
```

---

## 4. Policy Composition

### 4.1 Building Policy Sets

Combine policies untuk specific context:

```yaml
# security-complete.yaml
policies:
  - sec-tls-13
  - sec-mfa-required
  - sec-no-secrets
  - sec-dependency-scan
  - sec-audit-logging
  - sec-encryption-rest

# fintech-policies.yaml
profiles:
  - security-complete
  - compliance-pci-dss
  - compliance-gdpr
  - test-coverage-80
  - test-security
  - deploy-approval
```

### 4.2 Policy Inheritance

Create policy hierarchies:

```yaml
# sec-base.yaml
policy:
  id: sec-base
  name: "Base Security"
  rules:
    - encryption_required: true
    - audit_logging: true

# sec-high.yaml
policy:
  id: sec-high
  name: "High Security"
  inherits_from: sec-base
  adds:
    - mfa_required: true
    - penetration_testing: true
```

---

## 5. Evaluation Best Practices

### 5.1 Automated Evaluation

Use automated tools when possible:

```yaml
evaluation:
  method: automated
  
  # Define exact tools and commands
  tools:
    - name: sonarqube
      version: "9.0+"
      command: "sonar-scanner"
      config: "sonar-project.properties"
    
    - name: trivy
      version: "0.40+"
      command: "trivy image {image}"
  
  # Clear success criteria
  baseline:
    success_criteria: |
      - No high/critical vulnerabilities
      - Code complexity < 10
      - Coverage > 80%
    measurement: "percentage"
```

### 5.2 Manual Review

For policies needing human judgment:

```yaml
evaluation:
  method: manual
  
  reviewers:
    - role: "architect"
      min_experience: "5 years"
    - role: "security-lead"
  
  review_checklist:
    - "Design aligns with requirements"
    - "Security implications addressed"
    - "Performance impact acceptable"
    - "Documentation complete"
  
  review_timeline: "5 business days"
  
  escalation:
    if_no_consensus: "cto"
```

### 5.3 Audit Trail

Always track policy evaluations:

```yaml
evaluation:
  audit_trail:
    - timestamp: "2026-07-10T14:32:00Z"
      policy: "sec-tls-13"
      result: "pass"
      evaluated_by: "ci-pipeline"
      details: "TLS 1.3 verified for 42 services"
    
    - timestamp: "2026-07-10T14:35:00Z"
      policy: "test-coverage-80"
      result: "fail"
      evaluated_by: "codecov"
      details: "Coverage 76% < 80% threshold"
```

---

## 6. Exception Management

### 6.1 Structured Exceptions

Always document exceptions:

```yaml
policy:
  id: sec-tls-13
  
  exceptions:
    - id: exception-legacy-system
      description: "Legacy payment system requires TLS 1.2"
      
      # What is the exception
      applies_to:
        services: [legacy-payment-api]
        endpoints: [/v1/process]
      
      # Why is it needed
      business_justification: |
        Legacy system vendor doesn't support TLS 1.3.
        Vendor has migration roadmap to TLS 1.3 by 2027.
      
      # Who approved
      approved_by: "ciso"
      approval_date: "2026-01-15"
      
      # When does it expire
      valid_from: "2026-01-15"
      valid_until: "2027-12-31"
      
      # Review plan
      review_date: "2027-09-30"
      mitigation_controls:
        - "Limited network access"
        - "Enhanced monitoring"
        - "Regular scanning"
```

### 6.2 Exception Tracking

Monitor all exceptions:

```bash
# List all active exceptions
naeos policy list-exceptions --active

# Output:
# Policy              Service             Expires       Status
# sec-tls-13          legacy-payment-api  2027-12-31   Active
# test-coverage-80    mobile-app          2027-06-30   Active

# Alert when expiring soon
naeos policy check-expiring-exceptions --days-until 30
```

---

## 7. Policy Versioning

### 7.1 Semantic Versioning

```yaml
policy:
  version: "1.2.3"  # MAJOR.MINOR.PATCH
  
  # PATCH: Bug fixes, clarifications
  # MINOR: New optional features
  # MAJOR: Breaking changes
```

### 7.2 Breaking Changes

Document clearly:

```yaml
policy:
  version: "2.0.0"  # MAJOR bump
  
  changelog:
    breaking_changes:
      - "TLS 1.2 no longer supported (was optional in v1.x)"
      - "Coverage threshold raised from 75% to 80%"
      - "MFA now mandatory (was optional)"
    
    deprecations:
      - "old-format deprecated, use new-format"
    
    migration_guide: |
      1. Update TLS configuration to 1.3
      2. Increase test coverage to 80%
      3. Enable MFA for admins
      See: https://wiki.company.com/policy-v2-migration
```

### 7.3 Deprecation

```yaml
policy:
  id: sec-legacy-protocol
  version: "1.0.0"
  status: deprecated
  
  deprecation:
    reason: "Replaced by sec-modern-protocol"
    removal_date: "2027-01-01"
    replacement: "sec-modern-protocol"
    migration_guide: "https://wiki.company.com/migration"
```

---

## 8. Policy Testing

### 8.1 Test Policy Definitions

```bash
# Validate policy syntax
naeos policy validate policy.yaml

# Check policy completeness
naeos policy validate policy.yaml --check-completeness

# Verify evaluation tools exist
naeos policy validate policy.yaml --check-tools

# Preview policy impact
naeos policy preview policy.yaml \
  --target-service payment-api \
  --show-impact
```

### 8.2 Test in Staging

```bash
# Apply policy to staging first
naeos policy apply policy.yaml --environment staging

# Monitor for issues
naeos policy check-compliance \
  --policy sec-tls-13 \
  --environment staging \
  --monitoring-period "1 week"

# If all good, promote to production
naeos policy promote \
  --policy sec-tls-13 \
  --from staging \
  --to production
```

---

## 9. Policy Communication

### 9.1 Announcement Template

```markdown
# Policy Update: [Policy Name]

## Summary
[One sentence about what's changing]

## Effective Date
[Date when policy takes effect]

## What's Changing
- Change 1
- Change 2
- Change 3

## Why This Change
[Business rationale]

## Your Action Items
- Action 1 (deadline)
- Action 2 (deadline)
- Action 3 (deadline)

## Help & Questions
- Contact: [email]
- FAQ: [link]
- Migration Guide: [link]

## Exceptions
Need an exception? [Apply here]
```

### 9.2 Impact Analysis

Before deploying policy:

```bash
# Analyze impact
naeos policy impact-analysis policy.yaml

# Output:
# Policy: sec-tls-13
# 
# Current State:
#   - 87 services already compliant (90%)
#   - 10 services need updates (10%)
#
# Estimated Effort:
#   - Engineering: 40 hours
#   - Security review: 10 hours
#   - Testing: 15 hours
#   Total: 65 hours
#
# Affected Teams:
#   - Platform: 30 hours
#   - Payment: 25 hours
#   - Auth: 10 hours
#
# Timeline Recommendation:
#   - Phase 1 (Platform): 2 weeks
#   - Phase 2 (Payment): 2 weeks
#   - Phase 3 (Auth): 1 week
```

---

## 10. Common Pitfalls to Avoid

### 10.1 ❌ Too Strict

```yaml
# BAD: Will break everything
policy:
  id: code-coverage
  rule:
    coverage_threshold: 100  # Impossible!
```

```yaml
# GOOD: Reasonable threshold
policy:
  id: code-coverage
  rule:
    coverage_threshold: 80
    exceptions:
      - condition: legacy_code
        threshold: 50
```

### 10.2 ❌ Vague Evaluation

```yaml
# BAD: How to evaluate?
policy:
  id: code-quality
  rule:
    quality: "good"
  evaluation:
    method: manual  # No reviewers!
```

```yaml
# GOOD: Clear evaluation
policy:
  id: code-quality
  rule:
    complexity_threshold: 10
  evaluation:
    method: automated
    tools: [sonarqube]
    baseline:
      success_criteria: "complexity < 10"
```

### 10.3 ❌ No Exceptions

```yaml
# BAD: No exceptions ever
policy:
  id: deployment
  rule:
    approval_required: true
    exceptions: []  # What about emergencies?
```

```yaml
# GOOD: Structured exceptions
policy:
  id: deployment
  rule:
    approval_required: true
  exceptions:
    - condition: "critical_security_fix"
      approval_required: false
      requires: "immediate_notification"
```

---

## 11. Governance Best Practices

### 11.1 Policy Ownership

```yaml
policy:
  id: sec-tls-13
  
  owner: security-team
  owner_email: security-team@company.com
  
  maintainers:
    - name: Alice (Security Lead)
      email: alice@company.com
    - name: Bob (Compliance Officer)
      email: bob@company.com
  
  stakeholders:
    - platform-engineering@company.com
    - architecture-team@company.com
```

### 11.2 Review Process

```yaml
policy:
  governance:
    review:
      required: true
      frequency: "annually"
      reviewers: ["security-team"]
      last_reviewed: "2026-06-01"
      next_review: "2027-06-01"
    
    change_management:
      approval_required: true
      change_advisory_board: true
      documentation_required: true
```

---

## 12. References

- [NAEOS-POL-001.md](NAEOS-POL-001.md) - Policy Compiler
- [NAEOS-POL-002.md](NAEOS-POL-002.md) - Policy Definition & Authoring
- [NAEOS-POL-003.md](NAEOS-POL-003.md) - Policy Examples
- [NAEOS-POL-005.md](NAEOS-POL-005.md) - Policy Compiler & Engine
