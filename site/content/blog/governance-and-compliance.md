---
title: "Governance & Compliance in NAEOS: Built-In Policy, Audit, and SSO"
description: "How NAEOS helps teams enforce policies, maintain audit trails, and meet SOC 2, HIPAA, and GDPR requirements out of the box."
date: 2026-07-25
author: "NAEOS Foundation"
categories: ["tutorial"]
---

Enterprise software projects face a growing list of governance requirements: policy enforcement, audit trails, access control, and regulatory compliance. Most teams bolt these on after the fact.

NAEOS embeds governance into the pipeline itself. Policies are evaluated at every stage, every action is audited with tamper-proof chains, and SSO integrations work out of the box.

## Policy System

NAEOS ships with five built-in policy rules that run during the validation stage:

| Rule | What It Does | Action |
|------|-------------|--------|
| `project-required` | Blocks specs without a project name | block |
| `modules-required` | Blocks specs without any modules | block |
| `architecture-pattern-valid` | Warns on unknown patterns | warn |
| `deployment-strategy-valid` | Warns on unknown strategies | warn |
| `service-port-positive` | Warns on non-positive ports | warn |

### Custom Rules

You can define project-specific rules in Go:

```go
import "github.com/NAEOS-foundation/naeos/internal/governance/policy"

rules := []policy.Rule{
    {
        RuleID:    "require-testing",
        Condition: "exists:testing",
        Priority:  1,
        Action:    "block",
    },
}
```

Rules support condition operators like `exists`, `not_empty`, `contains`, `gt`, `lt`, and `in` — making it possible to encode almost any organizational policy.

## RBAC: Role-Based Access Control

NAEOS includes a hierarchical RBAC system with three built-in roles:

| Role | Permissions |
|------|------------|
| **admin** | Full access to all features |
| **developer** | Create and run pipelines, manage specs |
| **viewer** | Read-only access to specs and audit logs |

RBAC supports **parent chains** (role inheritance) and **deny rules** (explicit override that wins over allow):

```bash
naeos auth create-role devops \
  --parent developer --parent viewer \
  --permission pipeline:run --permission spec:read \
  --deny spec:delete
```

Four compliance templates are available: `auditor`, `soc2_auditor`, `gdpr_admin`, and `hipaa_admin`.

## Audit Trail

Every pipeline execution, policy evaluation, and artifact generation is logged. The audit system supports three backends:

### Standard Auditor
Sequential JSON log with timestamps and action metadata.

### Hashed Chain Auditor
Each entry includes `SHA256(previous_hash + payload)`. This creates a tamper-evident chain:

```go
auditor := audit.NewHashedFileAuditor("audit.log")
defer auditor.Close()

auditor.Log(audit.Entry{
    Action:    "pipeline.run",
    Entity:    "spec:blog-platform",
    Subject:   "user:deploy-bot",
    Metadata:  map[string]any{"status": "completed"},
})
```

Verify the chain at any time:
```bash
naeos compliance verify --audit-file audit.log
```

### Encrypted Auditor
Entries are encrypted with AES-256-GCM before writing to disk. Decrypt for review:

```go
auditor := audit.NewEncryptedFileAuditor("audit.enc", encryptionKey)
reader := auditor.NewDecryptedReader()
```

### Cloud Export

Export audit logs to AWS S3, Google Cloud Storage, or Azure Blob Storage:

```bash
naeos compliance cloud-export \
  --provider s3 \
  --bucket naeos-audit-logs \
  --prefix audit/
```

## Compliance Frameworks

NAEOS provides built-in compliance evaluation for three major frameworks:

| Framework | Controls | Coverage |
|-----------|----------|----------|
| SOC 2 | CC1.1–CC8.1 | 8 controls across security, availability, and confidentiality |
| HIPAA | 164.308–164.312 | 11 controls for administrative, physical, and technical safeguards |
| GDPR | Articles 5, 7, 17, 25, 30, 32, 33, 35 | 8 articles covering data protection, consent, breach notification |

Generate a compliance report:

```bash
naeos compliance report --framework soc2 --output report.json
naeos compliance report --framework hipaa --output hipaa-report.json
```

## SSO Integration

NAEOS supports three enterprise SSO protocols:

| Protocol | What It Provides |
|----------|-----------------|
| **OIDC** | Full OpenID Connect provider with discovery, JWKS, and auth code flow |
| **SAML 2.0** | XML Response parsing, NameID extraction, attribute mapping |
| **LDAP** | TCP/TLS bind, ASN.1 BER search, group membership resolution |

Configure SSO via CLI:

```bash
naeos auth sso configure \
  --provider-type oidc \
  --issuer https://accounts.google.com \
  --client-id your-client-id \
  --client-secret your-client-secret
```

The SSO registry supports multiple providers simultaneously, so you can mix OIDC for employees with SAML for partners.

## Governance in CI/CD

Integrate governance checks into your CI pipeline:

```yaml
# .github/workflows/ci.yml
- name: Validate Spec
  run: naeos validate --input-file spec.yaml

- name: Lint Spec
  run: naeos lint --input-file spec.yaml

- name: Audit
  run: naeos audit spec.yaml

- name: Compliance Check
  run: naeos compliance verify --audit-file ~/.naeos/audit.log
```

## What's Next

Governance in NAEOS is evolving toward fully automated compliance monitoring. Upcoming features include:

- **Real-time policy violation alerts** via webhook
- **Automated evidence collection** for SOC 2 and HIPAA audits
- **Multi-environment policy profiles** (stricter rules for production)

For complete documentation, see the [Governance docs](/docs/governance/) and [Security policy](/security/).
