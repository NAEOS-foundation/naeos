# NAEOS-PRO-004: Profile Best Practices

## Status
- Status: Draft
- Version: 1.0
- Owner: NAEOS Foundation
- Last Updated: 2026-07-10

---

## 1. Executive Summary

Dokumentasi ini berisi best practices untuk merancang, mengimplementasikan, dan mengelola Profile di NAEOS. Mengikuti praktik terbaik ini akan memastikan profile Anda scalable, maintainable, dan efektif.

---

## 2. Design Principles

### 2.1 Principle 1: Single Responsibility

Setiap profile harus memiliki satu tanggung jawab yang jelas.

**❌ BAD**: Profile yang mencakup enterprise + fintech + microservices + saas

```yaml
profile:
  id: megaprofile  # Too many concerns
  inherits: []
  policies:
    - enterprise-governance
    - fintech-compliance
    - microservices-patterns
    - saas-requirements
```

**✅ GOOD**: Separate profiles dengan satu tanggung jawab

```yaml
# profile-enterprise.yaml
profile:
  id: enterprise
  inherits: [base]
  policies: [enterprise-governance]

# profile-fintech.yaml
profile:
  id: fintech
  inherits: [enterprise]
  policies: [fintech-compliance]

# Combine dalam specification
profiles: [enterprise, fintech, microservices, saas-product]
```

### 2.2 Principle 2: Clear Hierarchy

Maintain strict hierarchy untuk memudahkan reasoning dan resolution.

```
BASE (universal)
  ↓
ORGANIZATION (enterprise, startup, agency)
  ↓
INDUSTRY (fintech, healthcare, e-commerce)
  ↓
ARCHITECTURE (monolith, microservices, serverless)
  ↓
PROJECT-TYPE (api, frontend, mobile, library)
  ↓
PROJECT (specific project)
```

### 2.3 Principle 3: Explicit Over Implicit

Selalu explicit dalam policies. Jangan mengandalkan defaults yang tidak jelas.

**❌ BAD**: Vague requirement

```yaml
policies:
  quality_gates:
    - id: gate-quality
      name: "Good Quality"
      description: "Code should be good quality"
```

**✅ GOOD**: Explicit dan measurable

```yaml
policies:
  quality_gates:
    - id: gate-test-coverage
      name: "Minimum 80% Test Coverage"
      description: "All code must have 80% test coverage measured by coverage.py"
      threshold: 80
      metric: coverage_percent
      tool: coverage.py
      enforced: true
```

### 2.4 Principle 4: Composition Over Inheritance

Prefer composition (combine multiple profiles) over deep inheritance chains.

**❌ BAD**: Deep inheritance (hard to understand)

```
Base
  ↓ inherits
Enterprise
  ↓ inherits
Fintech
  ↓ inherits
Payments
  ↓ inherits
PaymentGateway
```

**✅ GOOD**: Composition (clear intent)

```yaml
specification:
  profiles:
    - base              # Universal standards
    - enterprise        # Organization governance
    - fintech           # Industry compliance
    - payment-processing # Project-specific domain
    - kubernetes        # Infrastructure
```

---

## 3. Policy Design Best Practices

### 3.1 Policy Versioning

Version policies consistently untuk track changes dan backward compatibility.

```yaml
policies:
  standards:
    - id: std-go-fmt
      name: "Go Format Standard"
      description: "All Go code must pass gofmt"
      version: "1.0.0"        # ← Version the policy
      introduced: "2026-01-01"
      deprecated: null
      tool: gofmt
      enforced: true
```

### 3.2 Policy Documentation

Setiap policy harus didokumentasikan dengan jelas.

```yaml
policies:
  security:
    - id: sec-encryption-tls13
      name: "TLS 1.3 Minimum"
      description: "All connections must use TLS 1.3"
      
      # WHY is this required?
      rationale: |
        TLS 1.0-1.2 are deprecated. TLS 1.3 provides better security
        and performance. Required for PCI DSS compliance.
      
      # WHAT should implementers do?
      implementation_guide: |
        1. Update web server configuration
        2. Run security scan with: security-scan.sh
        3. Test with: tls-test.sh
      
      # WHERE can I learn more?
      link: "https://wiki.company.com/tls-1.3-migration"
      
      # WHO is responsible?
      owner: security-team
      
      version: "1.0.0"
      enforced: true
```

