# NAEOS-POL-003: Policy Examples & Templates

## Status
- Status: Draft
- Version: 1.0
- Owner: NAEOS Foundation
- Last Updated: 2026-07-10

---

## 1. Executive Summary

Dokumentasi ini menyediakan contoh konkret dan template untuk berbagai jenis policy di NAEOS:
- **Security Policies** - Policies untuk keamanan sistem
- **Testing Policies** - Policies untuk quality assurance
- **Code Quality Policies** - Policies untuk code standards
- **Documentation Policies** - Policies untuk documentation
- **Deployment Policies** - Policies untuk deployment & operations
- **Compliance Policies** - Policies untuk regulatory compliance

Setiap contoh dapat digunakan langsung atau sebagai template untuk custom policies.

---

## 2. Security Policies

### 2.1 No Secrets in Code

```yaml
policy:
  id: sec-no-secrets
  name: "No Secrets in Code"
  short_name: "SEC-NO-SEC"
  category: security
  type: mandatory
  priority: critical
  version: "1.0.0"
  status: active
  
  description: |
    Secrets (passwords, API keys, tokens) must never be committed to code.
    This policy prevents accidental exposure of sensitive credentials.
  
  rationale: |
    Leaked secrets can be used to gain unauthorized access.
    Prevention is easier than remediation after exposure.
    Tools can detect secrets automatically.
  
  owner: security-team@company.com
  
  conditions:
    target:
      type: code_repository
      all_branches: true
    scope:
      level: organization
  
  rule:
    secret_detection:
      patterns:
        - "password"
        - "api_key"
        - "token"
        - "secret"
        - "private_key"
      must_not_contain: true
  
  evaluation:
    method: automated
    tools:
      - name: truffleHog
        url: "https://github.com/trufflesecurity/trufflehog"
        command: "trufflehog git https://github.com/org/repo"
      - name: detect-secrets
        url: "https://github.com/Yelp/detect-secrets"
        command: "detect-secrets scan ."
  
  on_violation:
    action: "block"
    severity: "critical"
    prevents: "commit"
    message: "Secrets detected. Remove and try again."
  
  tags: ["security", "secrets", "code"]
```

### 2.2 TLS 1.3 Minimum

```yaml
policy:
  id: sec-tls-13
  name: "TLS 1.3 or Higher"
  category: security
  type: mandatory
  priority: critical
  version: "2.0.0"
  
  description: "All network connections must use TLS 1.3 or higher."
  
  conditions:
    target:
      type: network_connection
      environments: [production, staging]
    temporal:
      active_from: "2026-01-01"
  
  rule:
    tls_configuration:
      protocol: TLS
      minimum_version: "1.3"
      cipher_suites:
        allowed:
          - TLS_AES_256_GCM_SHA384
          - TLS_CHACHA20_POLY1305_SHA256
          - TLS_AES_128_CCM_SHA256
  
  evaluation:
    method: automated
    tools:
      - name: ssl-test
        command: "testssl.sh --full {host}"
      - name: nmap
        command: "nmap --script ssl-enum-ciphers -p 443 {host}"
    baseline:
      success_criteria: "All connections use TLS >= 1.3"
  
  on_violation:
    action: "block"
    severity: "critical"
    prevents: "deployment"
  
  exceptions:
    allowed:
      - id: legacy-system
        services: [legacy-payment-api]
        valid_until: "2027-12-31"
        approved_by: "ciso"
```

### 2.3 Multi-Factor Authentication

```yaml
policy:
  id: sec-mfa-required
  name: "Multi-Factor Authentication Required"
  category: security
  type: mandatory
  priority: high
  
  description: "All admin and sensitive access must require MFA."
  
  conditions:
    target:
      type: admin_access
      environments: [production]
  
  rule:
    authentication:
      mfa_required: true
      mfa_methods:
        - "totp"  # Time-based one-time password
        - "webauthn"
        - "hardware_key"
      no_sms: true  # SMS not allowed (vulnerable)
  
  evaluation:
    method: manual
    reviewers: [security-team]
  
  on_violation:
    action: "warn"
    severity: "high"
    deadline: "7 days"
```

### 2.4 Dependency Vulnerability Scanning

