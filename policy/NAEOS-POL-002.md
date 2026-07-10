# NAEOS-POL-002: Policy Definition & Authoring Guide

## Status
- Status: Draft
- Version: 1.0
- Owner: NAEOS Foundation
- Last Updated: 2026-07-10

---

## 1. Executive Summary

Panduan ini menjelaskan cara menulis, mendefinisikan, dan mengorganisir policy di NAEOS. Policy adalah aturan yang dapat dieksekusi untuk mengarahkan perilaku sistem, generators, planners, dan AI agents.

Dokumentasi ini ditujukan untuk:
- **Policy Authors** yang menulis policy baru
- **Engineering Leads** yang merancang kebijakan organisasi
- **Compliance Officers** yang memastikan adherence
- **Platform Teams** yang mengimplementasikan policy engine

---

## 2. Konsep Dasar

### 2.1 Apa itu Policy?

Policy adalah **declarative rule** yang:

| Aspek | Deskripsi |
|-------|-----------|
| **Definisi** | Aturan yang menentukan apa yang harus/tidak boleh dilakukan |
| **Format** | Deklaratif (bukan imperatif) - fokus pada "apa" bukan "bagaimana" |
| **Eksekusi** | Dapat dievaluasi dan dipaksakan secara otomatis |
| **Scope** | Berlaku untuk sistem, components, agents, atau artifacts tertentu |
| **Auditability** | Setiap policy decision tercatat dan dapat ditelusuri |

**Contoh policy:**

```yaml
policy:
  id: sec-tls-minimum
  name: "Minimum TLS Version"
  description: "All connections must use TLS 1.3 or higher"
  
  category: security
  
  # Kondisi kapan policy berlaku
  conditions:
    - target: network_connection
    - environment: production
  
  # Aturan yang harus dipenuhi
  rule:
    tls_version:
      minimum: "1.3"
  
  # Action jika rule dilanggar
  action:
    on_violation: "block"
    severity: "critical"
```

### 2.2 Policy vs Rule vs Constraint

Perbedaan penting:

| Aspek | Policy | Rule | Constraint |
|-------|--------|------|-----------|
| **Level** | High-level intent | Concrete implementation | Technical limit |
| **Flexibility** | Adaptable per context | Fixed | Hardware/system limit |
| **Example** | "Secure data transmission" | "Use TLS 1.3" | "CPU max 8 cores" |

Policy diharapkan dipecah menjadi rules yang concrete.

### 2.3 Policy Graph

Policy tidak berdiri sendiri - mereka membentuk graph dengan relationships.

```
Security Policy
    ├── requires: Authentication Policy
    │     └── requires: Identity Provider
    ├── requires: Encryption Policy
    │     └── suggests: Key Management
    └── conflicts_with: Legacy Insecure Protocol
```

Relationships dalam policy graph:

| Relationship | Meaning |
|--------------|---------|
| **inherits** | Child policy extends parent policy |
| **requires** | This policy depends on another |
| **suggests** | Recommended but not required |
| **conflicts** | Cannot coexist with another policy |
| **overrides** | This policy takes precedence |
| **extends** | Adds functionality to existing policy |

---

## 3. Policy Structure

### 3.1 Complete Policy Document

