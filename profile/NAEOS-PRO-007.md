# NAEOS-PRO-007: Profile Troubleshooting & FAQ

## Status
- Status: Draft
- Version: 1.0
- Owner: NAEOS Foundation
- Last Updated: 2026-07-10

---

## 1. Troubleshooting Guide

### 1.1 Profile Not Found

**Error:**
```
Error: Profile 'enterprise' not found
```

**Causes:**
1. Profile ID salah
2. Profile belum diregister
3. Registry tidak accessible
4. Version spesifik tidak ada

**Solutions:**

```bash
# List available profiles
naeos profile list

# Search untuk profile yang mirip
naeos profile list | grep enterp  # Cari 'enterp'

# Check registry connection
naeos profile registry-ping https://registry.naeos.io

# Verify specific version exists
naeos profile show enterprise:1.2.0 --check-version
```

**If still not found:**
- Pastikan profile sudah diregister: `naeos profile register --file enterprise.yaml`
- Check registry URL configuration: `naeos config get registry.url`
- Test connectivity: `curl https://registry.naeos.io/health`

---

### 1.2 Circular Dependency

**Error:**
```
Error: Circular dependency detected in profile inheritance
  enterprise → fintech → enterprise
```

**Cause:**
Profile inheritance membentuk loop (A → B → A).

**Solution:**

```bash
# Visualize dependency tree
naeos profile tree enterprise --show-cycles

# Output:
# ⚠️ CIRCULAR DEPENDENCY DETECTED:
# enterprise → fintech → enterprise

# Review profile inheritance
naeos profile show enterprise
# Check "inherits" field

naeos profile show fintech
# Remove inheritance of enterprise if not needed

# Fix: Remove cycle
# Edit fintech profile, remove "inherits: [enterprise]"
naeos profile update fintech --remove-inheritance enterprise

# Verify
naeos profile tree enterprise --check-cycles
```

---

### 1.3 Conflicting Policies

**Error:**
```
Error: Policy conflict detected between 'enterprise' and 'fintech'
  Policy: gate-test-coverage
  enterprise: threshold=80%
  fintech: threshold=90%
```

**Cause:**
Dua profile mendefinisikan policy yang sama dengan nilai berbeda dan tidak ada clear resolution.

**Solution:**

```bash
# Check conflict details
naeos profile diff enterprise fintech

# Resolve manually
naeos profile compile \
  --profiles enterprise,fintech \
  --report-conflicts \
  --verbose

# Option 1: Change order (later profile wins)
# specification.yaml
profiles:
  - enterprise
  - fintech  # Will override enterprise

# Option 2: Explicit override
policies:
  quality_gates:
    - id: gate-test-coverage
      threshold: 90  # Explicit override

# Verify resolution
naeos profile compile --profiles enterprise,fintech --report-conflicts
```

---

### 1.4 Profile Validation Fails

**Error:**
```
Error: Profile validation failed
  - Invalid policy 'gate-coverage': threshold must be 0-100
  - Missing required field 'owner'
  - Unknown policy type 'unknown-type'
```

**Solution:**

```bash
# Detailed validation report
naeos profile validate profiles/enterprise.yaml --verbose

# Fix issues one by one
# Example: Fix threshold
# Edit profiles/enterprise.yaml
# Change: threshold: 150  →  threshold: 80

# Add missing owner
# Edit profiles/enterprise.yaml
# Add: owner: platform-team

# Remove unknown policy types
# Review policy definitions

# Validate again
naeos profile validate profiles/enterprise.yaml

# Full validation with fixes
naeos profile validate profiles/enterprise.yaml --auto-fix
```

---

### 1.5 Compliance Check Fails

**Error:**
```
Error: Compliance check failed
Project compliance: 65% (29/45 policies)

Failed policies:
  ✗ gate-test-coverage: coverage 60% < 80%
  ✗ sec-no-secrets: credentials found in code
  ✗ doc-api-openapi: openapi.yaml not found
```

**Solution:**

```bash
# Get detailed report
naeos profile check-compliance \
  --profile enterprise \
  --project ./ \
  --verbose \
  --output compliance-report.json

# For each failed policy:

# 1. Increase test coverage
naeos profile check-compliance \
  --profile enterprise \
  --project ./ \
  --policy gate-test-coverage

# Add tests
# Run again to verify

# 2. Remove secrets from code
naeos profile check-compliance \
  --profile enterprise \
  --project ./ \
  --policy sec-no-secrets --verbose

# Review findings, remove secrets
git rm --cached credentials.json  # Don't commit secrets
echo "credentials.json" >> .gitignore

# 3. Add missing documentation
# Create openapi.yaml
# Validate again

# Final check
naeos profile check-compliance \
  --profile enterprise \
  --project ./
```

