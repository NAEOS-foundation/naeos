# NAEOS-PRO-002: Profile Implementation & Setup Guide

## Status
- Status: Draft
- Version: 1.0
- Owner: NAEOS Foundation
- Last Updated: 2026-07-10

---

## 1. Executive Summary

Panduan ini menjelaskan cara mengimplementasikan, mengaktifkan, dan mengelola Profile di NAEOS. Profile memungkinkan Anda untuk mengemas kebijakan engineering (rules, standards, gates, policies) menjadi konfigurasi yang dapat digunakan kembali di berbagai proyek dan konteks organisasi.

Dokumentasi ini ditujukan untuk:
- **Engineering Leads** yang ingin mendefinisikan profil organisasi
- **Project Managers** yang ingin mengaktifkan profil untuk proyek
- **Platform Engineers** yang mengintegrasikan NAEOS ke dalam pipeline
- **DevOps Teams** yang mengelola profil di multiple environment

---

## 2. Konsep Dasar

### 2.1 Apa itu Profile?

Profile adalah **Engineering Policy Package** yang berisi:

| Elemen | Deskripsi |
|--------|-----------|
| **Rules** | Aturan yang harus dipenuhi (misal: naming convention, file structure) |
| **Standards** | Standar kode dan praktik yang diharapkan |
| **Quality Gates** | Threshold kualitas yang harus dicapai sebelum merge/deployment |
| **Security Policies** | Kebijakan keamanan (authentication, encryption, audit) |
| **AI Policies** | Aturan untuk penggunaan AI dalam development |
| **Documentation Policies** | Standar dokumentasi (format, coverage, review process) |
| **Testing Policies** | Requirement untuk testing (unit, integration, e2e coverage) |
| **DevOps Policies** | Aturan deployment, versioning, dan release management |

### 2.2 Profile Hierarchy

Profile mengikuti hirarki inheritance yang jelas:

```
BASE
  ├── ORGANIZATION (enterprise, startup, agency)
  │    └── INDUSTRY (fintech, healthcare, e-commerce, saas)
  │         └── PROJECT-TYPE (api, frontend, mobile, infra)
  │              └── PROJECT (payment-gateway, auth-service, dashboard)
```

**Prinsip Resolusi**: Ketika terjadi konflik, hirarki yang lebih bawah (lebih spesifik) mengambil alih.

**Contoh**:
```
BASE
 ↓
ENTERPRISE
 ↓
FINTECH
 ↓
PAYMENTS
 ↓
PAYMENT-GATEWAY
```

---

## 3. Implementasi Profile

### 3.1 Struktur File Profile

Setiap profile memiliki struktur standar:

```yaml
# naeos-profile.yaml

profile:
  id: payment-gateway
  name: "Payment Gateway Profile"
  version: 1.0.0
  description: "Profile untuk payment gateway services di fintech organization"
  
  # Inheritance
  inherits:
    - enterprise
    - fintech
    - payments
  
  # Extensions
  extends:
    - security-high
    - compliance-pci
    - audit-full
  
  # Owner & metadata
  owner: fintech-platform-team
  created: 2026-07-01
  last_modified: 2026-07-10
  
  # Policies
  policies:
    rules:
      - rule-naming-go
      - rule-error-handling
      - rule-logging-structured
    
    standards:
      - std-go-fmt
      - std-go-lint
      - std-api-rest
    
    quality_gates:
      - gate-coverage-80
      - gate-lint-zero
      - gate-security-scan
    
    security:
      - sec-encryption-tls13
      - sec-auth-oauth2
      - sec-audit-logging
    
    testing:
      - test-unit-coverage-80
      - test-integration-required
      - test-e2e-critical-paths
    
    documentation:
      - doc-api-swagger
      - doc-architecture-adr
      - doc-readme-required
    
    devops:
      - deploy-helm
      - deploy-approval-required
      - version-semver
```

### 3.2 Membangun Profile Baru

#### Step 1: Tentukan Kebutuhan

Identifikasi:
- Level organisasi (Base, Organization, Industry, Project Type, atau Project)
- Policies yang diperlukan
- Standards yang harus dipenuhi
- Constraints khusus untuk konteks Anda

#### Step 2: Buat File Profile

```bash
# Buat directory untuk profile
mkdir -p profiles/myorg

# Buat file naeos-profile.yaml
cat > profiles/myorg/naeos-profile.yaml << EOF
profile:
  id: myorg-base
  name: "MyOrg Base Profile"
  version: 1.0.0
  description: "Foundation profile untuk MyOrg"
  
  inherits: []
  extends: []
  
  owner: platform-team
  
  policies:
    rules: []
    standards: []
    quality_gates: []
    security: []
    testing: []
    documentation: []
    devops: []
EOF
```

#### Step 3: Definisikan Policies

```yaml
policies:
  rules:
    # Naming conventions
    - id: rule-naming-go
      name: "Go Naming Convention"
      description: "Enforce Go naming conventions from Effective Go"
      enforced: true
    
  standards:
    # Code formatting
    - id: std-go-fmt
      name: "Go Format Standard"
      description: "All Go code must pass gofmt"
      tool: gofmt
      severity: ERROR
    
  quality_gates:
    # Test coverage
    - id: gate-coverage-80
      name: "Test Coverage 80%"
      description: "Minimum 80% test coverage required"
      threshold: 80
      metric: coverage_percent
      enforced: true
  
  security:
    # Encryption standard
    - id: sec-encryption-tls13
      name: "TLS 1.3"
      description: "All connections must use TLS 1.3 or higher"
      version: "1.3"
      enforced: true
```

#### Step 4: Register Profile

