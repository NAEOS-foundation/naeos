# NAEOS-PRO-003: Profile Examples

## Status
- Status: Draft
- Version: 1.0
- Owner: NAEOS Foundation
- Last Updated: 2026-07-10

---

## 1. Executive Summary

Dokumentasi ini menyediakan contoh konkret profile untuk berbagai skenario organisasi, industri, dan jenis project. Setiap contoh dapat dijadikan template atau baseline untuk custom profile Anda.

Contoh yang disertakan:
1. **Base Profile** - Foundation untuk semua profile
2. **Startup Profile** - Untuk early-stage organization
3. **Enterprise Profile** - Untuk large organization
4. **Fintech Profile** - Industry-specific (financial technology)
5. **Healthcare Profile** - Industry-specific (healthcare/biotech)
6. **SaaS Product Profile** - Project-type specific
7. **Microservices Profile** - Architecture-specific

---

## 2. BASE Profile

Profile dasar untuk semua organisasi. Berisi standar universal yang minimum.

```yaml
# profiles/base/naeos-profile.yaml

profile:
  id: base
  name: "NAEOS Base Profile"
  version: 1.0.0
  description: "Foundation profile dengan standar minimum universal"
  
  inherits: []
  extends: []
  
  owner: naeos-core-team
  tags:
    - foundation
    - universal
  
  policies:
    
    # === RULES ===
    rules:
      - id: rule-semver
        name: "Semantic Versioning"
        description: "All releases must follow semantic versioning (MAJOR.MINOR.PATCH)"
        enforced: true
        link: "https://semver.org"
      
      - id: rule-commit-message
        name: "Commit Message Convention"
        description: "Commit messages must follow conventional commits format"
        format: "type(scope): subject\n\nbody\n\nfooter"
        enforced: true
      
      - id: rule-file-structure
        name: "Project File Structure"
        description: "Projects must have standard directory structure"
        enforced: true
    
    # === STANDARDS ===
    standards:
      - id: std-readme
        name: "README Required"
        description: "Every project must have a README.md file"
        enforced: true
      
      - id: std-gitignore
        name: "Gitignore"
        description: "Project must have .gitignore file"
        enforced: true
      
      - id: std-license
        name: "License"
        description: "Project must declare license"
        options: ["MIT", "Apache-2.0", "GPL-3.0", "AGPL-3.0"]
        enforced: true
      
      - id: std-changelog
        name: "Changelog"
        description: "Maintain CHANGELOG.md following Keep a Changelog format"
        enforced: false
        link: "https://keepachangelog.com"
    
    # === QUALITY GATES ===
    quality_gates:
      - id: gate-code-review
        name: "Code Review Required"
        description: "All changes must be reviewed before merge"
        enforced: true
        requirement: "minimum 1 approval"
      
      - id: gate-ci-passing
        name: "CI Pipeline Passing"
        description: "All CI checks must pass before merge"
        enforced: true
      
      - id: gate-test-coverage
        name: "Minimum Test Coverage"
        description: "Code must have minimum test coverage"
        threshold: 60
        metric: coverage_percent
        enforced: true
    
    # === SECURITY ===
    security:
      - id: sec-no-secrets
        name: "No Secrets in Code"
        description: "Secrets must not be committed"
        tools: ["truffleHog", "detect-secrets"]
        enforced: true
      
      - id: sec-dependency-scan
        name: "Dependency Security Scan"
        description: "All dependencies must be scanned for vulnerabilities"
        tools: ["dependabot", "snyk"]
        enforced: true
      
      - id: sec-sast
        name: "SAST (Static Analysis)"
        description: "Code must pass static security analysis"
        tools: ["sonarqube", "codeql"]
        enforced: false
    
    # === TESTING ===
    testing:
      - id: test-unit
        name: "Unit Tests Required"
        description: "All packages/modules must have unit tests"
        coverage_minimum: 60
        enforced: true
      
      - id: test-ci-automation
        name: "CI Test Automation"
        description: "Tests must run automatically in CI pipeline"
        enforced: true
    
    # === DOCUMENTATION ===
    documentation:
      - id: doc-readme
        name: "README Documentation"
        description: "README.md must include project description, setup, and usage"
        sections: ["Description", "Setup", "Usage", "Contributing"]
        enforced: true
      
      - id: doc-code-comments
        name: "Code Comments"
        description: "Public APIs must have documentation comments"
        enforced: false
      
      - id: doc-adr
        name: "Architecture Decision Records"
        description: "Significant architectural decisions should be documented as ADR"
        enforced: false
        link: "https://adr.github.io"
    
    # === DEVOPS ===
    devops:
      - id: deploy-version-tag
        name: "Version Tagging"
        description: "Releases must be tagged with semantic version"
        format: "vMAJOR.MINOR.PATCH"
        enforced: true
      
      - id: deploy-reproducible
        name: "Reproducible Builds"
        description: "Build artifacts must be reproducible"
        enforced: false
      
      - id: deploy-release-notes
        name: "Release Notes"
        description: "Each release must have release notes"
        enforced: false
```