```yaml
# ============================================================================
# POLICY METADATA
# ============================================================================

policy:
  # Identifiers
  id: sec-encryption-tls13            # Unique identifier
  name: "TLS 1.3 Encryption"          # Human-readable name
  short_name: "SEC-TLS13"             # Short form
  
  # Classification
  category: security                   # Category: security, testing, documentation, etc
  type: mandatory                      # mandatory, recommended, optional
  priority: critical                   # critical, high, medium, low
  
  # Lifecycle
  version: "1.0.0"                    # Semantic versioning
  status: active                       # active, deprecated, archived
  created: "2026-01-15"               # Creation date
  last_modified: "2026-07-10"         # Last update
  
  # Ownership
  owner: security-team@company.com    # Primary owner
  stakeholders:                        # Affected parties
    - platform-team@company.com
    - architecture-team@company.com
  
  # Documentation
  description: |
    All production connections must use TLS 1.3.
    This policy ensures modern encryption standards.
  
  rationale: |
    TLS 1.0-1.2 are deprecated and have known vulnerabilities.
    TLS 1.3 is more secure and performant.
    Required for PCI DSS compliance.
  
  references:
    - https://datatracker.ietf.org/doc/html/rfc8446
    - https://wiki.company.com/security/tls
    - NAEOS-PRO-001.md
  
  tags:
    - compliance
    - encryption
    - production

# ============================================================================
# POLICY CONDITIONS (When does this policy apply?)
# ============================================================================

conditions:
  # Target: what does this policy apply to?
  target:
    type: network_connection         # Target type
    services:                         # Specific services (optional)
      - api-gateway
      - payment-service
    environments:                     # Specific environments
      - production
      - staging
  
  # Scope: level of applicability
  scope:
    level: organization              # global, organization, team, project
    teams:                           # If team level
      - platform-eng
    projects:                        # If project level
      - payment-gateway
  
  # Temporal: time-based conditions
  temporal:
    active_from: "2026-01-15"
    active_until: null               # null = indefinite
    exceptions:
      - date: "2026-12-24"
        reason: "Holiday exception"

# ============================================================================
# POLICY RULE (What exactly should be enforced?)
# ============================================================================

rule:
  # Rule body - defines actual requirement
  tls_configuration:
    protocol: "TLS"
    minimum_version: "1.3"
    maximum_version: null             # null = no maximum
    
    # Allowed cipher suites
    cipher_suites:
      allowed:
        - TLS_AES_256_GCM_SHA384
        - TLS_CHACHA20_POLY1305_SHA256
      forbidden:
        - TLS_RSA_WITH_AES_128_CBC_SHA  # Legacy, insecure
    
    # Certificate requirements
    certificates:
      must_be_valid: true
      must_not_be_self_signed: true
      minimum_key_length: 2048
      maximum_key_length: null

# ============================================================================
# POLICY EVALUATION (How to check if rule is satisfied?)
# ============================================================================

evaluation:
  # Method: how to evaluate this policy
  method: automated                   # automated, manual, hybrid
  
  # Tools/scanners that can evaluate this
  tools:
    - name: ssl-scanner
      url: https://github.com/nmap/nmap
      command: "nmap --script ssl-enum-ciphers -p 443 {host}"
    - name: testssl
      url: https://github.com/drwetter/testssl.sh
      command: "testssl.sh {host}"
  
  # Automated checks
  automated_checks:
    - type: "service_config_scan"
      parameters:
        check_tls_version: true
        check_cipher_suites: true
  
  # Baseline: expected passing criteria
  baseline:
    success_criteria: "TLS version >= 1.3"
    sample_size: "100%"  # Must check all instances
    failure_threshold: 0  # No failures allowed (for critical)

# ============================================================================
# POLICY ACTIONS (What happens if rule is violated?)
# ============================================================================

actions:
  # On violation
  on_violation:
    action: "block"                   # block, warn, remediate, report
    severity: "critical"              # critical, high, medium, low
    
    # Blocking behavior
    if_action_is_block:
      prevents: "deployment"          # Prevents what?
      message: "TLS 1.3 required for production"
    
    # Warning behavior
    if_action_is_warn:
      notification: "platform-team"
      deadline: "7 days"
    
    # Remediation behavior
    if_action_is_remediate:
      auto_remediate: false           # Auto fix? (usually false for security)
      suggested_fix: "Update TLS configuration"
  
  # On compliance
  on_compliance:
    action: "allow"                   # allow, record
    logging: "audit_trail"
  
  # Reporting
  reporting:
    frequency: "daily"                # How often to report
    format: "json"
    recipients:
      - security-team@company.com

# ============================================================================
# POLICY DEPENDENCIES
# ============================================================================

dependencies:
  # Required policies
  requires:
    - policy_id: sec-certificate-management
      reason: "Need certificate management infrastructure"
    - policy_id: sec-key-rotation
      reason: "Regular key rotation required"
  
  # Policies this depends on
  depends_on:
    - component: certificate_authority
      minimum_version: "1.0.0"
    - component: kms_service
      minimum_version: "2.0.0"
  
  # Conflicting policies
  conflicts_with:
    - policy_id: legacy-tls-support
      reason: "Legacy TLS 1.2 contradicts TLS 1.3 requirement"

# ============================================================================
# POLICY EXCEPTIONS & OVERRIDES
# ============================================================================

exceptions:
  # Allowed exceptions
  allowed:
    - id: exception-legacy-system
      description: "Legacy payment system requires TLS 1.2"
      services:
        - legacy-payment-api
      approved_by: "security-lead"
      valid_until: "2027-12-31"
      condition: "Will upgrade system by end of 2027"
  
  # Override capability
  can_be_overridden_by:
    - role: "ciso"
    - role: "vp-security"
  
  override_requires:
    approval_count: 2
    audit_trail: true
    notification: "audit-team"

# ============================================================================
# POLICY RELATIONSHIPS
# ============================================================================

relationships:
  # This policy within hierarchy
  inherits_from:
    - policy_id: sec-base
      percentage: 80                  # 80% of rules from parent
  
  # Extended by
  extended_by:
    - policy_id: sec-high-assurance
      reason: "Adds stricter requirements"
  
  # Related policies (for context)
  related_to:
    - policy_id: sec-certificate-pinning
    - policy_id: sec-mutual-tls
  
  # Part of profile/standard
  part_of:
    - profile_id: enterprise
    - standard_id: pci-dss
    - standard_id: iso-27001

# ============================================================================
# POLICY GOVERNANCE
# ============================================================================

governance:
  # Review process
  review:
    required: true
    frequency: "annually"
    reviewers: ["security-team"]
    last_reviewed: "2026-07-01"
    next_review: "2027-07-01"
  
  # Change management
  change_management:
    approval_required: true
    change_advisory_board: true
    documentation_required: true
  
  # Deprecation
  deprecation:
    status: null                      # Null if active
    # If deprecated:
    # - reason: "Superseded by policy-v2"
    # - removal_date: "2027-01-01"
    # - replacement_policy: "sec-encryption-tls13-v2"

# ============================================================================
# POLICY METADATA FOR PROCESSING
# ============================================================================

metadata:
  # Automation support
  machine_readable: true
  supports_remediation: false
  supports_audit: true
  
  # Integration
  integrations:
    - system: terraform
      support: partial
      notes: "Can validate via terraform checks"
    - system: kubernetes
      support: yes
      resources:
        - networkpolicy
        - tlspolicy
  
  # Compliance mappings
  compliance_mappings:
    pci_dss_v3_2_1:
      - requirement: "3.4"
        description: "Encryption of cardholder data in transit"
    iso_27001:
      - control: "A.10.1.1"
        description: "Cryptographic controls"
    gdpr:
      - article: "32"
        description: "Security of processing"
```