```bash
# Validate profile syntax
naeos profile validate profiles/myorg/naeos-profile.yaml

# Register ke profile registry
naeos profile register \
  --file profiles/myorg/naeos-profile.yaml \
  --registry-url https://registry.naeos.io
```

---

## 4. Mengaktifkan Profile

### 4.1 Aktivasi di Project Specification

Setiap project mendeklarasikan profile yang digunakan di `specification.yaml`:

```yaml
# specification.yaml

project:
  name: payment-gateway
  version: 1.0.0

profile: payment-gateway  # Single profile ID

# atau

profiles:  # Multiple profiles dengan composition
  - enterprise
  - fintech
  - payments
  - security-high
```

### 4.2 Resolusi & Compilation

Saat NAEOS memproses specification:

```
1. Load specification.yaml
   ↓
2. Resolve profile references
   ↓
3. Load profile definitions (dari registry atau local)
   ↓
4. Trace inheritance tree
   ↓
5. Merge policies (dengan conflict resolution)
   ↓
6. Generate effective policies
   ↓
7. Validate consistency
   ↓
8. Output compiled profile
```

### 4.3 Verification

```bash
# Check profile yang aktif
naeos profile show payment-gateway

# Verify inheritance
naeos profile tree payment-gateway

# Validate against specification
naeos validate --specification specification.yaml --profile payment-gateway

# Generate report
naeos profile report --profile payment-gateway --output report.json
```

---

## 5. Profile Composition

### 5.1 Single Profile

Menggunakan satu profile:

```yaml
profile: enterprise
```

### 5.2 Multiple Profiles

Menggabungkan beberapa profile:

```yaml
profiles:
  - base
  - enterprise
  - cloud-native
  - pci-compliance
```

NAEOS akan:
1. Load masing-masing profile
2. Merge policies sesuai urutan (yang belakang override yang depan)
3. Resolve conflicts menggunakan conflict resolution rules
4. Generate warning jika ada konflik yang terdeteksi

### 5.3 Profile Extension

Extend profile dengan policies tambahan:

```yaml
profile: enterprise

extends:
  - security-high  # Add extra security policies
  - audit-full     # Add audit requirements
```

---

## 6. Conflict Resolution

### 6.1 Prioritas Resolusi

Ketika dua policy bertentangan:

```
Project Level Policy       (Priority 1 - Highest)
    ↑
Project Type Policy        (Priority 2)
    ↑
Industry Policy           (Priority 3)
    ↑
Organization Policy       (Priority 4)
    ↑
Base Policy              (Priority 5 - Lowest)
```

### 6.2 Contoh Conflict Resolution

```yaml
# base profile
quality_gates:
  - id: coverage
    threshold: 70

# enterprise profile (inherits base)
quality_gates:
  - id: coverage
    threshold: 80  # Override base value

# fintech profile (inherits enterprise)
quality_gates:
  - id: coverage
    threshold: 90  # Override enterprise value
```

Hasil final: coverage 90%

### 6.3 Conflict Report

Compiler menghasilkan laporan jika terjadi konflik:

```bash
naeos profile compile --specification spec.yaml --verbose

# Output:
# ⚠️  CONFLICT DETECTED: policy 'coverage-threshold'
#   - base: 70%
#   - enterprise: 80%
#   - fintech: 90% ← APPLIED
#   Reason: Project-level override takes precedence
```

---

## 7. Command Reference

### 7.1 Profile Management

```bash
# Create profile
naeos profile create \
  --id myorg-base \
  --name "MyOrg Base" \
  --version 1.0.0

# Register profile
naeos profile register --file profile.yaml

# Update profile
naeos profile update --id myorg-base --file profile.yaml

# Delete profile
naeos profile delete --id myorg-base

# List profiles
naeos profile list
naeos profile list --filter tag:enterprise
```

### 7.2 Profile Inspection

```bash
# Show profile details
naeos profile show payment-gateway

# Show profile tree (inheritance)
naeos profile tree payment-gateway

# Compare profiles
naeos profile diff profile1 profile2

# Export profile
naeos profile export --id payment-gateway --format yaml > exported.yaml
```

### 7.3 Validation & Compliance

```bash
# Validate profile syntax
naeos profile validate --file profile.yaml

# Check profile compliance
naeos profile check-compliance --id payment-gateway

# Generate compliance report
naeos profile compliance-report --id payment-gateway --output report.pdf
```

---

## 8. Best Practices

### 8.1 Profile Organization

- **Base Profile**: Hanya untuk defaults universal
- **Organization Profile**: Policies spesifik organisasi
- **Industry Profile**: Standar industri (fintech, healthcare, etc)
- **Project Type Profile**: Standar untuk jenis project tertentu
- **Project Profile**: Customization spesifik project

### 8.2 Versioning

- Gunakan semantic versioning: `MAJOR.MINOR.PATCH`
- Document breaking changes di release notes
- Maintain backward compatibility dalam PATCH version
- Use deprecation warnings untuk planned breaking changes

### 8.3 Documentation

- Document setiap policy dengan clear description
- Jelaskan rationale di balik setiap requirement
- Include examples untuk policies yang kompleks
- Maintain changelog untuk setiap update

---

## 9. Troubleshooting

Lihat [NAEOS-PRO-007.md](NAEOS-PRO-007.md) untuk troubleshooting dan FAQ.

---

## 10. References

- [NAEOS-PRO-001.md](NAEOS-PRO-001.md) - Profile System Specification
- [NAEOS-PRO-003.md](NAEOS-PRO-003.md) - Profile Examples
- [NAEOS-PRO-004.md](NAEOS-PRO-004.md) - Profile Best Practices
- [NAEOS-PRO-005.md](NAEOS-PRO-005.md) - Profile API & CLI Reference