---

## 3. STARTUP Profile

Profile untuk early-stage organization/startup. Fokus pada kecepatan dan fleksibilitas.

```yaml
# profiles/startup/naeos-profile.yaml

profile:
  id: startup
  name: "Startup Profile"
  version: 1.0.0
  description: "Profile untuk early-stage organization - fokus pada kecepatan"
  
  inherits:
    - base
  
  owner: startup-team
  tags:
    - organization
    - startup
    - early-stage
  
  policies:
    # More lenient than base
    quality_gates:
      - id: gate-test-coverage-startup
        name: "Minimum Test Coverage"
        threshold: 40  # Lower than base (60%)
        metric: coverage_percent
        enforced: true
      
      - id: gate-ci-passing
        name: "CI Pipeline Passing"
        enforced: true
    
    # Flexible review process
    # (inherits gate-code-review from base but can be relaxed)
    
    testing:
      - id: test-unit-startup
        name: "Unit Tests"
        coverage_minimum: 40
        enforced: true
    
    # Less strict documentation
    documentation:
      - id: doc-readme-startup
        name: "Minimal README"
        sections: ["Description", "Quick Start"]
        enforced: true
```

---

## 4. ENTERPRISE Profile

Profile untuk large organization dengan strict governance dan compliance.

```yaml
# profiles/enterprise/naeos-profile.yaml

profile:
  id: enterprise
  name: "Enterprise Profile"
  version: 1.0.0
  description: "Profile untuk large organization - fokus pada governance dan compliance"
  
  inherits:
    - base
  
  owner: enterprise-architecture-team
  tags:
    - organization
    - enterprise
    - governance
  
  policies:
    
    quality_gates:
      - id: gate-test-coverage-enterprise
        name: "High Test Coverage"
        threshold: 80  # Higher than base
        metric: coverage_percent
        enforced: true
      
      - id: gate-code-review-enterprise
        name: "Code Review (2+ approvals)"
        enforced: true
        requirement: "minimum 2 approvals"
        exclude_owner: true
      
      - id: gate-architecture-review
        name: "Architecture Review"
        description: "Architectural changes need formal review"
        enforced: true
      
      - id: gate-security-approval
        name: "Security Review"
        description: "Security-related changes need security team approval"
        enforced: true
    
    security:
      - id: sec-no-secrets-scan
        name: "Automated Secret Scanning"
        tools: ["truffleHog", "detect-secrets"]
        enforced: true
      
      - id: sec-dependency-scan-enterprise
        name: "Dependency Security Scan"
        enforced: true
        frequency: daily
      
      - id: sec-sast-enterprise
        name: "SAST (Static Analysis)"
        tools: ["sonarqube", "codeql"]
        enforced: true
      
      - id: sec-dast
        name: "DAST (Dynamic Analysis)"
        description: "Dynamic security testing for web services"
        enforced: true
      
      - id: sec-pen-testing
        name: "Penetration Testing"
        description: "Regular penetration testing for critical systems"
        frequency: annual
        enforced: true
    
    testing:
      - id: test-unit-enterprise
        name: "Unit Tests (80%)"
        coverage_minimum: 80
        enforced: true
      
      - id: test-integration
        name: "Integration Tests"
        description: "All integrations must have integration tests"
        enforced: true
      
      - id: test-e2e-critical
        name: "E2E Tests (Critical Paths)"
        description: "All critical user paths must have E2E tests"
        enforced: true
      
      - id: test-regression
        name: "Regression Tests"
        enforced: true
    
    documentation:
      - id: doc-comprehensive
        name: "Comprehensive Documentation"
        sections: ["Overview", "Architecture", "API", "Setup", "Deployment", "Troubleshooting"]
        enforced: true
      
      - id: doc-adr-mandatory
        name: "Architecture Decision Records"
        description: "All significant decisions must be documented as ADR"
        enforced: true
      
      - id: doc-api-openapi
        name: "API Documentation (OpenAPI)"
        description: "All APIs must be documented in OpenAPI/Swagger"
        enforced: true
    
    devops:
      - id: deploy-approval-required
        name: "Deployment Approval"
        description: "Production deployments require formal approval"
        enforced: true
      
      - id: deploy-audit-logging
        name: "Audit Logging"
        description: "All deployments must be logged for audit"
        enforced: true
      
      - id: deploy-rollback-plan
        name: "Rollback Plan"
        description: "Every deployment must have rollback procedure documented"
        enforced: true
      
      - id: deploy-health-check
        name: "Health Check"
        description: "Services must implement health check endpoints"
        enforced: true
```