---

### 1.6 Registry Authentication Failed

**Error:**
```
Error: Authentication failed
  Status: 401 Unauthorized
  Registry: https://registry.company.com
```

**Solution:**

```bash
# Check authentication
naeos profile registry-auth \
  --registry https://registry.company.com \
  --token <token>

# If token-based:
naeos profile register \
  --file enterprise.yaml \
  --registry https://registry.company.com \
  --token $(cat ~/.naeos/token.txt)

# If OAuth2:
naeos profile auth \
  --registry https://registry.company.com \
  --oauth2
# Browser akan open untuk login

# If certificate-based:
naeos profile register \
  --file enterprise.yaml \
  --registry https://registry.company.com \
  --cert ~/.naeos/client-cert.pem \
  --key ~/.naeos/client-key.pem

# Verify connection
naeos profile registry-ping \
  --registry https://registry.company.com
```

---

### 1.7 Performance Issue

**Error:**
```
Warning: Profile compilation took 5.2 seconds (expected < 1 second)
```

**Cause:**
- Inheritance tree terlalu kompleks
- Terlalu banyak policies
- Registry response lambat

**Solution:**

```bash
# Profile compilation profiling
naeos profile compile \
  --specification specification.yaml \
  --profile enterprise \
  --profile fintech \
  --profile microservices \
  --timing

# Output:
# Loading profiles:      245ms
#   - enterprise:         120ms
#   - fintech:            85ms
#   - microservices:      40ms
# Resolving inheritance: 1200ms
# Merging policies:      340ms
# Validation:            520ms
# Total:                2305ms

# Optimization:
# 1. Flatten inheritance (reduce nesting)
# 2. Remove unused profiles
# 3. Cache compiled profiles
# 4. Use local registry mirror

# Test optimizations
naeos profile compile --specification spec.yaml --timing
```

---

### 1.8 Version Compatibility Issue

**Error:**
```
Error: Profile 'enterprise' (v2.0.0) requires NAEOS >= 0.3.0
  Current NAEOS version: 0.2.5
```

**Solution:**

```bash
# Check NAEOS version
naeos version

# Upgrade NAEOS
curl -fsSL https://install.naeos.io/cli.sh | bash

# Or use package manager
brew upgrade naeos-cli

# Downgrade profile if needed
naeos profile downgrade enterprise --to 1.5.0

# Verify compatibility
naeos profile compatibility-check \
  --profile enterprise:2.0.0 \
  --naeos-version $(naeos version --short)
```

---

## 2. FAQ

### 2.1 Profile Management

**Q: Bagaimana membuat profile pertama saya?**

A: Ikuti langkah-langkah di [NAEOS-PRO-002.md](NAEOS-PRO-002.md) Section 3.

Ringkas:
```bash
# 1. Create file
mkdir -p profiles/myorg
cat > profiles/myorg/naeos-profile.yaml << EOF
profile:
  id: myorg-base
  name: "MyOrg Base"
  version: 1.0.0
  inherits: []
  policies:
    standards: []
EOF

# 2. Validate
naeos profile validate profiles/myorg/naeos-profile.yaml

# 3. Register
naeos profile register --file profiles/myorg/naeos-profile.yaml
```

---

**Q: Berapa banyak profile yang harus saya buat?**

A: Tergantung struktur organisasi Anda.

Rekomendasi:
- **Base**: 1 (universal)
- **Organization**: 1-3 (enterprise, startup, agency)
- **Industry**: 1-5 (fintech, healthcare, etc)
- **Project Type**: 1-5 (api, frontend, mobile, library)
- **Project**: Sesuai kebutuhan

Total: 5-15 profile adalah typical.

**Terlalu banyak?** Combine beberapa profiles.

---

**Q: Kapan harus membuat profile vs kapan harus extend?**

A: Gunakan extend untuk add-ons, inherit untuk hierarchy.