### 3.3 Policy with Options

Gunakan options untuk policies yang memiliki variasi.

```yaml
policies:
  standards:
    - id: std-license
      name: "Project License"
      description: "Every project must have a license"
      
      # ← Multiple options allowed
      options:
        - MIT
        - Apache-2.0
        - GPL-3.0
        - AGPL-3.0
      
      # ← Default if not specified
      default: Apache-2.0
      
      enforced: true
```

### 3.4 Policy Exceptions

Allow exceptions dengan dokumentasi yang jelas.

```yaml
policies:
  testing:
    - id: test-coverage-80
      name: "80% Test Coverage"
      threshold: 80
      enforced: true
      
      # ← Exceptions allowed for specific cases
      exceptions:
        - condition: "integration_test"
          threshold: 60  # Lower threshold for integration
          reason: "Integration tests are complex, 60% acceptable"
        
        - condition: "legacy_code"
          threshold: 40  # Much lower for legacy
          reason: "Legacy code approved for maintenance only"
          expires: "2027-12-31"  # Set expiration date
```

---

## 4. Conflict Resolution Best Practices

### 4.1 Document Conflicts

Jika ada kemungkinan konflik, dokumentasikan dengan jelas.

```yaml
# profile-enterprise.yaml
profile:
  id: enterprise
  policies:
    quality_gates:
      - id: coverage
        threshold: 80

# profile-startup.yaml
profile:
  id: startup
  inherits: [base]
  policies:
    quality_gates:
      - id: coverage
        threshold: 40  # Different from enterprise
        comment: |
          Startup allows lower coverage for speed.
          If combined with enterprise, enterprise wins (higher priority).
```

### 4.2 Use Clear Merge Strategies

Tentukan merge strategy untuk setiap kombinasi profile.

```yaml
# specification.yaml
profiles:
  # Order matters! Profile di belakang override yang depan
  - base
  - enterprise
  - fintech  # More strict than enterprise
  
# Merge strategy:
# - base: foundation (60% coverage)
# - enterprise: override ke 80% coverage
# - fintech: override ke 90% coverage
# Result: 90% coverage (fintech wins)
```

### 4.3 Conflict Resolution Report

Selalu request report untuk melihat conflict resolutions.

```bash
naeos profile compile \
  --specification specification.yaml \
  --report-conflicts \
  --output report.json
```

---

## 5. Profile Organization Best Practices

### 5.1 Directory Structure

Organize profiles dengan clear directory structure.

```
profiles/
├── base/
│   └── naeos-profile.yaml          # BASE profile
├── organization/
│   ├── enterprise/
│   │   └── naeos-profile.yaml      # ENTERPRISE profile
│   └── startup/
│       └── naeos-profile.yaml      # STARTUP profile
├── industry/
│   ├── fintech/
│   │   └── naeos-profile.yaml      # FINTECH profile
│   ├── healthcare/
│   │   └── naeos-profile.yaml      # HEALTHCARE profile
│   └── ecommerce/
│       └── naeos-profile.yaml      # ECOMMERCE profile
├── architecture/
│   ├── microservices/
│   │   └── naeos-profile.yaml
│   └── serverless/
│       └── naeos-profile.yaml
└── project/
    └── payment-gateway/
        └── naeos-profile.yaml
```

### 5.2 Version Control

Treat profiles sebagai code - version dalam git.

```bash
# Good commit message
git commit -m "feat(profile): add PCI-DSS compliance policies"

# Track profile changes
git log --oneline profiles/fintech/

# Tag releases
git tag profile-fintech-v1.0.0
```

### 5.3 Profile Versioning Strategy

Use semantic versioning consistently.

```yaml
profile:
  id: fintech
  version: 1.5.2  # MAJOR.MINOR.PATCH
  
  # MAJOR: Breaking changes (e.g., remove required policy)
  # MINOR: New features (e.g., add new optional policy)
  # PATCH: Bug fixes (e.g., fix policy configuration)
```

### 5.4 Breaking Change Policies

Manage breaking changes carefully.

```yaml
profile:
  id: fintech
  version: 2.0.0
  
  breaking_changes:
    - change: "TLS 1.3 now mandatory"
      version_affected: "1.9.0"
      version_enforced: "2.0.0"
      migration_guide: "https://wiki.company.com/tls-migration"
      deprecation_period: "6 months"
```