---

## 5. FINTECH Profile

Profile untuk financial technology industry dengan strict compliance.

```yaml
# profiles/fintech/naeos-profile.yaml

profile:
  id: fintech
  name: "FinTech Profile"
  version: 1.0.0
  description: "Profile untuk fintech industry - fokus pada security dan compliance"
  
  inherits:
    - enterprise
  
  extends:
    - compliance-pci
    - compliance-sox
  
  owner: fintech-security-team
  tags:
    - industry
    - fintech
    - financial
    - regulated
  
  policies:
    
    # === ENHANCED SECURITY ===
    security:
      - id: sec-encryption-tls13
        name: "TLS 1.3 Minimum"
        description: "All connections must use TLS 1.3"
        version: "1.3"
        enforced: true
      
      - id: sec-encryption-at-rest
        name: "Encryption at Rest"
        description: "Sensitive data must be encrypted at rest"
        algorithm: "AES-256"
        enforced: true
      
      - id: sec-mfa
        name: "Multi-Factor Authentication"
        description: "MFA required for all admin access"
        enforced: true
      
      - id: sec-audit-trail
        name: "Comprehensive Audit Trail"
        description: "All transactions and admin actions must be logged"
        retention: "7 years"
        enforced: true
      
      - id: sec-pen-testing-fintech
        name: "Quarterly Penetration Testing"
        frequency: quarterly
        enforced: true
      
      - id: sec-code-signing
        name: "Code Signing"
        description: "All releases must be cryptographically signed"
        enforced: true
    
    # === COMPLIANCE ===
    compliance:
      - id: comp-pci-dss
        name: "PCI DSS Compliance"
        description: "Payment processing must comply with PCI DSS v3.2.1"
        version: "3.2.1"
        enforced: true
      
      - id: comp-gdpr
        name: "GDPR Compliance"
        description: "Customer data handling must comply with GDPR"
        enforced: true
      
      - id: comp-aml-kyc
        name: "AML/KYC"
        description: "Anti-Money Laundering and Know Your Customer policies"
        enforced: true
      
      - id: comp-data-retention
        name: "Data Retention Policy"
        description: "Data must be retained per regulatory requirements"
        retention_min: "7 years"
        enforced: true
    
    testing:
      - id: test-fuzzing
        name: "Fuzzing Tests"
        description: "Critical components must include fuzzing tests"
        enforced: true
      
      - id: test-load
        name: "Load Testing"
        description: "Services must be load tested for peak capacity"
        enforced: true
      
      - id: test-chaos
        name: "Chaos Engineering"
        description: "Resilience must be validated with chaos engineering"
        enforced: false
    
    documentation:
      - id: doc-security-analysis
        name: "Security Analysis Documentation"
        description: "All security measures must be documented"
        enforced: true
      
      - id: doc-compliance-report
        name: "Compliance Report"
        description: "Quarterly compliance audit reports required"
        enforced: true
```

---

## 6. HEALTHCARE Profile

Profile untuk healthcare industry dengan HIPAA compliance.

```yaml
# profiles/healthcare/naeos-profile.yaml

profile:
  id: healthcare
  name: "Healthcare Profile"
  version: 1.0.0
  description: "Profile untuk healthcare industry - fokus pada HIPAA compliance"
  
  inherits:
    - enterprise
  
  extends:
    - compliance-hipaa
    - compliance-hl7
  
  owner: healthcare-compliance-team
  tags:
    - industry
    - healthcare
    - biotech
    - regulated
  
  policies:
    
    security:
      - id: sec-hipaa-encryption
        name: "HIPAA Encryption Standards"
        description: "All PHI (Protected Health Information) must be encrypted"
        algorithm: "AES-256"
        enforced: true
      
      - id: sec-access-control
        name: "Role-Based Access Control"
        description: "RBAC required for all systems handling PHI"
        enforced: true
      
      - id: sec-audit-hipaa
        name: "HIPAA Audit Trail"
        description: "All PHI access must be logged for 6 years"
        retention: "6 years"
        enforced: true
    
    compliance:
      - id: comp-hipaa
        name: "HIPAA Compliance"
        description: "Full HIPAA Privacy and Security Rule compliance"
        enforced: true
      
      - id: comp-hitech
        name: "HITECH Act Compliance"
        description: "HITECH Act requirements for breach notification"
        enforced: true
      
      - id: comp-hl7-fhir
        name: "HL7 FHIR Standards"
        description: "Healthcare data must use HL7 FHIR standards"
        version: "R4"
        enforced: true
    
    testing:
      - id: test-penetration-healthcare
        name: "Penetration Testing"
        frequency: "semi-annual"
        enforced: true
    
    documentation:
      - id: doc-hipaa-baa
        name: "Business Associate Agreement"
        description: "BAA documentation required with all vendors"
        enforced: true
```