```yaml
policy:
  id: sec-dependency-scan
  name: "Dependency Vulnerability Scanning"
  category: security
  type: mandatory
  priority: high
  
  description: "All dependencies must be scanned for known vulnerabilities."
  
  rule:
    scanning:
      tools:
        - name: snyk
          config: "snyk test"
        - name: dependabot
          config: "github_dependabot"
        - name: trivy
          config: "trivy scan"
      
      failure_criteria:
        high_severity: "block"           # Block on high
        medium_severity: "warn"          # Warn on medium
        low_severity: "report"           # Report low
  
  evaluation:
    method: automated
    frequency: "on_commit"
  
  on_violation:
    action: "block"
    prevents: "merge"
```

---

## 3. Testing Policies

### 3.1 Minimum Test Coverage

```yaml
policy:
  id: test-coverage-80
  name: "Minimum 80% Test Coverage"
  category: testing
  type: mandatory
  priority: high
  version: "1.0.0"
  
  description: "All code must have minimum 80% test coverage."
  
  conditions:
    target:
      type: source_code
      all_commits: true
    scope:
      level: organization
  
  rule:
    coverage:
      metric: "line_coverage"
      threshold: 80
      measured_by: "coverage.py"
      
      exceptions:
        - condition: "test_code"
          exclusion: true
        - condition: "third_party"
          threshold: 0
        - condition: "legacy_code"
          threshold: 40
          valid_until: "2027-12-31"
  
  evaluation:
    method: automated
    tools:
      - name: coverage.py
        command: "coverage run -m pytest && coverage report"
      - name: codecov
        url: "https://codecov.io"
    baseline:
      success_criteria: "coverage >= 80%"
  
  on_violation:
    action: "block"
    severity: "high"
    prevents: "merge"
    message: "Coverage must be >= 80%"
```

### 3.2 Integration Tests Required

```yaml
policy:
  id: test-integration
  name: "Integration Tests Required"
  category: testing
  type: mandatory
  priority: medium
  
  description: "All service-to-service integrations must have integration tests."
  
  conditions:
    target:
      type: service_integration
  
  rule:
    testing:
      integration_tests_required: true
      test_coverage: "all_interfaces"
      test_scenarios:
        - "happy_path"
        - "error_cases"
        - "edge_cases"
  
  evaluation:
    method: manual
    review_checklist:
      - "Integration tests exist"
      - "Tests cover happy path"
      - "Tests cover error scenarios"
      - "Tests pass consistently"
```

### 3.3 Security Testing (SAST/DAST)

```yaml
policy:
  id: test-security
  name: "Security Testing Required"
  category: testing
  type: mandatory
  priority: critical
  
  description: |
    All code must pass SAST (Static) and DAST (Dynamic) security testing.
  
  rule:
    security_testing:
      sast_required: true
      sast_tools:
        - sonarqube
        - codeql
        - semgrep
      
      dast_required: true
      dast_tools:
        - burp_suite
        - owasp_zap
      
      failure_on:
        - "high_severity"
        - "critical_severity"
  
  evaluation:
    method: automated
    tools:
      - name: sonarqube
        command: "sonar-scanner -Dsonar.projectKey=my-project"
      - name: codeql
        command: "codeql database create --language=go"
```

---

## 4. Code Quality Policies

### 4.1 Code Formatting

```yaml
policy:
  id: code-format
  name: "Code Formatting Standards"
  category: code_quality
  type: mandatory
  priority: medium
  
  description: "All code must follow language formatting standards."
  
  rule:
    formatting:
      go:
        formatter: gofmt
        linter: golangci-lint
      
      python:
        formatter: black
        linter: pylint
      
      javascript:
        formatter: prettier
        linter: eslint
  
  evaluation:
    method: automated
    frequency: "on_commit"
```

### 4.2 Code Complexity

```yaml
policy:
  id: code-complexity
  name: "Code Complexity Limits"
  category: code_quality
  type: mandatory
  priority: medium
  
  description: "Code complexity must stay within acceptable limits."
  
  rule:
    complexity_metrics:
      cyclomatic_complexity: 10          # Max 10
      cognitive_complexity: 15           # Max 15
      nesting_depth: 4                   # Max 4 levels
      function_length: 50                # Max 50 lines
      method_length: 30                  # Max 30 lines
  
  evaluation:
    method: automated
    tools:
      - name: sonarqube
        metric: "complexity"
      - name: radon
        command: "radon cc -s"
```