---

## 6. Policy Module Best Practices

### 6.1 Create Reusable Modules

Buat policy modules yang dapat digunakan ulang.

```yaml
# modules/compliance-pci.yaml
module:
  id: compliance-pci
  name: "PCI DSS Compliance Module"
  description: "Collection of PCI DSS compliance policies"
  version: 1.0.0
  
  policies:
    security:
      - id: sec-encryption-tls13
        name: "TLS 1.3"
      - id: sec-authentication-mfa
        name: "Multi-Factor Authentication"
      - id: sec-audit-logging
        name: "Audit Logging"
    
    quality_gates:
      - id: gate-security-scan
        name: "Security Scan"
```

### 6.2 Module Dependencies

Track dependencies antara modules.

```yaml
module:
  id: compliance-pci
  depends_on:
    - compliance-base  # Must include base compliance first
    - security-audit   # Requires audit capabilities
  
  conflicts_with:
    - compliance-hipaa  # Cannot combine (conflicting requirements)
```

### 6.3 Module Versioning

Track module versions dan compatibility.

```yaml
module:
  id: compliance-pci
  version: 1.5.0
  
  compatibility:
    min_profile_version: "1.0.0"
    max_profile_version: "2.0.0"
    
  requires:
    - module: compliance-base
      version: ">=1.0.0"
    - module: security-base
      version: ">=1.0.0"
```

---

## 7. Testing Best Practices

### 7.1 Profile Validation Testing

Test profile untuk memastikan valid dan konsisten.

```bash
# Validate syntax
naeos profile validate profiles/fintech/naeos-profile.yaml

# Validate inheritance
naeos profile validate --check-inheritance profiles/fintech/naeos-profile.yaml

# Validate conflicts
naeos profile validate --check-conflicts \
  profiles/enterprise/naeos-profile.yaml \
  profiles/fintech/naeos-profile.yaml

# Full validation with report
naeos profile validate profiles/fintech/naeos-profile.yaml \
  --report full \
  --output report.json
```

### 7.2 Profile Compliance Testing

Test bahwa project mematuhi profile.

```bash
# Check project compliance
naeos profile check-compliance \
  --profile enterprise \
  --project ./my-project

# Generate compliance report
naeos profile compliance-report \
  --profile enterprise \
  --project ./my-project \
  --format pdf \
  --output compliance-report.pdf
```

### 7.3 Profile Evolution Testing

Test sebelum update profile.

```bash
# Test profile upgrade
naeos profile test-upgrade \
  --old-version profiles/fintech/v1.9.0/naeos-profile.yaml \
  --new-version profiles/fintech/v2.0.0/naeos-profile.yaml \
  --test-projects ./test-projects/

# Check impact
naeos profile upgrade-impact \
  --old profiles/fintech/v1.9.0/naeos-profile.yaml \
  --new profiles/fintech/v2.0.0/naeos-profile.yaml \
  --affected-projects all
```

---

## 8. Maintenance Best Practices

### 8.1 Profile Review Process

Establish review process untuk profile changes.

```
1. Create RFC untuk significant changes
2. Code review oleh architecture team
3. Security review jika ada policy keamanan baru
4. Test dengan test projects
5. Approve dan merge
6. Tag release
7. Document changelog
```

### 8.2 Profile Deprecation

Deprecate policies yang sudah usang dengan grace period.

```yaml
policies:
  standards:
    - id: std-go-1.16
      name: "Go 1.16"
      deprecated: true
      deprecation_date: "2026-01-01"
      removal_date: "2026-12-31"
      replacement: "std-go-1.20"
      migration_guide: "https://wiki.company.com/go-upgrade"
```

### 8.3 Changelog Maintenance

Maintain detailed changelog untuk profile.

```markdown
# CHANGELOG - fintech Profile

## [2.0.0] - 2026-07-10
### Added
- New PCI DSS 4.0 compliance policies
- Quantum-safe encryption requirement

### Changed
- TLS 1.3 now mandatory (was optional)
- Test coverage threshold increased to 85% (was 80%)

### Deprecated
- Old OAuth1 authentication policy
- Legacy XML configuration support

### Breaking Changes
- Removed support for TLS 1.2
- Removed Python 2 compatibility

### Migration
See: https://wiki.company.com/fintech-v2-migration
```