### 3.2 Minimal Policy

Jika tidak perlu detail lengkap:

```yaml
policy:
  id: doc-readme-required
  name: "README Required"
  description: "Every project must have README.md"
  category: documentation
  type: mandatory
  
  rule:
    files:
      must_exist: ["README.md"]
  
  on_violation:
    action: "warn"
    severity: "high"
```

---

## 4. Policy Categories

### 4.1 Security Policies

```yaml
category: security

# Examples
policies:
  - id: sec-no-secrets          # Secrets not in code
  - id: sec-tls-minimum         # TLS version
  - id: sec-mfa-required        # Multi-factor auth
  - id: sec-encryption-rest     # Encryption at rest
  - id: sec-audit-logging       # Audit trails
  - id: sec-dependency-scan     # Dependency scanning
```

### 4.2 Testing Policies

```yaml
category: testing

# Examples
policies:
  - id: test-unit-coverage      # Unit test coverage
  - id: test-integration        # Integration tests
  - id: test-e2e-critical       # End-to-end tests
  - id: test-security           # Security testing
  - id: test-performance        # Performance testing
```

### 4.3 Documentation Policies

```yaml
category: documentation

# Examples
policies:
  - id: doc-readme-required     # README required
  - id: doc-api-documented      # API documentation
  - id: doc-architecture-adr    # ADR required
  - id: doc-changelog           # Changelog required
```

### 4.4 Code Quality Policies

```yaml
category: code_quality

# Examples
policies:
  - id: code-lint-required      # Linting
  - id: code-format-required    # Code formatting
  - id: code-complexity         # Complexity limits
  - id: code-duplication        # Duplication limits
```

### 4.5 Deployment Policies

```yaml
category: deployment

# Examples
policies:
  - id: deploy-approval         # Approval required
  - id: deploy-staging-first    # Stage before prod
  - id: deploy-version-tag      # Version tagging
  - id: deploy-health-check     # Health checks
```

---

## 5. Policy Evaluation Methods

### 5.1 Automated Evaluation

```yaml
evaluation:
  method: automated
  
  tools:
    - name: sonarqube
      type: "sast"
      check: "code-complexity"
    
    - name: coverage.py
      type: "coverage"
      check: "test-coverage >= 80"
    
    - name: trivy
      type: "dependency-scan"
      check: "no-high-severity-vulnerabilities"
```

### 5.2 Manual Evaluation

```yaml
evaluation:
  method: manual
  
  reviewers:
    - role: "architect"
    - role: "security-team"
  
  review_checklist:
    - "Design meets requirements"
    - "Security implications reviewed"
    - "Performance impact assessed"
  
  review_timeline: "5 business days"
```

### 5.3 Hybrid Evaluation

```yaml
evaluation:
  method: hybrid
  
  # First automated
  automated_checks:
    - tool: sonarqube
      must_pass: true
  
  # Then manual if automated passes
  if_automated_passes:
    manual_review_required: true
    reviewers: ["security-team"]
```