### 4.3 Code Review

```yaml
policy:
  id: code-review
  name: "Code Review Required"
  category: code_quality
  type: mandatory
  priority: high
  
  description: "All code changes must be reviewed before merge."
  
  rule:
    review_requirements:
      minimum_approvals: 1
      for_critical_files:
        minimum_approvals: 2
        security_review_required: true
      
      reviewers:
        must_have_different_author: true
        can_be_requested_from: "team-leads"
  
  evaluation:
    method: manual
    reviewers: "team-specific"
  
  on_violation:
    action: "block"
    prevents: "merge"
    message: "Requires at least 1 approval"
```

---

## 5. Documentation Policies

### 5.1 README Required

```yaml
policy:
  id: doc-readme
  name: "README Documentation Required"
  category: documentation
  type: mandatory
  priority: medium
  
  description: "Every project must have comprehensive README.md"
  
  rule:
    documentation:
      files_required:
        - README.md
      
      required_sections:
        - "Overview/Description"
        - "Installation/Setup"
        - "Usage/Getting Started"
        - "API Documentation"
        - "Contributing"
        - "License"
  
  evaluation:
    method: automated
    tools:
      - name: markdown-checker
        command: "check-readme-sections README.md"
  
  on_violation:
    action: "warn"
    severity: "medium"
```

### 5.2 API Documentation

```yaml
policy:
  id: doc-api
  name: "API Documentation Required"
  category: documentation
  type: mandatory
  priority: high
  
  description: "All public APIs must be documented in OpenAPI/Swagger format."
  
  rule:
    api_documentation:
      format: "openapi_3.0"
      required_for:
        - "all_rest_apis"
        - "all_grpc_services"
      
      must_include:
        - "endpoints"
        - "request_schemas"
        - "response_schemas"
        - "error_responses"
        - "authentication"
        - "rate_limits"
  
  evaluation:
    method: automated
    tools:
      - name: openapi-validator
        command: "swagger-cli validate openapi.yaml"
```

### 5.3 Architecture Decision Records

```yaml
policy:
  id: doc-adr
  name: "Architecture Decision Records"
  category: documentation
  type: recommended
  priority: medium
  
  description: "Significant architectural decisions should be documented as ADR."
  
  rule:
    adr_requirements:
      required_for:
        - "major_architecture_changes"
        - "technology_decisions"
        - "breaking_changes"
      
      adr_template:
        - "Status"
        - "Context"
        - "Decision"
        - "Consequences"
        - "Alternatives Considered"
  
  tags: ["documentation", "architecture"]
```

---

## 6. Deployment Policies

### 6.1 Deployment Approval

```yaml
policy:
  id: deploy-approval
  name: "Production Deployment Approval"
  category: deployment
  type: mandatory
  priority: critical
  
  description: "All production deployments must be approved."
  
  conditions:
    target:
      type: deployment
      environments: [production]
  
  rule:
    approval_workflow:
      requires_approval: true
      approval_levels: 2
      approvers:
        - role: "release-manager"
        - role: "vp-engineering"
      
      approval_criteria:
        - "Code reviewed and merged"
        - "All tests passing"
        - "Security scan passed"
        - "Change log updated"
  
  on_violation:
    action: "block"
    prevents: "deployment"
```

### 6.2 Health Check

```yaml
policy:
  id: deploy-health-check
  name: "Health Check Endpoints Required"
  category: deployment
  type: mandatory
  priority: high
  
  description: "All services must implement health check endpoints."
  
  rule:
    health_checks:
      endpoints_required:
        - path: "/health"
          response_time_ms: 100
        - path: "/health/deep"
          includes: ["database", "cache", "external_services"]
      
      response_format:
        status: "ok|degraded|unhealthy"
        timestamp: "ISO8601"
        version: "service_version"
```

### 6.3 Canary Deployment

```yaml
policy:
  id: deploy-canary
  name: "Canary Deployments for Production"
  category: deployment
  type: recommended
  priority: high
  
  description: "Use canary deployments for safer rollouts."
  
  rule:
    canary_strategy:
      percentage_initial: 5               # Start with 5%
      percentage_step: 10                 # Increase by 10%
      duration_per_step: "10 minutes"
      
      monitoring:
        error_rate_threshold: 1            # Stop if error rate > 1%
        latency_threshold_ms: 500          # Stop if latency > 500ms
```