---

## 7. SAAS-PRODUCT Profile

Profile untuk SaaS product development.

```yaml
# profiles/saas-product/naeos-profile.yaml

profile:
  id: saas-product
  name: "SaaS Product Profile"
  version: 1.0.0
  description: "Profile untuk SaaS product development"
  
  inherits:
    - enterprise
  
  owner: product-engineering-team
  tags:
    - project-type
    - saas
    - product
  
  policies:
    
    quality_gates:
      - id: gate-performance
        name: "Performance Baseline"
        description: "Performance must meet baseline (p95 < 200ms)"
        metric: "p95_latency_ms"
        threshold: 200
        enforced: true
      
      - id: gate-uptime
        name: "Uptime SLA"
        description: "Service must maintain 99.95% uptime SLA"
        metric: "uptime_percent"
        threshold: 99.95
        enforced: true
    
    testing:
      - id: test-soak
        name: "Soak Testing"
        description: "Service must pass extended soak tests"
        duration: "72 hours"
        enforced: true
      
      - id: test-multi-tenant
        name: "Multi-Tenant Testing"
        description: "Multi-tenant isolation must be tested"
        enforced: true
    
    devops:
      - id: deploy-canary
        name: "Canary Deployments"
        description: "Use canary deployments for production"
        enforced: true
      
      - id: deploy-monitoring
        name: "Production Monitoring"
        description: "Comprehensive monitoring in production"
        enforced: true
      
      - id: deploy-incident-response
        name: "Incident Response Plan"
        description: "Documented incident response procedures"
        enforced: true
      
      - id: deploy-disaster-recovery
        name: "Disaster Recovery Plan"
        description: "DR plan with regular testing"
        rto: "1 hour"
        rpo: "15 minutes"
        enforced: true
```

---

## 8. MICROSERVICES Profile

Profile untuk microservices architecture.

```yaml
# profiles/microservices/naeos-profile.yaml

profile:
  id: microservices
  name: "Microservices Profile"
  version: 1.0.0
  description: "Profile untuk microservices architecture"
  
  inherits:
    - enterprise
  
  owner: architecture-team
  tags:
    - architecture
    - microservices
    - distributed
  
  policies:
    
    standards:
      - id: std-api-rest
        name: "REST API Standard"
        description: "APIs must follow REST conventions"
        link: "https://restfulapi.net"
        enforced: true
      
      - id: std-async-messaging
        name: "Async Messaging"
        description: "Inter-service communication uses message queues"
        tools: ["kafka", "rabbitmq", "nats"]
        enforced: true
      
      - id: std-service-discovery
        name: "Service Discovery"
        description: "Services must be discoverable"
        tools: ["consul", "eureka", "kubernetes"]
        enforced: true
    
    testing:
      - id: test-integration-ms
        name: "Microservice Integration Tests"
        description: "Tests for service-to-service communication"
        enforced: true
      
      - id: test-contract-driven
        name: "Contract-Driven Testing"
        description: "API contracts must be tested"
        tools: ["pact", "spring-cloud-contract"]
        enforced: true
    
    devops:
      - id: deploy-container
        name: "Container Deployment"
        description: "Services must be containerized"
        enforced: true
      
      - id: deploy-orchestration
        name: "Container Orchestration"
        description: "Use container orchestration platform"
        tools: ["kubernetes"]
        enforced: true
      
      - id: deploy-service-mesh
        name: "Service Mesh (Optional)"
        description: "Service mesh for advanced traffic management"
        tools: ["istio", "linkerd"]
        enforced: false
      
      - id: deploy-circuit-breaker
        name: "Circuit Breaker Pattern"
        description: "Implement circuit breaker for resilience"
        enforced: true
```

---

## 9. Cara Menggunakan Contoh Profile

### 9.1 Mengadopsi Profile Standar

```bash
# Copy base profile
cp -r profiles/base/naeos-profile.yaml my-org-profile.yaml

# Customize sesuai kebutuhan
# Edit my-org-profile.yaml
```

### 9.2 Menggabungkan Beberapa Profile

```yaml
# specification.yaml
profiles:
  - enterprise      # Base governance
  - fintech         # Industry standards
  - microservices   # Architecture pattern
  - saas-product    # Product requirements
```

### 9.3 Extending Profile

```yaml
# my-custom-profile.yaml
profile:
  id: my-org-fintech
  inherits:
    - fintech
    - microservices
  
  extends:
    - custom-compliance  # Custom policies
    - custom-security    # Custom security
```

---

## 10. References

- [NAEOS-PRO-001.md](NAEOS-PRO-001.md) - Profile System Specification
- [NAEOS-PRO-002.md](NAEOS-PRO-002.md) - Profile Implementation & Setup
- [NAEOS-PRO-004.md](NAEOS-PRO-004.md) - Profile Best Practices