---

## 6. Policy Authoring Best Practices

### 6.1 Clarity & Precision

✅ GOOD: Clear and specific
```yaml
policy:
  name: "Minimum Test Coverage 80%"
  rule:
    coverage_threshold: 80
    metric: "line_coverage"
    measured_by: "coverage.py"
```

❌ BAD: Vague
```yaml
policy:
  name: "Good Testing"
  rule:
    coverage: "sufficient"  # What is "sufficient"?
```

### 6.2 Measurability

✅ GOOD: Measurable
```yaml
rule:
  tls_minimum_version: "1.3"
  test_coverage: ">= 80%"
  deployment_time: "<= 5 minutes"
```

❌ BAD: Not measurable
```yaml
rule:
  security: "good"
  performance: "fast"
```

### 6.3 Scope Definition

Always define scope clearly:

```yaml
conditions:
  target:
    type: "service"
    environment: "production"
    services:
      - api-gateway
      - payment-service
```

### 6.4 Action Clarity

Be explicit about actions:

```yaml
on_violation:
  action: "block"            # Not just "fail"
  severity: "critical"
  prevents: "deployment"     # What is prevented?
```

---

## 7. Policy Versioning

### 7.1 Semantic Versioning

```yaml
policy:
  version: "1.2.3"  # MAJOR.MINOR.PATCH
```

- **MAJOR**: Breaking changes (e.g., stricter requirement)
- **MINOR**: New features (e.g., additional checks)
- **PATCH**: Bugfixes (e.g., fix tool configuration)

### 7.2 Deprecation

```yaml
policy:
  version: "1.0.0"
  status: deprecated
  
  deprecation:
    reason: "Superseded by sec-tls-v2"
    removal_date: "2027-01-01"
    replacement_policy: "sec-tls-v2"
    migration_guide: "https://wiki.company.com/tls-migration"
```

---

## 8. Common Policy Patterns

### Pattern 1: Threshold Policy

Policy with measurable threshold:

```yaml
rule:
  coverage_threshold: 80
  metric: coverage_percent
  tool: coverage.py
  
evaluation:
  baseline:
    success_criteria: "coverage >= 80%"
```

### Pattern 2: Enforcement Policy

Policy that blocks/allows actions:

```yaml
rule:
  requirement: "approval_required"
  approvers_needed: 2
  
on_violation:
  action: "block"
  prevents: "merge_to_main"
```

### Pattern 3: Configuration Policy

Policy that enforces specific configuration:

```yaml
rule:
  configuration:
    key: tls_version
    value: "1.3"
    location: "/etc/tls/config.yaml"
    
evaluation:
  method: config_verification
```

### Pattern 4: Scanning Policy

Policy that runs security scans:

```yaml
rule:
  scans:
    - name: "dependency_scan"
      tool: "trivy"
      fail_on: "high_severity"
    
    - name: "sast"
      tool: "sonarqube"
      fail_on: "blocker_issues"
```

---

## 9. Policy Documentation Template

```markdown
# [POLICY NAME]

## 1. Summary
[Brief description of what policy requires]

## 2. Motivation
[Why is this policy needed?]

## 3. Requirements
- Requirement 1
- Requirement 2
- Requirement 3

## 4. Scope
- Applies to: [services/systems/teams]
- Environment: [production/all/staging]
- Exceptions: [if any]

## 5. Evaluation
- Method: [automated/manual/hybrid]
- Tools: [tool names]
- Success criteria: [clear criteria]

## 6. Timeline
- Effective from: [date]
- Deadline for compliance: [date]

## 7. References
- Link to related policies
- External standards
- Implementation guides

## 8. Contact
- Owner: [name/team]
- Questions: [contact]
```

---

## 10. CLI for Policy Authoring

```bash
# Create new policy from template
naeos policy create \
  --id my-security-policy \
  --template security \
  --output policy.yaml

# Validate policy syntax
naeos policy validate policy.yaml

# Check policy completeness
naeos policy validate policy.yaml --check-completeness

# Preview how policy will work
naeos policy preview policy.yaml --target-service payment-api

# Publish policy
naeos policy publish policy.yaml --registry https://registry.naeos.io
```

---

## 11. References

- [NAEOS-POL-001.md](NAEOS-POL-001.md) - Policy Compiler
- [NAEOS-POL-003.md](NAEOS-POL-003.md) - Policy Examples & Templates
- [NAEOS-POL-004.md](NAEOS-POL-004.md) - Policy Best Practices
- [NAEOS-POL-005.md](NAEOS-POL-005.md) - Policy Compiler & Engine
- [docs/NES-012-Policy.md](docs/NES-012-Policy.md) - Policy Model Specification