```yaml
# Use inheritance untuk hierarchy
fintech:
  inherits: [enterprise]  # Fintech is a type of enterprise

# Use extends untuk add-ons
profile:
  extends:
    - security-high       # Add-on, not part of hierarchy
    - compliance-pci      # Add-on, not part of hierarchy
```

---

### 2.2 Policy Management

**Q: Apa perbedaan antara enforced dan optional policies?**

A:
- **enforced: true**: Harus memenuhi, compliance check akan fail
- **enforced: false**: Recommended tapi tidak wajib

```yaml
policies:
  standards:
    - id: std-readme
      enforced: true      # Must have README
    
    - id: doc-adr
      enforced: false     # Recommended but optional
```

---

**Q: Bisakah saya override policy dari inherited profile?**

A: Ya, gunakan prioritas inheritance.

```yaml
# Inherited profile (enterprise)
quality_gates:
  - id: coverage
    threshold: 80

# Your project profile (inherits enterprise)
quality_gates:
  - id: coverage
    threshold: 90  # Override dengan threshold lebih tinggi
```

Profile Anda (lebih spesifik) akan override enterprise (lebih umum).

---

**Q: Bagaimana jika saya ingin exception untuk policy tertentu?**

A: Gunakan exceptions field dalam policy.

```yaml
policies:
  testing:
    - id: test-coverage-80
      threshold: 80
      exceptions:
        - condition: "legacy_code"
          threshold: 40
          expires: "2027-12-31"
```

---

### 2.3 Compliance

**Q: Bagaimana jika compliance check gagal?**

A: Lihat 1.5 Compliance Check Fails untuk solusi.

Ringkas:
1. Dapatkan detailed report: `naeos profile check-compliance --verbose`
2. Review setiap failed policy
3. Fix issue satu per satu
4. Re-run compliance check

---

**Q: Berapa sering harus run compliance check?**

A:
- **Development**: Setiap commit (via CI)
- **Staging**: Setiap deploy
- **Production**: Minimal daily
- **Critical systems**: Setiap jam

```bash
# CI/CD example (GitHub Actions)
naeos profile check-compliance \
  --profile enterprise \
  --project ./ \
  --exit-on-fail  # Fail CI if non-compliant
```

---

**Q: Bagaimana monitor compliance over time?**

A:
```bash
# Generate daily report
naeos profile compliance-report \
  --profile enterprise \
  --project ./ \
  --format json \
  --output compliance-$(date +%Y%m%d).json

# Analyze trend
naeos profile compliance-trend \
  --profile enterprise \
  --days 30 \
  --output trend.html
```

---

### 2.4 Upgrade & Migration

**Q: Kapan harus upgrade profile?**

A:
- **PATCH**: Langsung upgrade (safe)
- **MINOR**: Cek release notes, biasanya safe
- **MAJOR**: Plan dengan hati-hati, ada breaking changes

---

**Q: Bagaimana jika upgrade gagal?**

A: Rollback.

```bash
# Quick rollback
naeos profile downgrade enterprise

# Or explicit
naeos profile downgrade enterprise --to 1.8.0

# Verify
naeos profile show enterprise
```

---

**Q: Berapa lama harus test sebelum production upgrade?**

A:
- **PATCH**: 1-2 jam (automated tests)
- **MINOR**: 4-8 jam (manual testing)
- **MAJOR**: 24-48 jam (full integration test)

---

### 2.5 Performance

**Q: Profile compilation lambat, apa yang bisa dilakukan?**

A: Lihat 1.7 Performance Issue.

Quick fixes:
```bash
# 1. Reduce inheritance depth (flatten tree)
# 2. Remove unused profiles
# 3. Use profile cache:
naeos profile compile \
  --specification spec.yaml \
  --cache ~/.naeos/profile-cache
# 4. Use local registry mirror
```

---

**Q: Bagaimana optimize large organizations dengan banyak profiles?**

A:
1. **Organize hierarki dengan jelas**
   ```
   Base → Org → Industry → Architecture → Project
   ```

2. **Use composition bukan inheritance**
   ```yaml
   # Instead of:
   fintech → payments → payment-gateway → payment-gateway-prod
   
   # Use:
   profiles:
     - base
     - enterprise
     - fintech
     - microservices
     - payment-processing
   ```

3. **Create profile modules**
   ```yaml
   # compliance-pci.yaml (reusable module)
   # compliance-gdpr.yaml (reusable module)
   # Then combine:
   profiles:
     - enterprise
     - compliance-pci
     - compliance-gdpr
   ```