---

## 9. Governance Best Practices

### 9.1 Profile Ownership

Assign clear ownership untuk setiap profile.

```yaml
profile:
  id: fintech
  owner: fintech-platform-team@company.com
  
  maintainers:
    - name: John Doe
      email: john@company.com
      role: Security Lead
    - name: Jane Smith
      email: jane@company.com
      role: Compliance Lead
  
  stakeholders:
    - fintech-business-team@company.com
    - security-team@company.com
    - compliance-team@company.com
```

### 9.2 Approval Process

Require approvals untuk profile changes.

```yaml
profile:
  id: fintech
  approval_process:
    required_approvals: 2
    approvers:
      - role: architecture-lead
      - role: security-lead
    
    # Different approvals untuk different types of changes
    for_breaking_changes:
      required_approvals: 3
      additional_approvers:
        - role: compliance-lead
```

### 9.3 Change Tracking

Track semua changes ke profile.

```bash
# View change history
naeos profile history fintech

# View who changed what
naeos profile blame profiles/fintech/naeos-profile.yaml

# Generate change report
naeos profile changes \
  --profile fintech \
  --from v1.0.0 \
  --to v2.0.0 \
  --format markdown \
  --output changes.md
```

---

## 10. Common Pitfalls to Avoid

### 10.1 ❌ Over-Specification

Jangan specify details yang terlalu kecil.

```yaml
# BAD: Too specific
policies:
  standards:
    - id: std-indent-spaces
      indent_size: 2  # Too specific, breaks flexibility
```

```yaml
# GOOD: Focus on outcome
policies:
  standards:
    - id: std-code-formatting
      tool: prettier  # Tool enforces specifics
      enforced: true
```

### 10.2 ❌ Circular Dependencies

Avoid circular dependencies antara profiles.

```yaml
# BAD: profile-a inherits from profile-b, profile-b inherits from profile-a
profile-a:
  inherits: [profile-b]

profile-b:
  inherits: [profile-a]  # Circular!
```

### 10.3 ❌ Too Many Profiles

Jangan combine terlalu banyak profiles.

```yaml
# BAD: Too many profiles
profiles:
  - base
  - enterprise
  - fintech
  - microservices
  - kubernetes
  - aws
  - gcp
  - ci-github
  - ci-gitlab
  # ... 20 more profiles!
```

```yaml
# GOOD: Combine strategically
profiles:
  - base
  - enterprise
  - fintech
  - microservices
  # Additional profiles as needed
```

### 10.4 ❌ Ambiguous Policies

Avoid ambiguous atau conflicting policies.

```yaml
# BAD: Ambiguous
policies:
  testing:
    - id: test-sufficient
      description: "Test coverage should be sufficient"  # Vague!
```

```yaml
# GOOD: Clear and measurable
policies:
  testing:
    - id: test-coverage-minimum
      description: "Minimum 80% test coverage"
      threshold: 80
      metric: coverage_percent
```

---

## 11. Real-World Examples

### 11.1 Scenario: Enterprise SaaS FinTech

```yaml
# specification.yaml
profiles:
  - base              # Universal standards
  - enterprise        # Organization governance
  - fintech           # Industry compliance (PCI DSS, GDPR)
  - saas-product      # Product requirements (SLA, monitoring)
  - microservices     # Architecture (service discovery, messaging)

# Result: 
# - Strict security (fintech)
# - High availability (saas-product)
# - Flexible architecture (microservices)
# - Enterprise governance (enterprise)
```

### 11.2 Scenario: Early-Stage Startup

```yaml
# specification.yaml
profiles:
  - base              # Universal standards
  - startup           # Flexible, fast-moving
  - cloud-native      # Leverage cloud

# Result:
# - Lower overhead (startup)
# - Fast iteration (startup)
# - Cloud benefits (cloud-native)
```

---

## 12. References

- [NAEOS-PRO-001.md](NAEOS-PRO-001.md) - Profile System Specification
- [NAEOS-PRO-002.md](NAEOS-PRO-002.md) - Profile Implementation & Setup
- [NAEOS-PRO-003.md](NAEOS-PRO-003.md) - Profile Examples
- [NAEOS-PRO-005.md](NAEOS-PRO-005.md) - Profile API & CLI Reference