---

## 7. Compliance Policies

### 7.1 GDPR Compliance

```yaml
policy:
  id: comp-gdpr
  name: "GDPR Compliance"
  category: compliance
  type: mandatory
  priority: critical
  
  description: "All systems must comply with GDPR regulations."
  
  rule:
    data_protection:
      personal_data:
        encryption_required: true
        encryption_at_rest: "AES-256"
        encryption_in_transit: "TLS1.3"
      
      data_retention:
        policy: "delete_after_period"
        default_retention_days: 2555    # ~7 years
      
      consent:
        explicit_consent_required: true
        consent_logging: true
      
      right_to_be_forgotten:
        implementation: "required"
        deletion_deadline_days: 30
  
  conditions:
    scope:
      applies_to: ["EU residents", "EU data"]
```

### 7.2 PCI DSS Compliance

```yaml
policy:
  id: comp-pci-dss
  name: "PCI DSS 3.2.1 Compliance"
  category: compliance
  type: mandatory
  priority: critical
  
  description: |
    Payment Card Industry Data Security Standard compliance required
    for all systems handling credit card data.
  
  rule:
    pci_requirements:
      encryption:
        tls_version: "1.2+"
        key_length: 2048
      
      network_security:
        firewall_required: true
        intrusion_detection: true
      
      access_control:
        principle_of_least_privilege: true
        unique_user_ids: true
        mfa_for_admin: true
      
      monitoring:
        audit_logging: true
        log_retention_days: 365
        vulnerability_scanning: "quarterly"
        penetration_testing: "annual"
```

---

## 8. Custom Policy Template

```yaml
# Save as policy-template.yaml and customize

policy:
  # === IDENTIFIERS ===
  id: your-policy-id
  name: "Your Policy Name"
  short_name: "SHORT"
  
  # === CLASSIFICATION ===
  category: security              # security, testing, documentation, etc
  type: mandatory                 # mandatory, recommended, optional
  priority: high                  # critical, high, medium, low
  
  # === LIFECYCLE ===
  version: "1.0.0"
  status: active
  created: "2026-07-10"
  
  # === OWNERSHIP ===
  owner: your-team@company.com
  
  # === DESCRIPTION ===
  description: |
    [One sentence summary of what this policy requires]
  
  rationale: |
    [Why is this policy needed?]
  
  # === CONDITIONS ===
  conditions:
    target:
      type: "[artifact_type]"
      environments: [production]
    
    scope:
      level: organization
  
  # === RULE ===
  rule:
    # Define the actual requirement here
    key_config:
      parameter: "value"
  
  # === EVALUATION ===
  evaluation:
    method: automated              # automated, manual, hybrid
    tools:
      - name: "[tool_name]"
        command: "[command]"
  
  # === ACTIONS ===
  on_violation:
    action: block                  # block, warn, remediate, report
    severity: high
    prevents: "[what?]"
  
  # === TAGS ===
  tags: ["tag1", "tag2"]
```

---

## 9. Combining Policies

Example: Policy composition for fintech payment service

```yaml
# payment-gateway-policies.yaml

policies:
  # Base security
  - sec-no-secrets
  - sec-tls-13
  - sec-mfa-required
  - sec-dependency-scan
  
  # Testing
  - test-coverage-80
  - test-integration
  - test-security
  
  # Code quality
  - code-format
  - code-review
  
  # Documentation
  - doc-readme
  - doc-api
  
  # Deployment
  - deploy-approval
  - deploy-health-check
  - deploy-canary
  
  # Compliance
  - comp-pci-dss
  - comp-gdpr
```

---

## 10. References

- [NAEOS-POL-001.md](NAEOS-POL-001.md) - Policy Compiler
- [NAEOS-POL-002.md](NAEOS-POL-002.md) - Policy Definition & Authoring
- [NAEOS-POL-004.md](NAEOS-POL-004.md) - Policy Best Practices
- [NAEOS-POL-005.md](NAEOS-POL-005.md) - Policy Compiler & Engine
- [docs/NES-012-Policy.md](docs/NES-012-Policy.md) - Policy Model Specification