---

### 2.6 Best Practices

**Q: Apa profile structure yang recommended?**

A: Lihat [NAEOS-PRO-004.md](NAEOS-PRO-004.md) Section 5.1 for directory structure.

```
profiles/
├── base/
├── organization/
├── industry/
├── architecture/
└── project/
```

---

**Q: Bagaimana version profile dengan git?**

A:
```bash
# Commit profile changes
git add profiles/enterprise/naeos-profile.yaml
git commit -m "feat(profile): add performance monitoring policy"

# Tag releases
git tag profile-enterprise-v1.2.0

# View history
git log --oneline profiles/enterprise/
```

---

**Q: Bagaimana handle profile across teams?**

A:
1. **Centralized registry**: Single source of truth
2. **Clear ownership**: Each profile has owner
3. **Approval process**: Reviews before changes
4. **Version control**: All in git
5. **Documentation**: Clear rationale for each policy

---

### 2.7 Integration

**Q: Bagaimana integrate profile dengan CI/CD pipeline?**

A: Lihat [NAEOS-PRO-005.md](NAEOS-PRO-005.md) Section 9 untuk CI/CD examples.

---

**Q: Bisakah combine NAEOS profile dengan tool lain (e.g., Helm, Terraform)?**

A: Ya, NAEOS profile bisa generate configuration untuk tools lain.

```bash
# Generate Helm values
naeos profile generate-helm \
  --profile enterprise \
  --output helm-values.yaml

# Generate Terraform variables
naeos profile generate-terraform \
  --profile enterprise \
  --output terraform.tfvars
```

---

**Q: Bagaimana use profile dengan infrastructure as code (Terraform, CloudFormation)?**

A:
```hcl
# terraform/main.tf

variable "profile_version" {
  default = "enterprise:1.2.0"
}

resource "local_file" "naeos_profile" {
  filename = "${path.module}/.naeos/profile.yaml"
  
  content = <<-EOT
    profile: ${var.profile_version}
    policies:
      security:
        - tls-version: "1.3"
        - mfa: "required"
  EOT
}
```

---

## 3. Common Issues Reference

| Issue | Error | Solution | Link |
|-------|-------|----------|------|
| Profile not found | E001 | Check registry, list profiles | 1.1 |
| Circular dependency | E003 | Review inheritance, remove cycles | 1.2 |
| Conflicting policies | E004 | Resolve conflicts, order profiles | 1.3 |
| Validation fails | E002 | Fix syntax, validate again | 1.4 |
| Compliance fails | E005 | Review report, fix issues | 1.5 |
| Auth failed | E401 | Check token, verify registry | 1.6 |
| Performance slow | Perf | Flatten tree, reduce profiles | 1.7 |
| Version incompatible | E999 | Upgrade NAEOS or downgrade profile | 1.8 |

---

## 4. Getting Help

### 4.1 Debug Mode

```bash
# Enable debug output
naeos profile show enterprise --debug

# Enable verbose logging
naeos profile check-compliance \
  --profile enterprise \
  --project ./ \
  --verbose

# Show timing information
naeos profile compile \
  --specification spec.yaml \
  --profile enterprise \
  --timing
```

### 4.2 Support Channels

- **Issues**: https://github.com/NAEOS-foundation/naeos/issues
- **Discussions**: https://github.com/NAEOS-foundation/naeos/discussions
- **Documentation**: https://naeos.io/docs
- **Community Slack**: https://slack.naeos.io

### 4.3 Provide Diagnostic Info

When asking for help:

```bash
# Collect diagnostic info
naeos profile diagnose enterprise \
  --output diagnostics.json

# Include in issue/question
# Also share:
# - NAEOS version: naeos version
# - Profile version: naeos profile show enterprise
# - Minimal reproduction steps
```

---

## 5. References

- [NAEOS-PRO-001.md](NAEOS-PRO-001.md) - Profile System Specification
- [NAEOS-PRO-002.md](NAEOS-PRO-002.md) - Profile Implementation & Setup
- [NAEOS-PRO-004.md](NAEOS-PRO-004.md) - Profile Best Practices
- [NAEOS-PRO-005.md](NAEOS-PRO-005.md) - Profile API & CLI Reference
- [NAEOS-PRO-006.md](NAEOS-PRO-006.md) - Profile Migration & Upgrade Guide
